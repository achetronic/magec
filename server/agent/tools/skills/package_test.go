// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"
)

func TestSlugify(t *testing.T) {
	cases := []struct{ in, want string }{
		{"Returns Policy", "returns-policy"},
		{"  Hello,  World! ", "hello-world"},
		{"already-good", "already-good"},
		{"---weird--", "weird"},
		{"DROP TABLE users;", "drop-table-users"},
		{"", ""},
		{"@#%@#%", ""},
		{strings.Repeat("a", 100), strings.Repeat("a", 64)},
	}
	for _, c := range cases {
		got := Slugify(c.in)
		if got != c.want {
			t.Errorf("Slugify(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestUniqueSlug(t *testing.T) {
	taken := map[string]struct{}{
		"alpha":   {},
		"alpha-2": {},
	}
	got := UniqueSlug("alpha", taken)
	if got != "alpha-3" {
		t.Errorf("UniqueSlug(alpha) = %q, want alpha-3", got)
	}
	got = UniqueSlug("beta", taken)
	if got != "beta" {
		t.Errorf("UniqueSlug(beta) = %q, want beta", got)
	}
}

// TestParsePackage_StandaloneSKILL confirms a bare SKILL.md upload is
// accepted and produces an empty resource map.
func TestParsePackage_StandaloneSKILL(t *testing.T) {
	body := []byte("---\nname: foo\ndescription: hello\n---\nthe body\n")
	pkg, err := ParsePackage("SKILL.md", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Frontmatter.Name != "foo" {
		t.Fatalf("expected name=foo, got %q", pkg.Frontmatter.Name)
	}
	if !strings.Contains(pkg.Instructions, "the body") {
		t.Fatalf("expected body, got %q", pkg.Instructions)
	}
	if len(pkg.Files) != 0 {
		t.Fatalf("expected no resources, got %v", pkg.Files)
	}
}

// TestParsePackage_RejectsInvalidFrontmatter exercises the contract
// surfaced to the operator: the admin API expects the error message to
// be human-readable and explicit. We only assert it mentions
// "frontmatter" so changes to ADK's wording don't break this test.
func TestParsePackage_RejectsInvalidFrontmatter(t *testing.T) {
	body := []byte("no frontmatter here\n")
	if _, err := ParsePackage("SKILL.md", body); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// TestParsePackage_Zip checks the happy path for a zip archive that
// includes resources spread across the three allowed sub-directories.
func TestParsePackage_Zip(t *testing.T) {
	zipBody := buildZip(t, map[string][]byte{
		"SKILL.md":            []byte("---\nname: pkg-zip\ndescription: zip\n---\nbody\n"),
		"references/note.md":  []byte("note"),
		"assets/template.txt": []byte("tpl"),
		"scripts/run.sh":      []byte("#!/bin/sh"),
	})
	pkg, err := ParsePackage("foo.zip", zipBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(pkg.Files) != 3 {
		t.Fatalf("expected 3 resources, got %d (%v)", len(pkg.Files), keys(pkg.Files))
	}
	for _, want := range []string{"references/note.md", "assets/template.txt", "scripts/run.sh"} {
		if _, ok := pkg.Files[want]; !ok {
			t.Errorf("missing resource %q", want)
		}
	}
}

// TestParsePackage_RejectsForeignTopLevel verifies we don't accept an
// archive that drops files outside the three allowed sub-trees. The
// operator gets a 400 with the bad path embedded.
func TestParsePackage_RejectsForeignTopLevel(t *testing.T) {
	zipBody := buildZip(t, map[string][]byte{
		"SKILL.md":     []byte("---\nname: foo\ndescription: x\n---\nbody\n"),
		"random/x.txt": []byte("x"),
	})
	_, err := ParsePackage("foo.zip", zipBody)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "random/x.txt") {
		t.Fatalf("error should mention offending path, got %q", err.Error())
	}
}

// TestParsePackage_StripsTopLevelDir verifies archives produced by
// "zip -r foo.zip skill-dir/" (i.e. wrapped in a single top-level
// directory) are accepted and treated as if SKILL.md were at the root.
func TestParsePackage_StripsTopLevelDir(t *testing.T) {
	zipBody := buildZip(t, map[string][]byte{
		"my-skill/SKILL.md":           []byte("---\nname: stripped\ndescription: x\n---\nbody\n"),
		"my-skill/references/note.md": []byte("note"),
	})
	pkg, err := ParsePackage("foo.zip", zipBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Frontmatter.Name != "stripped" {
		t.Fatalf("name = %q, want stripped", pkg.Frontmatter.Name)
	}
	if _, ok := pkg.Files["references/note.md"]; !ok {
		t.Fatalf("expected references/note.md, got %v", keys(pkg.Files))
	}
}

func TestParsePackage_TarGz(t *testing.T) {
	tgzBody := buildTarGz(t, map[string][]byte{
		"SKILL.md":           []byte("---\nname: tgz-pkg\ndescription: x\n---\nbody\n"),
		"references/note.md": []byte("note"),
	})
	pkg, err := ParsePackage("foo.tar.gz", tgzBody)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Frontmatter.Name != "tgz-pkg" {
		t.Fatalf("name = %q", pkg.Frontmatter.Name)
	}
	if len(pkg.Files) != 1 {
		t.Fatalf("expected 1 resource, got %d", len(pkg.Files))
	}
}

// TestWritePackageRoundTrip confirms WritePackage produces a directory
// that can be re-archived with PackageAsTarGz and re-parsed back to the
// same logical content. This is the contract the Download button on
// the UI relies on (download → re-upload → same skill).
func TestWritePackageRoundTrip(t *testing.T) {
	dir := t.TempDir()
	pkg := &PackageInfo{
		Frontmatter:  mustFrontmatter(t, []byte("---\nname: round\ndescription: trip\n---\n")),
		Instructions: "the body\n",
		Files: map[string][]byte{
			"references/r.md": []byte("ref"),
			"assets/a.txt":    []byte("ass"),
			"scripts/run.sh":  []byte("script"),
		},
	}
	if err := WritePackage(dir, "round", pkg); err != nil {
		t.Fatalf("WritePackage: %v", err)
	}

	// Confirm SKILL.md exists and contains the frontmatter+body.
	body, err := os.ReadFile(filepath.Join(dir, "round", "SKILL.md"))
	if err != nil {
		t.Fatalf("read SKILL.md: %v", err)
	}
	if !bytes.Contains(body, []byte("name: round")) || !bytes.Contains(body, []byte("the body")) {
		t.Fatalf("SKILL.md missing expected content: %s", body)
	}

	// Round-trip via tar.gz.
	tgz, err := PackageAsTarGz(dir, "round")
	if err != nil {
		t.Fatalf("PackageAsTarGz: %v", err)
	}
	parsed, err := ParsePackage("round.tar.gz", tgz)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	if parsed.Frontmatter.Name != "round" {
		t.Fatalf("name lost in round trip: %q", parsed.Frontmatter.Name)
	}
	if len(parsed.Files) != 3 {
		t.Fatalf("expected 3 resources after round trip, got %d", len(parsed.Files))
	}
}

// --- helpers ---

func buildZip(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for name, content := range files {
		f, err := w.Create(name)
		if err != nil {
			t.Fatalf("zip create %q: %v", name, err)
		}
		if _, err := f.Write(content); err != nil {
			t.Fatalf("zip write %q: %v", name, err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func buildTarGz(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for name, content := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		if err := tw.WriteHeader(hdr); err != nil {
			t.Fatalf("tar header %q: %v", name, err)
		}
		if _, err := tw.Write(content); err != nil {
			t.Fatalf("tar write %q: %v", name, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("tar close: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("gz close: %v", err)
	}
	return buf.Bytes()
}

func keys(m map[string][]byte) []string {
	out := []string{}
	for k := range m {
		out = append(out, k)
	}
	return out
}

func mustFrontmatter(t *testing.T, body []byte) *adkskill.Frontmatter {
	t.Helper()
	r := bufio.NewReader(bytes.NewReader(body))
	fm, err := adkskill.Parse(r)
	if err != nil {
		t.Fatalf("parse frontmatter: %v", err)
	}
	return fm
}

// TestParseFrontmatterPermissive_KeepsExtraKeys covers the contract
// the admin viewer relies on: a SKILL.md with non-canonical frontmatter
// keys (`version:`, `author:`, `tags:`) parses cleanly and surfaces
// every key, not just the ones ADK's strict struct knows about.
func TestParseFrontmatterPermissive_KeepsExtraKeys(t *testing.T) {
	body := []byte(`---
name: foo
description: hello
version: 2.5.1
author: Alby
tags:
  - rest
  - api
---
the body
`)
	fm, instr, ok := ParseFrontmatterPermissive(body)
	if !ok {
		t.Fatalf("expected ok=true")
	}
	if instr != "the body\n" {
		t.Errorf("instructions = %q", instr)
	}
	for _, want := range []string{"name", "description", "version", "author", "tags"} {
		if _, present := fm[want]; !present {
			t.Errorf("expected key %q in frontmatter, got keys=%v", want, mapKeys(fm))
		}
	}
	if v, _ := fm["version"].(string); v != "2.5.1" {
		t.Errorf("version = %q, want 2.5.1", v)
	}
	tags, ok := fm["tags"].([]any)
	if !ok || len(tags) != 2 {
		t.Errorf("tags = %v, want 2-element slice", fm["tags"])
	}
}

// TestParseFrontmatterPermissive_NoFrontmatterReturnsRaw covers the
// fallback contract: when the body has no `---\n...---\n` envelope at
// all, the function returns ok=false but also returns the raw text as
// the body so the UI can still render it. Without that, a malformed
// SKILL.md would render as a blank viewer.
func TestParseFrontmatterPermissive_NoFrontmatterReturnsRaw(t *testing.T) {
	body := []byte("just some markdown, no frontmatter\n")
	fm, instr, ok := ParseFrontmatterPermissive(body)
	if ok {
		t.Fatalf("expected ok=false on missing frontmatter")
	}
	if fm != nil {
		t.Errorf("expected nil frontmatter, got %v", fm)
	}
	if instr != string(body) {
		t.Errorf("body should round-trip when frontmatter is missing, got %q", instr)
	}
}

func mapKeys(m map[string]any) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
