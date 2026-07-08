// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package skills

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	adkskill "google.golang.org/adk/v2/tool/skilltoolset/skill"
	"gopkg.in/yaml.v3"
)

// ResourceKinds enumerates the three sub-directories ADK's FileSystemSource
// recognises inside a skill package. Anything outside these paths is
// rejected by ADK itself; we mirror that contract in the upload handler so
// the operator gets a 400 long before runtime.
var ResourceKinds = []string{"references", "assets", "scripts"}

// IsResourceKind reports whether kind is one of the three ADK-allowed
// sub-directories of a skill package.
func IsResourceKind(kind string) bool {
	for _, k := range ResourceKinds {
		if k == kind {
			return true
		}
	}
	return false
}

// slugRe matches one or more characters that are NOT a valid ADK slug
// character. ADK's frontmatter validator only accepts lowercase letters,
// digits and hyphens (no consecutive, no leading/trailing). We collapse
// every non-conforming run into a single hyphen and then trim.
var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

// Slugify converts a human-readable name into a slug ADK will accept as a
// frontmatter `name`. It returns an empty string when the input has no
// usable characters, in which case the caller must invent a fallback
// (typically based on the entity ID).
func Slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = slugRe.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	if len(s) > 64 {
		s = strings.TrimRight(s[:64], "-")
	}
	return s
}

// UniqueSlug returns a slug guaranteed not to collide with any of the
// given existing slugs by appending a numeric suffix when needed
// ("foo" → "foo-2" → "foo-3" …). The suffix is bounded by ADK's
// 64-char name limit.
func UniqueSlug(base string, taken map[string]struct{}) string {
	if _, clash := taken[base]; !clash {
		return base
	}
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", base, i)
		if len(candidate) > 64 {
			// Trim base just enough so the suffix fits; ADK enforces
			// a hard 64-char ceiling on frontmatter names.
			trim := len(candidate) - 64
			if trim >= len(base) {
				continue
			}
			candidate = fmt.Sprintf("%s-%d", strings.TrimRight(base[:len(base)-trim], "-"), i)
		}
		if _, clash := taken[candidate]; !clash {
			return candidate
		}
	}
	return base // pathological fallback; caller should detect collisions upstream
}

// PackageInfo captures the parsed contents of an uploaded skill
// package (or a SKILL.md uploaded standalone). Files maps the relative
// path to raw bytes; Frontmatter holds the parsed metadata;
// Instructions holds the SKILL.md body that follows the frontmatter;
// SkillMD holds the raw bytes of the original SKILL.md (preserving
// operator formatting AND extra non-canonical frontmatter keys ADK's
// strict parser cannot round-trip).
type PackageInfo struct {
	Frontmatter  *adkskill.Frontmatter
	Instructions string
	// SkillMD is the verbatim bytes of the SKILL.md file as the
	// operator wrote it (after slug-rewrite, when needed). WritePackage
	// uses this directly so we never lose extra frontmatter keys
	// (`version:`, `author:`, `tags:` …) that the canonical
	// Frontmatter struct doesn't model.
	SkillMD []byte
	// Files is the cleaned, slash-separated relative path of every
	// non-SKILL.md file in the upload. Paths are constrained to live
	// under references/, assets/ or scripts/; anything else returns an
	// error during extraction.
	Files map[string][]byte
}

// ParsePackage inspects an uploaded payload and returns its PackageInfo.
// It accepts three formats:
//
//   - A standalone SKILL.md (filename must end with .md or .markdown
//     and the body must start with a YAML frontmatter block). No
//     resources, just the skill text.
//   - A .zip archive containing SKILL.md at the root plus optional
//     resource subtrees.
//   - A .tar.gz / .tgz archive with the same layout.
//
// Top-level wrapper directories (a single common prefix shared by every
// archive entry, e.g. when zipping the parent folder) are stripped so
// the layout matches ADK's expected SKILL.md-at-root convention.
//
// The function rejects (with a descriptive error) any of the following:
//   - SKILL.md missing or unreadable
//   - frontmatter invalid per ADK's own validator
//   - resource files outside the three allowed sub-trees
//   - path traversal attempts
//
// The error messages are written to be surfaced verbatim through the
// admin API so the operator knows exactly what to fix.
func ParsePackage(filename string, body []byte) (*PackageInfo, error) {
	lower := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(lower, ".zip"):
		return parseFromArchive(extractZip, body)
	case strings.HasSuffix(lower, ".tar.gz"), strings.HasSuffix(lower, ".tgz"):
		return parseFromArchive(extractTarGz, body)
	case strings.HasSuffix(lower, ".md"), strings.HasSuffix(lower, ".markdown"):
		return parseStandaloneSKILL(body)
	default:
		return nil, fmt.Errorf("unsupported skill upload %q: expected SKILL.md, .zip or .tar.gz", filename)
	}
}

