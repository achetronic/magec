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

package runrecorder

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/session"
)

// fakeSink records calls. SaveRun can be made to fail via failSave.
type fakeSink struct {
	mu        sync.Mutex
	saved     []RunRecord
	runErrors map[string]string
	failSave  bool
}

func newFakeSink() *fakeSink {
	return &fakeSink{runErrors: map[string]string{}}
}

func (s *fakeSink) SaveRun(record RunRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failSave {
		return errors.New("sink unavailable")
	}
	s.saved = append(s.saved, record)
	return nil
}

func (s *fakeSink) SetRunError(runID, message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runErrors[runID] = message
	return nil
}

func (s *fakeSink) lastSaved(t *testing.T) RunRecord {
	t.Helper()
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.saved) == 0 {
		t.Fatal("expected at least one saved run")
	}
	return s.saved[len(s.saved)-1]
}

// Interface-embedding fakes: only the methods the recorder touches are
// implemented. Any other call panics, which doubles as a canary if the
// recorder grows new dependencies on the context.
type fakeAgent struct {
	adkagent.Agent
	name string
}

func (f *fakeAgent) Name() string { return f.name }

type fakeSession struct {
	session.Session
	id     string
	userID string
}

func (f *fakeSession) ID() string     { return f.id }
func (f *fakeSession) UserID() string { return f.userID }

type fakeIC struct {
	adkagent.InvocationContext
	invocationID string
	agent        adkagent.Agent
	session      session.Session
}

func (f *fakeIC) InvocationID() string     { return f.invocationID }
func (f *fakeIC) Agent() adkagent.Agent    { return f.agent }
func (f *fakeIC) Session() session.Session { return f.session }

func newIC(invocationID, appName, sessionID, userID string) *fakeIC {
	return &fakeIC{
		invocationID: invocationID,
		agent:        &fakeAgent{name: appName},
		session:      &fakeSession{id: sessionID, userID: userID},
	}
}

func newEvent(author string, routes ...string) *session.Event {
	ev := &session.Event{}
	ev.ID = "evt_" + author
	ev.Timestamp = time.Now()
	ev.Author = author
	ev.Routes = routes
	return ev
}

// startRecorder builds a recorder whose sweeper is stopped on test cleanup.
func startRecorder(t *testing.T, sink Sink, opts ...Option) *Recorder {
	t.Helper()
	r := New(sink, opts...)
	t.Cleanup(r.Close)
	return r
}

func TestRecorder_GroupsEventsPerInvocationInOrder(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink)
	ic := newIC("inv_1", "flow-demo", "sess_1", "user_1")

	if _, err := r.beforeRun(ic); err != nil {
		t.Fatalf("beforeRun: %v", err)
	}
	for _, author := range []string{"writer", "router_1", "writer"} {
		if _, err := r.onEvent(ic, newEvent(author)); err != nil {
			t.Fatalf("onEvent: %v", err)
		}
	}
	r.afterRun(ic)

	rec := sink.lastSaved(t)
	if rec.RunID != "inv_1" || rec.AppName != "flow-demo" || rec.SessionID != "sess_1" || rec.UserID != "user_1" {
		t.Fatalf("run metadata mismatch: %+v", rec)
	}
	if rec.Status != StatusCompleted {
		t.Fatalf("expected completed, got %s", rec.Status)
	}
	if len(rec.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(rec.Events))
	}
	for i, ev := range rec.Events {
		if ev.Seq != i {
			t.Fatalf("event %d has seq %d", i, ev.Seq)
		}
	}
	if rec.Events[1].Author != "router_1" {
		t.Fatalf("event order not preserved: %+v", rec.Events)
	}
}

func TestRecorder_OnEventForUnknownInvocationIsUntouchedPassthrough(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink)
	ev := newEvent("writer")

	got, err := r.onEvent(newIC("inv_unknown", "app", "s", "u"), ev)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got != ev {
		t.Fatal("expected the same event pointer back")
	}
}

