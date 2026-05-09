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
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	adkskill "google.golang.org/adk/tool/skilltoolset/skill"
)

// MigrateLegacySkills rewrites the skills section of a raw store.json
// payload from the v0.x layout (Skill struct carrying Name, Description,
// Instructions, References + flat data/skills/{ID}/{filename}) to the
// new on-disk layout (data/skills/{slug}/SKILL.md + references/, with
// the store keeping only {id, slug}).
//
// The function is idempotent: skills that already only have {id, slug}
// (and an existing data/skills/{slug}/SKILL.md) are left alone.
//
// dataDir is the absolute path containing store.json (i.e. the parent
// of "skills/"). Pass an empty string to skip filesystem operations
// (the function will still strip legacy JSON fields, useful for tests).
//
// Returns the rewritten JSON. If parsing fails the original bytes are
// returned unchanged with a non-nil error so the caller can decide
// whether to abort or fall back.
//
// TODO(v0.X): remove this migrator after at least two releases. Tracked
// in .agents/TODO.md under "Low Priority".
func MigrateLegacySkills(dataDir string, raw []byte) ([]byte, error) {
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		return raw, err
	}

	skillsAny, ok := doc["skills"].([]any)
	if !ok || len(skillsAny) == 0 {
		return raw, nil
	}

	taken := collectExistingSlugs(skillsAny)
	rewrote := false

	for i, item := range skillsAny {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		legacy, hasLegacy := readLegacyFields(entry)
		hasSlug := stringField(entry, "slug") != ""
		if !hasLegacy && hasSlug {
			// Already migrated. Strip any vestigial legacy keys
			// (defensive — a partial migration could leave them).
			if stripLegacyKeys(entry) {
				rewrote = true
			}
			continue
		}

		id := stringField(entry, "id")
		if id == "" {
			// Synthesize a deterministic id placeholder so the
			// caller can detect and recover; without an id we have
			// no stable reference to migrate around.
			slog.Warn("skills migrator: skill without id, leaving as-is")
			continue
		}

		slug := chooseSlug(legacy, taken)
		taken[slug] = struct{}{}

		if dataDir != "" {
			if err := materialiseLegacy(dataDir, id, slug, legacy); err != nil {
				return raw, fmt.Errorf("materialise skill %s: %w", id, err)
			}
		}

		// Replace the entry with the minimal new shape.
		skillsAny[i] = map[string]any{
			"id":   id,
			"slug": slug,
		}
		rewrote = true
		slog.Info("skills migrator: migrated skill",
			"id", id, "slug", slug,
			"references", len(legacy.References),
			"hasInstructions", legacy.Instructions != "")
	}

	if !rewrote {
		return raw, nil
	}
	doc["skills"] = skillsAny
	out, err := json.Marshal(doc)
	if err != nil {
		return raw, err
	}
	return out, nil
}

// legacySkill mirrors the v0.x on-disk JSON shape. Defined locally so
// the migrator stays a pure function and does not depend on store types
// (which already moved to the new shape).
type legacySkill struct {
	Name         string
	Description  string
	Instructions string
	References   []string
}

func readLegacyFields(entry map[string]any) (legacySkill, bool) {
	var l legacySkill
	hasAny := false
	if v, ok := entry["name"].(string); ok && v != "" {
		l.Name = v
		hasAny = true
	}
	if v, ok := entry["description"].(string); ok && v != "" {
		l.Description = v
		hasAny = true
	}
	if v, ok := entry["instructions"].(string); ok && v != "" {
		l.Instructions = v
		hasAny = true
	}
	if refs, ok := entry["references"].([]any); ok {
		for _, r := range refs {
			if rm, ok := r.(map[string]any); ok {
				if fn, ok := rm["filename"].(string); ok && fn != "" {
					l.References = append(l.References, fn)
				}
			}
		}
		if len(l.References) > 0 {
			hasAny = true
		}
	}
	return l, hasAny
}

func stripLegacyKeys(entry map[string]any) bool {
	changed := false
	for _, k := range []string{"name", "description", "instructions", "references"} {
		if _, ok := entry[k]; ok {
			delete(entry, k)
			changed = true
		}
	}
	return changed
}

