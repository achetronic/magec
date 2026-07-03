package admin

import (
	"encoding/json"
	"net/http"
	"strings"
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
	// NodeType is the flow node type this activation's node had at execution
	// time, resolved from the run's node-type snapshot. Empty for agent-only
	// runs and for runs recorded before the snapshot existed.
	NodeType string `json:"nodeType,omitempty"`
	// InputPreview is derived, not captured: a node's input is the output of
	// the activation that preceded it (routers and joins pass input through).
	InputPreview string `json:"inputPreview,omitempty"`
	// StateDelta holds the flow-state keys this activation wrote, prefix
	// stripped, internal bookkeeping keys hidden.
	StateDelta map[string]any `json:"stateDelta,omitempty"`
	// StateAfter is the accumulated flow state after this activation,
	// reconstructed by folding every StateDelta in event order.
	StateAfter map[string]any `json:"stateAfter,omitempty"`
}

type runDetail struct {
	RunID       string            `json:"runId"`
	AppName     string            `json:"appName"`
	SessionID   string            `json:"sessionId"`
	UserID      string            `json:"userId,omitempty"`
	ClientID    string            `json:"clientId,omitempty"`
	Source      string            `json:"source,omitempty"`
	Input       string            `json:"input,omitempty"`
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
		Input:       run.Input,
		StartedAt:   run.StartedAt,
		EndedAt:     run.EndedAt,
		Status:      run.Status,
		Error:       run.Error,
		Activations: projectActivations(run.Events, run.AppName, run.Input, run.NodeTypes),
		Events:      eventsRaw,
	}

	writeJSON(w, http.StatusOK, detail)
}

// projectActivations collapses raw event sequence into node activations.
// It groups sequential events that represent a single execution phase, then
// chains derived inputs (the run input for the first activation, the previous
// activation's output for the rest) and folds state deltas into an
// accumulated flow-state snapshot per activation.
func projectActivations(events []runrecorder.EventRecord, runAppName, runInput string, nodeTypes map[string]string) []runActivation {
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
			// Pure workflow nodes (joins) report the workflow name as Author,
			// so their key is the composite NodeInfo path; only the node's own
			// segment is meaningful to an operator.
			key = nodePathSegment(ev.NodePath)
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

	for i := range activations {
		activations[i].NodeType = nodeTypeFor(nodeTypes, activations[i].Node)
	}
	chainDerivedState(activations, runInput)
	return activations
}

// nodePathSegment extracts the node name from a composite NodeInfo path,
// "<flow>@1/<node>@2" reads <node>. A path without separators comes back
// unchanged.
func nodePathSegment(path string) string {
	segment := path
	if idx := strings.LastIndex(segment, "/"); idx >= 0 {
		segment = segment[idx+1:]
	}
	if idx := strings.Index(segment, "@"); idx >= 0 {
		segment = segment[:idx]
	}
	if segment == "" {
		return path
	}
	return segment
}

// nodeTypeFor resolves an activation's node type from the run's snapshot.
// The activation key is the node ID for Magec-built nodes and the node
// segment of the NodeInfo path for pure workflow nodes (joins), so a direct
// lookup covers both; the segment fallback guards records persisted before
// keys were shortened.
func nodeTypeFor(nodeTypes map[string]string, key string) string {
	if len(nodeTypes) == 0 {
		return ""
	}
	if t, ok := nodeTypes[key]; ok {
		return t
	}
	return nodeTypes[nodePathSegment(key)]
}

// chainDerivedState walks activations in order, setting each InputPreview to
// the previous activation's output (seeded with the run's own input for the
// first one) and StateAfter to the flow state folded up to that point. The
// input is an approximation for fan-out branches, where several successors
// share the same upstream output.
func chainDerivedState(activations []runActivation, runInput string) {
	state := map[string]any{}
	prevOutput := runInput
	for i := range activations {
		activations[i].InputPreview = prevOutput
		for k, v := range activations[i].StateDelta {
			state[k] = v
		}
		if len(state) > 0 {
			snapshot := make(map[string]any, len(state))
			for k, v := range state {
				snapshot[k] = v
			}
			activations[i].StateAfter = snapshot
		}
		if activations[i].OutputPreview != "" {
			prevOutput = activations[i].OutputPreview
		}
	}
}

// payloadValue reads a key from an event payload tolerating both casings:
// session.Event marshals with Go field names (no json tags), so keys arrive
// capitalized, but older payloads may carry lowercase.
func payloadValue(payload map[string]any, capitalized, lowercase string) (any, bool) {
	if v, ok := payload[capitalized]; ok && v != nil {
		return v, true
	}
	if v, ok := payload[lowercase]; ok && v != nil {
		return v, true
	}
	return nil, false
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
	stateDelta := map[string]any{}
	for _, ev := range events {
		if len(ev.Payload) > 0 {
			var payloadMap map[string]any
			if err := json.Unmarshal(ev.Payload, &payloadMap); err == nil {
				if val, ok := payloadValue(payloadMap, "Output", "output"); ok {
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
				} else if text := contentText(payloadMap); text != "" {
					// Agent events carry their answer as model content text, not
					// as a workflow Output value.
					lastOutput = text
				}
				if val, ok := payloadValue(payloadMap, "ErrorMessage", "errorMessage"); ok {
					if s, ok := val.(string); ok && s != "" {
						lastError = s
					}
				}
				collectFlowStateDelta(payloadMap, stateDelta)
			}
		}
	}

	if len([]rune(lastOutput)) > 200 {
		lastOutput = string([]rune(lastOutput)[:200]) + "..."
	}

	act := runActivation{
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
	if len(stateDelta) > 0 {
		act.StateDelta = stateDelta
	}
	return act
}

// contentText extracts and concatenates the text parts of an event's model
// content. The inner genai types marshal with lowercase json tags, unlike the
// event envelope.
func contentText(payload map[string]any) string {
	content, ok := payloadValue(payload, "Content", "content")
	if !ok {
		return ""
	}
	cm, ok := content.(map[string]any)
	if !ok {
		return ""
	}
	parts, ok := cm["parts"].([]any)
	if !ok {
		return ""
	}
	var out strings.Builder
	for _, p := range parts {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		if s, ok := pm["text"].(string); ok {
			out.WriteString(s)
		}
	}
	return out.String()
}

// collectFlowStateDelta copies the operator-visible flow-state writes of an
// event payload into dst: keys under the "flow:" namespace with the prefix
// stripped, skipping internal "__" bookkeeping such as router iteration
// counters.
func collectFlowStateDelta(payload map[string]any, dst map[string]any) {
	actions, ok := payload["Actions"].(map[string]any)
	if !ok {
		if actions, ok = payload["actions"].(map[string]any); !ok {
			return
		}
	}
	delta, ok := actions["StateDelta"].(map[string]any)
	if !ok {
		if delta, ok = actions["stateDelta"].(map[string]any); !ok {
			return
		}
	}
	for k, v := range delta {
		key, ok := strings.CutPrefix(k, "flow:")
		if !ok || strings.HasPrefix(key, "__") {
			continue
		}
		dst[key] = v
	}
}
