// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/achetronic/magec/server/admin"
	"github.com/achetronic/magec/server/agent"
	"github.com/achetronic/magec/server/clients/telegram"
	"github.com/achetronic/magec/server/config"
	"github.com/achetronic/magec/server/logging"
	"github.com/achetronic/magec/server/store"
	"github.com/achetronic/magec/server/userapi"
	"github.com/achetronic/magec/server/voice"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/achetronic/magec/server/userapi/docs"

	_ "github.com/achetronic/magec/server/admin/docs"
	_ "github.com/achetronic/magec/server/client/device"
	_ "github.com/achetronic/magec/server/client/telegram"
	_ "github.com/achetronic/magec/server/memory/postgres"
	_ "github.com/achetronic/magec/server/memory/redis"
)

var configFile = flag.String("config", "config.yaml", "Path to config file")

func main() {
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Setup logger
	logging.Setup(cfg.Log.Level, cfg.Log.Format)

	ctx := context.Background()

	// Initialize store with JSON persistence
	dataStore, err := store.New("data/store.json")
	if err != nil {
		slog.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}
	slog.Info("Store initialized", "agents", len(dataStore.Data().Agents), "backends", len(dataStore.Data().Backends))

	// Admin API — start first so it's available even if agent init fails
	adminHandler := admin.New(dataStore)

	adminMux := http.NewServeMux()
	adminMux.Handle("/api/v1/admin/", http.StripPrefix("/api/v1/admin", adminHandler))
	adminMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	adminMux.Handle("/", http.FileServer(http.Dir("admin-ui")))

	adminAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.AdminPort)
	adminServer := &http.Server{
		Addr:         adminAddr,
		Handler:      accessLogMiddleware(corsMiddleware(adminMux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("Admin server started", "addr", adminAddr, "url", fmt.Sprintf("http://%s", adminAddr))
		if err := adminServer.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Admin server error", "error", err)
		}
	}()

	// Swappable handler for agent-related routes (hot-reloaded on store changes)
	agentRouter := &agentRouterHandler{}
	agentRouter.rebuild(ctx, dataStore)

	httpMux := http.NewServeMux()
	httpMux.Handle("/api/v1/agent/", agentRouter)
	httpMux.Handle("/api/v1/voice/", newVoiceHandler(dataStore, agentRouter))

	userAPI := userapi.New(dataStore)
	httpMux.HandleFunc("/api/v1/health", userAPI.Health)
	httpMux.HandleFunc("/api/v1/device/info", userAPI.DeviceInfo)

	httpMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.InstanceName("userapi"),
	))

	// Voice events WebSocket handler (wake word + VAD)
	const (
		voiceModelsPath        = "models"
		voicePretrainedPath    = "pretrained"
		defaultOnnxLibraryPath = "/usr/lib/libonnxruntime.so"
	)
	onnxLibraryPath := defaultOnnxLibraryPath
	if cfg.Server.OnnxLibraryPath != "" {
		onnxLibraryPath = cfg.Server.OnnxLibraryPath
	}
	var voiceDetector *voice.Detector
	wakeWordModelsCfg, err := config.LoadWakeWordModels(voiceModelsPath)
	if err != nil {
		slog.Warn("Wake word models not available", "error", err)
	} else if len(wakeWordModelsCfg.Models) == 0 {
		slog.Warn("No wake word models configured in wakewords.yaml")
	} else {
		models := make([]voice.ModelConfig, len(wakeWordModelsCfg.Models))
		for i, m := range wakeWordModelsCfg.Models {
			models[i] = voice.ModelConfig{
				ID:        m.ID,
				Name:      m.Name,
				File:      fmt.Sprintf("%s/%s", voiceModelsPath, m.File),
				Phrase:    m.Phrase,
				Threshold: m.Threshold,
			}
		}

		voiceDetector = voice.NewDetector(voice.DetectorConfig{
			MelspecModelPath:   fmt.Sprintf("%s/mel-spectrogram.onnx", voicePretrainedPath),
			EmbeddingModelPath: fmt.Sprintf("%s/speech-embedding.onnx", voicePretrainedPath),
			VADModelPath:       fmt.Sprintf("%s/silero-vad.onnx", voicePretrainedPath),
			Models:             models,
			OnnxLibraryPath:    onnxLibraryPath,
		}, slog.Default())

		if err := voiceDetector.Load(); err != nil {
			slog.Warn("Failed to load voice detection models", "error", err)
		} else {
			voiceHandler := voice.NewHandler(voiceDetector, slog.Default())
			httpMux.Handle("/api/v1/voice/events", voiceHandler)
			slog.Info("Voice detection enabled", "wakeWordModels", len(models), "vadEnabled", true)
		}
	}

	// Watch for store changes and hot-reload the agent
	storeChanged := dataStore.OnChange()
	go func() {
		for range storeChanged {
			time.Sleep(500 * time.Millisecond)
			slog.Info("Store changed, reloading agent...")
			agentRouter.rebuild(ctx, dataStore)
		}
	}()

	// Static files
	httpMux.Handle("/", http.FileServer(http.Dir("voice-ui")))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      accessLogMiddleware(corsMiddleware(clientAuthMiddleware(httpMux, dataStore))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start Telegram clients from store
	var telegramClients []*telegram.Client
	for _, cl := range dataStore.ListClients() {
		if cl.Type != "telegram" || !cl.Enabled || cl.Config.Telegram == nil {
			continue
		}
		if len(cl.AllowedAgents) == 0 {
			slog.Warn("Telegram client has no allowed agents, skipping", "client", cl.Name)
			continue
		}
		agentID := cl.AllowedAgents[0]
		agentDef, ok := dataStore.GetAgent(agentID)
		if !ok {
			slog.Warn("Telegram client references unknown agent, skipping", "client", cl.Name, "agent", agentID)
			continue
		}

		agentURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/agent", cfg.Server.Port)
		ttsURL := ""
		if agentDef.TTS.Backend != "" {
			if b, ok := dataStore.GetBackend(agentDef.TTS.Backend); ok && b.URL != "" {
				ttsURL = fmt.Sprintf("http://127.0.0.1:%d/api/v1/voice/%s/speech", cfg.Server.Port, agentID)
			}
		}

		tgClient, err := telegram.New(*cl.Config.Telegram, agentURL, ttsURL, agentDef.TTS, slog.Default())
		if err != nil {
			slog.Error("Failed to create Telegram client", "client", cl.Name, "error", err)
			continue
		}
		telegramClients = append(telegramClients, tgClient)
		go func(name string) {
			time.Sleep(500 * time.Millisecond)
			if err := tgClient.Start(ctx); err != nil {
				slog.Error("Telegram client error", "client", name, "error", err)
			}
		}(cl.Name)
	}

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down...")
		for _, tc := range telegramClients {
			tc.Stop()
		}
		if voiceDetector != nil {
			voiceDetector.Close()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		adminServer.Shutdown(ctx)
		server.Shutdown(ctx)
	}()

	slog.Info("Server started", "addr", addr, "url", fmt.Sprintf("http://%s", addr))
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Error("Server error", "error", err)
		os.Exit(1)
	}
}

