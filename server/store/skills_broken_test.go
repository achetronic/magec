// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDetectBrokenSkills_LegacyEntryFlagged confirms the detector
// fires on a v0.x-shaped skill entry (instructions + references on
// the entry itself). The reason string is checked loosely so future
// wording tweaks don't churn the test.
func TestDetectBrokenSkills_LegacyEntryFlagged(t *testing.T) {
	raw := []byte(`{"skills":[
		{"id":"abc","slug":"foo","instructions":"old","references":[{"filename":"x.md","size":3}]}
	]}`)
	got := detectBrokenSkills(raw)
	if len(got) != 1 {
		t.Fatalf("expected 1 broken skill, got %d (%v)", len(got), got)
	}
	reason, ok := got["abc"]
	if !ok {
		t.Fatalf("expected id=abc to be flagged, got %v", got)
	}
	if !strings.Contains(reason, "instructions") || !strings.Contains(reason, "references") {
		t.Errorf("reason should mention legacy fields, got %q", reason)
	}
}

// TestDetectBrokenSkills_CleanEntryUntouched is the noise-floor
// check: a {id, slug} entry stays out of the broken set.
func TestDetectBrokenSkills_CleanEntryUntouched(t *testing.T) {
	raw := []byte(`{"skills":[{"id":"abc","slug":"foo"}]}`)
	got := detectBrokenSkills(raw)
	if len(got) != 0 {
		t.Fatalf("expected no broken skills, got %v", got)
	}
}

// TestDetectBrokenSkills_EmptyLegacyFieldsIgnored guards against a
// false positive when the legacy keys are present but empty (an
// over-zealous JSON marshaller could leave behind "" / [] / {}).
// Empty values are not legacy data — they're the same as absence.
func TestDetectBrokenSkills_EmptyLegacyFieldsIgnored(t *testing.T) {
	raw := []byte(`{"skills":[
		{"id":"abc","slug":"foo","instructions":"","references":[],"name":""}
	]}`)
	got := detectBrokenSkills(raw)
	if len(got) != 0 {
		t.Fatalf("expected no broken skills, got %v", got)
	}
}

// TestDetectBrokenSkills_NoSkillsSection covers the case where
// the store has no skills at all. Detector must return an empty
// map, not nil-panic on the iteration.
func TestDetectBrokenSkills_NoSkillsSection(t *testing.T) {
	got := detectBrokenSkills([]byte(`{"agents":[]}`))
	if len(got) != 0 {
		t.Fatalf("expected no broken skills, got %v", got)
	}
}

// TestDetectBrokenSkills_MalformedJSONReturnsEmpty ensures a
// corrupt payload doesn't bring down loadFromDisk through a
// detector panic; we just shrug and report nothing broken so the
// real unmarshal further down can produce the actual error.
func TestDetectBrokenSkills_MalformedJSONReturnsEmpty(t *testing.T) {
	got := detectBrokenSkills([]byte(`not json`))
	if len(got) != 0 {
		t.Fatalf("expected no broken skills on bad JSON, got %v", got)
	}
}

// TestStore_HidesBrokenSkillsFromAccessors is the integration
// check: a store loaded with one legacy + one clean skill must
// only surface the clean one through ListSkills/GetSkill/
// GetSkillBySlug/ListRawSkills/GetRawSkill. The legacy skill
// remains in the underlying StoreData (we don't rewrite the file
// behind the operator's back) but every accessor pretends it
// isn't there.
func TestStore_HidesBrokenSkillsFromAccessors(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "store.json")
	doc := map[string]any{
		"skills": []any{
			map[string]any{"id": "broken-1", "slug": "old", "instructions": "legacy"},
			map[string]any{"id": "good-1", "slug": "new"},
		},
	}
	body, _ := json.Marshal(doc)
	if err := os.WriteFile(storePath, body, 0o644); err != nil {
		t.Fatalf("write store: %v", err)
	}

	s, err := New(storePath, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	list := s.ListSkills()
	if len(list) != 1 || list[0].ID != "good-1" {
		t.Fatalf("ListSkills should hide broken entries, got %+v", list)
	}
	if _, ok := s.GetSkill("broken-1"); ok {
		t.Errorf("GetSkill should hide broken-1")
	}
	if _, ok := s.GetSkill("good-1"); !ok {
		t.Errorf("GetSkill should expose good-1")
	}
	if _, ok := s.GetSkillBySlug("old"); ok {
		t.Errorf("GetSkillBySlug should hide broken slug")
	}
	if _, ok := s.GetSkillBySlug("new"); !ok {
		t.Errorf("GetSkillBySlug should expose clean slug")
	}
	if _, ok := s.GetRawSkill("broken-1"); ok {
		t.Errorf("GetRawSkill should hide broken-1")
	}
	if rl := s.ListRawSkills(); len(rl) != 1 || rl[0].ID != "good-1" {
		t.Errorf("ListRawSkills should hide broken entries, got %+v", rl)
	}
	if !s.IsSkillBroken("broken-1") {
		t.Errorf("IsSkillBroken should return true for broken-1")
	}
	if s.IsSkillBroken("good-1") {
		t.Errorf("IsSkillBroken should return false for good-1")
	}
}

// TestStore_BrokenSkillSlugFreeForReupload makes sure GetSkillBySlug
// hiding broken entries lets the operator re-upload a slug whose
// legacy entry still sits in store.json: the upload handler probes
// GetSkillBySlug to decide between create-vs-replace, and a hidden
// broken entry must NOT block the create path.
func TestStore_BrokenSkillSlugFreeForReupload(t *testing.T) {
	dir := t.TempDir()
	storePath := filepath.Join(dir, "store.json")
	doc := map[string]any{
		"skills": []any{
			map[string]any{"id": "legacy", "slug": "study-mode", "instructions": "old"},
		},
	}
	body, _ := json.Marshal(doc)
	if err := os.WriteFile(storePath, body, 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	s, err := New(storePath, "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, taken := s.GetSkillBySlug("study-mode"); taken {
		t.Fatalf("a broken legacy entry must not occupy the slug for new uploads")
	}
}