func parseStandaloneSKILL(body []byte) (*PackageInfo, error) {
	fm, instructions, err := parseSkillTolerant(body)
	if err != nil {
		return nil, fmt.Errorf("invalid SKILL.md: %w", err)
	}
	return &PackageInfo{
		Frontmatter:  fm,
		Instructions: instructions,
		SkillMD:      body,
		Files:        map[string][]byte{},
	}, nil
}

func parseFromArchive(extract func([]byte) (map[string][]byte, error), body []byte) (*PackageInfo, error) {
	files, err := extract(body)
	if err != nil {
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}
	files = stripTopLevelDir(files)

	skillMD, ok := files["SKILL.md"]
	if !ok {
		return nil, errors.New("archive must contain a SKILL.md at the root")
	}
	delete(files, "SKILL.md")

	fm, instructions, err := parseSkillTolerant(skillMD)
	if err != nil {
		return nil, fmt.Errorf("invalid SKILL.md: %w", err)
	}

	resources := make(map[string][]byte, len(files))
	for relPath, content := range files {
		clean, err := validateResourcePath(relPath)
		if err != nil {
			return nil, err
		}
		resources[clean] = content
	}

	return &PackageInfo{
		Frontmatter:  fm,
		Instructions: instructions,
		SkillMD:      skillMD,
		Files:        resources,
	}, nil
}

// parseSkillTolerant returns the canonical Frontmatter struct + body
// for a SKILL.md, accepting frontmatters that include extra non-spec
// keys (e.g. `version:`, `author:`, `tags:`). ADK's own ParseBytes
// uses `decoder.KnownFields(true)` and rejects every unknown key,
// which is too strict for real-world skills authored against other
// agent runtimes (Claude Code, Open Code, …). We try the strict path
// first; on failure we strip anything that isn't in the canonical
// schema and re-parse. The Validate() step still runs on the result
// so genuinely broken frontmatters (missing description, bad name)
// are rejected with the original ADK error message.
func parseSkillTolerant(body []byte) (*adkskill.Frontmatter, string, error) {
	if fm, instructions, err := adkskill.ParseBytes(body); err == nil {
		return fm, instructions, nil
	}

	// Fallback: hand-parse the raw frontmatter, strip extras, and
	// rebuild a canonical SKILL.md the strict parser will accept.
	rawFM, instructions, ok := SplitFrontmatter(body)
	if !ok {
		// No frontmatter at all — surface ADK's original error.
		_, _, err := adkskill.ParseBytes(body)
		return nil, "", err
	}
	pruned := pruneUnknownFrontmatterKeys(rawFM)
	rebuilt := assembleSkillMD(pruned, instructions)
	fm, _, err := adkskill.ParseBytes(rebuilt)
	if err != nil {
		return nil, "", err
	}
	return fm, instructions, nil
}

