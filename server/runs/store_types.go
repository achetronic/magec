// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package runs

import (
	"database/sql"
	"time"
)

// schema holds the DDL statements applied on Open to bring a fresh or
// existing database file up to the expected shape.
const schema = `
CREATE TABLE IF NOT EXISTS runs (
	run_id     TEXT PRIMARY KEY,
	app_name   TEXT NOT NULL,
	session_id TEXT,
	user_id    TEXT,
	client_id  TEXT,
	source     TEXT,
	input      TEXT,
	node_types TEXT,
	started_at INTEGER NOT NULL,
	ended_at   INTEGER,
	status     TEXT NOT NULL,
	error      TEXT
);

CREATE TABLE IF NOT EXISTS events (
	run_id    TEXT NOT NULL,
	seq       INTEGER NOT NULL,
	timestamp INTEGER,
	author    TEXT,
	branch    TEXT,
	node_path TEXT,
	routes    TEXT,
	payload   BLOB,
	PRIMARY KEY (run_id, seq)
);

CREATE INDEX IF NOT EXISTS idx_runs_app_started ON runs(app_name, started_at DESC);
`

// Store persists run records and their events in a SQLite database and
// implements runrecorder.Sink.
type Store struct {
	db *sql.DB

	stopSweep chan struct{}
	doneSweep chan struct{}
}

// RunFilter narrows down ListRuns results by app name and/or status, with
// pagination via Limit/Offset. Zero Limit means no bound is applied by the
// caller side default (callers should set a sane limit).
type RunFilter struct {
	AppName string
	Status  string
	Limit   int
	Offset  int
}

// RunSummary is a lightweight projection of a run row, without its events,
// suitable for list views.
type RunSummary struct {
	RunID      string
	AppName    string
	SessionID  string
	UserID     string
	ClientID   string
	Source     string
	StartedAt  time.Time
	EndedAt    time.Time
	Status     string
	Error      string
	EventCount int
}
