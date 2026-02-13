package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listTriggers(w http.ResponseWriter, r *http.Request) {
	triggers := h.store.ListTriggers()
	writeJSON(w, http.StatusOK, triggers)
}

func (h *Handler) getTrigger(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	t, ok := h.store.GetTrigger(id)
	if !ok {
		writeError(w, http.StatusNotFound, "trigger not found")
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (h *Handler) createTrigger(w http.ResponseWriter, r *http.Request) {
	var t store.Trigger
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if t.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if t.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	switch t.Type {
	case store.TriggerTypeCron:
		if t.Cron == nil || t.Cron.Schedule == "" {
			writeError(w, http.StatusBadRequest, "cron trigger requires cron.schedule")
			return
		}
		if t.CommandID == "" {
			writeError(w, http.StatusBadRequest, "cron trigger requires commandId")
			return
		}
		if t.AgentID == "" {
			writeError(w, http.StatusBadRequest, "cron trigger requires agentId")
			return
		}
	case store.TriggerTypeWebhook:
		if t.Webhook == nil {
			t.Webhook = &store.WebhookConfig{}
		}
		if !t.Webhook.Passthrough {
			if t.CommandID == "" {
				writeError(w, http.StatusBadRequest, "non-passthrough webhook requires commandId")
				return
			}
			if t.AgentID == "" {
				writeError(w, http.StatusBadRequest, "non-passthrough webhook requires agentId")
				return
			}
		}
	default:
		writeError(w, http.StatusBadRequest, "unknown trigger type: "+t.Type)
		return
	}
	created, err := h.store.CreateTrigger(t)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateTrigger(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var t store.Trigger
	if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateTrigger(id, t); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetTrigger(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteTrigger(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteTrigger(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
