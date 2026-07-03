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

// Package runrecorder observes every runner invocation through the adk plugin
// API and persists the raw event stream of each run into a Sink. It is a pure
// observer: it never mutates events and never fails a run, whatever happens
// to its own internal state or to the sink.
package runrecorder

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/plugin"
	"google.golang.org/adk/v2/session"
	"google.golang.org/genai"
)

const (
	defaultMaxEventPayloadBytes = 64 * 1024
	defaultOrphanTimeout        = 30 * time.Minute
	defaultAttributionTTL       = 10 * time.Minute
	sweepInterval               = time.Minute
)

// truncationMarker replaces content that would push an event payload over the
// configured cap. The event itself is always kept.
const truncationMarker = "[truncated by runrecorder]"

// Option customises a Recorder.
type Option func(*Recorder)

// Recorder accumulates the events of in-flight runs and flushes each run to
// the sink when it finishes. Run-fatal errors do not surface through the
// plugin API, so callers report them via MarkRunError.
type Recorder struct {
	sink   Sink
	logger *slog.Logger

	maxPayload     int
	orphanTimeout  time.Duration
	attributionTTL time.Duration

	mu           sync.Mutex
	live         map[string]*runAccumulator
	attributions map[string]attribution
	// nodeTypes holds one node ID -> node type map per app name (flow ID),
	// consulted when a run starts. Replaced wholesale on agent rebuilds;
	// inner maps are never mutated after being set, so accumulators may
	// share them.
	nodeTypes map[string]map[string]string

	stop chan struct{}
	done chan struct{}
}

// WithMaxEventPayloadBytes caps the persisted size of a single event payload.
func WithMaxEventPayloadBytes(n int) Option {
	return func(r *Recorder) { r.maxPayload = n }
}

// WithOrphanTimeout sets how long a run may stay silent before it is flushed
// as interrupted.
func WithOrphanTimeout(d time.Duration) Option {
	return func(r *Recorder) { r.orphanTimeout = d }
}

// WithLogger sets the logger used for swallowed sink and marshal errors.
func WithLogger(l *slog.Logger) Option {
	return func(r *Recorder) { r.logger = l }
}

// New builds a Recorder around the given sink and starts the background
// sweeper that evicts orphaned runs and expired attributions.
func New(sink Sink, opts ...Option) *Recorder {
	r := &Recorder{
		sink:           sink,
		logger:         slog.Default(),
		maxPayload:     defaultMaxEventPayloadBytes,
		orphanTimeout:  defaultOrphanTimeout,
		attributionTTL: defaultAttributionTTL,
		live:           map[string]*runAccumulator{},
		attributions:   map[string]attribution{},
		stop:           make(chan struct{}),
		done:           make(chan struct{}),
	}
	for _, opt := range opts {
		opt(r)
	}
	go r.sweepLoop()
	return r
}

// Close stops the background sweeper. Live accumulators are left in place;
// the process is going away with them.
func (r *Recorder) Close() {
	close(r.stop)
	<-r.done
}

// Annotate records client attribution for a session ahead of its next run.
// The middleware layer calls it when it can resolve the caller identity from
// the HTTP request; the plugin cannot see HTTP.
func (r *Recorder) Annotate(sessionID, clientID, source string) {
	if sessionID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.attributions[sessionID] = attribution{
		clientID:  clientID,
		source:    source,
		expiresAt: time.Now().Add(r.attributionTTL),
	}
}

// SetNodeTypes replaces the per-app node type snapshots. Keyed by app name
// (the flow ID), each inner map goes from node ID to flow node type. The
// agent builder calls it after every rebuild; runs of apps without an entry
// (plain agents) record no types.
func (r *Recorder) SetNodeTypes(types map[string]map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nodeTypes = types
}

