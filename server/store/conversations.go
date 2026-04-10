package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

// ConversationMessage represents a single message in a conversation.
type ConversationMessage struct {
	Role      string                 `json:"role"`
	Agent     string                 `json:"agent,omitempty"`
	Content   string                 `json:"content"`
	Timestamp time.Time              `json:"timestamp"`
	ToolCalls []ToolCallInfo         `json:"toolCalls,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ToolCallInfo captures info about a tool invocation.
type ToolCallInfo struct {
	Name   string      `json:"name"`
	Args   interface{} `json:"args,omitempty"`
	Result interface{} `json:"result,omitempty"`
}

// Conversation represents a full conversation session.
type Conversation struct {
	ID          string                `json:"id"`
	SessionID   string                `json:"sessionId"`
	AgentID     string                `json:"agentId"`
	AgentName   string                `json:"agentName,omitempty"`
	FlowID      string                `json:"flowId,omitempty"`
	FlowName    string                `json:"flowName,omitempty"`
	ClientID    string                `json:"clientId,omitempty"`
	ClientName  string                `json:"clientName,omitempty"`
	Source      string                `json:"source"`
	Perspective string                `json:"perspective,omitempty"`
	UserID      string                `json:"userId,omitempty"`
	Messages    []ConversationMessage `json:"messages"`
	StartedAt   time.Time             `json:"startedAt"`
	EndedAt     *time.Time            `json:"endedAt,omitempty"`
	Closed      bool                  `json:"closed,omitempty"`
	Summary     string                `json:"summary,omitempty"`
	Preview     string                `json:"preview,omitempty"`
	ParentID    string                `json:"parentId,omitempty"`
	RawEvents   []interface{}         `json:"rawEvents,omitempty"`
}

// ConversationStore manages conversation logs with JSON persistence.
type ConversationStore struct {
	mu            sync.RWMutex
	conversations []Conversation
	filePath      string
}

// NewConversationStore creates a conversation store backed by a JSON file.
func NewConversationStore(filePath string) (*ConversationStore, error) {
	cs := &ConversationStore{
		conversations: []Conversation{},
		filePath:      filePath,
	}

	if filePath != "" {
		if _, err := os.Stat(filePath); err == nil {
			if err := cs.loadFromDisk(); err != nil {
				return nil, fmt.Errorf("failed to load conversations from %s: %w", filePath, err)
			}
		}
	}

	return cs, nil
}

// Append adds a new conversation to the store.
func (cs *ConversationStore) Append(c Conversation) (Conversation, error) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if c.ID == "" {
		c.ID = generateID()
	}
	cs.conversations = append(cs.conversations, c)

	if err := cs.persist(); err != nil {
		return c, err
	}
	return c, nil
}

// AppendMessage adds a message to an existing conversation.
func (cs *ConversationStore) AppendMessage(conversationID string, msg ConversationMessage) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i := range cs.conversations {
		if cs.conversations[i].ID == conversationID {
			cs.conversations[i].Messages = append(cs.conversations[i].Messages, msg)
			return cs.persist()
		}
	}
	return fmt.Errorf("conversation %q not found", conversationID)
}

// EndConversation marks a conversation as ended.
func (cs *ConversationStore) EndConversation(conversationID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i := range cs.conversations {
		if cs.conversations[i].ID == conversationID {
			now := time.Now()
			cs.conversations[i].EndedAt = &now
			return cs.persist()
		}
	}
	return fmt.Errorf("conversation %q not found", conversationID)
}

// SetSummary sets the summary for a conversation (used by future summarization feature).
func (cs *ConversationStore) SetSummary(conversationID, summary string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i := range cs.conversations {
		if cs.conversations[i].ID == conversationID {
			cs.conversations[i].Summary = summary
			return cs.persist()
		}
	}
	return fmt.Errorf("conversation %q not found", conversationID)
}

// PaginatedResult wraps a paginated response with total count.
type PaginatedResult[T any] struct {
	Items []T `json:"items"`
	Total int `json:"total"`
}

// List returns conversations sorted by start time (newest first) with pagination.
// Optional filters: agentID, source, clientID. Use limit=0 for no limit.
func (cs *ConversationStore) List(agentID, source, clientID, perspective string, limit, offset int) PaginatedResult[Conversation] {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var filtered []Conversation
	for _, c := range cs.conversations {
		if agentID != "" && c.AgentID != agentID && c.FlowID != agentID {
			continue
		}
		if source != "" && c.Source != source {
			continue
		}
		if clientID != "" && c.ClientID != clientID {
			continue
		}
		if perspective != "" && c.Perspective != perspective {
			continue
		}
		cCopy := c
		cCopy.Preview = conversationPreview(c.Messages)
		cCopy.Messages = nil
		cCopy.RawEvents = nil
		filtered = append(filtered, cCopy)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].StartedAt.After(filtered[j].StartedAt)
	})

	total := len(filtered)

	if offset > total {
		offset = total
	}
	filtered = filtered[offset:]

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}

	if filtered == nil {
		filtered = []Conversation{}
	}

	return PaginatedResult[Conversation]{Items: filtered, Total: total}
}

// FindExactPair finds the corresponding conversation for the other perspective
// that was created around the same time. This is used when viewing historical
// (closed) conversations to ensure we don't accidentally link to a newer session.
func (cs *ConversationStore) FindExactPair(id, sessionID, agentID, perspective string) (Conversation, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var sourceConvo Conversation
	foundSource := false
	for _, c := range cs.conversations {
		if c.ID == id {
			sourceConvo = c
			foundSource = true
			break
		}
	}

	if !foundSource {
		return Conversation{}, false
	}

	var bestMatch Conversation
	var minDiff time.Duration = 24 * time.Hour
	foundMatch := false

	otherPerspective := "admin"
	if perspective == "admin" {
		otherPerspective = "user"
	}

	for _, c := range cs.conversations {
		if c.SessionID == sessionID && c.AgentID == agentID && c.Perspective == otherPerspective {
			diff := c.StartedAt.Sub(sourceConvo.StartedAt)
			if diff < 0 {
				diff = -diff
			}
			// Pairs are usually created within milliseconds of each other
			if diff < minDiff && diff < 5*time.Second {
				minDiff = diff
				bestMatch = c
				foundMatch = true
			}
		}
	}

	return bestMatch, foundMatch
}

// FindBySession returns the most recent conversation matching the given
// sessionID, agentID, and perspective. This is used to append messages to an
// existing conversation instead of creating a new one for every /run call.
func (cs *ConversationStore) FindBySession(sessionID, agentID, perspective string) (Conversation, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for i := len(cs.conversations) - 1; i >= 0; i-- {
		c := cs.conversations[i]
		if c.SessionID == sessionID && c.AgentID == agentID && c.Perspective == perspective && !c.Closed {
			return c, true
		}
	}
	return Conversation{}, false
}

// CloseBySession marks a conversation as closed so new messages trigger a new conversation record.
func (cs *ConversationStore) CloseBySession(sessionID, agentID, perspective string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i := len(cs.conversations) - 1; i >= 0; i-- {
		c := cs.conversations[i]
		if c.SessionID == sessionID && c.AgentID == agentID && c.Perspective == perspective && !c.Closed {
			cs.conversations[i].Closed = true
			now := time.Now()
			cs.conversations[i].EndedAt = &now
			return cs.persist()
		}
	}
	return fmt.Errorf("active conversation for session %q not found", sessionID)
}

// Close marks a specific conversation as closed.
func (cs *ConversationStore) Close(id string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i, c := range cs.conversations {
		if c.ID == id {
			if cs.conversations[i].Closed {
				return nil
			}
			cs.conversations[i].Closed = true
			now := time.Now()
			cs.conversations[i].EndedAt = &now
			return cs.persist()
		}
	}
	return fmt.Errorf("conversation %q not found", id)
}

// AppendMessages adds multiple messages and raw events to an existing conversation.
func (cs *ConversationStore) AppendMessages(conversationID string, msgs []ConversationMessage, rawEvents []interface{}) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i := range cs.conversations {
		if cs.conversations[i].ID == conversationID {
			cs.conversations[i].Messages = append(cs.conversations[i].Messages, msgs...)
			cs.conversations[i].RawEvents = append(cs.conversations[i].RawEvents, rawEvents...)
			now := time.Now()
			cs.conversations[i].EndedAt = &now
			return cs.persist()
		}
	}
	return fmt.Errorf("conversation %q not found", conversationID)
}

// Get returns a single conversation by ID. If msgLimit > 0, only the latest
// msgLimit messages starting from msgOffset (counted from the end) are returned,
// along with totalMessages so the client can paginate.
func (cs *ConversationStore) Get(id string, msgLimit, msgOffset int) (Conversation, int, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	for _, c := range cs.conversations {
		if c.ID == id {
			totalMsgs := len(c.Messages)
			totalEvents := len(c.RawEvents)

			if msgLimit > 0 {
				end := totalMsgs - msgOffset
				if end < 0 {
					end = 0
				}
				start := end - msgLimit
				if start < 0 {
					start = 0
				}
				c.Messages = c.Messages[start:end]

				evEnd := totalEvents - msgOffset
				if evEnd < 0 {
					evEnd = 0
				}
				evStart := evEnd - msgLimit
				if evStart < 0 {
					evStart = 0
				}
				c.RawEvents = c.RawEvents[evStart:evEnd]
			}

			return c, totalMsgs, true
		}
	}
	return Conversation{}, 0, false
}

// Delete removes a conversation and its paired perspective (user↔admin).
func (cs *ConversationStore) Delete(id string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	var sessionID, agentID string
	found := false
	for _, c := range cs.conversations {
		if c.ID == id {
			sessionID = c.SessionID
			agentID = c.AgentID
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("conversation %q not found", id)
	}

	filtered := cs.conversations[:0]
	for _, c := range cs.conversations {
		if c.ID == id || (c.SessionID == sessionID && c.AgentID == agentID) {
			continue
		}
		filtered = append(filtered, c)
	}
	cs.conversations = filtered
	return cs.persist()
}

// Clear removes all conversations.
func (cs *ConversationStore) Clear() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.conversations = []Conversation{}
	return cs.persist()
}

// Count returns the total number of conversations.
func (cs *ConversationStore) Count() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.conversations)
}

// Reload re-reads conversations from disk, replacing all in-memory data.
func (cs *ConversationStore) Reload() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.conversations = []Conversation{}
	if cs.filePath != "" {
		if _, err := os.Stat(cs.filePath); err == nil {
			return cs.loadFromDisk()
		}
	}
	return nil
}

// persist writes conversations to disk as formatted JSON.
func (cs *ConversationStore) persist() error {
	if cs.filePath == "" {
		return nil
	}

	dir := filepath.Dir(cs.filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create conversations directory: %w", err)
	}

	data, err := json.MarshalIndent(cs.conversations, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal conversations: %w", err)
	}

	if err := os.WriteFile(cs.filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write conversations file: %w", err)
	}
	return nil
}

// loadFromDisk reads conversations from the JSON file.
func (cs *ConversationStore) loadFromDisk() error {
	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		return err
	}

	var convos []Conversation
	if err := json.Unmarshal(data, &convos); err != nil {
		return err
	}
	if convos == nil {
		convos = []Conversation{}
	}
	cs.conversations = convos
	return nil
}

var magecCommentRegex = regexp.MustCompile(`<!--MAGEC_[A-Z_]+:.*?:MAGEC_[A-Z_]+-->\n?`)

func stripMagecComments(s string) string {
	return strings.TrimSpace(magecCommentRegex.ReplaceAllString(s, ""))
}

func conversationPreview(msgs []ConversationMessage) string {
	for _, m := range msgs {
		if m.Role == "user" && m.Content != "" {
			clean := stripMagecComments(m.Content)
			if clean == "" {
				continue
			}
			if len(clean) > 120 {
				return clean[:120] + "…"
			}
			return clean
		}
	}
	return ""
}
