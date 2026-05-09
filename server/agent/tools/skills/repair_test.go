// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	adkskill "google.golang.org/adk/tool/skilltoolset/skill"
)

// TestRepairBrokenSkills_StackedFrontmatter reproduces the exact bug
// the buggy migrator left in production: a SKILL.md file whose body
// is itself a complete SKILL.md, leading to two `---` blocks at the
// top. The repair must drop the synthesised wrapper and pin the inner
// frontmatter's `name` to the directory slug.
func TestRepairBrokenSkills_StackedFrontmatter(t *testing.T) {
	dataDir := t.TempDir()
	skillDir := filepath.Join(dataDir, "skills", "study-mode")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	stacked := "---\n" +
		"name: study-mode\n" +
		"description: Study Mode\n" +
		"---\n" +
		"---\n" +
		"name: study-mode\n" +
		"description: >\n" +
		"  Deep study mode. Activate when Alby asks to enter\n" +
		"  \"modo estudio\" or similar.\n" +
		"license: MIT\n" +
		"---\n" +
		"You are Magec in Study Mode.\n"
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(stacked), 0o644); err != nil {
		t.Fatalf("write SKILL.md: %v", err)
	}

	if err := RepairBrokenSkills(dataDir); err != nil {
		t.Fatalf("RepairBrokenSkills: %v", err)
	}

	body, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	got := string(body)

	// Exactly one frontmatter block at the top.
	if strings.HasPrefix(got, "---\n---\n") {
		t.Fatalf("stacked frontmatter not healed:\n%s", got)
	}
	if !strings.HasPrefix(got, "---\n") {
		t.Fatalf("missing leading frontmatter delimiter:\n%s", got)
	}
	for _, want := range []string{
		"name: study-mode",
		"description: >",
		"Deep study mode. Activate when Alby",
		"license: MIT",
		"You are Magec in Study Mode.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("repaired SKILL.md missing %q. Got:\n%s", want, got)
		}
	}
	// ADK validator must accept the result.
	if _, _, err := adkskill.ParseBytes(body); err != nil {
		t.Fatalf("ADK rejected repaired SKILL.md: %v\n%s", err, got)
	}
}

// TestRepairBrokenSkills_NoOpOnGoodSkill confirms we do not touch a
// SKILL.md that is already correct.
func TestRepairBrokenSkills_NoOpOnGoodSkill(t *testing.T) {
	dataDir := t.TempDir()
	skillDir := filepath.Join(dataDir, "skills", "good")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	good := "---\nname: good\ndescription: ok\n---\nbody\n"
	skillPath := filepath.Join(skillDir, "SKILL.md")
	if err := os.WriteFile(skillPath, []byte(good), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
	beforeStat, _ := os.Stat(skillPath)
	beforeMod := beforeStat.ModTime()

	if err := RepairBrokenSkills(dataDir); err != nil {
		t.Fatalf("RepairBrokenSkills: %v", err)
	}

	body, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(body) != good {
		t.Fatalf("good SKILL.md was modified:\n%s", body)
	}
	afterStat, _ := os.Stat(skillPath)
	if !afterStat.ModTime().Equal(beforeMod) {
		t.Errorf("good SKILL.md was rewritten on disk (mtime changed)")
	}
}

// TestRepairBrokenSkills_MisplacedResources covers the "everything got
// dumped into references/" pathology. Files that name a canonical
// sub-tree (references/foo, assets/bar, scripts/baz) under references/
// must be lifted up to their intended place.
func TestRepairBrokenSkills_MisplacedResources(t *testing.T) {
	dataDir := t.TempDir()
	skillDir := filepath.Join(dataDir, "skills", "mixed")
	mustWrite(t, filepath.Join(skillDir, "SKILL.md"),
		"---\nname: mixed\ndescription: x\n---\nbody\n")
	mustWrite(t, filepath.Join(skillDir, "references", "actual-ref.md"), "stay")
	mustWrite(t, filepath.Join(skillDir, "references", "references", "deep.md"), "ref")
	mustWrite(t, filepath.Join(skillDir, "references", "assets", "tpl.txt"), "asset")
	mustWrite(t, filepath.Join(skillDir, "references", "scripts", "go.sh"), "script")

	if err := RepairBrokenSkills(dataDir); err != nil {
		t.Fatalf("RepairBrokenSkills: %v", err)
	}

	mustExist(t, filepath.Join(skillDir, "references", "actual-ref.md"))
	mustExist(t, filepath.Join(skillDir, "references", "deep.md"))
	mustExist(t, filepath.Join(skillDir, "assets", "tpl.txt"))
	mustExist(t, filepath.Join(skillDir, "scripts", "go.sh"))

	// The mis-rooted intermediate dirs must be gone.
	for _, gone := range []string{
		filepath.Join(skillDir, "references", "references"),
		filepath.Join(skillDir, "references", "assets"),
		filepath.Join(skillDir, "references", "scripts"),
	} {
		if _, err := os.Stat(gone); err == nil {
			t.Errorf("expected %s to be removed, still present", gone)
		}
	}
}

// TestRepairBrokenSkills_Idempotent covers a guarantee the operator
// relies on: running the repair twice is exactly equivalent to running
// it once. Especially important because we call it on every store
// load.
func TestRepairBrokenSkills_Idempotent(t *testing.T) {
	dataDir := t.TempDir()
	skillDir := filepath.Join(dataDir, "skills", "idem")
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	stacked := "---\nname: idem\ndescription: x\n---\n---\nname: idem\ndescription: y\n---\nbody\n"
	skillPath := filepath.Join(skillDir, "SKILL.md")
	mustWrite(t, skillPath, stacked)

	if err := RepairBrokenSkills(dataDir); err != nil {
		t.Fatalf("first repair: %v", err)
	}
	first, _ := os.ReadFile(skillPath)

	if err := RepairBrokenSkills(dataDir); err != nil {
		t.Fatalf("second repair: %v", err)
	}
	second, _ := os.ReadFile(skillPath)
	if string(first) != string(second) {
		t.Fatalf("repair not idempotent:\nfirst : %s\nsecond: %s", first, second)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Errorf("expected %s to exist: %v", path, err)
	}
}
