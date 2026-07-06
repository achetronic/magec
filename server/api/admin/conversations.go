// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Conversations are a pure projection over recorded runs (decision #31 phase
// 2): a conversation is every run sharing a session, each run is one turn,
// and the user/admin perspective is an on-read filter. Nothing conversation-
// shaped is persisted.

package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/adk/v2/session"

	"github.com/achetronic/magec/server/agent/runrecorder"
	"github.com/achetronic/magec/server/runs"
)

// conversationSummary is one list row: an aggregated session.
type conversationSummary struct {
	ID         string    `json:"id"`
	SessionID  string    `json:"sessionId"`
	AgentID    string    `json:"agentId"`
	AgentName  string    `json:"agentName,omitempty"`
	FlowID     string    `json:"flowId,omitempty"`
	FlowName   string    `json:"flowName,omitempty"`
	ClientID   string    `json:"clientId,omitempty"`
	ClientName string    `json:"clientName,omitempty"`
	Source     string    `json:"source,omitempty"`
	UserID     string    `json:"userId,omitempty"`
	StartedAt  time.Time `json:"startedAt"`
	LastAt     time.Time `json:"lastAt"`
	Turns      int       `json:"turns"`
	Preview    string    `json:"preview,omitempty"`
}

// conversationMessage is one projected message of a conversation detail.
type conversationMessage struct {
	Role      string         `json:"role"`
	Agent     string         `json:"agent,omitempty"`
	Content   string         `json:"content,omitempty"`
	Timestamp time.Time      `json:"timestamp"`
	ToolCalls []toolCallView `json:"toolCalls,omitempty"`
	RunID     string         `json:"runId,omitempty"`
	Error     string         `json:"error,omitempty"`
}

// toolCallView captures a tool invocation extracted from event content parts.
type toolCallView struct {
	Name   string `json:"name"`
	Args   any    `json:"args,omitempty"`
	Result any    `json:"result,omitempty"`
}

// conversationDetail is the full projection of one session.
type conversationDetail struct {
	conversationSummary
	View     string                `json:"view"`
	Messages []conversationMessage `json:"messages"`
}

// conversationID encodes the composite key of a conversation for URLs. The
// app name is always a UUID (no colon), so splitting on the first colon is
// unambiguous whatever the session ID contains.
func conversationID(appName, sessionID string) string {
	return appName + ":" + sessionID
}

func splitConversationID(id string) (appName, sessionID string, ok bool) {
	parts := strings.SplitN(id, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", false
	}
	return parts[0], parts[1], true
}

// resolveAppNames fills agent/flow identity fields from the store.
func (h *Handler) decorateConversation(sum *conversationSummary) {
	if a, ok := h.store.GetAgent(sum.AgentID); ok {
		sum.AgentName = a.Name
	} else if f, ok := h.store.GetFlow(sum.AgentID); ok {
		sum.FlowID = f.ID
		sum.FlowName = f.Name
		sum.AgentName = f.Name
	}
	if sum.ClientID != "" {
		if cl, ok := h.store.GetClient(sum.ClientID); ok {
			sum.ClientName = cl.Name
		}
	}
}

// listConversations returns aggregated sessions, newest activity first.
// @Summary      List conversations
// @Description  Returns conversations projected from recorded runs (one conversation per session), newest activity first
// @Tags         conversations
// @Produce      json
// @Param        agentId   query     string  false  "Filter by agent or flow ID"
// @Param        source    query     string  false  "Filter by source (telegram, voice-ui, webhook...)"
// @Param        clientId  query     string  false  "Filter by client ID"
// @Param        limit     query     int     false  "Max items to return (default 30)"
// @Param        offset    query     int     false  "Items to skip (default 0)"
// @Success      200  {object}  map[string]interface{}  "items + total"
// @Security     AdminAuth
// @Router       /conversations [get]
func (h *Handler) listConversations(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeJSON(w, http.StatusOK, map[string]any{"items": []conversationSummary{}, "total": 0})
		return
	}

	filter := runs.ConversationFilter{
		AppName:  r.URL.Query().Get("agentId"),
		Source:   r.URL.Query().Get("source"),
		ClientID: r.URL.Query().Get("clientId"),
		Limit:    queryInt(r, "limit", 30),
		Offset:   queryInt(r, "offset", 0),
	}
	rows, total, err := h.runs.ListConversations(filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	items := make([]conversationSummary, 0, len(rows))
	for _, row := range rows {
		sum := conversationSummary{
			ID:        conversationID(row.AppName, row.SessionID),
			SessionID: row.SessionID,
			AgentID:   row.AppName,
			ClientID:  row.ClientID,
			Source:    row.Source,
			UserID:    row.UserID,
			StartedAt: row.StartedAt,
			LastAt:    row.LastAt,
			Turns:     row.Turns,
			Preview:   previewText(row.FirstInput),
		}
		h.decorateConversation(&sum)
		items = append(items, sum)
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total})
}

