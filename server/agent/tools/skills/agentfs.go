// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

// Package skills wires Magec's on-disk skill packages
// (data/skills/{slug}/SKILL.md plus optional references/, assets/, scripts/
// subtrees) into ADK's skilltoolset. The bridge is intentionally thin: ADK
// owns the SKILL.md format, frontmatter validation, and resource loading
// via skill.NewFileSystemSource. This package only contributes a per-agent
// fs.FS wrapper that exposes a whitelist of skill slugs to a single agent
// — so an agent's list_skills/load_skill calls only see the skills the
// operator linked to that agent in the store.
//
// Design notes:
//
//   - We do NOT implement a custom skill.Source. That would duplicate
//     ADK's frontmatter parsing and reintroduce drift the moment ADK
//     adds a field. Whatever ADK reads from SKILL.md is what Magec
//     surfaces to the LLM, period.
//   - The whitelist is a slice of slugs (NOT skill IDs). Slugs are the
//     directory names on disk and the names ADK exposes to the LLM.
//     IDs live in the store and are translated to slugs once at
//     agent-build time by the caller.
//   - The wrapper rejects any path whose first segment is not in the
//     whitelist. fs.WalkDir / fs.ReadDir on the root only see the
//     whitelisted entries. This is the smallest surface that prevents
//     ADK's filesystem source from leaking other skills' frontmatters
//     through ListFrontmatters.
package skills

import (
	"errors"
	"io/fs"
	"path"
	"strings"
	"time"
)

// AgentFS wraps a root fs.FS (typically os.DirFS(data/skills)) and only
// exposes a fixed list of skill slugs to its consumer. It implements
// io/fs.FS and io/fs.ReadDirFS so it satisfies every interface ADK's
// skill.NewFileSystemSource exercises.
//
// Reads outside the whitelist return fs.ErrNotExist — never permission
// errors — so ADK's source treats them as missing and skips them. This
// is the correct semantics: from the agent's perspective, those skills
// simply do not exist.
type AgentFS struct {
	root    fs.FS
	allowed map[string]struct{}
}

// NewAgentFS returns an fs.FS that exposes only the directories named by
// allowedSlugs. Empty/duplicate entries in allowedSlugs are ignored. The
// returned filesystem is safe for concurrent use as long as the
// underlying root is.
func NewAgentFS(root fs.FS, allowedSlugs []string) *AgentFS {
	a := &AgentFS{
		root:    root,
		allowed: make(map[string]struct{}, len(allowedSlugs)),
	}
	for _, s := range allowedSlugs {
		if s == "" {
			continue
		}
		a.allowed[s] = struct{}{}
	}
	return a
}

// firstSegment returns the first path component of name, or "" if name
// is "." (the root). Paths in fs.FS are slash-separated and never start
// with a leading slash, per the io/fs contract.
func firstSegment(name string) string {
	clean := path.Clean(name)
	if clean == "." || clean == "" {
		return ""
	}
	if i := strings.Index(clean, "/"); i >= 0 {
		return clean[:i]
	}
	return clean
}

// allows reports whether name is reachable through this filesystem.
// The root itself is always allowed; everything else must descend into
// a whitelisted slug.
func (a *AgentFS) allows(name string) bool {
	seg := firstSegment(name)
	if seg == "" {
		return true
	}
	_, ok := a.allowed[seg]
	return ok
}

// Open implements fs.FS. It enforces the whitelist before delegating to
// the underlying root. Listing the root returns a synthetic directory
// entry (filteredRootDir) so ReadDir on "." only surfaces whitelisted
// slugs even when the underlying filesystem has more.
func (a *AgentFS) Open(name string) (fs.File, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return &filteredRoot{fs: a}, nil
	}
	if !a.allows(name) {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrNotExist}
	}
	return a.root.Open(name)
}

// ReadDir implements fs.ReadDirFS. ADK's fileSystemSource calls
// fs.ReadDir(filesystem, ".") to enumerate skills, so we must filter
// the root listing here too. Non-root listings only succeed when the
// path lives under an allowed slug — the underlying root handles those.
func (a *AgentFS) ReadDir(name string) ([]fs.DirEntry, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return a.rootEntries()
	}
	if !a.allows(name) {
		return nil, &fs.PathError{Op: "readdir", Path: name, Err: fs.ErrNotExist}
	}
	return fs.ReadDir(a.root, name)
}

