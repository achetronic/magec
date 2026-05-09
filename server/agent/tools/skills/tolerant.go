// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package skills

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"

	adkskill "google.golang.org/adk/tool/skilltoolset/skill"
	"gopkg.in/yaml.v3"
)

// TolerantSource wraps another skill.Source and prevents a single
// malformed SKILL.md (typically an extra non-canonical frontmatter
// field like `version:`, `author:`, `tags:`) from poisoning the
// entire toolset.
//
// Background: ADK's FileSystemSource parses SKILL.md frontmatter with
// `decoder.KnownFields(true)`, which means any key ADK doesn't know
// about returns an error. ListFrontmatters then fails the whole call,
// and `skilltoolset.ProcessRequest` propagates that error out of the
// LLM request — which the user perceives as "the agent didn't answer".
//
// We can't change ADK's strict parser without forking the library, but
// we can fall back to a permissive parse when the strict one fails,
// strip the unknown keys, and feed the cleaned bytes to ADK's parser
// instead. From the LLM's perspective the skill loses its custom
// fields (which it was never going to see anyway — `version:` doesn't
// affect behaviour) but keeps name/description/instructions and stays
// callable.
//
// All other Source methods delegate verbatim to the inner source. The
// permissive fallback only triggers in the discovery path; once the
// LLM calls `load_skill` for a specific skill, the underlying
// FileSystemSource's LoadInstructions reads bytes directly without
// going through frontmatter validation, so the skill body is served
// verbatim.
type TolerantSource struct {
	inner adkskill.Source
	// rawSource is the raw fs.FS used by the underlying FileSystemSource
	// (typically per-agent, see AgentFS). When we need to fall back to
	// the permissive path we re-read SKILL.md ourselves through this
	// FS so the wrapper doesn't have to know about file system details.
	rawSource fs.FS
}

// NewTolerantSource builds a Source proxy. fsRoot is the same fs.FS
// you would pass to skill.NewFileSystemSource — the wrapper uses it
// directly to read SKILL.md when the strict parser rejects it.
func NewTolerantSource(fsRoot fs.FS) *TolerantSource {
	return &TolerantSource{
		inner:     adkskill.NewFileSystemSource(fsRoot),
		rawSource: fsRoot,
	}
}

// ListFrontmatters delegates to the strict source first; on failure
// it walks every immediate sub-directory ourselves, parsing each
// SKILL.md tolerantly. Skills whose frontmatter cannot be salvaged
// even after stripping unknown keys are skipped with a warn-level log
// — they simply don't appear in the LLM's skill catalogue.
func (s *TolerantSource) ListFrontmatters(ctx context.Context) ([]*adkskill.Frontmatter, error) {
	if fms, err := s.inner.ListFrontmatters(ctx); err == nil {
		return fms, nil
	}

	// Fallback: enumerate skills ourselves.
	entries, err := fs.ReadDir(s.rawSource, ".")
	if err != nil {
		return nil, fmt.Errorf("list skills root: %w", err)
	}
	var out []*adkskill.Frontmatter
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		fm, ok := s.tolerantFrontmatter(e.Name())
		if !ok {
			continue
		}
		out = append(out, fm)
	}
	return out, nil
}

// LoadFrontmatter mirrors ListFrontmatters' two-step approach for the
// single-skill case. The LLM hits this when it calls `load_skill`,
// so it must keep working even for skills with non-canonical
// frontmatter.
func (s *TolerantSource) LoadFrontmatter(ctx context.Context, name string) (*adkskill.Frontmatter, error) {
	if fm, err := s.inner.LoadFrontmatter(ctx, name); err == nil {
		return fm, nil
	}
	fm, ok := s.tolerantFrontmatter(name)
	if !ok {
		return nil, fmt.Errorf("%w: %q", adkskill.ErrSkillNotFound, name)
	}
	return fm, nil
}