// MarkRunError flags a run as failed. If the run is still live the flag is
// applied on flush; if it was already flushed the sink record is updated.
func (r *Recorder) MarkRunError(invocationID, message string) {
	if invocationID == "" {
		return
	}
	r.mu.Lock()
	acc := r.live[invocationID]
	r.mu.Unlock()

	if acc != nil {
		acc.mu.Lock()
		acc.status = StatusFailed
		acc.errMsg = message
		acc.mu.Unlock()
		return
	}
	if err := r.sink.SetRunError(invocationID, message); err != nil {
		r.logger.Warn("runrecorder: failed to set run error on sink", "runId", invocationID, "error", err)
	}
}

// Plugin returns the adk plugin that feeds this recorder. Every callback is
// panic-safe and side-effect free towards the run it observes.
func (r *Recorder) Plugin() (*plugin.Plugin, error) {
	return plugin.New(plugin.Config{
		Name:                  "runrecorder",
		OnUserMessageCallback: r.onUserMessage,
		BeforeRunCallback:     r.beforeRun,
		OnEventCallback:       r.onEvent,
		AfterRunCallback:      r.afterRun,
	})
}

// ensureAccumulator returns the live accumulator for the invocation, creating
// it on first sight. Both onUserMessage and beforeRun can be the first caller
// depending on the runner's internal ordering, so creation is idempotent.
func (r *Recorder) ensureAccumulator(ictx adkagent.InvocationContext) *runAccumulator {
	r.mu.Lock()
	defer r.mu.Unlock()

	if acc, ok := r.live[ictx.InvocationID()]; ok {
		return acc
	}
	acc := &runAccumulator{
		runID:     ictx.InvocationID(),
		startedAt: time.Now(),
		lastSeen:  time.Now(),
		status:    StatusCompleted,
	}
	if ag := ictx.Agent(); ag != nil {
		acc.appName = ag.Name()
		acc.nodeTypes = r.nodeTypes[acc.appName]
	}
	if sess := ictx.Session(); sess != nil {
		acc.sessionID = sess.ID()
		acc.userID = sess.UserID()
	}
	if attr, ok := r.attributions[acc.sessionID]; ok && time.Now().Before(attr.expiresAt) {
		acc.clientID = attr.clientID
		acc.source = attr.source
	}
	r.live[acc.runID] = acc
	return acc
}

// onUserMessage captures the text of the message that starts the invocation.
// The user message never travels through OnEventCallback, so this is the only
// hook where the run's input is visible to the plugin.
func (r *Recorder) onUserMessage(ictx adkagent.InvocationContext, msg *genai.Content) (*genai.Content, error) {
	defer r.recoverPanic("onUserMessage")

	if msg == nil {
		return nil, nil
	}
	var text strings.Builder
	for _, part := range msg.Parts {
		if part != nil && part.Text != "" {
			text.WriteString(part.Text)
		}
	}
	if text.Len() == 0 {
		return nil, nil
	}

	acc := r.ensureAccumulator(ictx)
	acc.mu.Lock()
	acc.input = text.String()
	acc.lastSeen = time.Now()
	acc.mu.Unlock()
	return nil, nil
}

// beforeRun opens the accumulator for a fresh invocation and consumes any
// pending attribution for its session.
func (r *Recorder) beforeRun(ictx adkagent.InvocationContext) (*genai.Content, error) {
	defer r.recoverPanic("beforeRun")

	r.ensureAccumulator(ictx)
	return nil, nil
}

// onEvent appends the event to its run accumulator. It always returns the
// event untouched and a nil error: recording must never alter or break a run.
func (r *Recorder) onEvent(ictx adkagent.InvocationContext, ev *session.Event) (*session.Event, error) {
	defer r.recoverPanic("onEvent")

	if ev == nil {
		return ev, nil
	}
	r.mu.Lock()
	acc := r.live[ictx.InvocationID()]
	r.mu.Unlock()
	if acc == nil {
		return ev, nil
	}

	rec := EventRecord{
		Timestamp: ev.Timestamp,
		Author:    ev.Author,
		Branch:    ev.Branch,
		Routes:    append([]string(nil), ev.Routes...),
		Payload:   r.marshalEvent(ev),
	}
	if ev.NodeInfo != nil {
		rec.NodePath = ev.NodeInfo.Path
	}

	acc.mu.Lock()
	rec.Seq = acc.nextSeq
	acc.nextSeq++
	acc.events = append(acc.events, rec)
	acc.lastSeen = time.Now()
	acc.mu.Unlock()
	return ev, nil
}

