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

package runs

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/achetronic/magec/server/agent/runrecorder"
)

func openStore(t *testing.T) *Store {
	t.Helper()
	s, err := Open(filepath.Join(t.TempDir(), "runs.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func sampleRun(runID, appName string, startedAt time.Time, events int) runrecorder.RunRecord {
	rec := runrecorder.RunRecord{
		RunID:     runID,
		AppName:   appName,
		SessionID: "sess_" + runID,
		UserID:    "user_1",
		ClientID:  "client_1",
		Source:    "telegram",
		StartedAt: startedAt,
		EndedAt:   startedAt.Add(3 * time.Second),
		Status:    runrecorder.StatusCompleted,
	}
	for i := 0; i < events; i++ {
		rec.Events = append(rec.Events, runrecorder.EventRecord{
			Seq:       i,
			Timestamp: startedAt.Add(time.Duration(i) * time.Second),
			Author:    "node_a",
			Branch:    "b",
			NodePath:  "p",
			Routes:    []string{"done"},
			Payload:   []byte(`{"author":"node_a"}`),
		})
	}
	return rec
}

func TestStore_SaveAndGetRoundtrip(t *testing.T) {
	s := openStore(t)
	base := time.Now().Truncate(time.Millisecond)
	want := sampleRun("run_1", "flow-demo", base, 3)

	if err := s.SaveRun(want); err != nil {
		t.Fatalf("SaveRun: %v", err)
	}
	got, ok, err := s.GetRun("run_1")
	if err != nil || !ok {
		t.Fatalf("GetRun: ok=%v err=%v", ok, err)
	}

	if got.AppName != want.AppName || got.SessionID != want.SessionID ||
		got.UserID != want.UserID || got.ClientID != want.ClientID ||
		got.Source != want.Source || got.Status != want.Status {
		t.Fatalf("run fields mismatch: %+v", got)
	}
	if !got.StartedAt.Equal(want.StartedAt) {
		t.Fatalf("startedAt mismatch: %v vs %v", got.StartedAt, want.StartedAt)
	}
	if len(got.Events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(got.Events))
	}
	for i, ev := range got.Events {
		if ev.Seq != i {
			t.Fatalf("events out of order at %d: seq=%d", i, ev.Seq)
		}
	}
	if got.Events[0].Routes[0] != "done" || got.Events[0].NodePath != "p" {
		t.Fatalf("event fields mismatch: %+v", got.Events[0])
	}
	if string(got.Events[0].Payload) != `{"author":"node_a"}` {
		t.Fatalf("payload mismatch: %s", got.Events[0].Payload)
	}
}

func TestStore_GetMissingRun(t *testing.T) {
	s := openStore(t)

	_, ok, err := s.GetRun("nope")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected ok=false for a missing run")
	}
}

func TestStore_ListRunsFiltersAndPaginates(t *testing.T) {
	s := openStore(t)
	base := time.Now().Add(-time.Hour)
	for i := 0; i < 5; i++ {
		run := sampleRun(runID(i), "flow-a", base.Add(time.Duration(i)*time.Minute), 1)
		if i == 4 {
			run.Status = runrecorder.StatusFailed
		}
		if err := s.SaveRun(run); err != nil {
			t.Fatalf("SaveRun %d: %v", i, err)
		}
	}
	if err := s.SaveRun(sampleRun("other_1", "flow-b", base, 1)); err != nil {
		t.Fatalf("SaveRun other: %v", err)
	}

	all, total, err := s.ListRuns(RunFilter{AppName: "flow-a", Limit: 2})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if total != 5 || len(all) != 2 {
		t.Fatalf("expected total 5 page 2, got total %d page %d", total, len(all))
	}
	if all[0].RunID != "run_4" {
		t.Fatalf("expected newest first, got %s", all[0].RunID)
	}
	if all[0].EventCount != 1 {
		t.Fatalf("expected event count 1, got %d", all[0].EventCount)
	}

	failed, total, err := s.ListRuns(RunFilter{Status: runrecorder.StatusFailed})
	if err != nil {
		t.Fatalf("ListRuns failed filter: %v", err)
	}
	if total != 1 || len(failed) != 1 || failed[0].RunID != "run_4" {
		t.Fatalf("status filter broken: total=%d %+v", total, failed)
	}

	page2, _, err := s.ListRuns(RunFilter{AppName: "flow-a", Limit: 2, Offset: 4})
	if err != nil {
		t.Fatalf("ListRuns page 2: %v", err)
	}
	if len(page2) != 1 || page2[0].RunID != "run_0" {
		t.Fatalf("pagination broken: %+v", page2)
	}
}

func TestStore_SetRunError(t *testing.T) {
	s := openStore(t)
	if err := s.SaveRun(sampleRun("run_e", "app", time.Now(), 0)); err != nil {
		t.Fatalf("SaveRun: %v", err)
	}

	if err := s.SetRunError("run_e", "late boom"); err != nil {
		t.Fatalf("SetRunError: %v", err)
	}

	got, ok, err := s.GetRun("run_e")
	if err != nil || !ok {
		t.Fatalf("GetRun: ok=%v err=%v", ok, err)
	}
	if got.Status != runrecorder.StatusFailed || got.Error != "late boom" {
		t.Fatalf("expected failed/late boom, got %s/%s", got.Status, got.Error)
	}
}

func TestStore_SweepByAge(t *testing.T) {
	s := openStore(t)
	old := sampleRun("run_old", "app", time.Now().Add(-48*time.Hour), 2)
	fresh := sampleRun("run_new", "app", time.Now(), 2)
	for _, r := range []runrecorder.RunRecord{old, fresh} {
		if err := s.SaveRun(r); err != nil {
			t.Fatalf("SaveRun: %v", err)
		}
	}

	if err := s.Sweep(24*time.Hour, 0); err != nil {
		t.Fatalf("Sweep: %v", err)
	}

	if _, ok, _ := s.GetRun("run_old"); ok {
		t.Fatal("old run survived the age sweep")
	}
	got, ok, _ := s.GetRun("run_new")
	if !ok || len(got.Events) != 2 {
		t.Fatalf("fresh run damaged by sweep: ok=%v events=%d", ok, len(got.Events))
	}
}

func TestStore_SweepByMaxPerApp(t *testing.T) {
	s := openStore(t)
	base := time.Now().Add(-time.Hour)
	for i := 0; i < 5; i++ {
		if err := s.SaveRun(sampleRun(runID(i), "flow-a", base.Add(time.Duration(i)*time.Minute), 1)); err != nil {
			t.Fatalf("SaveRun: %v", err)
		}
	}
	if err := s.SaveRun(sampleRun("other_1", "flow-b", base, 1)); err != nil {
		t.Fatalf("SaveRun other: %v", err)
	}

	if err := s.Sweep(0, 2); err != nil {
		t.Fatalf("Sweep: %v", err)
	}

	survivors, total, err := s.ListRuns(RunFilter{AppName: "flow-a"})
	if err != nil {
		t.Fatalf("ListRuns: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected exactly 2 survivors for flow-a, got %d", total)
	}
	if survivors[0].RunID != "run_4" || survivors[1].RunID != "run_3" {
		t.Fatalf("wrong survivors: %+v", survivors)
	}
	if _, otherTotal, _ := s.ListRuns(RunFilter{AppName: "flow-b"}); otherTotal != 1 {
		t.Fatalf("flow-b should be untouched, got %d", otherTotal)
	}
}

func runID(i int) string {
	return "run_" + string(rune('0'+i))
}
