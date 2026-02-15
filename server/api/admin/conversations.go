package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"google.golang.org/adk/session"

	"github.com/achetronic/magec/server/store"
)

// listConversations returns a paginated list of conversation audit logs.
// @Summary      List conversations
// @Description  Returns a paginated list of conversation audit logs, newest first. Filters by agent, source, or client.
// @Tags         conversations
// @Produce      json
// @Param        agentId   query     string  false  "Filter by agent or flow ID"
// @Param        source    query     string  false  "Filter by source (voice-ui, telegram, executor, direct, cron, webhook)"
// @Param        clientId  query     string  false  "Filter by client ID"
// @Param        perspective query  string  false  "Filter by perspective (admin, user)"
// @Param        limit     query     int     false  "Max items to return (default 30, 0 for all)"
// @Param        offset    query     int     false  "Items to skip (default 0)"
// @Success      200  {object}  store.PaginatedResult[store.Conversation]
// @Router       /conversations [get]
func (h *Handler) listConversations(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeJSON(w, http.StatusOK, store.PaginatedResult[store.Conversation]{
			Items: []store.Conversation{},
			Total: 0,
		})
		return
	}

	agentID := r.URL.Query().Get("agentId")
	source := r.URL.Query().Get("source")
	clientID := r.URL.Query().Get("clientId")
	perspective := r.URL.Query().Get("perspective")
	limit := queryInt(r, "limit", 30)
	offset := queryInt(r, "offset", 0)

	result := h.conversations.List(agentID, source, clientID, perspective, limit, offset)
	writeJSON(w, http.StatusOK, result)
}

// getConversation returns a single conversation with paginated messages.
// @Summary      Get conversation
// @Description  Returns a conversation by ID with paginated messages (latest first). Includes totalMessages for client-side pagination.
// @Tags         conversations
// @Produce      json
// @Param        id         path   string  true   "Conversation ID"
// @Param        msgLimit   query  int     false  "Max messages to return (default 50, 0 for all)"
// @Param        msgOffset  query  int     false  "Messages to skip from the end (default 0)"
// @Success      200  {object}  map[string]interface{}  "conversation + totalMessages"
// @Failure      404  {object}  ErrorResponse
// @Router       /conversations/{id} [get]
func (h *Handler) getConversation(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeError(w, http.StatusNotFound, "conversation store not initialized")
		return
	}

	id := mux.Vars(r)["id"]
	msgLimit := queryInt(r, "msgLimit", 50)
	msgOffset := queryInt(r, "msgOffset", 0)

	convo, totalMsgs, ok := h.conversations.Get(id, msgLimit, msgOffset)
	if !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"conversation":  convo,
		"totalMessages": totalMsgs,
	})
}

// deleteConversation removes a conversation audit log.
// @Summary      Delete conversation
// @Description  Deletes a conversation audit log by ID. Does not affect the ADK session.
// @Tags         conversations
// @Param        id  path  string  true  "Conversation ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Router       /conversations/{id} [delete]
func (h *Handler) deleteConversation(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeError(w, http.StatusNotFound, "conversation store not initialized")
		return
	}

	id := mux.Vars(r)["id"]
	if err := h.conversations.Delete(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// clearConversations removes all conversation audit logs.
// @Summary      Clear all conversations
// @Description  Deletes all conversation audit logs. Does not affect ADK sessions.
// @Tags         conversations
// @Success      204
// @Failure      500  {object}  ErrorResponse
// @Router       /conversations/clear [delete]
func (h *Handler) clearConversations(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeError(w, http.StatusNotFound, "conversation store not initialized")
		return
	}

	if err := h.conversations.Clear(); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// conversationStats returns aggregate statistics about conversations.
// @Summary      Conversation statistics
// @Description  Returns total count and breakdowns by source and agent.
// @Tags         conversations
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "total, bySources, byAgents"
// @Router       /conversations/stats [get]
func (h *Handler) conversationStats(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"total": 0})
		return
	}

	all := h.conversations.List("", "", "", "", 0, 0)

	sourceCount := map[string]int{}
	agentCount := map[string]int{}
	for _, c := range all.Items {
		sourceCount[c.Source]++
		if c.AgentName != "" {
			agentCount[c.AgentName]++
		} else {
			agentCount[c.AgentID]++
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total":     all.Total,
		"bySources": sourceCount,
		"byAgents":  agentCount,
	})
}

// updateConversationSummary sets or updates the summary for a conversation.
// @Summary      Update conversation summary
// @Description  Sets the summary text for a conversation. Used for context window summarization.
// @Tags         conversations
// @Accept       json
// @Produce      json
// @Param        id    path  string  true  "Conversation ID"
// @Param        body  body  object  true  "Summary text"
// @Success      200  {object}  store.Conversation
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /conversations/{id}/summary [put]
func (h *Handler) updateConversationSummary(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeError(w, http.StatusNotFound, "conversation store not initialized")
		return
	}

	id := mux.Vars(r)["id"]

	var body struct {
		Summary string `json:"summary"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.conversations.SetSummary(id, body.Summary); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	convo, _, _ := h.conversations.Get(id, 0, 0)
	writeJSON(w, http.StatusOK, convo)
}

// resetConversationSession deletes the ADK session associated with a conversation.
// @Summary      Reset ADK session
// @Description  Deletes the ADK session (in Redis or in-memory) for the agent/user/session referenced by this conversation. The user will start a fresh session on their next message. The conversation audit log is preserved.
// @Tags         conversations
// @Produce      json
// @Param        id  path  string  true  "Conversation ID"
// @Success      200  {object}  map[string]interface{}  "message, agentId, sessionId"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Failure      503  {object}  ErrorResponse
// @Router       /conversations/{id}/reset-session [post]
func (h *Handler) resetConversationSession(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeError(w, http.StatusNotFound, "conversation store not initialized")
		return
	}
	if h.sessionService == nil {
		writeError(w, http.StatusServiceUnavailable, "session service not available")
		return
	}

	id := mux.Vars(r)["id"]
	convo, _, ok := h.conversations.Get(id, 0, 0)
	if !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}

	if convo.AgentID == "" || convo.SessionID == "" {
		writeError(w, http.StatusBadRequest, "conversation has no agent or session ID")
		return
	}

	userID := convo.UserID
	if userID == "" {
		userID = "user"
	}

	if err := h.sessionService.Delete(r.Context(), &session.DeleteRequest{
		AppName:   convo.AgentID,
		UserID:    userID,
		SessionID: convo.SessionID,
	}); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete session: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"message":   "Session reset successfully",
		"agentId":   convo.AgentID,
		"sessionId": convo.SessionID,
	})
}

func (h *Handler) findPerspectivePair(w http.ResponseWriter, r *http.Request) {
	if h.conversations == nil {
		writeError(w, http.StatusNotFound, "conversation store not initialized")
		return
	}

	id := mux.Vars(r)["id"]
	convo, _, ok := h.conversations.Get(id, 0, 0)
	if !ok {
		writeError(w, http.StatusNotFound, "conversation not found")
		return
	}

	otherPerspective := "admin"
	if convo.Perspective == "admin" {
		otherPerspective = "user"
	}

	pair, found := h.conversations.FindBySession(convo.SessionID, convo.AgentID, otherPerspective)
	if !found {
		writeJSON(w, http.StatusOK, map[string]interface{}{"pairId": nil})
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"pairId": pair.ID})
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 0 {
		return defaultVal
	}
	return v
}
