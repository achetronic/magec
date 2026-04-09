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
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/achetronic/adk-utils-go/plugin/contextguard"
	"github.com/achetronic/magec/server/agent"
	mageca2a "github.com/achetronic/magec/server/a2a"
	"github.com/achetronic/magec/server/api/admin"
	user "github.com/achetronic/magec/server/api/user"
	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/clients/cron"
	discordclient "github.com/achetronic/magec/server/clients/discord"
	slackclient "github.com/achetronic/magec/server/clients/slack"
	"github.com/achetronic/magec/server/clients/telegram"
	"github.com/achetronic/magec/server/clients/webhook"
	"github.com/achetronic/magec/server/config"
	"github.com/achetronic/magec/server/frontend"
	"github.com/achetronic/magec/server/logging"
	"github.com/achetronic/magec/server/middleware"
	"github.com/achetronic/magec/server/models"
	"github.com/achetronic/magec/server/store"
	"github.com/achetronic/magec/server/voice"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	_ "github.com/achetronic/magec/server/api/user/docs"

	_ "github.com/achetronic/magec/server/api/admin/docs"
	_ "github.com/achetronic/magec/server/clients/direct"
	_ "github.com/achetronic/magec/server/clients/discord"
	_ "github.com/achetronic/magec/server/clients/slack"
	_ "github.com/achetronic/magec/server/memory/postgres"
	_ "github.com/achetronic/magec/server/memory/redis"
	_ "github.com/achetronic/magec/server/voice/gemini"
	_ "github.com/achetronic/magec/server/voice/openai"
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

	if cfg.Server.AdminPassword == "" {
		slog.Warn("Admin API is unprotected — set server.adminPassword in config")
	}

	// Check runtime dependencies
	checkDependencies(cfg)

	ctx := context.Background()

	// Initialize store with JSON persistence
	dataStore, err := store.New("data/store.json", cfg.Server.EncryptionKey)
	if err != nil {
		slog.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}
	slog.Info("Store initialized", "agents", len(dataStore.Data().Agents), "backends", len(dataStore.Data().Backends))

	if cfg.Server.EncryptionKey == "" && len(dataStore.Data().Secrets) > 0 {
		slog.Warn("Secrets are stored without encryption — set server.encryptionKey in config")
	}

	// Initialize conversation store for audit logging
	convoStore, err := store.NewConversationStore("data/conversations.json")
	if err != nil {
		slog.Warn("Failed to initialize conversation store", "error", err)
		convoStore, _ = store.NewConversationStore("")
	}
	slog.Info("Conversation store initialized", "conversations", convoStore.Count())

	// Admin API — start first so it's available even if agent init fails
	adminHandler := admin.New(dataStore)
	adminHandler.SetConversationStore(convoStore)

	adminMux := http.NewServeMux()
	adminMux.Handle("/api/v1/admin/", http.StripPrefix("/api/v1/admin", adminHandler))
	adminMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))
	adminFS, err := frontend.AdminUI()
	if err != nil {
		slog.Error("Failed to load admin UI", "error", err)
		os.Exit(1)
	}
	adminMux.Handle("/", http.FileServer(http.FS(adminFS)))

	adminAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.AdminPort)
	adminServer := &http.Server{
		Addr:         adminAddr,
		Handler:      middleware.AccessLog(middleware.CORS(middleware.AdminAuth(adminMux, cfg.Server.AdminPassword))),
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

	// cwRegistry provides LLM context window sizes from catwalk's embedded database.
	cwRegistry := contextguard.NewCrushRegistry()

	// A2A (Agent-to-Agent) protocol handler
	a2aPublicURL := cfg.Server.PublicURL
	if a2aPublicURL == "" {
		a2aPublicURL = fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	}
	a2aHandler := mageca2a.NewHandler(a2aPublicURL)

	// Swappable handler for agent-related routes (hot-reloaded on store changes)
	agentRouter := &agentRouterHandler{adminHandler: adminHandler, a2aHandler: a2aHandler, cwRegistry: cwRegistry}
	agentRouter.rebuild(ctx, dataStore)

	// Executor for running commands against agents (cron, webhooks, etc.)
	agentURL := fmt.Sprintf("http://127.0.0.1:%d/api/v1/agent", cfg.Server.Port)
	executor := clients.NewExecutor(dataStore, agentURL, slog.Default())
	executor.SetConversationStore(convoStore)

	httpMux := http.NewServeMux()
	// Chain: Client ← RecorderUser ← FlowFilter ← RecorderAdmin ← SessionEnsure ← SessionStateSeed ← SSEIdleTimeout ← ADK
	idleGuarded := middleware.SSEIdleTimeout(agentRouter, 15*time.Minute)
	seeded := middleware.SessionEnsure(middleware.SessionStateSeed(idleGuarded, dataStore))
	adminRecorded := middleware.ConversationRecorder(
		middleware.ConversationRecorderSSE(seeded, executor, dataStore, "admin"),
		executor, dataStore, "admin",
	)
	filtered := middleware.FlowResponseFilter(adminRecorded, dataStore)
	userRecorded := middleware.ConversationRecorder(
		middleware.ConversationRecorderSSE(filtered, executor, dataStore, "user"),
		executor, dataStore, "user",
	)
	httpMux.Handle("/api/v1/agent/", middleware.SnakeCaseNormalize(userRecorded))
	httpMux.Handle("/api/v1/voice/", newVoiceHandler(dataStore, agentRouter))

	// A2A protocol endpoints (global discovery + per-agent card + JSON-RPC invoke)
	httpMux.HandleFunc("/api/v1/a2a/", a2aHandler.ServeA2A)

	userAPI := user.New(dataStore)
	httpMux.HandleFunc("/api/v1/health", userAPI.Health)
	httpMux.HandleFunc("/api/v1/client/info", userAPI.ClientInfo)

	httpMux.Handle("/swagger/", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.InstanceName("userapi"),
	))

	// Voice events WebSocket handler (wake word + VAD)
	var voiceDetector *voice.Detector
	if *cfg.Voice.UI.Enabled {
		const defaultOnnxLibraryPath = "/usr/lib/libonnxruntime.so"
		onnxLibraryPath := defaultOnnxLibraryPath
		if cfg.Voice.OnnxLibraryPath != "" {
			onnxLibraryPath = cfg.Voice.OnnxLibraryPath
		}

		wakewordYAML, err := models.WakewordConfig()
		if err != nil {
			slog.Warn("Wake word config not available", "error", err)
		} else {
			wakeWordModelsCfg, err := config.LoadWakeWordModels(wakewordYAML)
			if err != nil {
				slog.Warn("Wake word models not available", "error", err)
			} else if len(wakeWordModelsCfg.Models) == 0 {
				slog.Warn("No wake word models configured in wakewords.yaml")
			} else {
				voiceModels := make([]voice.ModelConfig, 0, len(wakeWordModelsCfg.Models))
				for _, m := range wakeWordModelsCfg.Models {
					data, err := models.ReadWakewordModel(m.File)
					if err != nil {
						slog.Warn("Failed to read wake word model", "model", m.ID, "error", err)
						continue
					}
					voiceModels = append(voiceModels, voice.ModelConfig{
						ID:        m.ID,
						Name:      m.Name,
						Data:      data,
						Phrase:    m.Phrase,
						Threshold: m.Threshold,
					})
				}

				melData, err1 := models.ReadAuxiliaryModel("mel-spectrogram.onnx")
				embData, err2 := models.ReadAuxiliaryModel("speech-embedding.onnx")
				vadData, err3 := models.ReadAuxiliaryModel("silero-vad.onnx")
				if err1 != nil || err2 != nil || err3 != nil {
					slog.Warn("Failed to read auxiliary models", "mel", err1, "embedding", err2, "vad", err3)
				} else {
					voiceDetector = voice.NewDetector(voice.DetectorConfig{
						MelspecModelData:   melData,
						EmbeddingModelData: embData,
						VADModelData:       vadData,
						Models:             voiceModels,
						OnnxLibraryPath:    onnxLibraryPath,
					}, slog.Default())

					if err := voiceDetector.Load(); err != nil {
						slog.Warn("Failed to load voice detection models", "error", err)
					} else {
						voiceHandler := voice.NewHandler(voiceDetector, slog.Default())
						httpMux.Handle("/api/v1/voice/events", voiceHandler)
						slog.Info("Voice detection enabled", "wakeWordModels", len(voiceModels), "vadEnabled", true)
					}
				}
			}
		}
	} else {
		slog.Info("Voice UI disabled via config")
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
	webhookHandler := webhook.NewHandler(executor, dataStore, slog.Default())
	httpMux.Handle("/api/v1/webhooks/", http.StripPrefix("/api/v1/webhooks", webhookHandler))

	// Static files
	if *cfg.Voice.UI.Enabled {
		voiceFS, err := frontend.VoiceUI()
		if err != nil {
			slog.Error("Failed to load voice UI", "error", err)
			os.Exit(1)
		}
		httpMux.Handle("/", http.FileServer(http.FS(voiceFS)))
	}

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

	// Start Telegram, Slack, and Discord clients (hot-reloaded on store changes)
	cm := newClientManager(dataStore, cfg.Server.Port, slog.Default())
	cm.start(ctx)

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		slog.Info("Shutting down...")
		cronScheduler.Stop()
		cm.stop()
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
			// If not found as agent, try resolving as a flow
			flow, flowOk := dataStore.GetFlow(agentID)
			if !flowOk {
				http.Error(w, `{"error":"agent not found"}`, http.StatusNotFound)
				return
			}
			firstID := flow.FirstAgentID()
			if firstID == "" {
				http.Error(w, `{"error":"flow has no agents"}`, http.StatusNotFound)
				return
			}
			agentDef, ok = dataStore.GetAgent(firstID)
			if !ok {
				http.Error(w, `{"error":"flow agent not found"}`, http.StatusNotFound)
				return
			}
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
// model/voice/speed from the agent's store config, and dispatches to the voice
// provider registered for the backend type.
func serveSpeechProxy(w http.ResponseWriter, r *http.Request, agentDef store.AgentDefinition, dataStore *store.Store) {
	if agentDef.TTS.Backend == "" {
		http.Error(w, `{"error":"TTS not configured for this agent"}`, http.StatusServiceUnavailable)
		return
	}

	backend, ok := dataStore.GetBackend(agentDef.TTS.Backend)
	if !ok {
		http.Error(w, `{"error":"TTS backend not found"}`, http.StatusServiceUnavailable)
		return
	}

	provider := voice.Get(backend.Type)
	if provider == nil || !provider.SupportsTTS() {
		http.Error(w, `{"error":"TTS not supported for this backend type"}`, http.StatusBadRequest)
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

	inputStr, _ := clientBody["input"].(string)
	rfStr, _ := clientBody["response_format"].(string)

	req := voice.TTSRequest{
		Input:          inputStr,
		Model:          agentDef.TTS.Model,
		Voice:          agentDef.TTS.Voice,
		ResponseFormat: rfStr,
		Config:         agentDef.TTS.Config,
	}

	resp, err := provider.SynthesizeSpeech(r.Context(), req, backend)
	if err != nil {
		slog.Error("TTS proxy error", "error", err)
		http.Error(w, "TTS service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Audio.Close()

	w.Header().Set("Content-Type", resp.ContentType)
	w.WriteHeader(http.StatusOK)
	io.Copy(w, resp.Audio)
}

// serveTranscriptionProxy forwards a speech-to-text request to the backend
// configured for the agent. It dispatches to the voice provider registered
// for the backend type.
func serveTranscriptionProxy(w http.ResponseWriter, r *http.Request, agentDef store.AgentDefinition, dataStore *store.Store) {
	if agentDef.Transcription.Backend == "" {
		http.Error(w, `{"error":"transcription not configured for this agent"}`, http.StatusServiceUnavailable)
		return
	}

	backend, ok := dataStore.GetBackend(agentDef.Transcription.Backend)
	if !ok {
		http.Error(w, `{"error":"transcription backend not found"}`, http.StatusServiceUnavailable)
		return
	}

	provider := voice.Get(backend.Type)
	if provider == nil || !provider.SupportsSTT() {
		http.Error(w, `{"error":"transcription not supported for this backend type"}`, http.StatusBadRequest)
		return
	}

	req := voice.STTRequest{
		Audio:       r.Body,
		ContentType: r.Header.Get("Content-Type"),
		Model:       agentDef.Transcription.Model,
		Config:      agentDef.Transcription.Config,
	}

	text, err := provider.TranscribeAudio(r.Context(), req, backend)
	if err != nil {
		slog.Error("Transcription proxy error", "error", err)
		http.Error(w, "Transcription service unavailable", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"text": text})
}

// agentRouterHandler is an HTTP handler that can be atomically swapped at
// runtime. It wraps the ADK agent handler and is rebuilt every time the store
// changes (agents added, removed, or modified).
type agentRouterHandler struct {
	mu           sync.RWMutex
	agentHandler http.Handler
	adminHandler *admin.Handler
	a2aHandler   *mageca2a.Handler
	// cwRegistry is passed through to agent.New so the ContextGuard plugin
	// can look up each model's context window at runtime.
	cwRegistry *contextguard.CrushRegistry
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
		svc, err := agent.New(ctx, storeData.Agents, storeData.Backends, storeData.MemoryProviders, storeData.MCPServers, storeData.Skills, storeData.Flows, storeData.Settings, h.cwRegistry)
		if err != nil {
			slog.Warn("Failed to initialize agents", "error", err)
		} else {
			agentHandler = http.StripPrefix("/api/v1/agent", svc.Handler())
			if h.adminHandler != nil {
				h.adminHandler.SetSessionService(svc.SessionService())
			}
			if h.a2aHandler != nil {
				h.a2aHandler.Rebuild(storeData.Agents, storeData.Flows, svc.ADKAgents(), svc.SessionService(), svc.MemoryService())
			}
		}
	} else {
		slog.Warn("No agents defined in store")
	}

	h.mu.Lock()
	h.agentHandler = agentHandler
	h.mu.Unlock()
}

// checkDependencies verifies that required external runtime dependencies are
// available and logs warnings for any that are missing.
func checkDependencies(cfg *config.Config) {
	var missing []string

	// ffmpeg — required for Telegram/Discord voice messages (audio→WAV conversion)
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		missing = append(missing, "ffmpeg (required for Telegram/Discord voice messages)")
	}

	// ONNX Runtime — required for voice UI (wake word + VAD detection)
	if *cfg.Voice.UI.Enabled {
		onnxPath := "/usr/lib/libonnxruntime.so"
		if cfg.Voice.OnnxLibraryPath != "" {
			onnxPath = cfg.Voice.OnnxLibraryPath
		}
		if _, err := os.Stat(onnxPath); os.IsNotExist(err) {
			missing = append(missing, fmt.Sprintf("libonnxruntime.so at %s (required for voice detection)", onnxPath))
		}
	}

	for _, dep := range missing {
		slog.Warn("Missing dependency", "dependency", dep)
	}
}

// clientManager manages the lifecycle of long-running clients (Telegram, Slack).
// It subscribes to store changes and reconciles running clients: stopping
// removed/disabled ones and starting new/re-enabled ones automatically.
type clientManager struct {
	store    *store.Store
	agentURL string
	logger   *slog.Logger

	mu      sync.Mutex
	running map[string]*managedClient
}

type managedClient struct {
	stop   func()
	cancel context.CancelFunc
	hash   string
}

func newClientManager(s *store.Store, port int, logger *slog.Logger) *clientManager {
	return &clientManager{
		store:    s,
		agentURL: fmt.Sprintf("http://127.0.0.1:%d/api/v1/agent", port),
		logger:   logger,
		running:  make(map[string]*managedClient),
	}
}

func (m *clientManager) start(ctx context.Context) {
	m.reconcile(ctx)

	changeCh := m.store.OnChange()
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-changeCh:
				time.Sleep(500 * time.Millisecond)
				m.reconcile(ctx)
			}
		}
	}()
}

