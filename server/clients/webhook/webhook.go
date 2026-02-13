package webhook

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/store"
)

// Handler serves webhook client endpoints.
// Each webhook client has a unique URL: /api/v1/webhooks/{clientID}
// Authentication uses the client's token via Authorization: Bearer header.
type Handler struct {
	executor *clients.Executor
	store    *store.Store
	logger   *slog.Logger
	router   *mux.Router
}

type webhookRequest struct {
	Prompt string `json:"prompt,omitempty"`
}

type webhookResponse struct {
	OK       bool   `json:"ok"`
	Response string `json:"response,omitempty"`
	Error    string `json:"error,omitempty"`
}

// NewHandler creates the webhook HTTP handler.
func NewHandler(executor *clients.Executor, s *store.Store, logger *slog.Logger) *Handler {
	h := &Handler{
		executor: executor,
		store:    s,
		logger:   logger,
	}
	h.router = mux.NewRouter()
	h.router.HandleFunc("/{id}", h.handle).Methods("POST")
	return h
}

// ServeHTTP implements http.Handler.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	clientID := mux.Vars(r)["id"]

	cl, ok := h.store.GetClient(clientID)
	if !ok || cl.Type != "webhook" {
		writeError(w, http.StatusNotFound, "webhook not found")
		return
	}

	if !cl.Enabled {
		writeError(w, http.StatusForbidden, "webhook client is disabled")
		return
	}

	token := r.Header.Get("Authorization")
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	} else {
		token = ""
	}

	if token == "" || token != cl.Token {
		writeError(w, http.StatusUnauthorized, "invalid or missing token")
		return
	}

	var req webhookRequest
	if r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}
	}

	h.logger.Info("Webhook client firing", "client", cl.Name, "id", cl.ID)

	result, err := h.executor.RunClient(r.Context(), cl, req.Prompt)
	if err != nil {
		h.logger.Error("Webhook client failed", "client", cl.Name, "error", err)
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logger.Info("Webhook client completed", "client", cl.Name, "responseLen", len(result))
	writeJSON(w, http.StatusOK, webhookResponse{OK: true, Response: result})
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, webhookResponse{OK: false, Error: message})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
