package trigger

import (
	"crypto/subtle"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// WebhookHandler serves webhook trigger endpoints.
// Each trigger has a unique URL: /api/v1/webhooks/{triggerID}
type WebhookHandler struct {
	executor *Executor
	store    *store.Store
	logger   *slog.Logger
	router   *mux.Router
}

// webhookRequest is the JSON body expected from incoming webhook calls.
type webhookRequest struct {
	Prompt  string `json:"prompt,omitempty"`
	AgentID string `json:"agentId,omitempty"`
	Secret  string `json:"secret,omitempty"`
}

type webhookResponse struct {
	OK       bool   `json:"ok"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}

// NewWebhookHandler creates the webhook HTTP handler.
func NewWebhookHandler(executor *Executor, s *store.Store, logger *slog.Logger) *WebhookHandler {
	h := &WebhookHandler{
		executor: executor,
		store:    s,
		logger:   logger,
	}
	h.router = mux.NewRouter()
	h.router.HandleFunc("/{id}", h.handle).Methods("POST")
	return h
}

// ServeHTTP implements http.Handler.
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *WebhookHandler) handle(w http.ResponseWriter, r *http.Request) {
	triggerID := mux.Vars(r)["id"]

	trigger, ok := h.store.GetTrigger(triggerID)
	if !ok || trigger.Type != store.TriggerTypeWebhook {
		writeWebhookError(w, http.StatusNotFound, "webhook not found")
		return
	}

	if !trigger.Enabled {
		writeWebhookError(w, http.StatusForbidden, "trigger is disabled")
		return
	}

	var req webhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWebhookError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}

	if trigger.Webhook != nil && trigger.Webhook.Secret != "" {
		secret := req.Secret
		if secret == "" {
			secret = r.Header.Get("X-Webhook-Secret")
		}
		if subtle.ConstantTimeCompare([]byte(trigger.Webhook.Secret), []byte(secret)) != 1 {
			writeWebhookError(w, http.StatusUnauthorized, "invalid secret")
			return
		}
	}

	h.logger.Info("Webhook trigger firing", "trigger", trigger.Name, "id", trigger.ID)

	result, err := h.executor.RunTrigger(r.Context(), trigger, req.Prompt, req.AgentID)
	if err != nil {
		h.logger.Error("Webhook trigger failed", "trigger", trigger.Name, "error", err)
		writeWebhookError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("Webhook trigger completed", "trigger", trigger.Name, "responseLen", len(result))
	writeWebhookJSON(w, http.StatusOK, webhookResponse{OK: true, Response: result})
}

func writeWebhookError(w http.ResponseWriter, status int, message string) {
	writeWebhookJSON(w, status, webhookResponse{OK: false, Error: message})
}

func writeWebhookJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
