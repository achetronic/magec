// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"sort"
	"testing"
	"testing/fstest"

	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"
)

// fakeRoot is a fstest.MapFS that mimics data/skills/, populated with
// two skills. Tests build on this fixture to verify that AgentFS only
// exposes the whitelisted entries to ADK's filesystem source.
func fakeRoot() fstest.MapFS {
	return fstest.MapFS{
		"alpha/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: alpha\ndescription: alpha skill\n---\nalpha body\n"),
		},
		"alpha/references/note.md": &fstest.MapFile{
			Data: []byte("hello"),
		},
		"beta/SKILL.md": &fstest.MapFile{
			Data: []byte("---\nname: beta\ndescription: beta skill\n---\nbeta body\n"),
		},
	}
}

func TestAgentFS_FiltersRootListing(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	entries, err := fs.ReadDir(a, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) != 1 || names[0] != "alpha" {
		t.Fatalf("expected only [alpha], got %v", names)
	}
}

func TestAgentFS_DeniesNonWhitelistedSlug(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	if _, err := a.Open("beta/SKILL.md"); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected fs.ErrNotExist for non-whitelisted skill, got %v", err)
	}
}

func TestAgentFS_AllowsWhitelistedSlugReads(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	f, err := a.Open("alpha/SKILL.md")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected SKILL.md body, got empty")
	}
}

// TestAgentFS_ListAllowsResourceWalk exercises the path used by ADK's
// fileSystemSource.ListResources: fs.WalkDir over a sub-tree must only
// see the whitelisted skill's files.
func TestAgentFS_ListAllowsResourceWalk(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	got := []string{}
	err := fs.WalkDir(a, "alpha", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			got = append(got, p)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WalkDir: %v", err)
	}
	sort.Strings(got)
	want := []string{"alpha/SKILL.md", "alpha/references/note.md"}
	if !equalStrings(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}

// TestAgentFS_EmptyWhitelist guarantees a skill-less agent sees no
// frontmatters at all (used as a fast-path: callers don't even need to
// instantiate the skilltoolset in that case, but the wrapper must still
// behave correctly if they do).
func TestAgentFS_EmptyWhitelist(t *testing.T) {
	a := NewAgentFS(fakeRoot(), nil)
	entries, err := fs.ReadDir(a, ".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty listing, got %v", entries)
	}
}

// TestAgentFS_BridgesADKFileSystemSource is the integration check: feed
// the wrapped fs to ADK's filesystem source and verify ListFrontmatters
// only surfaces the whitelisted skill. If this ever breaks, ADK has
// changed how its source walks the tree and the wrapper needs revisiting.
func TestAgentFS_BridgesADKFileSystemSource(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	src := adkskill.NewFileSystemSource(a)
	fms, err := src.ListFrontmatters(context.Background())
	if err != nil {
		t.Fatalf("ListFrontmatters: %v", err)
	}
	if len(fms) != 1 || fms[0].Name != "alpha" {
		got := []string{}
		for _, f := range fms {
			got = append(got, f.Name)
		}
		t.Fatalf("expected [alpha], got %v", got)
	}
}

// TestAgentFS_RejectsTraversal ensures path traversal attempts that
// cross the slug boundary are rejected by the underlying fs.ValidPath
// check before they reach the whitelist.
func TestAgentFS_RejectsTraversal(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	for _, bad := range []string{"alpha/../beta/SKILL.md", "/alpha/SKILL.md", "alpha//SKILL.md"} {
		if _, err := a.Open(bad); err == nil {
			t.Fatalf("expected error opening %q, got nil", bad)
		}
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestAgentFS_ReadDirReturnsEOFWhenExhausted locks in the io.ReadDirFile
// contract: with n>0, after the cursor has been exhausted, ReadDir
// must signal io.EOF so callers can detect end-of-iteration.
// Important because an earlier version returned (nil, nil), which
// looks like "no entries this batch but maybe more later" and could
// trip future consumers.
func TestAgentFS_ReadDirReturnsEOFWhenExhausted(t *testing.T) {
	a := NewAgentFS(fakeRoot(), []string{"alpha"})
	f, err := a.Open(".")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	rdr, ok := f.(interface {
		ReadDir(int) ([]fs.DirEntry, error)
	})
	if !ok {
		t.Fatalf("filteredRoot does not implement ReadDir(int)")
	}
	first, err := rdr.ReadDir(10)
	if err != nil {
		t.Fatalf("first ReadDir: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("first batch len = %d", len(first))
	}
	// The underlying cursor is now at end. A subsequent positive-n
	// call must return (nil, io.EOF).
	second, err := rdr.ReadDir(10)
	if err == nil {
		t.Fatalf("expected io.EOF after exhaustion, got err=nil len=%d", len(second))
	}
	if !errors.Is(err, io.EOF) {
		t.Fatalf("expected io.EOF, got %v", err)
	}
	if second != nil {
		t.Errorf("expected nil entries on EOF, got %v", second)
	}
}
