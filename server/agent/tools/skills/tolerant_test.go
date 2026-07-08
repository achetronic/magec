// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package skills

import (
	"context"
	"strings"
	"testing"
	"testing/fstest"
)

// humanizerSKILL is a minimal real-world reproduction of the SKILL.md
// that broke the runtime: a `version:` field plus other extras ADK's
// strict KnownFields parser rejects.
const humanizerSKILL = "---\n" +
	"name: humanizer\n" +
	"version: 2.5.1\n" +
	"description: Remove signs of AI-generated writing.\n" +
	"license: MIT\n" +
	"compatibility: claude-code opencode\n" +
	"allowed-tools:\n" +
	"  - Read\n" +
	"  - Write\n" +
	"---\n" +
	"# Humanizer\n\nBody goes here.\n"

// TestTolerantSource_HumanizerCase locks in the symptom the operator
// reported: a SKILL.md with extra non-canonical keys must remain
// visible to ListFrontmatters and LoadFrontmatter so the agent can
// keep responding.
func TestTolerantSource_HumanizerCase(t *testing.T) {
	root := fstest.MapFS{
		"humanizer/SKILL.md": &fstest.MapFile{Data: []byte(humanizerSKILL)},
	}
	src := NewTolerantSource(root)
	ctx := context.Background()

	fms, err := src.ListFrontmatters(ctx)
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	if len(fms) != 1 {
		t.Fatalf("expected 1 frontmatter, got %d", len(fms))
	}
	if fms[0].Name != "humanizer" {
		t.Errorf("name = %q, want humanizer", fms[0].Name)
	}
	if fms[0].Description == "" {
		t.Errorf("description was lost")
	}
	if fms[0].License != "MIT" {
		t.Errorf("license = %q, want MIT", fms[0].License)
	}

	fm, err := src.LoadFrontmatter(ctx, "humanizer")
	if err != nil {
		t.Fatalf("LoadFrontmatter: %v", err)
	}
	if fm.Name != "humanizer" {
		t.Fatalf("LoadFrontmatter name = %q", fm.Name)
	}

	body, err := src.LoadInstructions(ctx, "humanizer")
	if err != nil {
		t.Fatalf("LoadInstructions: %v", err)
	}
	if !strings.Contains(body, "# Humanizer") {
		t.Errorf("body missing heading: %q", body)
	}
}

// TestTolerantSource_StrictPathStillWorks confirms the wrapper does
// not regress the happy path: a fully canonical SKILL.md goes through
// ADK's strict source verbatim and lands in the frontmatter intact.
func TestTolerantSource_StrictPathStillWorks(t *testing.T) {
	canonical := "---\nname: clean\ndescription: ok\n---\nbody\n"
	root := fstest.MapFS{
		"clean/SKILL.md": &fstest.MapFile{Data: []byte(canonical)},
	}
	src := NewTolerantSource(root)

	fms, err := src.ListFrontmatters(context.Background())
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	if len(fms) != 1 || fms[0].Name != "clean" {
		t.Fatalf("unexpected: %+v", fms)
	}
}

// TestTolerantSource_DropsHopelesslyBrokenSkills makes sure we don't
// confidently surface skills whose frontmatter cannot be parsed at
// all. Those should be skipped (with a warn-level log we don't assert
// on here) rather than break the whole listing.
func TestTolerantSource_DropsHopelesslyBrokenSkills(t *testing.T) {
	root := fstest.MapFS{
		"good/SKILL.md":   &fstest.MapFile{Data: []byte("---\nname: good\ndescription: ok\n---\nbody\n")},
		"broken/SKILL.md": &fstest.MapFile{Data: []byte("not a valid frontmatter at all")},
	}
	src := NewTolerantSource(root)
	fms, err := src.ListFrontmatters(context.Background())
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	if len(fms) != 1 || fms[0].Name != "good" {
		names := []string{}
		for _, f := range fms {
			names = append(names, f.Name)
		}
		t.Fatalf("expected only [good], got %v", names)
	}
}

// TestTolerantSource_NameMismatchPinsToDir covers the case where the
// frontmatter `name:` does not satisfy ADK's slug validator (e.g.
// "Foo Bar") but the directory IS a valid slug. We force the name to
// the directory before re-parsing so the skill stays usable.
func TestTolerantSource_NameMismatchPinsToDir(t *testing.T) {
	body := "---\nname: Foo Bar\nversion: 1\ndescription: x\n---\nbody\n"
	root := fstest.MapFS{
		"foo-bar/SKILL.md": &fstest.MapFile{Data: []byte(body)},
	}
	src := NewTolerantSource(root)
	fms, err := src.ListFrontmatters(context.Background())
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	if len(fms) != 1 || fms[0].Name != "foo-bar" {
		t.Fatalf("expected [foo-bar], got %+v", fms)
	}
}
