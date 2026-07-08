// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/achetronic/magec/server/agent/runrecorder"
	"github.com/achetronic/magec/server/runs"
	"github.com/achetronic/magec/server/store"
)

// conversationFixture wires an admin handler over a store with one flow
// (writer is the response agent) and a runs DB holding one two-turn session.
func conversationFixture(t *testing.T) (*Handler, string) {
	t.Helper()
	s, err := store.New(filepath.Join(t.TempDir(), "store.json"), "")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	flow, err := s.CreateFlow(store.FlowDefinition{
		Name:  "Pipeline",
		Entry: "researcher",
		Nodes: []store.FlowNode{
			{ID: "researcher", Type: store.FlowNodeAgent, AgentID: "a1"},
			{ID: "writer", Type: store.FlowNodeAgent, AgentID: "a2", ResponseAgent: true},
		},
		Edges: []store.FlowEdge{{From: "researcher", To: "writer"}},
	})
	if err != nil {
		t.Fatalf("CreateFlow: %v", err)
	}

	db, err := runs.Open(filepath.Join(t.TempDir(), "runs.db"))
	if err != nil {
		t.Fatalf("runs.Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	base := time.Now().Add(-time.Hour).Truncate(time.Millisecond)
	event := func(seq int, author, text string) runrecorder.EventRecord {
		payload, _ := json.Marshal(map[string]any{
			"Content": map[string]any{"parts": []any{map[string]any{"text": text}}},
		})
		return runrecorder.EventRecord{Seq: seq, Timestamp: base.Add(time.Duration(seq) * time.Second), Author: author, Payload: payload}
	}
	for i, input := range []string{"first question", "second question"} {
		rec := runrecorder.RunRecord{
			RunID:     "run_" + string(rune('a'+i)),
			AppName:   flow.ID,
			SessionID: "sess_x",
			UserID:    "u1",
			ClientID:  "c1",
			Source:    "telegram",
			Input:     input,
			StartedAt: base.Add(time.Duration(i) * 10 * time.Minute),
			EndedAt:   base.Add(time.Duration(i)*10*time.Minute + time.Minute),
			Status:    runrecorder.StatusCompleted,
			Events: []runrecorder.EventRecord{
				event(0, "researcher", "internal research notes"),
				event(1, "writer", "final answer"),
				event(2, "__meta__", "bookkeeping"),
			},
		}
		if err := db.SaveRun(rec); err != nil {
			t.Fatalf("SaveRun: %v", err)
		}
	}

	h := New(s)
	h.SetRunsStore(db)
	return h, flow.ID
}

func getJSON(t *testing.T, h *Handler, url string) map[string]any {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, url, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s = %d: %s", url, rec.Code, rec.Body.String())
	}
	var out map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return out
}

func TestConversations_ListAggregatesSession(t *testing.T) {
	h, flowID := conversationFixture(t)

	out := getJSON(t, h, "/conversations")
	items := out["items"].([]any)
	if len(items) != 1 || out["total"].(float64) != 1 {
		t.Fatalf("expected 1 conversation, got %v", out)
	}
	c := items[0].(map[string]any)
	if c["turns"].(float64) != 2 {
		t.Fatalf("turns = %v, want 2", c["turns"])
	}
	if c["id"] != flowID+":sess_x" {
		t.Fatalf("id = %v", c["id"])
	}
	if c["flowName"] != "Pipeline" || c["preview"] != "first question" {
		t.Fatalf("decoration missing: %v", c)
	}
}

// TestConversations_PerspectiveIsOnReadFilter is the core canary of decision
// #31 phase 2: the same recorded runs yield the admin view (every agent) and
// the user view (response agents only), with internal __ authors never shown.
func TestConversations_PerspectiveIsOnReadFilter(t *testing.T) {
	h, flowID := conversationFixture(t)
	id := flowID + ":sess_x"

	admin := getJSON(t, h, "/conversations/"+id)
	adminMsgs := admin["conversation"].(map[string]any)["messages"].([]any)
	// Per run: user input + researcher + writer (never __meta__) = 3, twice.
	if len(adminMsgs) != 6 {
		t.Fatalf("admin messages = %d, want 6: %v", len(adminMsgs), adminMsgs)
	}
	for _, m := range adminMsgs {
		if m.(map[string]any)["agent"] == "__meta__" {
			t.Fatalf("internal author leaked into projection")
		}
	}

	user := getJSON(t, h, "/conversations/"+id+"?view=user")
	userMsgs := user["conversation"].(map[string]any)["messages"].([]any)
	// Per run: user input + writer only = 2, twice.
	if len(userMsgs) != 4 {
		t.Fatalf("user messages = %d, want 4: %v", len(userMsgs), userMsgs)
	}
	for _, m := range userMsgs {
		mm := m.(map[string]any)
		if mm["role"] == "assistant" && mm["agent"] != "writer" {
			t.Fatalf("non-response agent %v leaked into user view", mm["agent"])
		}
	}
}

func TestConversations_DeleteRemovesRuns(t *testing.T) {
	h, flowID := conversationFixture(t)
	id := flowID + ":sess_x"

	req := httptest.NewRequest(http.MethodDelete, "/conversations/"+id, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete = %d", rec.Code)
	}

	out := getJSON(t, h, "/conversations")
	if out["total"].(float64) != 0 {
		t.Fatalf("conversation survived: %v", out)
	}
	// Deleting a conversation deletes its runs: the Runs view loses them too.
	req = httptest.NewRequest(http.MethodGet, "/runs", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var runsOut map[string]any
	json.Unmarshal(rec.Body.Bytes(), &runsOut)
	if runsOut["total"].(float64) != 0 {
		t.Fatalf("runs survived the conversation delete: %v", runsOut)
	}
}

func TestConversations_MessagePagination(t *testing.T) {
	h, flowID := conversationFixture(t)
	id := flowID + ":sess_x"

	out := getJSON(t, h, "/conversations/"+id+"?msgLimit=2")
	msgs := out["conversation"].(map[string]any)["messages"].([]any)
	if len(msgs) != 2 || out["totalMessages"].(float64) != 6 {
		t.Fatalf("pagination: %d msgs total=%v", len(msgs), out["totalMessages"])
	}
	// The window is anchored at the end: last message of the last run.
	last := msgs[1].(map[string]any)
	if last["agent"] != "writer" || last["runId"] != "run_b" {
		t.Fatalf("expected the newest messages, got %v", msgs)
	}
}

// TestPreviewText_StripsMetadataBeforeTruncating is the canary for a real
// bug: a MAGEC_META block truncated at 120 chars loses its closing marker
// and leaked into the conversation list preview as raw comment text.
func TestPreviewText_StripsMetadataBeforeTruncating(t *testing.T) {
	long := `<!--MAGEC_META:{"source":"telegram","chat_id":123456789,"user":"somebody","first_name":"Some","last_name":"Body","username":"nick"}:MAGEC_META-->` + "\nhello agent"
	if got := previewText(long); got != "hello agent" {
		t.Fatalf("previewText = %q, want metadata gone", got)
	}
	history := `<!--MAGEC_THREAD_HISTORY:[{"a":1}]:MAGEC_THREAD_HISTORY-->` + "\nquestion"
	if got := previewText(history); got != "question" {
		t.Fatalf("previewText = %q, want thread history gone", got)
	}
	if got := previewText("plain text"); got != "plain text" {
		t.Fatalf("previewText = %q, plain text must pass through", got)
	}
}