// responseRecorder wraps http.ResponseWriter to capture status code and bytes written.
type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	n, err := r.ResponseWriter.Write(b)
	r.bytes += n
	return n, err
}

// Hijack implements http.Hijacker for WebSocket support.
func (r *responseRecorder) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := r.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// accessLogMiddleware logs all HTTP requests with method, path, status, duration and size.
func accessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &responseRecorder{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(rec, r)

		duration := time.Since(start)
		logFn := slog.Info
		if rec.status >= 400 {
			logFn = slog.Warn
		}
		logFn("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rec.status,
			"duration", duration.Round(time.Millisecond),
			"bytes", rec.bytes,
		)
	})
}

// clientAuthMiddleware protects API endpoints on port 8080 with client token auth.
// Static files, health checks, CORS preflight, and voice-events pass through.
// If no clients exist in the store, all requests pass through (open mode).
func clientAuthMiddleware(next http.Handler, dataStore *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if r.Method == http.MethodOptions ||
			path == "/api/v1/health" ||
			path == "/api/v1/voice/events" ||
			!strings.HasPrefix(path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		clients := dataStore.ListClients()
		if len(clients) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		hasToken := strings.HasPrefix(token, "Bearer ")

		if hasToken {
			token = strings.TrimPrefix(token, "Bearer ")
			cl, ok := dataStore.GetClientByToken(token)
			if !ok || !cl.Enabled {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid or disabled client token"}`, http.StatusUnauthorized)
				return
			}
			r.Header.Set("X-Client-Name", cl.Name)
			next.ServeHTTP(w, r)
			return
		}

		if path == "/api/v1/device/info" {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
	})
}

