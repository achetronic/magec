package mcp

import (
	"context"
	"fmt"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/session"

	"github.com/achetronic/magec/server/store"
)

type listConversationsInput struct {
	AgentID     string `json:"agentId,omitempty" jsonschema:"filter by agent or flow id"`
	Source      string `json:"source,omitempty" jsonschema:"filter by source (voice-ui, telegram, executor, direct, cron, webhook)"`
	ClientID    string `json:"clientId,omitempty"`
	Perspective string `json:"perspective,omitempty" jsonschema:"filter by perspective (admin or user)"`
	Limit       int    `json:"limit,omitempty" jsonschema:"max items to return (default 30, 0 for all)"`
	Offset      int    `json:"offset,omitempty"`
}

func (h *Handler) listConversations(_ context.Context, _ *sdk.CallToolRequest, in listConversationsInput) (*sdk.CallToolResult, store.PaginatedResult[store.Conversation], error) {
	if h.conversations == nil {
		return nil, store.PaginatedResult[store.Conversation]{Items: []store.Conversation{}}, nil
	}
	limit := in.Limit
	if limit == 0 {
		limit = 30
	}
	return nil, h.conversations.List(in.AgentID, in.Source, in.ClientID, in.Perspective, limit, in.Offset), nil
}

type getConversationInput struct {
	ID        string `json:"id" jsonschema:"conversation id"`
	MsgLimit  int    `json:"msgLimit,omitempty" jsonschema:"max messages to return (default 50, 0 for all)"`
	MsgOffset int    `json:"msgOffset,omitempty"`
}

type getConversationOutput struct {
	Conversation  store.Conversation `json:"conversation"`
	TotalMessages int                `json:"totalMessages"`
}

func (h *Handler) getConversation(_ context.Context, _ *sdk.CallToolRequest, in getConversationInput) (*sdk.CallToolResult, getConversationOutput, error) {
	if h.conversations == nil {
		return nil, getConversationOutput{}, fmt.Errorf("get conversation: %w", errValidation("conversation store not initialized"))
	}
	limit := in.MsgLimit
	if limit == 0 {
		limit = 50
	}
	convo, total, ok := h.conversations.Get(in.ID, limit, in.MsgOffset)
	if !ok {
		return nil, getConversationOutput{}, fmt.Errorf("get conversation: %w", errValidation("not found: "+in.ID))
	}
	return nil, getConversationOutput{Conversation: convo, TotalMessages: total}, nil
}

func (h *Handler) deleteConversation(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, emptyOutput, error) {
	if h.conversations == nil {
		return nil, emptyOutput{}, fmt.Errorf("delete conversation: %w", errValidation("conversation store not initialized"))
	}
	if err := h.conversations.Delete(in.ID); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("delete conversation: %w", err)
	}
	return nil, okOutput, nil
}

func (h *Handler) clearConversations(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, emptyOutput, error) {
	if h.conversations == nil {
		return nil, emptyOutput{}, fmt.Errorf("clear conversations: %w", errValidation("conversation store not initialized"))
	}
	if err := h.conversations.Clear(); err != nil {
		return nil, emptyOutput{}, fmt.Errorf("clear conversations: %w", err)
	}
	return nil, okOutput, nil
}

type conversationStatsOutput struct {
	Total     int            `json:"total"`
	BySources map[string]int `json:"bySources"`
	ByAgents  map[string]int `json:"byAgents"`
}

func (h *Handler) conversationStats(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, conversationStatsOutput, error) {
	if h.conversations == nil {
		return nil, conversationStatsOutput{BySources: map[string]int{}, ByAgents: map[string]int{}}, nil
	}
	all := h.conversations.List("", "", "", "", 0, 0)
	source := map[string]int{}
	agent := map[string]int{}
	for _, c := range all.Items {
		source[c.Source]++
		if c.AgentName != "" {
			agent[c.AgentName]++
		} else {
			agent[c.AgentID]++
		}
	}
	return nil, conversationStatsOutput{Total: all.Total, BySources: source, ByAgents: agent}, nil
}

type updateConversationSummaryInput struct {
	ID      string `json:"id" jsonschema:"conversation id"`
	Summary string `json:"summary" jsonschema:"new summary text"`
}

