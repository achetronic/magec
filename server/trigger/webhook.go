package trigger

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// WebhookHandler serves webhook client endpoints.
// Each webhook client has a unique URL: /api/v1/webhooks/{clientID}
// Authentication uses the client's token via Authorization: Bearer header.
type WebhookHandler struct {
	executor *Executor
	store    *store.Store
	logger   *slog.Logger
	router   *mux.Router
}

// webhookRequest is the JSON body expected from incoming webhook calls.
type webhookRequest struct {
	Prompt string `json:"prompt,omitempty"`
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
	clientID := mux.Vars(r)["id"]

	cl, ok := h.store.GetClient(clientID)
	if !ok || cl.Type != "webhook" {
		writeWebhookError(w, http.StatusNotFound, "webhook not found")
		return
	}

	if !cl.Enabled {
		writeWebhookError(w, http.StatusForbidden, "webhook client is disabled")
		return
	}

	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	} else {
		token = ""
	}

	if token == "" || token != cl.Token {
		writeWebhookError(w, http.StatusUnauthorized, "invalid or missing token")
		return
	}

	var req webhookRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeWebhookError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
	}

	h.logger.Info("Webhook client firing", "client", cl.Name, "id", cl.ID)

	result, err := h.executor.RunClient(r.Context(), cl, req.Prompt)
	if err != nil {
		h.logger.Error("Webhook client failed", "client", cl.Name, "error", err)
		writeWebhookError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("Webhook client completed", "client", cl.Name, "responseLen", len(result))
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
