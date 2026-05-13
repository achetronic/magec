// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package admin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	skillsmod "github.com/achetronic/magec/server/agent/tools/skills"
	"github.com/achetronic/magec/server/store"
)

// SkillSummary is the shape returned by the list endpoint. It carries
// only the fields the admin UI's card grid uses (id, slug, name,
// description) so the GET /skills response stays cheap regardless of
// how many skills the operator uploaded or how big each SKILL.md is.
// The full SkillView (with instructions and resource walk) is only
// hydrated by GET /skills/{id} when the operator opens the viewer.
type SkillSummary struct {
	ID          string `json:"id"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SkillView is the shape returned by GET /skills/{id} for a single
// skill. All fields except `id` and `slug` are read live from disk
// so the API never serves stale metadata after an out-of-band edit.
//
// `name` and `description` are the frontmatter values (NOT a store
// override — decision #29 keeps the store at {id, slug} only). Other
// frontmatter fields are surfaced verbatim through the `frontmatter`
// map so the UI can render whatever the operator put in the SKILL.md
// — including non-canonical keys (`version`, `author`, etc.) that
// ADK's strict `KnownFields` parser would otherwise reject. We use a
// permissive YAML decode here so a single unrecognised key cannot
// blank out the whole viewer.
type SkillView struct {
	ID           string          `json:"id"`
	Slug         string          `json:"slug"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Instructions string          `json:"instructions"`
	Frontmatter  map[string]any  `json:"frontmatter"`
	Resources    []SkillResource `json:"resources"`
}

