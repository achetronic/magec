package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listCronJobs(w http.ResponseWriter, r *http.Request) {
	crons := h.store.ListCronJobs()
	writeJSON(w, http.StatusOK, crons)
}

func (h *Handler) getCronJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := h.store.GetCronJob(id)
	if !ok {
		writeError(w, http.StatusNotFound, "cron job not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

func (h *Handler) createCronJob(w http.ResponseWriter, r *http.Request) {
	var c store.CronJob
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if c.Schedule == "" {
		writeError(w, http.StatusBadRequest, "schedule is required")
		return
	}
	if c.AgentID == "" {
		writeError(w, http.StatusBadRequest, "agentId is required")
		return
	}
	created, err := h.store.CreateCronJob(c)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateCronJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c store.CronJob
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := h.store.UpdateCronJob(id, c); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetCronJob(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteCronJob(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteCronJob(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
