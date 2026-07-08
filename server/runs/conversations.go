// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package runs

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/achetronic/magec/server/agent/runrecorder"
)

// ConversationFilter narrows down ListConversations results, with pagination.
type ConversationFilter struct {
	AppName  string
	Source   string
	ClientID string
	Limit    int
	Offset   int
}

// ConversationSummary is one aggregated session: every run sharing a
// (session_id, app_name) pair, projected for list views. A conversation is
// not stored anywhere; it exists only as this aggregation (decision #31
// phase 2).
type ConversationSummary struct {
	SessionID  string
	AppName    string
	UserID     string
	ClientID   string
	Source     string
	StartedAt  time.Time
	LastAt     time.Time
	Turns      int
	FirstInput string
}

// ListConversations groups runs by session and returns conversation
// summaries ordered by most recent activity, plus the total number of
// matching conversations. Runs without a session_id cannot belong to a
// conversation and are skipped.
func (s *Store) ListConversations(f ConversationFilter) ([]ConversationSummary, int, error) {
	where := "WHERE session_id != ''"
	args := []any{}
	if f.AppName != "" {
		where += " AND app_name = ?"
		args = append(args, f.AppName)
	}
	if f.Source != "" {
		where += " AND source = ?"
		args = append(args, f.Source)
	}
	if f.ClientID != "" {
		where += " AND client_id = ?"
		args = append(args, f.ClientID)
	}

	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM (SELECT 1 FROM runs `+where+` GROUP BY session_id, app_name)`, args...,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count conversations: %w", err)
	}

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	// user_id/client_id/source are attribution fields constant within a
	// session in practice; MAX picks a deterministic representative without
	// grouping by them (which would split conversations on attribution gaps).
	query := `SELECT session_id, app_name, MAX(user_id), MAX(client_id), MAX(source),
	                 MIN(started_at), MAX(started_at), COUNT(*),
	                 (SELECT input FROM runs r2
	                  WHERE r2.session_id = runs.session_id AND r2.app_name = runs.app_name
	                  ORDER BY r2.started_at ASC LIMIT 1)
	          FROM runs ` + where + `
	          GROUP BY session_id, app_name
	          ORDER BY MAX(started_at) DESC LIMIT ? OFFSET ?`
	rows, err := s.db.Query(query, append(args, limit, f.Offset)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list conversations: %w", err)
	}
	defer rows.Close()

	summaries := []ConversationSummary{}
	for rows.Next() {
		var sum ConversationSummary
		var startedAt, lastAt int64
		var firstInput sql.NullString
		if err := rows.Scan(&sum.SessionID, &sum.AppName, &sum.UserID, &sum.ClientID,
			&sum.Source, &startedAt, &lastAt, &sum.Turns, &firstInput); err != nil {
			return nil, 0, fmt.Errorf("scan conversation: %w", err)
		}
		sum.StartedAt = time.UnixMilli(startedAt)
		sum.LastAt = time.UnixMilli(lastAt)
		sum.FirstInput = firstInput.String
		summaries = append(summaries, sum)
	}
	return summaries, total, rows.Err()
}

// GetSessionRuns rehydrates every run of one conversation, oldest first,
// events included. An empty slice means the conversation does not exist.
func (s *Store) GetSessionRuns(sessionID, appName string) ([]runrecorder.RunRecord, error) {
	rows, err := s.db.Query(
		`SELECT run_id FROM runs WHERE session_id = ? AND app_name = ? ORDER BY started_at ASC`,
		sessionID, appName)
	if err != nil {
		return nil, fmt.Errorf("list session runs: %w", err)
	}
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan run id: %w", err)
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	records := []runrecorder.RunRecord{}
	for _, id := range ids {
		record, ok, err := s.GetRun(id)
		if err != nil {
			return nil, err
		}
		if ok {
			records = append(records, record)
		}
	}
	return records, nil
}

// DeleteSession removes every run (and their events) of one conversation.
// The boolean is false when no run matched.
func (s *Store) DeleteSession(sessionID, appName string) (bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return false, fmt.Errorf("begin delete session: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`DELETE FROM events WHERE run_id IN (SELECT run_id FROM runs WHERE session_id = ? AND app_name = ?)`,
		sessionID, appName); err != nil {
		return false, fmt.Errorf("delete session events: %w", err)
	}
	res, err := tx.Exec(`DELETE FROM runs WHERE session_id = ? AND app_name = ?`, sessionID, appName)
	if err != nil {
		return false, fmt.Errorf("delete session runs: %w", err)
	}
	affected, _ := res.RowsAffected()
	if err := tx.Commit(); err != nil {
		return false, fmt.Errorf("commit delete session: %w", err)
	}
	return affected > 0, nil
}

// DeleteAllSessions removes every run that belongs to a session (and their
// events), returning how many conversations were cleared. Session-less runs
// are not conversations and stay.
func (s *Store) DeleteAllSessions() (int, error) {
	var total int
	if err := s.db.QueryRow(
		`SELECT COUNT(*) FROM (SELECT 1 FROM runs WHERE session_id != '' GROUP BY session_id, app_name)`,
	).Scan(&total); err != nil {
		return 0, fmt.Errorf("count sessions: %w", err)
	}

	tx, err := s.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("begin clear sessions: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM events WHERE run_id IN (SELECT run_id FROM runs WHERE session_id != '')`); err != nil {
		return 0, fmt.Errorf("clear session events: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM runs WHERE session_id != ''`); err != nil {
		return 0, fmt.Errorf("clear session runs: %w", err)
	}
	return total, tx.Commit()
}

// ConversationStats aggregates conversation counts by app and by source for
// the stats endpoint.
type ConversationStats struct {
	Total    int
	ByApp    map[string]int
	BySource map[string]int
}

// Stats returns aggregated conversation counts.
func (s *Store) Stats() (ConversationStats, error) {
	stats := ConversationStats{ByApp: map[string]int{}, BySource: map[string]int{}}

	rows, err := s.db.Query(
		`SELECT app_name, MAX(source), COUNT(*) FROM (
		   SELECT app_name, session_id, MAX(source) AS source
		   FROM runs WHERE session_id != ''
		   GROUP BY session_id, app_name
		 ) GROUP BY app_name`)
	if err != nil {
		return stats, fmt.Errorf("conversation stats: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var app, source string
		var count int
		if err := rows.Scan(&app, &source, &count); err != nil {
			return stats, fmt.Errorf("scan stats: %w", err)
		}
		stats.ByApp[app] += count
		stats.Total += count
	}
	if err := rows.Err(); err != nil {
		return stats, err
	}

	srcRows, err := s.db.Query(
		`SELECT source, COUNT(*) FROM (
		   SELECT session_id, app_name, MAX(source) AS source
		   FROM runs WHERE session_id != ''
		   GROUP BY session_id, app_name
		 ) GROUP BY source`)
	if err != nil {
		return stats, fmt.Errorf("conversation source stats: %w", err)
	}
	defer srcRows.Close()
	for srcRows.Next() {
		var source string
		var count int
		if err := srcRows.Scan(&source, &count); err != nil {
			return stats, fmt.Errorf("scan source stats: %w", err)
		}
		stats.BySource[source] += count
	}
	return stats, srcRows.Err()
}
