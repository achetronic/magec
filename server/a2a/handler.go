// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package a2a

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/a2aproject/a2a-go/a2a"
	"github.com/a2aproject/a2a-go/a2asrv"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/server/adka2a"
	"google.golang.org/adk/session"

	"github.com/achetronic/magec/server/store"
)

const protocolVersion = "0.2.5"

// Handler manages Agent-to-Agent (A2A) protocol endpoints.
// It maintains dynamic JSON-RPC handlers and Agent Cards for every agent
// or flow that has A2A enabled in its configuration.
type Handler struct {
	mu        sync.RWMutex
	handlers  map[string]http.Handler // agentID → JSON-RPC handler
	cards     map[string]*a2a.AgentCard
	publicURL string
}

// NewHandler creates a new A2A handler bound to the given public URL.
func NewHandler(publicURL string) *Handler {
	return &Handler{
		handlers:  make(map[string]http.Handler),
		cards:     make(map[string]*a2a.AgentCard),
		publicURL: strings.TrimRight(publicURL, "/"),
	}
}

// Rebuild regenerates the A2A endpoints and Agent Cards based on the current
// configuration of agents and flows. It should be called whenever the store changes.
func (h *Handler) Rebuild(agents []store.AgentDefinition, flows []store.FlowDefinition, adkAgents map[string]agent.Agent, sessionSvc session.Service, memorySvc memory.Service) {
	handlers := make(map[string]http.Handler)
	cards := make(map[string]*a2a.AgentCard)

	type a2aEntry struct {
		id, name, description string
		a2aCfg                *store.A2AConfig
	}

	var entries []a2aEntry
	for _, ag := range agents {
		entries = append(entries, a2aEntry{ag.ID, ag.Name, ag.Description, ag.A2A})
	}
	for _, fl := range flows {
		entries = append(entries, a2aEntry{fl.ID, fl.Name, fl.Description, fl.A2A})
	}

	for _, entry := range entries {
		if entry.a2aCfg == nil || !entry.a2aCfg.Enabled {
			continue
		}
		adkAgent, ok := adkAgents[entry.id]
		if !ok {
			slog.Warn("A2A: agent not found in ADK map", "agent", entry.id)
			continue
		}

		invokeURL := fmt.Sprintf("%s/api/v1/a2a/%s", h.publicURL, entry.id)

		card := &a2a.AgentCard{
			Name:               entry.name,
			Description:        entry.description,
			URL:                invokeURL,
			Version:            "1.0.0",
			ProtocolVersion:    protocolVersion,
			PreferredTransport: a2a.TransportProtocolJSONRPC,
			DefaultInputModes:  []string{"text/plain"},
			DefaultOutputModes: []string{"text/plain"},
			Capabilities: a2a.AgentCapabilities{
				Streaming: true,
			},
			Skills: adka2a.BuildAgentSkills(adkAgent),
			SecuritySchemes: a2a.NamedSecuritySchemes{
				"bearer": a2a.HTTPAuthSecurityScheme{
					Scheme:      "bearer",
					Description: "Magec client token",
				},
			},
			Security: []a2a.SecurityRequirements{
				{"bearer": a2a.SecuritySchemeScopes{}},
			},
		}
		cards[entry.id] = card

		execCfg := adka2a.ExecutorConfig{
			RunnerConfig: runner.Config{
				AppName:        entry.id,
				Agent:          adkAgent,
				SessionService: sessionSvc,
				MemoryService:  memorySvc,
			},
		}
		executor := adka2a.NewExecutor(execCfg)
		reqHandler := a2asrv.NewHandler(executor, a2asrv.WithLogger(slog.Default()))
		handlers[entry.id] = a2asrv.NewJSONRPCHandler(reqHandler)

		slog.Info("A2A endpoint enabled", "agent", entry.id, "name", entry.name)
	}

	h.mu.Lock()
	h.handlers = handlers
	h.cards = cards
	h.mu.Unlock()
}

const a2aPrefix = "/api/v1/a2a/"
const wellKnownSuffix = "/.well-known/agent-card.json"

// ServeA2A routes incoming HTTP requests to either the JSON-RPC execution
// handler or the well-known Agent Card discovery endpoints.
func (h *Handler) ServeA2A(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	rest := strings.TrimPrefix(path, a2aPrefix)

	if rest == ".well-known/agent-card.json" {
		h.serveGlobalCards(w, r)
		return
	}

	if strings.HasSuffix(path, wellKnownSuffix) {
		h.servePerAgentCard(w, r)
		return
	}

	h.serveJSONRPC(w, r)
}

func (h *Handler) serveGlobalCards(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent")

	h.mu.RLock()
	cards := h.cards
	h.mu.RUnlock()

	if agentID != "" {
		card, ok := cards[agentID]
		if !ok {
			http.Error(w, `{"error":"agent not found or A2A not enabled"}`, http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		json.NewEncoder(w).Encode(card)
		return
	}

	cardList := make([]*a2a.AgentCard, 0, len(cards))
	for _, card := range cards {
		cardList = append(cardList, card)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(cardList)
}

func (h *Handler) servePerAgentCard(w http.ResponseWriter, r *http.Request) {
	agentID := extractAgentID(r.URL.Path, a2aPrefix, wellKnownSuffix)

	h.mu.RLock()
	card, ok := h.cards[agentID]
	h.mu.RUnlock()

	if !ok {
		http.Error(w, `{"error":"agent not found or A2A not enabled"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(card)
}

func (h *Handler) serveJSONRPC(w http.ResponseWriter, r *http.Request) {
	agentID := extractAgentID(r.URL.Path, a2aPrefix, "")

	h.mu.RLock()
	handler, ok := h.handlers[agentID]
	h.mu.RUnlock()

	if !ok {
		w.Header().Set("Content-Type", "application/json")
		http.Error(w, `{"error":"agent not found or A2A not enabled"}`, http.StatusNotFound)
		return
	}

	handler.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), agentIDKey, agentID)))
}

type contextKey string

const agentIDKey contextKey = "a2a-agent-id"

func extractAgentID(path, prefix, suffix string) string {
	path = strings.TrimPrefix(path, prefix)
	if suffix != "" {
		path = strings.TrimSuffix(path, suffix)
	}
	path = strings.TrimSuffix(path, "/")
	if idx := strings.Index(path, "/"); idx >= 0 {
		path = path[:idx]
	}
	return path
}