// SplitFrontmatter slices a SKILL.md into (frontmatter YAML bytes,
// body string). Returns ok=false when the leading "---\n" delimiter
// or the closing one is missing — the caller is expected to surface
// the strict parser's original error in that case.
//
// Exported because both the runtime tolerant source and the admin API
// hydration path need to peek at the raw frontmatter without going
// through ADK's strict KnownFields parser. Decision #29 keeps the
// SKILL.md format owned by ADK; this helper is the single
// frontmatter-locator both Magec call sites share.
func SplitFrontmatter(body []byte) ([]byte, string, bool) {
	trimmed := bytes.TrimLeft(body, " \t\r\n")
	if !bytes.HasPrefix(trimmed, []byte("---\n")) && !bytes.HasPrefix(trimmed, []byte("---\r\n")) {
		return nil, "", false
	}
	rest := trimmed[len("---\n"):]
	if bytes.HasPrefix(trimmed, []byte("---\r\n")) {
		rest = trimmed[len("---\r\n"):]
	}
	if i := bytes.Index(rest, []byte("\n---\n")); i >= 0 {
		return rest[:i], string(rest[i+len("\n---\n"):]), true
	}
	if i := bytes.Index(rest, []byte("\n---\r\n")); i >= 0 {
		return rest[:i], string(rest[i+len("\n---\r\n"):]), true
	}
	return nil, "", false
}

// ParseFrontmatterPermissive splits a SKILL.md and decodes its
// frontmatter as a permissive map[string]any so callers can surface
// every operator-supplied key — including non-canonical ones
// (`version`, `author`, `tags`) that ADK's strict parser would
// reject. Returns ok=false when the frontmatter delimiters are
// missing OR the YAML doesn't decode; the body string is still
// returned in the YAML-decode-error case so the UI shows the prose
// even when the metadata is malformed.
func ParseFrontmatterPermissive(raw []byte) (map[string]any, string, bool) {
	yamlBytes, body, ok := SplitFrontmatter(raw)
	if !ok {
		return nil, string(raw), false
	}
	fm := map[string]any{}
	if err := yaml.Unmarshal(yamlBytes, &fm); err != nil {
		return nil, body, false
	}
	return fm, body, true
}

// canonicalFrontmatterKeys is the set of fields ADK's Frontmatter
// struct accepts. Anything outside this set is dropped during the
// tolerant fallback so the canonical parser can complete.
var canonicalFrontmatterKeys = map[string]struct{}{
	"name":          {},
	"description":   {},
	"license":       {},
	"compatibility": {},
	"metadata":      {},
	"allowed-tools": {},
}

// pruneUnknownFrontmatterKeys decodes the YAML, drops any key not in
// canonicalFrontmatterKeys, and re-encodes. We deliberately do NOT
// preserve unknown keys here — this byte stream is fed to ADK's
// strict parser only to satisfy validation. The verbatim original
// SKILL.md (with all extras intact) is what gets written to disk
// later by WritePackage via PackageInfo.SkillMD.
func pruneUnknownFrontmatterKeys(yamlBytes []byte) []byte {
	var raw map[string]any
	if err := yaml.Unmarshal(yamlBytes, &raw); err != nil {
		return yamlBytes
	}
	for k := range raw {
		if _, ok := canonicalFrontmatterKeys[k]; !ok {
			delete(raw, k)
		}
	}
	out, err := yaml.Marshal(raw)
	if err != nil {
		return yamlBytes
	}
	return out
}

// assembleSkillMD wraps a YAML frontmatter block + a markdown body in
// the canonical SKILL.md envelope. Used internally during tolerant
// parsing only — never written to disk.
func assembleSkillMD(fm []byte, body string) []byte {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.Write(fm)
	if !bytes.HasSuffix(fm, []byte("\n")) {
		buf.WriteByte('\n')
	}
	buf.WriteString("---\n")
	buf.WriteString(body)
	return buf.Bytes()
}

// validateResourcePath normalises a path to forward slashes and rejects
// any entry that escapes the three allowed sub-trees. Returns the
// cleaned path on success.
func validateResourcePath(p string) (string, error) {
	clean := path.Clean(filepath.ToSlash(p))
	if clean == "." || clean == "" {
		return "", fmt.Errorf("invalid resource path %q", p)
	}
	if strings.HasPrefix(clean, "/") || strings.Contains(clean, "..") {
		return "", fmt.Errorf("invalid resource path %q (must be relative, no traversal)", p)
	}
	first := clean
	if i := strings.Index(clean, "/"); i >= 0 {
		first = clean[:i]
	}
	if !IsResourceKind(first) {
		return "", fmt.Errorf("invalid resource path %q: top-level directory must be one of references/, assets/, scripts/", p)
	}
	return clean, nil
}

