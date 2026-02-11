package admin

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listCronJobs returns all cron jobs.
// @Summary      List cron jobs
// @Description  Returns all configured scheduled tasks
// @Tags         crons
// @Produce      json
// @Success      200  {array}  store.CronJob
// @Router       /crons [get]
func (h *Handler) listCronJobs(w http.ResponseWriter, r *http.Request) {
	crons := h.store.ListCronJobs()
	writeJSON(w, http.StatusOK, crons)
}

// getCronJob returns a single cron job by name.
// @Summary      Get cron job
// @Description  Returns a cron job by its unique name
// @Tags         crons
// @Produce      json
// @Param        name  path      string  true  "Cron job name"
// @Success      200   {object}  store.CronJob
// @Failure      404   {object}  ErrorResponse
// @Router       /crons/{name} [get]
func (h *Handler) getCronJob(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	c, ok := h.store.GetCronJob(name)
	if !ok {
		writeError(w, http.StatusNotFound, "cron job not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// createCronJob creates a new cron job.
// @Summary      Create cron job
// @Description  Creates a new scheduled task that sends a prompt to an agent
// @Tags         crons
// @Accept       json
// @Produce      json
// @Param        body  body      store.CronJob  true  "Cron job definition"
// @Success      201   {object}  store.CronJob
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Router       /crons [post]
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
	if err := h.store.CreateCronJob(c); err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, c)
}

// updateCronJob updates an existing cron job.
// @Summary      Update cron job
// @Description  Updates a cron job by name
// @Tags         crons
// @Accept       json
// @Produce      json
// @Param        name  path      string         true  "Cron job name"
// @Param        body  body      store.CronJob  true  "Cron job definition"
// @Success      200   {object}  store.CronJob
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Router       /crons/{name} [put]
func (h *Handler) updateCronJob(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	var c store.CronJob
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Name != "" && c.Name != name {
		if err := h.store.RenameCronJob(name, c.Name); err != nil {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		name = c.Name
	}
	if err := h.store.UpdateCronJob(name, c); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	c.Name = name
	writeJSON(w, http.StatusOK, c)
}

// deleteCronJob deletes a cron job.
// @Summary      Delete cron job
// @Description  Deletes a cron job by name
// @Tags         crons
// @Param        name  path  string  true  "Cron job name"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /crons/{name} [delete]
func (h *Handler) deleteCronJob(w http.ResponseWriter, r *http.Request) {
	name := mux.Vars(r)["name"]
	if err := h.store.DeleteCronJob(name); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