// getConversation returns one conversation with its projected messages.
// @Summary      Get conversation
// @Description  Projects a conversation from the runs of its session. view=user filters flow output to response agents; view=admin (default) shows every agent's messages.
// @Tags         conversations
// @Produce      json
// @Param        id         path   string  true   "Conversation ID (appId:sessionId)"
// @Param        view       query  string  false  "Perspective: admin (default) or user"
// @Param        msgLimit   query  int     false  "Max messages to return (default 50, 0 for all)"
// @Param        msgOffset  query  int     false  "Messages to skip from the end (default 0)"
// @Success      200  {object}  map[string]interface{}  "conversation + totalMessages"
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /conversations/{id} [get]
func (h *Handler) getConversation(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeError(w, http.StatusNotFound, "runs store not initialized")
		return
	}
	appName, sessionID, ok := splitConversationID(mux.Vars(r)["id"])
	if !ok {
		writeError(w, http.StatusBadRequest, "malformed conversation id")
		return
	}

	records, err := h.runs.GetSessionRuns(sessionID, appName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(records) == 0 {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}

	view := r.URL.Query().Get("view")
	if view != "user" {
		view = "admin"
	}
	detail := h.projectConversation(appName, sessionID, records, view)

	// Paginate messages from the end, mirroring the old behaviour the UI
	// relies on: offset skips backwards from the latest message.
	totalMsgs := len(detail.Messages)
	msgLimit := queryInt(r, "msgLimit", 50)
	msgOffset := queryInt(r, "msgOffset", 0)
	if msgLimit > 0 || msgOffset > 0 {
		end := totalMsgs - msgOffset
		if end < 0 {
			end = 0
		}
		start := 0
		if msgLimit > 0 && end-msgLimit > 0 {
			start = end - msgLimit
		}
		detail.Messages = detail.Messages[start:end]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"conversation":  detail,
		"totalMessages": totalMsgs,
	})
}

// projectConversation builds the conversation detail from its runs: each run
// is one turn (its input as the user message, its events as the assistant
// messages). In the user view of a flow with response agents, only their
// events surface, mirroring what external clients received.
func (h *Handler) projectConversation(appName, sessionID string, records []runrecorder.RunRecord, view string) conversationDetail {
	first := records[0]
	last := records[len(records)-1]

	sum := conversationSummary{
		ID:        conversationID(appName, sessionID),
		SessionID: sessionID,
		AgentID:   appName,
		ClientID:  first.ClientID,
		Source:    first.Source,
		UserID:    first.UserID,
		StartedAt: first.StartedAt,
		LastAt:    last.StartedAt,
		Turns:     len(records),
		Preview:   previewText(first.Input),
	}
	h.decorateConversation(&sum)

	// The response-agent filter only exists for flows and only in user view.
	responseOnly := map[string]bool{}
	if view == "user" {
		if f, ok := h.store.GetFlow(appName); ok {
			for _, name := range f.ResponseAgentNames() {
				responseOnly[name] = true
			}
		}
	}

	detail := conversationDetail{conversationSummary: sum, View: view, Messages: []conversationMessage{}}
	for _, rec := range records {
		if rec.Input != "" {
			detail.Messages = append(detail.Messages, conversationMessage{
				Role:      "user",
				Content:   rec.Input,
				Timestamp: rec.StartedAt,
				RunID:     rec.RunID,
			})
		}
		for _, ev := range rec.Events {
			if strings.HasPrefix(ev.Author, "__") {
				continue // internal nodes (meta prefilter) never surface
			}
			if len(responseOnly) > 0 && !responseOnly[ev.Author] {
				continue
			}
			msg, ok := projectEventMessage(ev)
			if !ok {
				continue
			}
			msg.RunID = rec.RunID
			detail.Messages = append(detail.Messages, msg)
		}
		if rec.Status == runrecorder.StatusFailed && rec.Error != "" {
			detail.Messages = append(detail.Messages, conversationMessage{
				Role:      "assistant",
				Timestamp: rec.EndedAt,
				RunID:     rec.RunID,
				Error:     rec.Error,
			})
		}
	}
	return detail
}

// projectEventMessage turns one recorded event into a conversation message:
// its model text plus any tool calls. Events with neither (transform nodes,
// bookkeeping) yield nothing; their detail lives in the Runs view.
func projectEventMessage(ev runrecorder.EventRecord) (conversationMessage, bool) {
	if len(ev.Payload) == 0 {
		return conversationMessage{}, false
	}
	var payload map[string]any
	if err := json.Unmarshal(ev.Payload, &payload); err != nil {
		return conversationMessage{}, false
	}

	text := contentText(payload)
	toolCalls := contentToolCalls(payload)
	if text == "" && len(toolCalls) == 0 {
		return conversationMessage{}, false
	}
	return conversationMessage{
		Role:      "assistant",
		Agent:     ev.Author,
		Content:   text,
		Timestamp: ev.Timestamp,
		ToolCalls: toolCalls,
	}, true
}