func (m *clientManager) stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for id, mc := range m.running {
		mc.cancel()
		mc.stop()
		delete(m.running, id)
	}
}

func (m *clientManager) reconcile(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	desired := make(map[string]store.ClientDefinition)
	for _, cl := range m.store.ListClients() {
		if (cl.Type == "telegram" || cl.Type == "slack" || cl.Type == "discord") && cl.Enabled && len(cl.AllowedAgents) > 0 {
			desired[cl.ID] = cl
		}
	}

	for id, mc := range m.running {
		cl, ok := desired[id]
		if !ok {
			m.logger.Info("Stopping removed/disabled client", "id", id)
			mc.cancel()
			mc.stop()
			delete(m.running, id)
			continue
		}
		if h := clientHash(cl); h != mc.hash {
			m.logger.Info("Restarting changed client", "client", cl.Name)
			mc.cancel()
			mc.stop()
			delete(m.running, id)
		}
	}

	for id, cl := range desired {
		if _, ok := m.running[id]; ok {
			continue
		}
		switch cl.Type {
		case "telegram":
			m.startTelegram(ctx, cl)
		case "slack":
			m.startSlack(ctx, cl)
		case "discord":
			m.startDiscord(ctx, cl)
		}
	}
}

func clientHash(cl store.ClientDefinition) string {
	b, _ := json.Marshal(cl)
	return string(b)
}

