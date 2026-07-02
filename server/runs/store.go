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

// Package runs persists recorded agent/flow runs in a SQLite database using
// the pure Go modernc.org/sqlite driver, keeping the build free of CGO. It
// implements runrecorder.Sink and offers the query API the admin endpoints
// project their views from.
package runs

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	_ "modernc.org/sqlite"

	"github.com/achetronic/magec/server/agent/runrecorder"
)

// Open opens (or creates) the run database at path, enables WAL and applies
// the schema. The single connection keeps modernc/sqlite in its comfort zone
// of one writer.
func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open runs db: %w", err)
	}
	db.SetMaxOpenConns(1)

	for _, pragma := range []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA busy_timeout=5000",
		"PRAGMA foreign_keys=ON",
	} {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("apply %q: %w", pragma, err)
		}
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("apply runs schema: %w", err)
	}
	return &Store{db: db}, nil
}

// Close stops the sweeper, when running, and closes the database.
func (s *Store) Close() error {
	if s.stopSweep != nil {
		close(s.stopSweep)
		<-s.doneSweep
	}
	return s.db.Close()
}

// SaveRun persists a run and its events in a single transaction.
func (s *Store) SaveRun(record runrecorder.RunRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO runs (run_id, app_name, session_id, user_id, client_id, source, started_at, ended_at, status, error)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.RunID, record.AppName, record.SessionID, record.UserID, record.ClientID,
		record.Source, record.StartedAt.UnixMilli(), record.EndedAt.UnixMilli(),
		record.Status, record.Error,
	); err != nil {
		return fmt.Errorf("insert run %s: %w", record.RunID, err)
	}

	stmt, err := tx.Prepare(
		`INSERT INTO events (run_id, seq, timestamp, author, branch, node_path, routes, payload)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return fmt.Errorf("prepare events insert: %w", err)
	}
	defer stmt.Close()

	for _, ev := range record.Events {
		routes, err := json.Marshal(ev.Routes)
		if err != nil {
			routes = []byte("[]")
		}
		if _, err := stmt.Exec(
			record.RunID, ev.Seq, ev.Timestamp.UnixMilli(), ev.Author, ev.Branch,
			ev.NodePath, string(routes), ev.Payload,
		); err != nil {
			return fmt.Errorf("insert event %d of run %s: %w", ev.Seq, record.RunID, err)
		}
	}
	return tx.Commit()
}

// SetRunError marks an already-persisted run as failed with the given message.
func (s *Store) SetRunError(runID, message string) error {
	_, err := s.db.Exec(`UPDATE runs SET status = ?, error = ? WHERE run_id = ?`,
		runrecorder.StatusFailed, message, runID)
	if err != nil {
		return fmt.Errorf("set run error %s: %w", runID, err)
	}
	return nil
}

// ListRuns returns run summaries newest first, filtered and paginated, along
// with the total number of matching runs.
func (s *Store) ListRuns(f RunFilter) ([]RunSummary, int, error) {
	where := "WHERE 1=1"
	args := []any{}
	if f.AppName != "" {
		where += " AND app_name = ?"
		args = append(args, f.AppName)
	}
	if f.Status != "" {
		where += " AND status = ?"
		args = append(args, f.Status)
	}

	var total int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM runs `+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count runs: %w", err)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	query := `SELECT r.run_id, r.app_name, r.session_id, r.user_id, r.client_id, r.source,
	                 r.started_at, r.ended_at, r.status, r.error,
	                 (SELECT COUNT(*) FROM events e WHERE e.run_id = r.run_id)
	          FROM runs r ` + where + ` ORDER BY r.started_at DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, append(args, limit, f.Offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list runs: %w", err)
	}
	defer rows.Close()

	summaries := []RunSummary{}
	for rows.Next() {
		var sum RunSummary
		var startedAt, endedAt int64
		if err := rows.Scan(&sum.RunID, &sum.AppName, &sum.SessionID, &sum.UserID,
			&sum.ClientID, &sum.Source, &startedAt, &endedAt, &sum.Status, &sum.Error,
			&sum.EventCount); err != nil {
			return nil, 0, fmt.Errorf("scan run: %w", err)
		}
		sum.StartedAt = time.UnixMilli(startedAt)
		sum.EndedAt = time.UnixMilli(endedAt)
		summaries = append(summaries, sum)
	}
	return summaries, total, rows.Err()
}