func collectExistingSlugs(entries []any) map[string]struct{} {
	out := map[string]struct{}{}
	for _, e := range entries {
		if m, ok := e.(map[string]any); ok {
			if s, ok := m["slug"].(string); ok && s != "" {
				out[s] = struct{}{}
			}
		}
	}
	return out
}

func chooseSlug(l legacySkill, taken map[string]struct{}) string {
	base := Slugify(l.Name)
	if base == "" {
		base = "skill"
	}
	return UniqueSlug(base, taken)
}

func stringField(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// materialiseLegacy moves the legacy on-disk files (data/skills/{id}/*)
// under data/skills/{slug}/references/ and writes a fresh SKILL.md
// reconstructed from the legacy Name/Description/Instructions fields.
// The frontmatter description is non-optional in ADK, so we fall back
// to a sensible default when the operator left it empty.
func materialiseLegacy(dataDir, id, slug string, l legacySkill) error {
	skillsRoot := filepath.Join(dataDir, "skills")
	legacyDir := filepath.Join(skillsRoot, id)
	newDir := filepath.Join(skillsRoot, slug)

	if _, err := os.Stat(filepath.Join(newDir, "SKILL.md")); err == nil {
		// Already migrated on disk; do not clobber.
		return nil
	}

	if err := os.MkdirAll(newDir, 0o755); err != nil {
		return fmt.Errorf("mkdir skill dir: %w", err)
	}

	// Move legacy reference files into the new package. The old UI
	// only labelled files as "references", but operators sometimes
	// uploaded archives whose internal layout already named the
	// canonical sub-directories (references/, assets/, scripts/).
	// We honour those prefixes when present and default to
	// references/ for everything else.
	//
	// Skip when the legacy dir IS the new dir (the slug happens to
	// equal the ID, e.g. tests that use UUID-shaped names): there is
	// nothing to move and we'd otherwise try to walk and unlink the
	// directory we just created.
	if legacyDir != newDir {
		if _, err := os.Stat(legacyDir); err == nil {
			if err := moveLegacyReferences(legacyDir, newDir, l.References); err != nil {
				return err
			}
			// Remove the legacy directory if it is now empty. Best
			// effort: a non-empty leftover means we left an unknown
			// file behind, which is preferable to deleting it blindly.
			_ = os.Remove(legacyDir)
		}
	}

	// Reconstruct SKILL.md. Two cases:
	//
	//   (1) Legacy Instructions already IS a full SKILL.md (frontmatter
	//       + body). The previous UI used to paste canonical skill
	//       packages straight into the Instructions textarea, so this
	//       is the common shape in real-world stores. We must NOT wrap
	//       it in a second frontmatter block — that would produce an
	//       invalid SKILL.md (two `---` headers) and lose every field
	//       the operator carefully wrote (license, allowed-tools,
	//       multi-line description, etc.).
	//
	//   (2) Legacy Instructions is plain text. We synthesise a minimal
	//       frontmatter from the legacy Name/Description fields.
	//
	// In case (1) we rewrite the frontmatter `name` to the slug we
	// chose (legacy Names that included spaces or capitals would not
	// satisfy ADK's slug validator, but the body's frontmatter still
	// uses the human-readable form). Description, license, etc. are
	// preserved verbatim.
	body, err := buildLegacySkillMD(slug, l)
	if err != nil {
		return fmt.Errorf("build SKILL.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(newDir, "SKILL.md"), body, 0o644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}
	return nil
}

// buildLegacySkillMD turns a legacySkill into the bytes of a valid
// SKILL.md. If l.Instructions already starts with a YAML frontmatter
// block, that frontmatter is reused (with `name` replaced by the
// migration slug) and the original body kept. Otherwise the legacy
// Name/Description/Instructions are stitched into a fresh minimal
// SKILL.md.
//
// We do NOT round-trip the parsed frontmatter through yaml.Marshal:
// that would re-quote multi-line strings, drop comments, and reorder
// keys. Instead we splice the new `name:` line into the raw frontmatter
// block by string surgery — the operator's formatting survives intact.
//
// Important: the validation step happens AFTER we rewrite the name.
// Real-world stores often contain frontmatters whose `name` field is
// not a valid ADK slug (e.g. "Study Mode", "studyMode") — those would
// fail ADK's validator if we tried to parse the raw bytes first.
// Rewriting the name to the chosen slug first gives the validator a
// fighting chance, and any remaining failure (truly malformed YAML,
// missing description, etc.) falls through to the wrap path.
func buildLegacySkillMD(slug string, l legacySkill) ([]byte, error) {
	if hasFrontmatterPrefix(l.Instructions) {
		patched := rewriteFrontmatterName([]byte(l.Instructions), slug)
		if _, _, err := adkskill.ParseBytes(patched); err == nil {
			return patched, nil
		}
		// Fall through: even with the slug forced into place, the
		// frontmatter is unparseable (broken YAML, missing
		// description, etc.). We treat the whole blob as a body so
		// at least the operator's prose survives — this is rare in
		// practice and the operator can re-upload a clean package.
	}

	desc := l.Description
	if desc == "" {
		desc = l.Name
	}
	if desc == "" {
		desc = "Migrated skill"
	}
	fm := &adkskill.Frontmatter{
		Name:        slug,
		Description: desc,
	}
	return adkskill.Build(fm, l.Instructions)
}

// hasFrontmatterPrefix reports whether the input begins with a YAML
// frontmatter delimiter line. We accept leading whitespace because
// some legacy operators pasted skills with a blank line on top.
func hasFrontmatterPrefix(s string) bool {
	trimmed := strings.TrimLeft(s, " \t\r\n")
	return strings.HasPrefix(trimmed, "---\n") || strings.HasPrefix(trimmed, "---\r\n")
}

// rewriteFrontmatterName replaces the value of the top-level `name:`
// key in the document's leading YAML frontmatter block with the given
// slug. It works on raw bytes — no YAML round-trip — so every other
// formatting detail (multi-line scalars, comments, key order) survives
// untouched. If the frontmatter has no `name:` key (would be invalid
// per ADK's validator, but we are forgiving here), the new line is
// inserted right after the opening delimiter.
//
// The function leaves the body that follows the closing delimiter
// completely alone.
func rewriteFrontmatterName(raw []byte, slug string) []byte {
	// Locate the opening delimiter (after optional leading
	// whitespace). The skill loader already validated that one
	// exists, so we just trust it.
	openIdx := indexAfterDelimiter(raw, 0)
	if openIdx < 0 {
		return raw
	}
	closeIdx := indexAfterDelimiter(raw, openIdx)
	if closeIdx < 0 {
		return raw
	}
	// closeIdx points just past the closing "---\n" line; the
	// frontmatter body sits in [openIdx, closeStart) where closeStart
	// is the start of that closing line.
	closeStart := closeIdx
	// Walk back from closeIdx past the closing delimiter line.
	closeStart = lineStartBefore(raw, closeStart-1)

	header := raw[:openIdx]
	frontmatter := raw[openIdx:closeStart]
	rest := raw[closeStart:]

	newFrontmatter := replaceTopLevelKey(frontmatter, "name", slug)

	out := make([]byte, 0, len(header)+len(newFrontmatter)+len(rest))
	out = append(out, header...)
	out = append(out, newFrontmatter...)
	out = append(out, rest...)
	return out
}

// indexAfterDelimiter returns the byte offset right after the next
// "---\n" (or "---\r\n") line found at or after start, treating only
// lines whose first non-whitespace content is exactly the delimiter.
// Returns -1 when no such line exists.
func indexAfterDelimiter(raw []byte, start int) int {
	i := start
	for i < len(raw) {
		// Find the next newline.
		nl := bytesIndexFromByte(raw, i, '\n')
		var line []byte
		if nl < 0 {
			line = raw[i:]
		} else {
			line = raw[i : nl+1]
		}
		trimmed := strings.TrimRight(string(line), "\r\n")
		trimmed = strings.TrimLeft(trimmed, " \t")
		if trimmed == "---" {
			if nl < 0 {
				return len(raw)
			}
			return nl + 1
		}
		if nl < 0 {
			return -1
		}
		i = nl + 1
	}
	return -1
}

// bytesIndexFromByte is a tiny helper that finds the first occurrence
// of c in raw at or after start. Implemented locally so the migrator
// stays free of bytes/strings.IndexByte boilerplate at every call.
func bytesIndexFromByte(raw []byte, start int, c byte) int {
	for i := start; i < len(raw); i++ {
		if raw[i] == c {
			return i
		}
	}
	return -1
}

// lineStartBefore returns the byte offset of the first character of
// the line that contains offset i. It walks backwards until a newline
// (or BOF) is found.
func lineStartBefore(raw []byte, i int) int {
	if i < 0 {
		return 0
	}
	for i > 0 {
		if raw[i-1] == '\n' {
			return i
		}
		i--
	}
	return 0
}

// replaceTopLevelKey rewrites the value of a top-level YAML key inside
// a frontmatter block. "Top-level" here means the line starts with the
// key (no indentation) followed by a colon. The new value is rendered
// as a plain scalar on the same line. Any existing folded/flow value
// (including multi-line continuations indented under the key) is
// removed.
//
// If the key is absent the new line is inserted at the top of the
// frontmatter so the resulting document is still valid.
func replaceTopLevelKey(frontmatter []byte, key, value string) []byte {
	prefix := key + ":"
	lines := splitLinesPreserve(frontmatter)
	out := make([][]byte, 0, len(lines))
	replaced := false
	skipContinuation := false
	for _, line := range lines {
		if skipContinuation {
			// A continuation belongs to the previous key when it
			// starts with whitespace. The first non-indented line
			// ends the suppression.
			trimmed := strings.TrimRight(string(line), "\r\n")
			if trimmed == "" || (len(trimmed) > 0 && (trimmed[0] == ' ' || trimmed[0] == '\t')) {
				continue
			}
			skipContinuation = false
		}
		raw := strings.TrimRight(string(line), "\r\n")
		if !replaced && strings.HasPrefix(strings.TrimLeft(raw, " \t"), prefix) && !strings.HasPrefix(raw, " ") && !strings.HasPrefix(raw, "\t") {
			out = append(out, []byte(fmt.Sprintf("%s %s\n", prefix, value)))
			replaced = true
			// Eat any folded/multi-line continuation that
			// followed the original key.
			skipContinuation = true
			continue
		}
		out = append(out, line)
	}
	if !replaced {
		out = append([][]byte{[]byte(fmt.Sprintf("%s %s\n", prefix, value))}, out...)
	}
	var buf []byte
	for _, l := range out {
		buf = append(buf, l...)
	}
	return buf
}

// splitLinesPreserve splits raw into lines while keeping the line
// terminator with each piece, so concatenating the slices yields the
// original byte stream verbatim.
func splitLinesPreserve(raw []byte) [][]byte {
	var lines [][]byte
	start := 0
	for i := 0; i < len(raw); i++ {
		if raw[i] == '\n' {
			lines = append(lines, raw[start:i+1])
			start = i + 1
		}
	}
	if start < len(raw) {
		lines = append(lines, raw[start:])
	}
	return lines
}

// moveLegacyReferences relocates every legacy reference file under
// the new skill directory. The destination is chosen per file:
//
//   - If the filename already lives under a canonical resource
//     sub-tree (`references/...`, `assets/...`, `scripts/...`) — for
//     instance because the operator pasted a real skill package into
//     the legacy "references" bucket — we preserve that layout.
//   - Otherwise we put the file under `references/` (the only bucket
//     the old UI exposed).
//
// We do NOT auto-classify by extension. The operator's intent, when
// available, wins.
func moveLegacyReferences(srcDir, newSkillDir string, declared []string) error {
	// Prefer the declared reference list; fall back to whatever is on
	// disk if the store didn't track it explicitly.
	candidates := declared
	if len(candidates) == 0 {
		entries, err := os.ReadDir(srcDir)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if !e.IsDir() {
				candidates = append(candidates, e.Name())
			}
		}
	}
	for _, fn := range candidates {
		src := filepath.Join(srcDir, fn)
		if _, err := os.Stat(src); err != nil {
			continue // declared but missing — skip silently
		}
		relTarget := destinationForLegacyFile(fn)
		dst := filepath.Join(newSkillDir, filepath.FromSlash(relTarget))
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return err
		}
		if err := os.Rename(src, dst); err != nil {
			// Cross-device fallback: copy + remove.
			if copyErr := copyFile(src, dst); copyErr != nil {
				return fmt.Errorf("move %s: rename failed (%v) and copy failed: %w", fn, err, copyErr)
			}
			_ = os.Remove(src)
		}
	}
	return nil
}

// destinationForLegacyFile decides where a legacy reference filename
// should land inside the new skill directory. Canonical prefixes
// (references/, assets/, scripts/) are kept verbatim; everything else
// is dropped into references/{filename}. The returned path is always
// slash-separated (caller converts to OS path with filepath.FromSlash).
func destinationForLegacyFile(name string) string {
	clean := strings.TrimLeft(filepath.ToSlash(name), "/")
	for _, kind := range ResourceKinds {
		prefix := kind + "/"
		if strings.HasPrefix(clean, prefix) {
			return clean
		}
	}
	return "references/" + clean
}

func copyFile(src, dst string) error {
	in, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, in, 0o644)
}