func (h *Handler) updateConversationSummary(_ context.Context, _ *sdk.CallToolRequest, in updateConversationSummaryInput) (*sdk.CallToolResult, store.Conversation, error) {
	if h.conversations == nil {
		return nil, store.Conversation{}, fmt.Errorf("update conversation summary: %w", errValidation("conversation store not initialized"))
	}
	if err := h.conversations.SetSummary(in.ID, in.Summary); err != nil {
		return nil, store.Conversation{}, fmt.Errorf("update conversation summary: %w", err)
	}
	convo, _, _ := h.conversations.Get(in.ID, 0, 0)
	return nil, convo, nil
}

type findPairOutput struct {
	PairID string `json:"pairId,omitempty"`
}

func (h *Handler) findConversationPair(_ context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, findPairOutput, error) {
	if h.conversations == nil {
		return nil, findPairOutput{}, fmt.Errorf("find conversation pair: %w", errValidation("conversation store not initialized"))
	}
	convo, _, ok := h.conversations.Get(in.ID, 0, 0)
	if !ok {
		return nil, findPairOutput{}, fmt.Errorf("find conversation pair: %w", errValidation("not found: "+in.ID))
	}
	pair, found := h.conversations.FindExactPair(convo.ID, convo.SessionID, convo.AgentID, convo.Perspective)
	if !found {
		return nil, findPairOutput{}, nil
	}
	return nil, findPairOutput{PairID: pair.ID}, nil
}

type resetSessionOutput struct {
	Message   string `json:"message"`
	AgentID   string `json:"agentId"`
	SessionID string `json:"sessionId"`
}

func (h *Handler) resetConversationSession(ctx context.Context, _ *sdk.CallToolRequest, in idInput) (*sdk.CallToolResult, resetSessionOutput, error) {
	if h.conversations == nil {
		return nil, resetSessionOutput{}, fmt.Errorf("reset session: %w", errValidation("conversation store not initialized"))
	}
	svc, _ := h.sessionService().(session.Service)
	if svc == nil {
		return nil, resetSessionOutput{}, fmt.Errorf("reset session: %w", errValidation("session service not available"))
	}
	convo, _, ok := h.conversations.Get(in.ID, 0, 0)
	if !ok {
		return nil, resetSessionOutput{}, fmt.Errorf("reset session: %w", errValidation("conversation not found: "+in.ID))
	}
	if convo.AgentID == "" || convo.SessionID == "" {
		return nil, resetSessionOutput{}, fmt.Errorf("reset session: %w", errValidation("conversation has no agent or session id"))
	}
	userID := convo.UserID
	if userID == "" {
		userID = "user"
	}
	if err := svc.Delete(ctx, &session.DeleteRequest{
		AppName:   convo.AgentID,
		UserID:    userID,
		SessionID: convo.SessionID,
	}); err != nil {
		return nil, resetSessionOutput{}, fmt.Errorf("reset session: %w", err)
	}
	_ = h.conversations.CloseBySession(convo.SessionID, convo.AgentID, "admin")
	_ = h.conversations.CloseBySession(convo.SessionID, convo.AgentID, "user")
	return nil, resetSessionOutput{
		Message:   "session reset successfully",
		AgentID:   convo.AgentID,
		SessionID: convo.SessionID,
	}, nil
}

func (h *Handler) registerConversationTools() {
	destructive := true
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_conversations", Title: "List conversations",
		Description: "Paginated list of conversation audit logs (newest first). Filters by agent, source, client, perspective.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listConversations)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_get_conversation", Title: "Get conversation",
		Description: "Return one conversation by id with paginated messages.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.getConversation)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_delete_conversation", Title: "Delete conversation",
		Description: "Delete a conversation by id (also deletes the paired perspective).",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive, IdempotentHint: true},
	}, h.deleteConversation)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_clear_conversations", Title: "Clear all conversations",
		Description: "Delete every conversation audit log. Does not affect ADK sessions.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive},
	}, h.clearConversations)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_conversation_stats", Title: "Conversation stats",
		Description: "Return totals broken down by source and agent.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.conversationStats)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_update_conversation_summary", Title: "Update conversation summary",
		Description: "Set the summary text used by context window compaction.",
	}, h.updateConversationSummary)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_find_conversation_pair", Title: "Find conversation pair",
		Description: "Return the id of the conversation that records the opposite perspective (admin/user) of a given conversation.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.findConversationPair)
	h.toolCount++
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_reset_conversation_session", Title: "Reset ADK session for a conversation",
		Description: "Delete the ADK session associated with a conversation so the next message starts a fresh one. Requires the ADK session service to be wired in.",
		Annotations: &sdk.ToolAnnotations{DestructiveHint: &destructive},
	}, h.resetConversationSession)
	h.toolCount++
}