// GetRun rehydrates a full run record, events in seq order. The boolean is
// false when the run does not exist.
func (s *Store) GetRun(runID string) (runrecorder.RunRecord, bool, error) {
	var record runrecorder.RunRecord
	var startedAt, endedAt int64
	err := s.db.QueryRow(
		`SELECT run_id, app_name, session_id, user_id, client_id, source, started_at, ended_at, status, error
		 FROM runs WHERE run_id = ?`, runID,
	).Scan(&record.RunID, &record.AppName, &record.SessionID, &record.UserID,
		&record.ClientID, &record.Source, &startedAt, &endedAt, &record.Status, &record.Error)
	if err == sql.ErrNoRows {
		return record, false, nil
	}
	if err != nil {
		return record, false, fmt.Errorf("get run %s: %w", runID, err)
	}
	record.StartedAt = time.UnixMilli(startedAt)
	record.EndedAt = time.UnixMilli(endedAt)

	rows, err := s.db.Query(
		`SELECT seq, timestamp, author, branch, node_path, routes, payload
		 FROM events WHERE run_id = ? ORDER BY seq`, runID)
	if err != nil {
		return record, false, fmt.Errorf("get events of %s: %w", runID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var ev runrecorder.EventRecord
		var ts int64
		var routes string
		if err := rows.Scan(&ev.Seq, &ts, &ev.Author, &ev.Branch, &ev.NodePath,
			&routes, &ev.Payload); err != nil {
			return record, false, fmt.Errorf("scan event: %w", err)
		}
		ev.Timestamp = time.UnixMilli(ts)
		if routes != "" {
			json.Unmarshal([]byte(routes), &ev.Routes)
		}
		record.Events = append(record.Events, ev)
	}
	return record, true, rows.Err()
}

// Sweep deletes runs older than maxAge and, per app, runs beyond the newest
// maxPerApp, together with their events. A zero value disables that bound.
func (s *Store) Sweep(maxAge time.Duration, maxPerApp int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin sweep: %w", err)
	}
	defer tx.Rollback()

	if maxAge > 0 {
		cutoff := time.Now().Add(-maxAge).UnixMilli()
		if _, err := tx.Exec(`DELETE FROM events WHERE run_id IN (SELECT run_id FROM runs WHERE started_at < ?)`, cutoff); err != nil {
			return fmt.Errorf("sweep old events: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM runs WHERE started_at < ?`, cutoff); err != nil {
			return fmt.Errorf("sweep old runs: %w", err)
		}
	}

	if maxPerApp > 0 {
		// Rank runs per app by recency; everything past the cap goes away.
		const overflow = `SELECT run_id FROM (
			SELECT run_id, ROW_NUMBER() OVER (PARTITION BY app_name ORDER BY started_at DESC) AS rn
			FROM runs
		) WHERE rn > ?`
		if _, err := tx.Exec(`DELETE FROM events WHERE run_id IN (`+overflow+`)`, maxPerApp); err != nil {
			return fmt.Errorf("sweep overflow events: %w", err)
		}
		if _, err := tx.Exec(`DELETE FROM runs WHERE run_id IN (`+overflow+`)`, maxPerApp); err != nil {
			return fmt.Errorf("sweep overflow runs: %w", err)
		}
	}
	return tx.Commit()
}

// StartSweeper runs Sweep on the given interval until Close is called.
func (s *Store) StartSweeper(interval, maxAge time.Duration, maxPerApp int) {
	s.stopSweep = make(chan struct{})
	s.doneSweep = make(chan struct{})
	go func() {
		defer close(s.doneSweep)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-s.stopSweep:
				return
			case <-ticker.C:
				if err := s.Sweep(maxAge, maxPerApp); err != nil {
					slog.Warn("runs: sweep failed", "error", err)
				}
			}
		}
	}()
}