func TestRecorder_AnnotationLandsInFlushedRecord(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink)
	ic := newIC("inv_2", "flow-demo", "sess_2", "u")

	r.Annotate("sess_2", "client_9", "telegram")
	r.beforeRun(ic)
	r.afterRun(ic)

	rec := sink.lastSaved(t)
	if rec.ClientID != "client_9" || rec.Source != "telegram" {
		t.Fatalf("attribution not applied: %+v", rec)
	}
}

func TestRecorder_MarkRunErrorBeforeFlushYieldsFailed(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink)
	ic := newIC("inv_3", "app", "s", "u")

	r.beforeRun(ic)
	r.MarkRunError("inv_3", "boom")
	r.afterRun(ic)

	rec := sink.lastSaved(t)
	if rec.Status != StatusFailed || rec.Error != "boom" {
		t.Fatalf("expected failed/boom, got %s/%s", rec.Status, rec.Error)
	}
}

func TestRecorder_MarkRunErrorAfterFlushReachesSink(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink)
	ic := newIC("inv_4", "app", "s", "u")

	r.beforeRun(ic)
	r.afterRun(ic)
	r.MarkRunError("inv_4", "late failure")

	sink.mu.Lock()
	defer sink.mu.Unlock()
	if sink.runErrors["inv_4"] != "late failure" {
		t.Fatalf("expected SetRunError call, got %+v", sink.runErrors)
	}
}

func TestRecorder_OrphanEvictionFlushesInterrupted(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink, WithOrphanTimeout(time.Millisecond))
	ic := newIC("inv_5", "app", "s", "u")

	r.beforeRun(ic)
	r.onEvent(ic, newEvent("writer"))
	r.sweep(time.Now().Add(time.Second))

	rec := sink.lastSaved(t)
	if rec.Status != StatusInterrupted {
		t.Fatalf("expected interrupted, got %s", rec.Status)
	}
	if len(rec.Events) != 1 {
		t.Fatalf("expected the accumulated event to survive eviction, got %d", len(rec.Events))
	}
}

func TestRecorder_OversizedPayloadIsTruncatedNotDropped(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink, WithMaxEventPayloadBytes(256))
	ic := newIC("inv_6", "app", "s", "u")

	ev := newEvent("code_1")
	ev.Output = strings.Repeat("x", 4096)

	r.beforeRun(ic)
	r.onEvent(ic, ev)
	r.afterRun(ic)

	rec := sink.lastSaved(t)
	if len(rec.Events) != 1 {
		t.Fatal("oversized event was dropped")
	}
	payload := string(rec.Events[0].Payload)
	if len(payload) > 4096 {
		t.Fatalf("payload not truncated: %d bytes", len(payload))
	}
	if !strings.Contains(payload, truncationMarker) {
		t.Fatalf("expected truncation marker in payload: %s", payload)
	}
}

func TestRecorder_SinkFailureDoesNotPanicOrPropagate(t *testing.T) {
	sink := newFakeSink()
	sink.failSave = true
	r := startRecorder(t, sink)
	ic := newIC("inv_7", "app", "s", "u")

	r.beforeRun(ic)
	r.afterRun(ic)
}

func TestRecorder_EventCarriesRoutesAndPayload(t *testing.T) {
	sink := newFakeSink()
	r := startRecorder(t, sink)
	ic := newIC("inv_8", "app", "s", "u")

	r.beforeRun(ic)
	r.onEvent(ic, newEvent("router_1", "done"))
	r.afterRun(ic)

	rec := sink.lastSaved(t)
	got := rec.Events[0]
	if len(got.Routes) != 1 || got.Routes[0] != "done" {
		t.Fatalf("routes not captured: %+v", got.Routes)
	}
	if !strings.Contains(string(got.Payload), "router_1") {
		t.Fatal("payload does not contain the raw event")
	}
}