// Stat implements fs.StatFS. ADK's source uses fs.Stat to confirm that
// resource directories exist before walking; we honour the whitelist
// on top of the underlying stat.
func (a *AgentFS) Stat(name string) (fs.FileInfo, error) {
	if !fs.ValidPath(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrInvalid}
	}
	if name == "." {
		return rootStat{}, nil
	}
	if !a.allows(name) {
		return nil, &fs.PathError{Op: "stat", Path: name, Err: fs.ErrNotExist}
	}
	return fs.Stat(a.root, name)
}

// Sub implements fs.SubFS. Used by ADK's source via fs.Sub to scope
// resource walks. We delegate to the underlying root after the
// whitelist check so the returned sub-FS is the natural one for the
// allowed slug.
func (a *AgentFS) Sub(dir string) (fs.FS, error) {
	if !fs.ValidPath(dir) {
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrInvalid}
	}
	if dir == "." {
		return a, nil
	}
	if !a.allows(dir) {
		return nil, &fs.PathError{Op: "sub", Path: dir, Err: fs.ErrNotExist}
	}
	return fs.Sub(a.root, dir)
}

// rootEntries returns the filtered listing of the root directory: only
// entries whose name is in the whitelist AND that exist on the
// underlying root are reported. We do not error out if the underlying
// root cannot be listed at all (e.g. data/skills/ does not exist yet)
// — instead we return an empty listing so a freshly-installed Magec
// with no skills configured behaves the same as one whose skills
// directory simply has no entries.
func (a *AgentFS) rootEntries() ([]fs.DirEntry, error) {
	if len(a.allowed) == 0 {
		return nil, nil
	}
	rootEntries, err := fs.ReadDir(a.root, ".")
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var out []fs.DirEntry
	for _, e := range rootEntries {
		if _, ok := a.allowed[e.Name()]; ok {
			out = append(out, e)
		}
	}
	return out, nil
}

// filteredRoot is the fs.File representation of "." returned by Open.
// It only needs to satisfy fs.ReadDirFile — that's all ADK reaches for
// when enumerating skills via the high-level fs.ReadDir helper.
type filteredRoot struct {
	fs     *AgentFS
	cursor int
	cached []fs.DirEntry
}

func (r *filteredRoot) Stat() (fs.FileInfo, error) { return rootStat{}, nil }
func (r *filteredRoot) Read([]byte) (int, error) {
	return 0, &fs.PathError{Op: "read", Path: ".", Err: fs.ErrInvalid}
}
func (r *filteredRoot) Close() error { return nil }

// ReadDir implements fs.ReadDirFile. n<=0 returns all remaining entries;
// n>0 returns up to n entries and io.EOF when exhausted. Following the
// stdlib convention exactly avoids surprising the caller (ADK uses the
// helper fs.ReadDir which always asks for everything in one call).
func (r *filteredRoot) ReadDir(n int) ([]fs.DirEntry, error) {
	if r.cached == nil {
		entries, err := r.fs.rootEntries()
		if err != nil {
			return nil, err
		}
		r.cached = entries
	}
	remaining := r.cached[r.cursor:]
	if n <= 0 {
		r.cursor = len(r.cached)
		return remaining, nil
	}
	if n > len(remaining) {
		n = len(remaining)
	}
	r.cursor += n
	return remaining[:n], nil
}

// rootStat satisfies fs.FileInfo for the synthetic root directory. The
// concrete values are uninteresting to ADK — it only checks IsDir.
type rootStat struct{}

func (rootStat) Name() string       { return "." }
func (rootStat) Size() int64        { return 0 }
func (rootStat) Mode() fs.FileMode  { return fs.ModeDir | 0o555 }
func (rootStat) ModTime() time.Time { return time.Time{} }
func (rootStat) IsDir() bool        { return true }
func (rootStat) Sys() any           { return nil }
