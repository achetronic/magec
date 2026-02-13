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
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/achetronic/magec/server/api/admin"
	"github.com/achetronic/magec/server/agent"
	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/clients/cron"
	"github.com/achetronic/magec/server/clients/telegram"
	"github.com/achetronic/magec/server/clients/webhook"
	"github.com/achetronic/magec/server/config"
	"github.com/achetronic/magec/server/logging"
	"github.com/achetronic/magec/server/middleware"
	"github.com/achetronic/magec/server/store"
	user "github.com/achetronic/magec/server/api/user"
	"github.com/achetronic/magec/server/voice"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/achetronic/magec/server/api/user/docs"

	_ "github.com/achetronic/magec/server/api/admin/docs"
	_ "github.com/achetronic/magec/server/client/cron"
	_ "github.com/achetronic/magec/server/client/direct"
	_ "github.com/achetronic/magec/server/client/telegram"
	_ "github.com/achetronic/magec/server/client/webhook"
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
	adminMux.Handle("/", http.FileServer(http.Dir("admin-ui/dist")))

	adminAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.AdminPort)
	adminServer := &http.Server{
		Addr:         adminAddr,
		Handler:      middleware.AccessLog(middleware.CORS(adminMux)),
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

	userAPI := user.New(dataStore)
	httpMux.HandleFunc("/api/v1/health", userAPI.Health)
	httpMux.HandleFunc("/api/v1/client/info", userAPI.ClientInfo)

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

	// Webhook handler for trigger endpoints
	agentURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/agent", cfg.Server.Port)
	executor := clients.NewExecutor(dataStore, agentURL, slog.Default())
	webhookHandler := webhook.NewHandler(executor, dataStore, slog.Default())
	httpMux.Handle("/api/v1/webhooks/", http.StripPrefix("/api/v1/webhooks", webhookHandler))

	// Static files
	httpMux.Handle("/", http.FileServer(http.Dir("voice-ui")))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      middleware.AccessLog(middleware.CORS(middleware.ClientAuth(httpMux, dataStore))),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 15 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	// Start cron scheduler
	cronScheduler := cron.NewScheduler(executor, dataStore, slog.Default())
	go cronScheduler.Start(ctx)

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

		agentURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/agent", cfg.Server.Port)

		var allowedAgents []telegram.AgentInfo
		for _, agentID := range cl.AllowedAgents {
			agentDef, ok := dataStore.GetAgent(agentID)
			if !ok {
				slog.Warn("Telegram client references unknown agent, skipping agent", "client", cl.Name, "agent", agentID)
				continue
			}
			allowedAgents = append(allowedAgents, telegram.AgentInfo{
				ID:   agentID,
				Name: agentDef.Name,
			})
		}

		if len(allowedAgents) == 0 {
			slog.Warn("Telegram client has no valid agents after resolution, skipping", "client", cl.Name)
			continue
		}

		tgClient, err := telegram.New(cl, agentURL, allowedAgents, slog.Default())
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
		cronScheduler.Stop()
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

// newVoiceHandler creates a router for /api/v1/voice/{agentId}/{action} routes.
// It extracts the agent ID and action from the URL path, resolves the agent
// from the store, and dispatches to the speech (TTS) or transcription (STT) proxy.
// The /api/v1/voice/events WebSocket endpoint is handled separately.
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

// serveSpeechProxy forwards a TTS request to the backend configured for the agent.
// It reads only "input" and "response_format" from the client body, injects
// model/voice/speed from the agent's store config, and proxies the request
// to the backend's /v1/audio/speech endpoint.
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

	var clientBody map[string]interface{}
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}
		if len(body) > 0 {
			if err := json.Unmarshal(body, &clientBody); err != nil {
				http.Error(w, "Invalid JSON body", http.StatusBadRequest)
				return
			}
		}
	}
	if clientBody == nil {
		clientBody = make(map[string]interface{})
	}

	proxyBody := map[string]interface{}{
		"input": clientBody["input"],
		"model": agentDef.TTS.Model,
		"voice": agentDef.TTS.Voice,
		"speed": agentDef.TTS.Speed,
	}
	if rf, ok := clientBody["response_format"]; ok {
		proxyBody["response_format"] = rf
	}

	newBody, err := json.Marshal(proxyBody)
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

// serveTranscriptionProxy forwards a speech-to-text request to the backend
// configured for the agent. It injects the transcription model from the agent's
// store config into the multipart form before proxying to the backend's
// /v1/audio/transcriptions endpoint.
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

	body, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	contentType := r.Header.Get("Content-Type")

	var proxyBody bytes.Buffer
	if agentDef.Transcription.Model != "" && strings.Contains(contentType, "multipart/form-data") {
		boundary := ""
		for _, param := range strings.Split(contentType, ";") {
			param = strings.TrimSpace(param)
			if strings.HasPrefix(param, "boundary=") {
				boundary = strings.TrimPrefix(param, "boundary=")
				break
			}
		}
		if boundary == "" {
			http.Error(w, "Missing multipart boundary", http.StatusBadRequest)
			return
		}

		closingBoundary := fmt.Sprintf("\r\n--%s--", boundary)
		trimmed := bytes.TrimSuffix(body, []byte(closingBoundary))
		trimmed = bytes.TrimSuffix(trimmed, []byte(fmt.Sprintf("--%s--\r\n", boundary)))
		trimmed = bytes.TrimSuffix(trimmed, []byte(fmt.Sprintf("--%s--", boundary)))

		proxyBody.Write(trimmed)
		proxyBody.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		proxyBody.WriteString("Content-Disposition: form-data; name=\"model\"\r\n\r\n")
		proxyBody.WriteString(agentDef.Transcription.Model)
		proxyBody.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else {
		proxyBody.Write(body)
	}

	proxyURL := *target
	proxyURL.Path = "/v1/audio/transcriptions"

	proxyReq, err := http.NewRequestWithContext(r.Context(), "POST", proxyURL.String(), &proxyBody)
	if err != nil {
		http.Error(w, "Failed to create proxy request", http.StatusInternalServerError)
		return
	}
	proxyReq.Header.Set("Content-Type", contentType)
	if backend.APIKey != "" {
		proxyReq.Header.Set("Authorization", "Bearer "+backend.APIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(proxyReq)
	if err != nil {
		slog.Error("Transcription proxy error", "error", err)
		http.Error(w, "Transcription service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

// agentRouterHandler is an HTTP handler that can be atomically swapped at
// runtime. It wraps the ADK agent handler and is rebuilt every time the store
// changes (agents added, removed, or modified).
type agentRouterHandler struct {
	mu           sync.RWMutex
	agentHandler http.Handler
}

// ServeHTTP delegates to the current agent handler, or returns 503 if no
// agents are configured.
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

// rebuild recreates the ADK agent handler from the current store data and
// swaps it in atomically. Called on startup and whenever the store changes.
func (h *agentRouterHandler) rebuild(ctx context.Context, dataStore *store.Store) {
	storeData := dataStore.Data()

	var agentHandler http.Handler
	if len(storeData.Agents) > 0 {
		svc, err := agent.New(ctx, storeData.Agents, storeData.Backends, storeData.MemoryProviders, storeData.MCPServers, storeData.Flows)
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