// corsMiddleware adds CORS headers to all responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "*")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// newVoiceHandler creates a handler for /api/v1/voice/ routes.
// Routes:
//   - /api/v1/voice/{agentId}/speech        → TTS proxy (resolved per-agent from store)
//   - /api/v1/voice/{agentId}/transcription  → STT proxy (resolved per-agent from store)
//   - /api/v1/voice/events                   → handled separately by WebSocket mux entry
func newVoiceHandler(dataStore *store.Store, agentRouter *agentRouterHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strip prefix: /api/v1/voice/  →  {agentId}/speech or {agentId}/transcription
		path := strings.TrimPrefix(r.URL.Path, "/api/v1/voice/")

		parts := strings.SplitN(path, "/", 2)
		if len(parts) != 2 {
			http.Error(w, `{"error":"invalid voice endpoint"}`, http.StatusBadRequest)
			return
		}
		agentID := parts[0]
		action := parts[1]

		agentDef, ok := dataStore.GetAgent(agentID)
		if !ok {
			http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
			return
		}

		switch action {
		case "speech":
			serveSpeechProxy(w, r, agentDef, dataStore)
		case "transcription":
			serveTranscriptionProxy(w, r, agentDef, dataStore)
		default:
			http.Error(w, `{"error":"unknown voice action"}`, http.StatusBadRequest)
		}
	})
}

func serveSpeechProxy(w http.ResponseWriter, r *http.Request, agentDef store.AgentDefinition, dataStore *store.Store) {
	if agentDef.TTS.Backend == "" {
		http.Error(w, `{"error":"TTS not configured for this agent"}`, http.StatusServiceUnavailable)
		return
	}

	backend, ok := dataStore.GetBackend(agentDef.TTS.Backend)
	if !ok || backend.URL == "" {
		http.Error(w, `{"error":"TTS backend not found"}`, http.StatusServiceUnavailable)
		return
	}

	target, err := url.Parse(backend.URL)
	if err != nil {
		http.Error(w, `{"error":"invalid TTS backend URL"}`, http.StatusInternalServerError)
		return
	}

	var inputBody map[string]interface{}
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &inputBody); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
		}
	}
	if inputBody == nil {
		inputBody = make(map[string]interface{})
	}

	inputBody["model"] = agentDef.TTS.Model
	inputBody["voice"] = agentDef.TTS.Voice
	inputBody["speed"] = agentDef.TTS.Speed

	newBody, err := json.Marshal(inputBody)
	if err != nil {
		http.Error(w, "Failed to build request", http.StatusInternalServerError)
		return
	}

	proxyURL := *target
	proxyURL.Path = "/v1/audio/speech"

	proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", proxyURL.String(), bytes.NewReader(newBody))
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")
	if backend.APIKey != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+backend.APIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		slog.Error("TTS proxy error", "error", err)
		http.Error(w, "TTS service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func serveTranscriptionProxy(w http.ResponseWriter, r *http.Request, agentDef store.AgentDefinition, dataStore *store.Store) {
	if agentDef.Transcription.Backend == "" {
		http.Error(w, `{"error":"transcription not configured for this agent"}`, http.StatusServiceUnavailable)
		return
	}

	backend, ok := dataStore.GetBackend(agentDef.Transcription.Backend)
	if !ok || backend.URL == "" {
		http.Error(w, `{"error":"transcription backend not found"}`, http.StatusServiceUnavailable)
		return
	}

	target, err := url.Parse(backend.URL)
	if err != nil {
		http.Error(w, `{"error":"invalid transcription backend URL"}`, http.StatusInternalServerError)
		return
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = "/v1/audio/transcriptions"
			req.Host = target.Host
			if backend.APIKey != "" {
				req.Header.Set("Authorization", "Bearer "+backend.APIKey)
			}
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("Transcription proxy error", "error", err, "path", r.URL.Path)
			http.Error(w, "Transcription service unavailable", http.StatusBadGateway)
		},
	}
	proxy.ServeHTTP(w, r)
}

// agentRouterHandler is a hot-swappable HTTP handler that routes agent requests.
// It is rebuilt whenever the store changes.
type agentRouterHandler struct {
	mu           sync.RWMutex
	agentHandler http.Handler
}

func (h *agentRouterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mu.RLock()
	handler := h.agentHandler
	h.mu.RUnlock()

	if handler != nil {
		handler.ServeHTTP(w, r)
	} else {
		http.Error(w, `{"error":"no agent configured"}`, http.StatusServiceUnavailable)
	}
}

func (h *agentRouterHandler) rebuild(ctx context.Context, dataStore *store.Store) {
	storeData := dataStore.Data()

	var agentHandler http.Handler
	if len(storeData.Agents) > 0 {
		svc, err := agent.New(ctx, storeData.Agents, storeData.Backends, storeData.MemoryProviders, storeData.MCPServers)
		if err != nil {
			slog.Warn("Failed to initialize agents", "error", err)
		} else {
			agentHandler = http.StripPrefix("/api/v1/agent", svc.Handler())
		}
	} else {
		slog.Warn("No agents defined in store")
	}

	h.mu.Lock()
	h.agentHandler = agentHandler
	h.mu.Unlock()
}