func (m *clientManager) startTelegram(ctx context.Context, cl store.ClientDefinition) {
	if cl.Config.Telegram == nil {
		return
	}

	var agents []telegram.AgentInfo
	for _, agentID := range cl.AllowedAgents {
		agentDef, ok := m.store.GetAgent(agentID)
		if !ok {
			if flowDef, ok := m.store.GetFlow(agentID); ok {
				agents = append(agents, telegram.AgentInfo{ID: agentID, Name: flowDef.Name})
				continue
			}
			m.logger.Warn("Telegram client references unknown agent", "client", cl.Name, "agent", agentID)
			continue
		}
		agents = append(agents, telegram.AgentInfo{ID: agentID, Name: agentDef.Name})
	}

	if len(agents) == 0 {
		m.logger.Warn("Telegram client has no valid agents", "client", cl.Name)
		return
	}

	tgClient, err := telegram.New(cl, m.agentURL, agents, m.store, m.logger)
	if err != nil {
		m.logger.Error("Failed to create Telegram client", "client", cl.Name, "error", err)
		return
	}

	clientCtx, cancel := context.WithCancel(ctx)
	m.running[cl.ID] = &managedClient{stop: tgClient.Stop, cancel: cancel, hash: clientHash(cl)}

	go func(name string) {
		time.Sleep(500 * time.Millisecond)
		if err := tgClient.Start(clientCtx); err != nil {
			if clientCtx.Err() == nil {
				m.logger.Error("Telegram client error", "client", name, "error", err)
			}
		}
	}(cl.Name)

	m.logger.Info("Started Telegram client", "client", cl.Name)
}