// RepairBrokenSkills walks every skill package under dataDir/skills/
// and fixes two known pathologies left behind by an earlier buggy
// version of the migrator:
//
//  1. SKILL.md files with stacked frontmatter blocks
//     ("---\n…\n---\n---\n…real frontmatter…\n---\nbody"). The earlier
//     migrator wrapped legacy Instructions that were already a full
//     SKILL.md inside a second frontmatter, producing an invalid file.
//     The repair drops the synthesised wrapper and pins the real
//     frontmatter's `name` to the directory slug.
//
//  2. Resource files mis-rooted under `references/<kind>/...` when
//     `<kind>` is one of references/, assets/, scripts/. The earlier
//     migrator dropped every legacy file into references/, even when
//     the original filename declared its own canonical sub-tree.
//     The repair moves those files up to their intended sub-tree.
//
// The repair is idempotent and safe to run unconditionally on every
// load: skills already in good shape are left alone. We log every
// healed skill at info level so operators can audit what was changed.
//
// TODO(v0.X): remove together with MigrateLegacySkills.
func RepairBrokenSkills(dataDir string) error {
	if dataDir == "" {
		return nil
	}
	root := filepath.Join(dataDir, "skills")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		slug := e.Name()
		skillDir := filepath.Join(root, slug)
		if err := repairOneSkill(slug, skillDir); err != nil {
			slog.Warn("skills repair: failed", "slug", slug, "error", err)
		}
	}
	return nil
}

