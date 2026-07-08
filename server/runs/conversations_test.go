// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package runs

import (
	"testing"
	"time"

	"github.com/achetronic/magec/server/agent/runrecorder"
)

// sessionRun builds a run pinned to an explicit session, so several runs can
// share one conversation (sampleRun derives the session from the run ID).
func sessionRun(runID, appName, sessionID, source, input string, startedAt time.Time) runrecorder.RunRecord {
	rec := sampleRun(runID, appName, startedAt, 2)
	rec.SessionID = sessionID
	rec.Source = source
	rec.Input = input
	return rec
}

// seedConversations persists two conversations for app A (one with two
// turns, one with a single turn) and one for app B, plus a session-less run
// that must never surface as a conversation.
func seedConversations(t *testing.T, s *Store) time.Time {
	t.Helper()
	base := time.Now().Add(-1 * time.Hour).Truncate(time.Millisecond)

	for _, rec := range []runrecorder.RunRecord{
		sessionRun("run_a1", "appA", "sess_1", "telegram", "hello there", base),
		sessionRun("run_a2", "appA", "sess_1", "telegram", "second turn", base.Add(10*time.Minute)),
		sessionRun("run_a3", "appA", "sess_2", "voice-ui", "another chat", base.Add(5*time.Minute)),
		sessionRun("run_b1", "appB", "sess_3", "webhook", "bee input", base.Add(2*time.Minute)),
	} {
		if err := s.SaveRun(rec); err != nil {
			t.Fatalf("SaveRun(%s): %v", rec.RunID, err)
		}
	}
	orphan := sampleRun("run_orphan", "appA", base, 1)
	orphan.SessionID = ""
	if err := s.SaveRun(orphan); err != nil {
		t.Fatalf("SaveRun(orphan): %v", err)
	}
	return base
}

func TestListConversations_GroupsRunsBySession(t *testing.T) {
	s := openStore(t)
	base := seedConversations(t, s)

	got, total, err := s.ListConversations(ConversationFilter{})
	if err != nil {
		t.Fatalf("ListConversations: %v", err)
	}
	if total != 3 || len(got) != 3 {
		t.Fatalf("total=%d len=%d, want 3 conversations (orphan run must not count)", total, len(got))
	}
	// Ordered by most recent activity: sess_1 (turn at +10m) first.
	first := got[0]
	if first.SessionID != "sess_1" || first.Turns != 2 {
		t.Fatalf("first conversation = %+v, want sess_1 with 2 turns", first)
	}
	if first.FirstInput != "hello there" {
		t.Fatalf("firstInput = %q, want the input of the OLDEST run", first.FirstInput)
	}
	if !first.StartedAt.Equal(base) {
		t.Fatalf("startedAt = %v, want the oldest run's start %v", first.StartedAt, base)
	}
	if !first.LastAt.Equal(base.Add(10 * time.Minute)) {
		t.Fatalf("lastAt = %v, want the newest run's start", first.LastAt)
	}
}

func TestListConversations_Filters(t *testing.T) {
	s := openStore(t)
	seedConversations(t, s)

	byApp, total, err := s.ListConversations(ConversationFilter{AppName: "appB"})
	if err != nil || total != 1 || len(byApp) != 1 || byApp[0].SessionID != "sess_3" {
		t.Fatalf("app filter: got %+v total=%d err=%v", byApp, total, err)
	}
	bySource, total, err := s.ListConversations(ConversationFilter{Source: "voice-ui"})
	if err != nil || total != 1 || bySource[0].SessionID != "sess_2" {
		t.Fatalf("source filter: got %+v total=%d err=%v", bySource, total, err)
	}
	paged, total, err := s.ListConversations(ConversationFilter{Limit: 1, Offset: 1})
	if err != nil || total != 3 || len(paged) != 1 {
		t.Fatalf("pagination: got %d rows total=%d err=%v", len(paged), total, err)
	}
}

func TestGetSessionRuns_OldestFirstWithEvents(t *testing.T) {
	s := openStore(t)
	seedConversations(t, s)

	records, err := s.GetSessionRuns("sess_1", "appA")
	if err != nil {
		t.Fatalf("GetSessionRuns: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("got %d runs, want 2", len(records))
	}
	if records[0].RunID != "run_a1" || records[1].RunID != "run_a2" {
		t.Fatalf("order mismatch: %s, %s (want oldest first)", records[0].RunID, records[1].RunID)
	}
	if len(records[0].Events) != 2 {
		t.Fatalf("events not rehydrated: %d", len(records[0].Events))
	}

	empty, err := s.GetSessionRuns("nope", "appA")
	if err != nil || len(empty) != 0 {
		t.Fatalf("missing session must yield empty slice, got %v err=%v", empty, err)
	}
}

func TestDeleteSession_RemovesRunsAndEvents(t *testing.T) {
	s := openStore(t)
	seedConversations(t, s)

	ok, err := s.DeleteSession("sess_1", "appA")
	if err != nil || !ok {
		t.Fatalf("DeleteSession: ok=%v err=%v", ok, err)
	}
	records, _ := s.GetSessionRuns("sess_1", "appA")
	if len(records) != 0 {
		t.Fatalf("runs survived the delete: %d", len(records))
	}
	// Other conversations untouched.
	if _, total, _ := s.ListConversations(ConversationFilter{}); total != 2 {
		t.Fatalf("expected 2 conversations left, got %d", total)
	}
	// Deleting again reports not found.
	ok, err = s.DeleteSession("sess_1", "appA")
	if err != nil || ok {
		t.Fatalf("second delete should be a no-op, ok=%v err=%v", ok, err)
	}
}

func TestConversationStats(t *testing.T) {
	s := openStore(t)
	seedConversations(t, s)

	stats, err := s.Stats()
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}
	if stats.Total != 3 {
		t.Fatalf("total = %d, want 3", stats.Total)
	}
	if stats.ByApp["appA"] != 2 || stats.ByApp["appB"] != 1 {
		t.Fatalf("byApp = %+v", stats.ByApp)
	}
	if stats.BySource["telegram"] != 1 || stats.BySource["voice-ui"] != 1 || stats.BySource["webhook"] != 1 {
		t.Fatalf("bySource = %+v", stats.BySource)
	}
}