// SkillResource is one file uploaded under the skill's references/,
// assets/ or scripts/ subtree. Path is the full relative path under
// the skill directory so the UI can preserve nested layouts.
type SkillResource struct {
	Kind string `json:"kind"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

// listSkills returns every skill known to the store as a lightweight
// summary (id, slug, name, description). Heavier fields — the
// instructions body and the resource walk — only travel through
// GET /skills/{id} where the viewer actually needs them.
//
// @Summary      List skills
// @Description  Returns all configured skills with their on-disk metadata
// @Tags         skills
// @Produce      json
// @Success      200  {array}  SkillSummary
// @Security     AdminAuth
// @Router       /skills [get]
func (h *Handler) listSkills(w http.ResponseWriter, r *http.Request) {
	skills := h.store.ListRawSkills()
	root := h.store.SkillsDir()
	out := make([]SkillSummary, 0, len(skills))
	for _, sk := range skills {
		out = append(out, hydrateSkillSummary(sk, root))
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Slug < out[j].Slug })
	writeJSON(w, http.StatusOK, out)
}

// getSkill returns a single skill view.
//
// @Summary      Get skill
// @Description  Returns a skill with its on-disk frontmatter, instructions and resources
// @Tags         skills
// @Produce      json
// @Param        id    path      string  true  "Skill ID"
// @Success      200   {object}  SkillView
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id} [get]
func (h *Handler) getSkill(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sk, ok := h.store.GetRawSkill(id)
	if !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	writeJSON(w, http.StatusOK, hydrateSkill(sk, h.store.SkillsDir()))
}

// uploadSkill is the single entry point for both creating a new skill
// and replacing an existing one. The accepted formats are documented
// in skillsmod.ParsePackage. Validation errors are surfaced verbatim
// so the operator can fix the SKILL.md / archive without guessing.
//
// Conflict resolution: the SKILL.md frontmatter `name` becomes the
// on-disk slug, which must be unique. If a skill with the same slug
// already exists, the request fails with 409 unless `?replace=true`
// is set, in which case the existing skill's directory is overwritten
// and its store ID is preserved (so agent links remain valid).
//
// @Summary      Upload a skill
// @Description  Creates or replaces a skill from a SKILL.md or a .zip/.tar.gz package
// @Tags         skills
// @Accept       multipart/form-data
// @Produce      json
// @Param        file     formData  file   true  "SKILL.md, .zip or .tar.gz"
// @Param        replace  query     bool   false "Overwrite if a skill with the same slug already exists"
// @Success      200      {object}  SkillView
// @Success      201      {object}  SkillView
// @Failure      400      {object}  ErrorResponse
// @Failure      409      {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/upload [post]
func (h *Handler) uploadSkill(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form: "+err.Error())
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	body, err := io.ReadAll(io.LimitReader(file, 50<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read upload: "+err.Error())
		return
	}

	pkg, err := skillsmod.ParsePackage(header.Filename, body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	slug := pkg.Frontmatter.Name
	replace, _ := strconv.ParseBool(r.URL.Query().Get("replace"))
	existing, slugTaken := h.store.GetSkillBySlug(slug)

	if slugTaken && !replace {
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":      fmt.Sprintf("a skill with slug %q already exists", slug),
			"existingId": existing.ID,
			"slug":       slug,
		})
		return
	}

	root := h.store.SkillsDir()
	if err := os.MkdirAll(root, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create skills directory: "+err.Error())
		return
	}
	if err := skillsmod.WritePackage(root, slug, pkg); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to write skill package: "+err.Error())
		return
	}

	var saved store.Skill
	status := http.StatusCreated
	if slugTaken {
		// Replace path: keep the same store ID so existing agent
		// links keep working. The on-disk directory has already been
		// rewritten above.
		if err := h.store.UpdateSkill(existing.ID, store.Skill{ID: existing.ID, Slug: slug}); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		saved = store.Skill{ID: existing.ID, Slug: slug}
		status = http.StatusOK
	} else {
		saved, err = h.store.CreateSkill(store.Skill{Slug: slug})
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, status, hydrateSkill(saved, root))
}

// deleteSkill removes the store record and the on-disk package.
//
// @Summary      Delete skill
// @Description  Deletes a skill and its on-disk package
// @Tags         skills
// @Param        id  path  string  true  "Skill ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id} [delete]
func (h *Handler) deleteSkill(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteSkill(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// downloadSkill streams the skill's directory as a tar.gz so operators
// can back it up or copy it between magec instances. Re-uploading the
// produced archive via /skills/upload?replace=true reconstructs the
// same skill verbatim.
//
// @Summary      Download skill package
// @Description  Streams the skill's on-disk directory as a tar.gz archive
// @Tags         skills
// @Produce      application/gzip
// @Param        id  path  string  true  "Skill ID"
// @Success      200
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id}/download [get]
func (h *Handler) downloadSkill(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	sk, ok := h.store.GetRawSkill(id)
	if !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	tgz, err := skillsmod.PackageAsTarGz(h.store.SkillsDir(), sk.Slug)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			writeError(w, http.StatusNotFound, "skill package not found on disk")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to package skill: "+err.Error())
		return
	}
	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", sk.Slug+".tar.gz"))
	_, _ = w.Write(tgz)
}

// hydrateSkillSummary is the cheap version of hydrateSkill used by
// the list endpoint. It reads only the SKILL.md frontmatter
// (skipping the body and the resource walk) so listing N skills
// costs O(N) small reads instead of O(N) full directory scans.
func hydrateSkillSummary(sk store.Skill, root string) SkillSummary {
	out := SkillSummary{ID: sk.ID, Slug: sk.Slug}
	if sk.Slug == "" || root == "" {
		return out
	}
	body, err := os.ReadFile(filepath.Join(root, sk.Slug, "SKILL.md"))
	if err != nil {
		return out
	}
	if fm, _, ok := skillsmod.ParseFrontmatterPermissive(body); ok {
		if name, _ := fm["name"].(string); name != "" {
			out.Name = name
		}
		out.Description = stringFromAny(fm["description"])
	}
	return out
}

// hydrateSkill reads the on-disk SKILL.md and resource tree for a
// stored skill and returns the full SkillView. When the on-disk
// package is missing or unreadable the view still contains the bare
// {id, slug} so the UI can show an "orphan" entry the operator can
// delete; the missing data is left at zero values.
//
// Frontmatter parsing is intentionally PERMISSIVE: ADK's own parser
// uses `KnownFields(true)` and rejects any extra key, but real-world
// SKILL.md files in the wild routinely carry `version`, `author`,
// `tags`, etc. We can't make those skills disappear from the admin
// UI just because of a stricter spec — so we decode the YAML as
// `map[string]any` here. Runtime (the agent's skilltoolset) uses a
// separate, equally-tolerant wrapper so a misnamed key cannot break
// agent rebuild either.
func hydrateSkill(sk store.Skill, root string) SkillView {
	v := SkillView{
		ID:        sk.ID,
		Slug:      sk.Slug,
		Resources: []SkillResource{},
	}
	if sk.Slug == "" || root == "" {
		return v
	}

	skillDir := filepath.Join(root, sk.Slug)
	body, err := os.ReadFile(filepath.Join(skillDir, "SKILL.md"))
	if err == nil {
		if fm, instructions, ok := skillsmod.ParseFrontmatterPermissive(body); ok {
			v.Frontmatter = fm
			if name, _ := fm["name"].(string); name != "" {
				v.Name = name
			}
			v.Description = stringFromAny(fm["description"])
			v.Instructions = instructions
		}
	}

	v.Resources = walkResources(context.Background(), skillDir)
	return v
}

// stringFromAny coerces map values to a clean string. Multi-line YAML
// scalars (`description: >\n  ...`) come back from the decoder with a
// trailing newline; trim it so the UI does not render visual
// whitespace.
func stringFromAny(v any) string {
	if v == nil {
		return ""
	}
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	default:
		return strings.TrimSpace(fmt.Sprint(x))
	}
}

// walkResources enumerates every regular file under the three allowed
// resource sub-directories of a skill package, returning their relative
// paths plus the kind they belong to. Anything outside those sub-trees
// is intentionally invisible — ADK's skilltoolset would refuse to read
// it anyway.
func walkResources(_ context.Context, skillDir string) []SkillResource {
	out := []SkillResource{}
	for _, kind := range skillsmod.ResourceKinds {
		base := filepath.Join(skillDir, kind)
		if _, err := os.Stat(base); err != nil {
			continue
		}
		_ = filepath.WalkDir(base, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			info, ierr := d.Info()
			if ierr != nil {
				return nil
			}
			rel, rerr := filepath.Rel(skillDir, p)
			if rerr != nil {
				return nil
			}
			out = append(out, SkillResource{
				Kind: kind,
				Path: filepath.ToSlash(rel),
				Size: info.Size(),
			})
			return nil
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