func repairOneSkill(slug, skillDir string) error {
	skillMDPath := filepath.Join(skillDir, "SKILL.md")
	body, err := os.ReadFile(skillMDPath)
	if err != nil {
		return err
	}

	healed := healStackedFrontmatter(body, slug)
	if !equalsBytes(body, healed) {
		if err := os.WriteFile(skillMDPath, healed, 0o644); err != nil {
			return fmt.Errorf("write SKILL.md: %w", err)
		}
		slog.Info("skills repair: rewrote SKILL.md with stacked frontmatter", "slug", slug)
	}

	// Look for misplaced resources: references/{references,assets,scripts}/...
	if err := healMisplacedResources(skillDir); err != nil {
		return err
	}
	return nil
}

// healStackedFrontmatter detects the doubled-frontmatter pathology
// produced by the earlier migrator and returns a clean SKILL.md. When
// the input does not exhibit the bug, the original bytes are returned
// unchanged so equalsBytes can short-circuit the on-disk write.
//
// Detection rule: the file starts with a frontmatter block whose body
// is itself another frontmatter block (i.e. "---\n…\n---\n---\n"
// appears at the very top, ignoring leading whitespace). When that
// pattern matches we strip the outer wrapper, then rewrite the inner
// frontmatter's `name` to the directory slug to satisfy ADK's
// dirname-equals-name invariant.
func healStackedFrontmatter(raw []byte, slug string) []byte {
	if !looksStacked(raw) {
		return raw
	}

	openIdx := indexAfterDelimiter(raw, 0)
	if openIdx < 0 {
		return raw
	}
	closeIdx := indexAfterDelimiter(raw, openIdx)
	if closeIdx < 0 {
		return raw
	}
	// closeIdx is just past the outer closing delimiter line. The
	// inner SKILL.md begins right there.
	inner := raw[closeIdx:]
	if !hasFrontmatterPrefix(string(inner)) {
		return raw
	}
	// Drop any leading blank lines so the result starts with the
	// real "---\n" line cleanly.
	inner = bytes.TrimLeft(inner, " \t\r\n")
	return rewriteFrontmatterName(inner, slug)
}