// LoadInstructions reads bytes directly from disk through the inner
// source. ADK's FileSystemSource only validates the frontmatter (to
// confirm the directory name matches `frontmatter.name`) — when that
// validation rejects an otherwise-fine SKILL.md because of extra
// keys, we fall back to splitting the file ourselves and returning
// the body verbatim.
func (s *TolerantSource) LoadInstructions(ctx context.Context, name string) (string, error) {
	if body, err := s.inner.LoadInstructions(ctx, name); err == nil {
		return body, nil
	}
	body, ok := s.readSkillBody(name)
	if !ok {
		return "", fmt.Errorf("%w: %q", adkskill.ErrSkillNotFound, name)
	}
	return body, nil
}

// LoadResource and ListResources have nothing to do with frontmatter
// validation, so we always delegate to the strict source. If those
// methods need their own fallback in the future, this is the place
// to add it.
func (s *TolerantSource) LoadResource(ctx context.Context, name, resourcePath string) (io.ReadCloser, error) {
	return s.inner.LoadResource(ctx, name, resourcePath)
}

func (s *TolerantSource) ListResources(ctx context.Context, name, subpath string) ([]string, error) {
	return s.inner.ListResources(ctx, name, subpath)
}

// tolerantFrontmatter reads a single skill directory, strips unknown
// frontmatter keys, and returns ADK's parsed Frontmatter. ok=false
// means we couldn't recover the skill (missing SKILL.md, totally
// malformed YAML, etc.) and the caller should leave it out.
func (s *TolerantSource) tolerantFrontmatter(name string) (*adkskill.Frontmatter, bool) {
	body, err := fs.ReadFile(s.rawSource, name+"/SKILL.md")
	if err != nil {
		return nil, false
	}
	fm, _, err := tolerantParse(body, name)
	if err != nil {
		slog.Warn("skill discarded: invalid frontmatter even after permissive parse",
			"skill", name, "error", err)
		return nil, false
	}
	return fm, true
}

// readSkillBody returns the markdown body of a skill (everything
// after the closing frontmatter delimiter) without going through
// ADK's strict validator.
func (s *TolerantSource) readSkillBody(name string) (string, bool) {
	body, err := fs.ReadFile(s.rawSource, name+"/SKILL.md")
	if err != nil {
		return "", false
	}
	_, instructions, err := tolerantParse(body, name)
	if err != nil {
		return "", false
	}
	return instructions, true
}

// tolerantParse strips non-canonical YAML keys from the SKILL.md
// frontmatter, then feeds the cleaned bytes to ADK's strict parser.
// expectedName is used to verify the dirname/frontmatter-name
// invariant ADK relies on. We keep the original body verbatim.
func tolerantParse(raw []byte, expectedName string) (*adkskill.Frontmatter, string, error) {
	rawFM, body, ok := splitFrontmatter(raw)
	if !ok {
		return nil, "", fmt.Errorf("missing frontmatter delimiters")
	}
	cleaned := pruneUnknownFrontmatterKeys(rawFM)
	rebuilt := assembleSkillMD(cleaned, body)
	fm, _, err := adkskill.ParseBytes(rebuilt)
	if err != nil {
		// Try once more after forcing the name to match the
		// directory — some operators write `name: Foo Bar` in the
		// frontmatter but the directory is `foo-bar` because the
		// migrator slugified it.
		forced := forceFrontmatterName(rawFM, expectedName)
		cleaned = pruneUnknownFrontmatterKeys(forced)
		rebuilt = assembleSkillMD(cleaned, body)
		fm2, _, err2 := adkskill.ParseBytes(rebuilt)
		if err2 != nil {
			return nil, "", err
		}
		return fm2, body, nil
	}
	return fm, body, nil
}

// forceFrontmatterName replaces the YAML `name:` key in raw with
// expected, dropping any pre-existing value. Round-tripping through
// the YAML encoder is fine here because the resulting bytes are only
// fed to ADK's parser, never written to disk.
func forceFrontmatterName(raw []byte, expected string) []byte {
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return raw
	}
	doc["name"] = expected
	out, err := yaml.Marshal(doc)
	if err != nil {
		return raw
	}
	return out
}

// Compile-time check: TolerantSource satisfies adkskill.Source.
var _ adkskill.Source = (*TolerantSource)(nil)

// (Avoid an "imported and not used" complaint when only some helpers
// reference these packages — the compiler will tell us if anything
// drifts.)
var _ = bytes.NewReader