func (m *clientManager) startSlack(ctx context.Context, cl store.ClientDefinition) {
	if cl.Config.Slack == nil {
		return
	}

	var agents []slackclient.AgentInfo
	for _, agentID := range cl.AllowedAgents {
		agentDef, ok := m.store.GetAgent(agentID)
		if !ok {
			if flowDef, ok := m.store.GetFlow(agentID); ok {
				agents = append(agents, slackclient.AgentInfo{ID: agentID, Name: flowDef.Name})
				continue
			}
			m.logger.Warn("Slack client references unknown agent", "client", cl.Name, "agent", agentID)
			continue
		}
		agents = append(agents, slackclient.AgentInfo{ID: agentID, Name: agentDef.Name})
	}

	if len(agents) == 0 {
		m.logger.Warn("Slack client has no valid agents", "client", cl.Name)
		return
	}

	skClient, err := slackclient.New(cl, m.agentURL, agents, m.store, m.logger)
	if err != nil {
		m.logger.Error("Failed to create Slack client", "client", cl.Name, "error", err)
		return
	}

	clientCtx, cancel := context.WithCancel(ctx)
	m.running[cl.ID] = &managedClient{stop: skClient.Stop, cancel: cancel, hash: clientHash(cl)}

	go func(name string) {
		time.Sleep(500 * time.Millisecond)
		if err := skClient.Start(clientCtx); err != nil {
			if clientCtx.Err() == nil {
				m.logger.Error("Slack client error", "client", name, "error", err)
			}
		}
	}(cl.Name)

	m.logger.Info("Started Slack client", "client", cl.Name)
}