// looksStacked is a cheap heuristic: the first non-whitespace bytes
// must form a "---" delimiter line, AND the line immediately after
// the next "---" delimiter must also be "---". That captures the
// outer-wraps-inner shape without false positives on normal SKILL.md
// files (whose body very rarely contains a literal "---" line at the
// start of a paragraph).
func looksStacked(raw []byte) bool {
	trimmed := bytes.TrimLeft(raw, " \t\r\n")
	if !bytes.HasPrefix(trimmed, []byte("---\n")) && !bytes.HasPrefix(trimmed, []byte("---\r\n")) {
		return false
	}
	// Find the next "---\n" line after the first.
	openIdx := indexAfterDelimiter(raw, 0)
	if openIdx < 0 {
		return false
	}
	closeIdx := indexAfterDelimiter(raw, openIdx)
	if closeIdx < 0 {
		return false
	}
	rest := raw[closeIdx:]
	rest = bytes.TrimLeft(rest, " \t\r\n")
	return bytes.HasPrefix(rest, []byte("---\n")) || bytes.HasPrefix(rest, []byte("---\r\n"))
}

// healMisplacedResources walks references/ in skillDir and moves any
// nested files whose path is references/<kind>/... (for kind in
// references, assets, scripts) up to skillDir/<kind>/...; this undoes
// the over-eager "everything goes to references/" behaviour of the
// earlier migrator.
func healMisplacedResources(skillDir string) error {
	refsDir := filepath.Join(skillDir, "references")
	info, err := os.Stat(refsDir)
	if err != nil || !info.IsDir() {
		return nil
	}
	for _, kind := range ResourceKinds {
		nested := filepath.Join(refsDir, kind)
		stat, err := os.Stat(nested)
		if err != nil || !stat.IsDir() {
			continue
		}
		dst := filepath.Join(skillDir, kind)
		if err := mergeDir(nested, dst); err != nil {
			return fmt.Errorf("merge %s/%s -> %s: %w", "references", kind, kind, err)
		}
		_ = os.Remove(nested) // best-effort cleanup of the now-empty wrapper
		slog.Info("skills repair: moved misplaced resources",
			"from", filepath.Join("references", kind),
			"to", kind,
			"dir", filepath.Base(skillDir))
	}
	// If references/ ends up empty after the dance, drop it too so
	// the layout matches a fresh upload.
	if entries, err := os.ReadDir(refsDir); err == nil && len(entries) == 0 {
		_ = os.Remove(refsDir)
	}
	return nil
}

// mergeDir moves every file under src into dst, preserving sub-paths.
// Existing files at the destination are NOT overwritten — we rename
// the incoming file with a "-1", "-2" suffix so nothing is lost. Empty
// directories left behind in src are removed best-effort.
func mergeDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return err
	}
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		target = uniquePath(target)
		if err := os.Rename(path, target); err != nil {
			if copyErr := copyFile(path, target); copyErr != nil {
				return copyErr
			}
			_ = os.Remove(path)
		}
		return nil
	})
}

// uniquePath returns p, or p with a numeric suffix when an entry
// already exists there. The suffix sits between the basename stem and
// the extension to keep filetype detection working downstream.
func uniquePath(p string) string {
	if _, err := os.Stat(p); err != nil {
		return p
	}
	dir, base := filepath.Split(p)
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)
	for i := 1; i < 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", stem, i, ext))
		if _, err := os.Stat(candidate); err != nil {
			return candidate
		}
	}
	return p
}

func equalsBytes(a, b []byte) bool { return bytes.Equal(a, b) }
