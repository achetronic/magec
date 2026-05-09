// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	adkskill "google.golang.org/adk/tool/skilltoolset/skill"
)

// TestMigrateLegacySkills_HappyPath simulates a v0.x store with one
// skill plus a reference file on disk, and asserts that:
//   - the JSON gets rewritten to the new minimal shape ({id, slug})
//   - SKILL.md is generated with the expected frontmatter
//   - the legacy data/skills/{id}/file is moved under references/
func TestMigrateLegacySkills_HappyPath(t *testing.T) {
	dataDir := t.TempDir()
	legacy := filepath.Join(dataDir, "skills", "abc-123")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatalf("mkdir legacy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "policy.md"), []byte("ref body"), 0o644); err != nil {
		t.Fatalf("write legacy file: %v", err)
	}

	raw := `{
		"skills": [
			{
				"id": "abc-123",
				"name": "Returns Policy",
				"description": "How returns work",
				"instructions": "Be empathetic.",
				"references": [{"filename": "policy.md", "size": 8}]
			}
		]
	}`
	out, err := MigrateLegacySkills(dataDir, []byte(raw))
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	skills := doc["skills"].([]any)
	entry := skills[0].(map[string]any)
	if entry["slug"] != "returns-policy" {
		t.Fatalf("expected slug=returns-policy, got %v", entry["slug"])
	}
	if entry["id"] != "abc-123" {
		t.Fatalf("expected id=abc-123, got %v", entry["id"])
	}
	for _, k := range []string{"name", "description", "instructions", "references"} {
		if _, ok := entry[k]; ok {
			t.Errorf("legacy field %q still present in entry: %v", k, entry)
		}
	}

	skillMD, err := os.ReadFile(filepath.Join(dataDir, "skills", "returns-policy", "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	body := string(skillMD)
	for _, want := range []string{"name: returns-policy", "description: How returns work", "Be empathetic."} {
		if !contains(body, want) {
			t.Errorf("SKILL.md missing %q. Got:\n%s", want, body)
		}
	}

	movedRef := filepath.Join(dataDir, "skills", "returns-policy", "references", "policy.md")
	if data, err := os.ReadFile(movedRef); err != nil {
		t.Fatalf("expected reference at %s, got %v", movedRef, err)
	} else if string(data) != "ref body" {
		t.Fatalf("reference content lost: %q", data)
	}

	// Legacy directory should have been removed (or at least emptied).
	if _, err := os.Stat(filepath.Join(dataDir, "skills", "abc-123")); err == nil {
		t.Errorf("legacy directory still exists")
	}
}

// TestMigrateLegacySkills_Idempotent ensures running the migrator twice
// does not change the second output and does not duplicate work.
func TestMigrateLegacySkills_Idempotent(t *testing.T) {
	dataDir := t.TempDir()
	raw := []byte(`{"skills":[{"id":"x","name":"X","description":"d","instructions":"i"}]}`)

	first, err := MigrateLegacySkills(dataDir, raw)
	if err != nil {
		t.Fatalf("first migration: %v", err)
	}
	second, err := MigrateLegacySkills(dataDir, first)
	if err != nil {
		t.Fatalf("second migration: %v", err)
	}
	if string(first) != string(second) {
		t.Errorf("idempotency broken:\nfirst : %s\nsecond: %s", first, second)
	}
}

// TestMigrateLegacySkills_NoSkills is a noise-floor test: a store with
// no skills section is returned untouched.
func TestMigrateLegacySkills_NoSkills(t *testing.T) {
	raw := []byte(`{"agents":[]}`)
	out, err := MigrateLegacySkills(t.TempDir(), raw)
	if err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if string(out) != string(raw) {
		t.Errorf("no-op migration changed bytes: %s", out)
	}
}

// TestMigrateLegacySkills_SlugCollision checks that two legacy skills
// with the same Name produce two distinct slugs (suffixed with -2).
func TestMigrateLegacySkills_SlugCollision(t *testing.T) {
	dataDir := t.TempDir()
	raw := []byte(`{"skills":[
		{"id":"id1","name":"Same","instructions":"i1"},
		{"id":"id2","name":"Same","instructions":"i2"}
	]}`)
	out, err := MigrateLegacySkills(dataDir, raw)
	if err != nil {
		t.Fatalf("migrate: %v", err)
	}
	var doc map[string]any
	_ = json.Unmarshal(out, &doc)
	skills := doc["skills"].([]any)
	slug1 := skills[0].(map[string]any)["slug"].(string)
	slug2 := skills[1].(map[string]any)["slug"].(string)
	if slug1 == slug2 {
		t.Fatalf("slugs collided: %q == %q", slug1, slug2)
	}
}

// TestMigrateLegacySkills_InstructionsAlreadyHaveFrontmatter exercises
// the most common real-world shape: the operator pasted a complete
// SKILL.md (frontmatter + body) into the legacy Instructions textarea.
// The migrator must NOT wrap that in a second frontmatter block — the
// resulting SKILL.md would otherwise have two `---` headers and be
// invalid. It also must preserve every frontmatter field the operator
// wrote (license, allowed-tools, multi-line description, etc.) and
// only rewrite the `name` to match the chosen slug.
func TestMigrateLegacySkills_InstructionsAlreadyHaveFrontmatter(t *testing.T) {
	dataDir := t.TempDir()
	canonical := "---\n" +
		"name: study-mode\n" +
		"description: >\n" +
		"  Deep study mode. Activate when Alby asks to enter \"modo estudio\",\n" +
		"  \"vamos a estudiar X\". DO NOT use when he is actively programming.\n" +
		"license: MIT\n" +
		"allowed-tools:\n" +
		"  - search_orders\n" +
		"---\n" +
		"You are Magec in Study Mode.\n\n" +
		"## Goals\n- Teach concepts\n- Stay out of production code\n"

	// Build the legacy JSON via Marshal so the embedded newlines and
	// quotes get escaped correctly without manual quoting gymnastics.
	rawDoc := map[string]any{
		"skills": []any{
			map[string]any{
				"id":           "abc-123",
				"name":         "Study Mode",
				"description":  "ignored override",
				"instructions": canonical,
			},
		},
	}
	rawBytes, err := json.Marshal(rawDoc)
	if err != nil {
		t.Fatalf("marshal raw: %v", err)
	}

	out, err := MigrateLegacySkills(dataDir, rawBytes)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var doc map[string]any
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatalf("unmarshal out: %v", err)
	}
	skills := doc["skills"].([]any)
	entry := skills[0].(map[string]any)
	if entry["slug"] != "study-mode" {
		t.Fatalf("slug = %v, want study-mode", entry["slug"])
	}

	skillMD, err := os.ReadFile(filepath.Join(dataDir, "skills", "study-mode", "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}

	body := string(skillMD)

	// 1. Exactly one opening frontmatter delimiter at the top.
	if strings.Count(body, "\n---\n") < 1 {
		t.Fatalf("expected at least one closing delimiter, got body:\n%s", body)
	}
	delimitersAtStart := 0
	if strings.HasPrefix(body, "---\n") {
		delimitersAtStart++
	}
	if strings.HasPrefix(body, "---\n---\n") {
		t.Fatalf("two stacked frontmatter blocks at the top — bug repro:\n%s", body)
	}
	if delimitersAtStart != 1 {
		t.Fatalf("expected exactly one leading frontmatter delimiter, got %d. Body:\n%s", delimitersAtStart, body)
	}

	// 2. Operator-supplied fields survived verbatim.
	for _, want := range []string{
		"license: MIT",
		"allowed-tools:",
		"  - search_orders",
		"Deep study mode. Activate when Alby",
		"## Goals",
		"You are Magec in Study Mode.",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("SKILL.md missing %q. Got:\n%s", want, body)
		}
	}

	// 3. The name is rewritten to the slug (which here matches
	//    because "study-mode" was already a valid slug, but if the
	//    legacy Name was "Study Mode" the slug came from slugify and
	//    differs from the frontmatter the operator wrote — we always
	//    pin name to slug).
	if !strings.Contains(body, "name: study-mode") {
		t.Fatalf("frontmatter name not rewritten to slug: %s", body)
	}

	// 4. Re-parsing the produced SKILL.md must succeed under ADK's
	//    own validator (the strongest guarantee we can offer).
	if _, _, err := adkParseBytes(t, skillMD); err != nil {
		t.Fatalf("ADK rejected migrated SKILL.md: %v\n%s", err, body)
	}
}

// TestMigrateLegacySkills_NameMismatchPinsToSlug covers the case where
// the operator's pasted frontmatter has a `name:` that doesn't match
// the slug we chose (e.g. legacy Name "Study Mode" produces slug
// "study-mode" but the embedded frontmatter said `name: studyMode`).
// The migrator must rewrite the name to the slug to satisfy ADK's
// "dirname == frontmatter.name" invariant.
func TestMigrateLegacySkills_NameMismatchPinsToSlug(t *testing.T) {
	dataDir := t.TempDir()
	canonical := "---\nname: studyMode\ndescription: legacy\n---\nbody\n"
	raw, _ := json.Marshal(map[string]any{
		"skills": []any{
			map[string]any{
				"id":           "x",
				"name":         "Study Mode",
				"instructions": canonical,
			},
		},
	})
	out, err := MigrateLegacySkills(dataDir, raw)
	if err != nil {
		t.Fatalf("Migrate: %v", err)
	}
	var doc map[string]any
	_ = json.Unmarshal(out, &doc)
	slug := doc["skills"].([]any)[0].(map[string]any)["slug"].(string)
	if slug != "study-mode" {
		t.Fatalf("slug = %q, want study-mode", slug)
	}
	body, err := os.ReadFile(filepath.Join(dataDir, "skills", slug, "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if !strings.Contains(string(body), "name: study-mode") {
		t.Fatalf("name not pinned to slug:\n%s", body)
	}
	if strings.Contains(string(body), "name: studyMode") {
		t.Fatalf("legacy name leaked into frontmatter:\n%s", body)
	}
}

func adkParseBytes(t *testing.T, raw []byte) (any, string, error) {
	t.Helper()
	fm, body, err := adkskill.ParseBytes(raw)
	return fm, body, err
}

// adkSkillParse keeps the migrate_test signature symmetric with the
// rest of the file's helpers — the test only cares about the error.
func adkSkillParse(raw []byte) (any, string, error) {
	fm, body, err := adkskill.ParseBytes(raw)
	return fm, body, err
}

func contains(s, sub string) bool {
	return strings.Contains(s, sub)
}
