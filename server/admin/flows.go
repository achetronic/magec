package admin

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

func (h *Handler) listFlows(w http.ResponseWriter, r *http.Request) {
	flows := h.store.ListFlows()
	writeJSON(w, http.StatusOK, flows)
}

func (h *Handler) getFlow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	f, ok := h.store.GetFlow(id)
	if !ok {
		writeError(w, http.StatusNotFound, "flow not found")
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (h *Handler) createFlow(w http.ResponseWriter, r *http.Request) {
	var f store.FlowDefinition
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if f.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if err := validateFlowStep(&f.Root); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.store.CreateFlow(f)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateFlow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var f store.FlowDefinition
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if err := validateFlowStep(&f.Root); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.store.UpdateFlow(id, f); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetFlow(id)
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteFlow(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteFlow(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateFlowStep(step *store.FlowStep) error {
	switch step.Type {
	case store.FlowStepAgent:
		if step.AgentID == "" {
			return fmt.Errorf("agent step requires agentId")
		}
	case store.FlowStepSequential, store.FlowStepParallel:
		if len(step.Steps) == 0 {
			return fmt.Errorf("%s step requires at least one child step", step.Type)
		}
		for i := range step.Steps {
			if err := validateFlowStep(&step.Steps[i]); err != nil {
				return err
			}
		}
	case store.FlowStepLoop:
		if len(step.Steps) == 0 {
			return fmt.Errorf("loop step requires at least one child step")
		}
		for i := range step.Steps {
			if err := validateFlowStep(&step.Steps[i]); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("unknown step type %q", step.Type)
	}
	return nil
}
