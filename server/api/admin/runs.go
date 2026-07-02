package admin

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/agent/runrecorder"
	"github.com/achetronic/magec/server/runs"
)

type runSummaryView struct {
	RunID      string    `json:"runId"`
	AppName    string    `json:"appName"`
	SessionID  string    `json:"sessionId"`
	UserID     string    `json:"userId,omitempty"`
	ClientID   string    `json:"clientId,omitempty"`
	Source     string    `json:"source,omitempty"`
	StartedAt  time.Time `json:"startedAt"`
	EndedAt    time.Time `json:"endedAt,omitempty"`
	Status     string    `json:"status"`
	Error      string    `json:"error,omitempty"`
	EventCount int       `json:"eventCount"`
}

type runActivation struct {
	Node          string    `json:"node"`
	Seq           int       `json:"seq"`
	StartedAt     time.Time `json:"startedAt"`
	EndedAt       time.Time `json:"endedAt"`
	Events        int       `json:"events"`
	Routes        []string  `json:"routes,omitempty"`
	Branch        string    `json:"branch,omitempty"`
	OutputPreview string    `json:"outputPreview,omitempty"`
	Error         string    `json:"error,omitempty"`
}

type runDetail struct {
	RunID       string            `json:"runId"`
	AppName     string            `json:"appName"`
	SessionID   string            `json:"sessionId"`
	UserID      string            `json:"userId,omitempty"`
	ClientID    string            `json:"clientId,omitempty"`
	Source      string            `json:"source,omitempty"`
	StartedAt   time.Time         `json:"startedAt"`
	EndedAt     time.Time         `json:"endedAt,omitempty"`
	Status      string            `json:"status"`
	Error       string            `json:"error,omitempty"`
	Activations []runActivation   `json:"activations"`
	Events      []json.RawMessage `json:"events,omitempty"`
}