func extractZip(data []byte) (map[string][]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte)
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		clean := filepath.ToSlash(filepath.Clean(f.Name))
		if strings.Contains(clean, "..") {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			return nil, err
		}
		files[clean] = content
	}
	return files, nil
}

func extractTarGz(data []byte) (map[string][]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	files := make(map[string][]byte)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		clean := filepath.ToSlash(filepath.Clean(hdr.Name))
		if strings.Contains(clean, "..") {
			continue
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[clean] = content
	}
	return files, nil
}

// stripTopLevelDir removes a single shared prefix directory from every
// entry in the map, so archives that wrap their content in a parent
// folder (the common case when zipping a directory) end up rooted at
// SKILL.md anyway.
func stripTopLevelDir(files map[string][]byte) map[string][]byte {
	if len(files) == 0 {
		return files
	}
	prefix := ""
	for name := range files {
		parts := strings.SplitN(name, "/", 2)
		if len(parts) < 2 {
			return files
		}
		if prefix == "" {
			prefix = parts[0]
		} else if parts[0] != prefix {
			return files
		}
	}
	stripped := make(map[string][]byte, len(files))
	for name, content := range files {
		stripped[name[len(prefix)+1:]] = content
	}
	return stripped
}

// WritePackage materialises a PackageInfo on disk under dir/{slug}/. It
// rebuilds the directory from scratch (existing contents under
// dir/{slug}/ are removed) so re-uploads always reflect exactly what
// the operator just sent. The on-disk SKILL.md is reconstructed from
// the parsed frontmatter and the instructions body so any irregularities
// in the uploaded SKILL.md (extra whitespace, alternate frontmatter
// delimiters) get normalised.
func WritePackage(dir, slug string, pkg *PackageInfo) error {
	if pkg == nil || pkg.Frontmatter == nil {
		return errors.New("nil package info")
	}
	skillDir := filepath.Join(dir, slug)
	if err := os.RemoveAll(skillDir); err != nil {
		return fmt.Errorf("clean skill dir: %w", err)
	}
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		return fmt.Errorf("create skill dir: %w", err)
	}

	// Prefer the verbatim SKILL.md the operator uploaded so we
	// preserve every frontmatter key, including non-canonical ones
	// (`version:`, `author:`, `tags:` …) that ADK's strict
	// Frontmatter struct does not model. Fall back to rebuilding
	// from the parsed Frontmatter only when the raw bytes are
	// missing — that path loses extras but is at least valid.
	skillMD := pkg.SkillMD
	if len(skillMD) == 0 {
		built, err := adkskill.Build(pkg.Frontmatter, pkg.Instructions)
		if err != nil {
			return fmt.Errorf("build SKILL.md: %w", err)
		}
		skillMD = built
	}
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), skillMD, 0o644); err != nil {
		return fmt.Errorf("write SKILL.md: %w", err)
	}

	for relPath, content := range pkg.Files {
		target := filepath.Join(skillDir, filepath.FromSlash(relPath))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("create %s parent dir: %w", relPath, err)
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", relPath, err)
		}
	}
	return nil
}

// PackageAsTarGz reads dir/{slug}/ and returns its contents as a
// tar.gz archive. The archive root is the SKILL.md file (no wrapping
// directory), so re-uploading the produced file via the upload endpoint
// reconstructs the same skill verbatim.
func PackageAsTarGz(dir, slug string) ([]byte, error) {
	skillDir := filepath.Join(dir, slug)
	if _, err := os.Stat(skillDir); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)

	err := filepath.WalkDir(skillDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(skillDir, p)
		if err != nil {
			return err
		}
		body, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		hdr := &tar.Header{
			Name:    filepath.ToSlash(rel),
			Mode:    0o644,
			Size:    int64(len(body)),
			ModTime: info.ModTime(),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if _, err := tw.Write(body); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		_ = tw.Close()
		_ = gz.Close()
		return nil, err
	}
	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := gz.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