// contentToolCalls extracts functionCall/functionResponse parts from an
// event's model content.
func contentToolCalls(payload map[string]any) []toolCallView {
	content, ok := payloadValue(payload, "Content", "content")
	if !ok {
		return nil
	}
	cm, ok := content.(map[string]any)
	if !ok {
		return nil
	}
	parts, ok := cm["parts"].([]any)
	if !ok {
		return nil
	}
	var calls []toolCallView
	for _, p := range parts {
		pm, ok := p.(map[string]any)
		if !ok {
			continue
		}
		if fc, ok := pm["functionCall"].(map[string]any); ok {
			calls = append(calls, toolCallView{Name: fmt.Sprintf("%v", fc["name"]), Args: fc["args"]})
		}
		if fr, ok := pm["functionResponse"].(map[string]any); ok {
			calls = append(calls, toolCallView{Name: fmt.Sprintf("%v", fr["name"]), Result: fr["response"]})
		}
	}
	return calls
}

// metadataBlockRE matches the inline metadata comment blocks client bots
// prepend to user messages. Previews must strip them BEFORE truncating:
// a block cut at 120 chars loses its closing marker and no downstream
// regex can recognise it anymore.
var metadataBlockRE = regexp.MustCompile(`<!--MAGEC_(?:META|THREAD_HISTORY):.*?:MAGEC_(?:META|THREAD_HISTORY)-->\n?`)

// previewText trims an input down to a one-line list preview, dropping the
// client metadata comment blocks first.
func previewText(s string) string {
	s = metadataBlockRE.ReplaceAllString(s, "")
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	if len([]rune(s)) > 120 {
		s = string([]rune(s)[:120]) + "..."
	}
	return s
}

// deleteConversation removes a conversation: its runs and their events.
// @Summary      Delete conversation
// @Description  Deletes every recorded run of the conversation's session. The ADK session itself is not affected.
// @Tags         conversations
// @Param        id  path  string  true  "Conversation ID (appId:sessionId)"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /conversations/{id} [delete]
func (h *Handler) deleteConversation(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeError(w, http.StatusNotFound, "runs store not initialized")
		return
	}
	appName, sessionID, ok := splitConversationID(mux.Vars(r)["id"])
	if !ok {
		writeError(w, http.StatusBadRequest, "malformed conversation id")
		return
	}
	deleted, err := h.runs.DeleteSession(sessionID, appName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if !deleted {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// clearConversations removes every conversation (all runs with a session).
// @Summary      Clear all conversations
// @Description  Deletes every recorded run that belongs to a session. Session-less runs are preserved.
// @Tags         conversations
// @Success      204
// @Failure      500  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /conversations/clear [delete]
func (h *Handler) clearConversations(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if _, err := h.runs.DeleteAllSessions(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// conversationStats returns aggregated conversation counts.
// @Summary      Conversation stats
// @Description  Returns total conversations plus per-agent and per-source counts, projected from recorded runs
// @Tags         conversations
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Security     AdminAuth
// @Router       /conversations/stats [get]
func (h *Handler) conversationStats(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeJSON(w, http.StatusOK, map[string]any{"total": 0})
		return
	}
	stats, err := h.runs.Stats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	byAgents := map[string]int{}
	for appName, count := range stats.ByApp {
		label := appName
		if a, ok := h.store.GetAgent(appName); ok {
			label = a.Name
		} else if f, ok := h.store.GetFlow(appName); ok {
			label = f.Name
		}
		byAgents[label] += count
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"total":     stats.Total,
		"bySources": stats.BySource,
		"byAgents":  byAgents,
	})
}

// resetConversationSession deletes the ADK session behind a conversation.
// @Summary      Reset ADK session
// @Description  Deletes the ADK session (in Redis or in-memory) for the agent/user/session of this conversation. The user starts a fresh session on their next message; the recorded runs are preserved.
// @Tags         conversations
// @Produce      json
// @Param        id  path  string  true  "Conversation ID (appId:sessionId)"
// @Success      200  {object}  map[string]interface{}  "message, agentId, sessionId"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      503  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /conversations/{id}/reset-session [post]
func (h *Handler) resetConversationSession(w http.ResponseWriter, r *http.Request) {
	if h.runs == nil {
		writeError(w, http.StatusNotFound, "runs store not initialized")
		return
	}
	if h.sessionService == nil {
		writeError(w, http.StatusServiceUnavailable, "session service not available")
		return
	}
	appName, sessionID, ok := splitConversationID(mux.Vars(r)["id"])
	if !ok {
		writeError(w, http.StatusBadRequest, "malformed conversation id")
		return
	}

	records, err := h.runs.GetSessionRuns(sessionID, appName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if len(records) == 0 {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	userID := records[0].UserID
	if userID == "" {
		userID = "user"
	}

	if err := h.sessionService.Delete(r.Context(), &session.DeleteRequest{
		AppName:   appName,
		UserID:    userID,
		SessionID: sessionID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete session: "+err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"message":   "Session reset successfully",
		"agentId":   appName,
		"sessionId": sessionID,
	})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return defaultVal
	}
	return n
}