func (m *clientManager) startDiscord(ctx context.Context, cl store.ClientDefinition) {
	if cl.Config.Discord == nil {
		return
	}

	var agents []discordclient.AgentInfo
	for _, agentID := range cl.AllowedAgents {
		agentDef, ok := m.store.GetAgent(agentID)
		if !ok {
			if flowDef, ok := m.store.GetFlow(agentID); ok {
				agents = append(agents, discordclient.AgentInfo{ID: agentID, Name: flowDef.Name})
				continue
			}
			m.logger.Warn("Discord client references unknown agent", "client", cl.Name, "agent", agentID)
			continue
		}
		agents = append(agents, discordclient.AgentInfo{ID: agentID, Name: agentDef.Name})
	}

	if len(agents) == 0 {
		m.logger.Warn("Discord client has no valid agents", "client", cl.Name)
		return
	}

	dcClient, err := discordclient.New(cl, m.agentURL, agents, m.store, m.logger)
	if err != nil {
		m.logger.Error("Failed to create Discord client", "client", cl.Name, "error", err)
		return
	}

	clientCtx, cancel := context.WithCancel(ctx)
	m.running[cl.ID] = &managedClient{stop: dcClient.Stop, cancel: cancel, hash: clientHash(cl)}

	go func(name string) {
		time.Sleep(500 * time.Millisecond)
		if err := dcClient.Start(clientCtx); err != nil {
			if clientCtx.Err() == nil {
				m.logger.Error("Discord client error", "client", name, "error", err)
			}
		}
	}(cl.Name)

	m.logger.Info("Started Discord client", "client", cl.Name)
}
