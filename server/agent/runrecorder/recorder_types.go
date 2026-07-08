// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package runrecorder

import (
	"sync"
	"time"
)

// Run status values persisted on RunRecord.Status.
const (
	StatusCompleted   = "completed"
	StatusFailed      = "failed"
	StatusInterrupted = "interrupted"
)

// RunRecord is the persisted representation of a single agent/flow run,
// including every workflow event observed during its execution.
type RunRecord struct {
	RunID     string        `json:"runId"`
	AppName   string        `json:"appName"`
	SessionID string        `json:"sessionId"`
	UserID    string        `json:"userId,omitempty"`
	ClientID  string        `json:"clientId,omitempty"`
	Source    string        `json:"source,omitempty"`
	Input     string        `json:"input,omitempty"`
	StartedAt time.Time     `json:"startedAt"`
	EndedAt   time.Time     `json:"endedAt,omitempty"`
	Status    string        `json:"status"`
	Error     string        `json:"error,omitempty"`
	// NodeTypes snapshots the flow's node ID -> node type map at execution
	// time, so the audit record stays truthful even if the flow is edited
	// later. Empty for runs of plain agents.
	NodeTypes map[string]string `json:"nodeTypes,omitempty"`
	Events    []EventRecord     `json:"events"`
}

// EventRecord is a single workflow/agent event captured during a run, in
// arrival order. Payload holds the full session.Event marshalled to JSON,
// possibly truncated when it exceeds MaxEventPayloadBytes.
type EventRecord struct {
	Seq       int       `json:"seq"`
	Timestamp time.Time `json:"timestamp"`
	Author    string    `json:"author,omitempty"`
	Branch    string    `json:"branch,omitempty"`
	NodePath  string    `json:"nodePath,omitempty"`
	Routes    []string  `json:"routes,omitempty"`
	Payload   []byte    `json:"payload,omitempty"`
}

// attribution carries client/source information captured by the middleware
// layer ahead of the actual run, keyed by session ID until BeforeRunCallback
// picks it up (or it expires).
type attribution struct {
	clientID  string
	source    string
	expiresAt time.Time
}

// runAccumulator collects events for a single in-flight run before it is
// flushed to the sink. Safe for concurrent use since workflow branches can
// emit events from multiple goroutines.
type runAccumulator struct {
	mu sync.Mutex

	runID     string
	appName   string
	sessionID string
	userID    string
	clientID  string
	source    string
	input     string
	nodeTypes map[string]string
	startedAt time.Time
	lastSeen  time.Time

	status string
	errMsg string

	events  []EventRecord
	nextSeq int
}
