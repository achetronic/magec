// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"testing"
	"time"
)

// buildStoreWithConvos returns an in-memory ConversationStore (filePath="")
// pre-populated with the given conversations.
func buildStoreWithConvos(convos []Conversation) *ConversationStore {
	return &ConversationStore{
		conversations: convos,
		filePath:      "",
	}
}

func TestConversationStoreDelete_SingleConversation(t *testing.T) {
	base := time.Now()
	cs := buildStoreWithConvos([]Conversation{
		{ID: "a", SessionID: "s1", AgentID: "agent1", Perspective: "admin", StartedAt: base},
	})

	if err := cs.Delete("a"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := cs.Count(); got != 0 {
		t.Fatalf("expected 0 conversations after delete, got %d", got)
	}
}

func TestConversationStoreDelete_RemovesPair(t *testing.T) {
	base := time.Now()
	cs := buildStoreWithConvos([]Conversation{
		{ID: "admin-1", SessionID: "s1", AgentID: "agent1", Perspective: "admin", StartedAt: base},
		{ID: "user-1", SessionID: "s1", AgentID: "agent1", Perspective: "user", StartedAt: base.Add(10 * time.Millisecond)},
	})

	if err := cs.Delete("admin-1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if got := cs.Count(); got != 0 {
		t.Fatalf("expected both pair conversations removed, got %d remaining", got)
	}
}

// TestConversationStoreDelete_PreservesHistoricalSplits is the regression test
// for the bug where deleting one conversation with sessionID S also wiped out
// every previous conversation with the same sessionID produced by /reset
// splits. Delete must only remove the target conversation and its immediate
// admin/user pair (created within a few seconds).
func TestConversationStoreDelete_PreservesHistoricalSplits(t *testing.T) {
	base := time.Now()
	cs := buildStoreWithConvos([]Conversation{
		// Old closed split (admin + user), 2 hours ago.
		{ID: "old-admin", SessionID: "s1", AgentID: "agent1", Perspective: "admin", StartedAt: base.Add(-2 * time.Hour), Closed: true},
		{ID: "old-user", SessionID: "s1", AgentID: "agent1", Perspective: "user", StartedAt: base.Add(-2*time.Hour + 10*time.Millisecond), Closed: true},
		// Current active pair.
		{ID: "cur-admin", SessionID: "s1", AgentID: "agent1", Perspective: "admin", StartedAt: base},
		{ID: "cur-user", SessionID: "s1", AgentID: "agent1", Perspective: "user", StartedAt: base.Add(5 * time.Millisecond)},
	})

	if err := cs.Delete("cur-admin"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	remaining := make(map[string]bool)
	for _, c := range cs.conversations {
		remaining[c.ID] = true
	}

	if len(remaining) != 2 {
		t.Fatalf("expected 2 remaining conversations, got %d: %+v", len(remaining), remaining)
	}
	if !remaining["old-admin"] || !remaining["old-user"] {
		t.Fatalf("historical closed conversations were wiped: %+v", remaining)
	}
	if remaining["cur-admin"] || remaining["cur-user"] {
		t.Fatalf("active pair not fully deleted: %+v", remaining)
	}
}

func TestConversationStoreDelete_HistoricalLeavesOthersAlone(t *testing.T) {
	base := time.Now()
	cs := buildStoreWithConvos([]Conversation{
		{ID: "old-admin", SessionID: "s1", AgentID: "agent1", Perspective: "admin", StartedAt: base.Add(-2 * time.Hour), Closed: true},
		{ID: "old-user", SessionID: "s1", AgentID: "agent1", Perspective: "user", StartedAt: base.Add(-2*time.Hour + 10*time.Millisecond), Closed: true},
		{ID: "cur-admin", SessionID: "s1", AgentID: "agent1", Perspective: "admin", StartedAt: base},
		{ID: "cur-user", SessionID: "s1", AgentID: "agent1", Perspective: "user", StartedAt: base.Add(5 * time.Millisecond)},
	})

	// Deleting the old admin must only take down its own pair (old-user),
	// leaving the current active pair untouched.
	if err := cs.Delete("old-admin"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	remaining := make(map[string]bool)
	for _, c := range cs.conversations {
		remaining[c.ID] = true
	}

	if len(remaining) != 2 {
		t.Fatalf("expected 2 remaining, got %d: %+v", len(remaining), remaining)
	}
	if !remaining["cur-admin"] || !remaining["cur-user"] {
		t.Fatalf("current pair was wiped when deleting a historical entry: %+v", remaining)
	}
}

func TestConversationStoreDelete_NotFound(t *testing.T) {
	cs := buildStoreWithConvos(nil)
	if err := cs.Delete("does-not-exist"); err == nil {
		t.Fatalf("expected error when deleting a non-existent id")
	}
}