// afterRun finalizes and flushes the accumulator. The callback carries no
// outcome information, so the status is whatever MarkRunError left behind.
func (r *Recorder) afterRun(ictx adkagent.InvocationContext) {
	defer r.recoverPanic("afterRun")

	r.mu.Lock()
	acc := r.live[ictx.InvocationID()]
	delete(r.live, ictx.InvocationID())
	r.mu.Unlock()
	if acc == nil {
		return
	}
	r.flush(acc, "")
}

// flush persists an accumulator, optionally overriding its status.
func (r *Recorder) flush(acc *runAccumulator, statusOverride string) {
	acc.mu.Lock()
	record := RunRecord{
		RunID:     acc.runID,
		AppName:   acc.appName,
		SessionID: acc.sessionID,
		UserID:    acc.userID,
		ClientID:  acc.clientID,
		Source:    acc.source,
		Input:     acc.input,
		StartedAt: acc.startedAt,
		EndedAt:   time.Now(),
		Status:    acc.status,
		Error:     acc.errMsg,
		NodeTypes: acc.nodeTypes,
		Events:    acc.events,
	}
	acc.mu.Unlock()

	if statusOverride != "" {
		record.Status = statusOverride
	}
	if err := r.sink.SaveRun(record); err != nil {
		r.logger.Warn("runrecorder: failed to save run", "runId", record.RunID, "error", err)
	}
}

// marshalEvent serializes an event, degrading gracefully when the payload
// exceeds the cap: first the heavy fields are replaced by a marker, and as a
// last resort a minimal stub is stored. The event is never dropped.
func (r *Recorder) marshalEvent(ev *session.Event) []byte {
	payload, err := json.Marshal(ev)
	if err == nil && len(payload) <= r.maxPayload {
		return payload
	}
	if err != nil {
		r.logger.Warn("runrecorder: failed to marshal event", "eventId", ev.ID, "error", err)
	}

	slim := *ev
	slim.Output = truncationMarker
	slim.Content = nil
	payload, err = json.Marshal(&slim)
	if err == nil && len(payload) <= r.maxPayload {
		return payload
	}

	stub := map[string]string{
		"id":           ev.ID,
		"invocationId": ev.InvocationID,
		"author":       ev.Author,
		"note":         truncationMarker,
	}
	payload, _ = json.Marshal(stub)
	return payload
}

// sweepLoop periodically evicts orphaned runs (flushed as interrupted) and
// expired attributions.
func (r *Recorder) sweepLoop() {
	defer close(r.done)
	ticker := time.NewTicker(sweepInterval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stop:
			return
		case <-ticker.C:
			r.sweep(time.Now())
		}
	}
}

// sweep flushes accumulators silent for longer than the orphan timeout and
// drops expired attributions. Exported logic kept separate for testability.
func (r *Recorder) sweep(now time.Time) {
	var orphans []*runAccumulator

	r.mu.Lock()
	for id, acc := range r.live {
		acc.mu.Lock()
		silent := now.Sub(acc.lastSeen)
		acc.mu.Unlock()
		if silent > r.orphanTimeout {
			orphans = append(orphans, acc)
			delete(r.live, id)
		}
	}
	for sid, attr := range r.attributions {
		if now.After(attr.expiresAt) {
			delete(r.attributions, sid)
		}
	}
	r.mu.Unlock()

	for _, acc := range orphans {
		r.flush(acc, StatusInterrupted)
	}
}

// recoverPanic keeps a recording failure from ever reaching the run.
func (r *Recorder) recoverPanic(where string) {
	if p := recover(); p != nil {
		r.logger.Error("runrecorder: recovered panic", "callback", where, "panic", fmt.Sprintf("%v", p))
	}
}