// listRuns handles the GET /runs endpoint to retrieve run summaries.
// @Summary      List runs
// @Description  Returns a paginated list of run audit logs, newest first. Filters by appName or status.
// @Tags         runs
// @Produce      json
// @Param        appName   query     string  false  "Filter by app name"
// @Param        status    query     string  false  "Filter by status"
// @Param        limit     query     int     false  "Max items to return (default 30, 0 for all)"
// @Param        offset    query     int     false  "Items to skip (default 0)"
// @Success      200       {object}  map[string]interface{}
// @Security     AdminAuth
// @Router       /runs [get]
func (h *Handler) listRuns(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"items": []runSummaryView{},
			"total": 0,
		})
		return
	}

	appName := r.URL.Query().Get("appName")
	status := r.URL.Query().Get("status")
	limit := queryInt(r, "limit", 30)
	offset := queryInt(r, "offset", 0)

	summaries, total, err := h.runs.ListRuns(runs.RunFilter{
		AppName: appName,
		Status:  status,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	views := []runSummaryView{}
	for _, s := range summaries {
		views = append(views, runSummaryView{
			RunID:      s.RunID,
			AppName:    s.AppName,
			SessionID:  s.SessionID,
			UserID:     s.UserID,
			ClientID:   s.ClientID,
			Source:     s.Source,
			StartedAt:  s.StartedAt,
			EndedAt:    s.EndedAt,
			Status:     s.Status,
			Error:      s.Error,
			EventCount: s.EventCount,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": views,
		"total": total,
	})
}

// getRun handles GET /runs/{id} to retrieve details of a specific run.
// @Summary      Get run
// @Description  Returns a run detail by ID, including its timeline of node activations and optionally raw events.
// @Tags         runs
// @Produce      json
// @Param        id        path      string  true   "Run ID"
// @Param        raw       query     bool    false  "Include raw events payload"
// @Success      200       {object}  map[string]interface{}
// @Failure      404       {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /runs/{id} [get]
func (h *Handler) getRun(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeError(w, http.StatusNotFound, "runs store not initialized")
		return
	}

	id := mux.Vars(r)["id"]
	raw := r.URL.Query().Get("raw") == "true"

	run, ok, err := h.runs.GetRun(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !ok {
		writeError(w, http.StatusNotFound, "run not found")
		return
	}

	var eventsRaw []json.RawMessage
	if raw {
		eventsRaw = make([]json.RawMessage, len(run.Events))
		for i, ev := range run.Events {
			eventsRaw[i] = json.RawMessage(ev.Payload)
		}
	}

	detail := runDetail{
		RunID:       run.RunID,
		AppName:     run.AppName,
		SessionID:   run.SessionID,
		UserID:      run.UserID,
		ClientID:    run.ClientID,
		Source:      run.Source,
		StartedAt:   run.StartedAt,
		EndedAt:     run.EndedAt,
		Status:      run.Status,
		Error:       run.Error,
		Activations: projectActivations(run.Events, run.AppName),
		Events:      eventsRaw,
	}

	writeJSON(w, http.StatusOK, detail)
}

// projectActivations collapses raw event sequence into node activations.
// It groups sequential events that represent a single execution phase.
func projectActivations(events []runrecorder.EventRecord, runAppName string) []runActivation {
	if len(events) == 0 {
		return nil
	}

	var activations []runActivation
	var currentEvents []runrecorder.EventRecord
	var currentKey string

	for _, ev := range events {
		key := ""
		if ev.Author != "" && ev.Author != runAppName {
			key = ev.Author
		} else if ev.NodePath != "" {
			key = ev.NodePath
		} else {
			key = "workflow"
		}

		if len(currentEvents) == 0 {
			currentKey = key
			currentEvents = append(currentEvents, ev)
		} else if key == currentKey {
			currentEvents = append(currentEvents, ev)
		} else {
			activations = append(activations, buildActivation(currentKey, currentEvents))
			currentKey = key
			currentEvents = []runrecorder.EventRecord{ev}
		}
	}
	if len(currentEvents) > 0 {
		activations = append(activations, buildActivation(currentKey, currentEvents))
	}

	return activations
}

// buildActivation groups sequential event attributes into a run activation summary.
// It analyzes payloads to extract outputs, preview strings, and node error status.
func buildActivation(node string, events []runrecorder.EventRecord) runActivation {
	first := events[0]
	last := events[len(events)-1]

	var branch string
	for _, ev := range events {
		if ev.Branch != "" {
			branch = ev.Branch
			break
		}
	}

	var routes []string
	seenRoutes := make(map[string]bool)
	for _, ev := range events {
		for _, r := range ev.Routes {
			if !seenRoutes[r] {
				seenRoutes[r] = true
				routes = append(routes, r)
			}
		}
	}

	var lastOutput string
	var lastError string
	for _, ev := range events {
		if len(ev.Payload) > 0 {
			var payloadMap map[string]any
			if err := json.Unmarshal(ev.Payload, &payloadMap); err == nil {
				if val, ok := payloadMap["output"]; ok && val != nil {
					var outStr string
					if s, ok := val.(string); ok {
						outStr = s
					} else {
						if bs, err := json.Marshal(val); err == nil {
							outStr = string(bs)
						}
					}
					if outStr != "" {
						lastOutput = outStr
					}
				}
				if val, ok := payloadMap["errorMessage"]; ok && val != nil {
					if s, ok := val.(string); ok && s != "" {
						lastError = s
					}
				}
			}
		}
	}

	if len([]rune(lastOutput)) > 200 {
		lastOutput = string([]rune(lastOutput)[:200]) + "..."
	}

	return runActivation{
		Node:          node,
		Seq:           first.Seq,
		StartedAt:     first.Timestamp,
		EndedAt:       last.Timestamp,
		Events:        len(events),
		Routes:        routes,
		Branch:        branch,
		OutputPreview: lastOutput,
		Error:         lastError,
	}
}
