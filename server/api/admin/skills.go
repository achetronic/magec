package admin

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"

	"github.com/achetronic/magec/server/store"
)

// listSkills returns all skills.
// @Summary      List skills
// @Description  Returns all configured skills
// @Tags         skills
// @Produce      json
// @Success      200  {array}  store.Skill
// @Security     AdminAuth
// @Router       /skills [get]
func (h *Handler) listSkills(w http.ResponseWriter, r *http.Request) {
	skills := h.store.ListRawSkills()
	writeJSON(w, http.StatusOK, skills)
}

// getSkill returns a single skill by ID.
// @Summary      Get skill
// @Description  Returns a skill by its unique ID
// @Tags         skills
// @Produce      json
// @Param        id    path      string  true  "Skill ID"
// @Success      200   {object}  store.Skill
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
	writeJSON(w, http.StatusOK, sk)
}

// createSkill creates a new skill.
// @Summary      Create skill
// @Description  Creates a new skill with instructions and optional references
// @Tags         skills
// @Accept       json
// @Produce      json
// @Param        body  body      store.Skill  true  "Skill definition"
// @Success      201   {object}  store.Skill
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills [post]
func (h *Handler) createSkill(w http.ResponseWriter, r *http.Request) {
	var sk store.Skill
	if err := json.NewDecoder(r.Body).Decode(&sk); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if sk.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if sk.Instructions == "" {
		writeError(w, http.StatusBadRequest, "instructions are required")
		return
	}
	sk.References = nil
	created, err := h.store.CreateSkill(sk)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateSkill updates an existing skill.
// @Summary      Update skill
// @Description  Updates a skill by ID
// @Tags         skills
// @Accept       json
// @Produce      json
// @Param        id    path      string       true  "Skill ID"
// @Param        body  body      store.Skill  true  "Skill definition"
// @Success      200   {object}  store.Skill
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id} [put]
func (h *Handler) updateSkill(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var sk store.Skill
	if err := json.NewDecoder(r.Body).Decode(&sk); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	existing, ok := h.store.GetRawSkill(id)
	if !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}
	sk.References = existing.References
	if err := h.store.UpdateSkill(id, sk); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetRawSkill(id)
	writeJSON(w, http.StatusOK, updated)
}

// deleteSkill deletes a skill.
// @Summary      Delete skill
// @Description  Deletes a skill by ID
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

// uploadSkillReference uploads a file as a skill reference.
// @Summary      Upload skill reference
// @Description  Uploads a file and registers it as a reference for the skill
// @Tags         skills
// @Accept       multipart/form-data
// @Produce      json
// @Param        id    path      string  true  "Skill ID"
// @Param        file  formData  file    true  "Reference file"
// @Success      201   {object}  store.SkillReference
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id}/references [post]
func (h *Handler) uploadSkillReference(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if _, ok := h.store.GetSkill(id); !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}

	r.ParseMultipartForm(10 << 20) // 10 MB limit
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	filename := filepath.Base(header.Filename)
	if filename == "" || filename == "." {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return
	}
	if strings.Contains(filename, "..") {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	dir := h.store.SkillDir(id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create directory: %v", err))
		return
	}

	dst, err := os.Create(filepath.Join(dir, filename))
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to create file: %v", err))
		return
	}
	defer dst.Close()

	written, err := io.Copy(dst, file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("failed to write file: %v", err))
		return
	}

	ref := store.SkillReference{
		Filename: filename,
		Size:     written,
	}

	if err := h.store.AddSkillReference(id, ref); err != nil {
		os.Remove(filepath.Join(dir, filename))
		if strings.Contains(err.Error(), "already exists") {
			writeError(w, http.StatusConflict, err.Error())
		} else {
			writeError(w, http.StatusNotFound, err.Error())
		}
		return
	}

	writeJSON(w, http.StatusCreated, ref)
}

// downloadSkillReference serves a skill reference file.
// @Summary      Download skill reference
// @Description  Downloads a reference file from a skill
// @Tags         skills
// @Produce      octet-stream
// @Param        id        path  string  true  "Skill ID"
// @Param        filename  path  string  true  "Reference filename"
// @Success      200
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id}/references/{filename} [get]
func (h *Handler) downloadSkillReference(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	filename := vars["filename"]

	if _, ok := h.store.GetSkill(id); !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	path := filepath.Join(h.store.SkillDir(id), filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		writeError(w, http.StatusNotFound, "reference file not found")
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	http.ServeFile(w, r, path)
}

// deleteSkillReference deletes a skill reference file.
// @Summary      Delete skill reference
// @Description  Removes a reference file from a skill
// @Tags         skills
// @Param        id        path  string  true  "Skill ID"
// @Param        filename  path  string  true  "Reference filename"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id}/references/{filename} [delete]
func (h *Handler) deleteSkillReference(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	filename := vars["filename"]

	if strings.Contains(filename, "..") || strings.Contains(filename, "/") {
		writeError(w, http.StatusBadRequest, "invalid filename")
		return
	}

	if err := h.store.RemoveSkillReference(id, filename); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	path := filepath.Join(h.store.SkillDir(id), filename)
	os.Remove(path)

	w.WriteHeader(http.StatusNoContent)
}

// uploadSkillPackage extracts a ZIP or tar.gz archive into a skill's directory.
// SKILL.md is required; its content populates the instructions field.
// If the SKILL.md has valid YAML frontmatter, name and description are extracted.
// All other files are registered as references.
// @Summary      Upload skill package
// @Description  Uploads a ZIP or tar.gz archive containing a SKILL.md and optional reference files
// @Tags         skills
// @Accept       multipart/form-data
// @Produce      json
// @Param        id    path      string  true  "Skill ID"
// @Param        file  formData  file    true  "Package archive (ZIP or tar.gz)"
// @Success      200   {object}  store.Skill
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /skills/{id}/package [post]
func (h *Handler) uploadSkillPackage(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	if _, ok := h.store.GetSkill(id); !ok {
		writeError(w, http.StatusNotFound, "skill not found")
		return
	}

	r.ParseMultipartForm(50 << 20) // 50 MB limit
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()

	buf, err := io.ReadAll(io.LimitReader(file, 50<<20))
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to read file: "+err.Error())
		return
	}

	var files map[string][]byte
	name := strings.ToLower(header.Filename)
	switch {
	case strings.HasSuffix(name, ".zip"):
		files, err = extractZip(buf)
	case strings.HasSuffix(name, ".tar.gz") || strings.HasSuffix(name, ".tgz"):
		files, err = extractTarGz(buf)
	default:
		writeError(w, http.StatusBadRequest, "unsupported format: use .zip or .tar.gz")
		return
	}
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to extract archive: "+err.Error())
		return
	}

	files = stripTopLevelDir(files)

	skillMD, ok := files["SKILL.md"]
	if !ok {
		writeError(w, http.StatusBadRequest, "archive must contain a SKILL.md at the root")
		return
	}

	dir := h.store.SkillDir(id)
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create directory: "+err.Error())
		return
	}

	var refs []store.SkillReference
	for relPath, content := range files {
		if relPath == "SKILL.md" {
			continue
		}
		target := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create directory: "+err.Error())
			return
		}
		if err := os.WriteFile(target, content, 0o644); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to write file: "+err.Error())
			return
		}
		refs = append(refs, store.SkillReference{Filename: relPath, Size: int64(len(content))})
	}

	instructions := string(skillMD)
	skillName, description := parseSkillFrontmatter(instructions)
	if skillName == "" {
		skillName = archiveBaseName(header.Filename)
	}

	sk := store.Skill{
		Name:         skillName,
		Description:  description,
		Instructions: instructions,
		References:   refs,
	}
	if err := h.store.UpdateSkill(id, sk); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	updated, _ := h.store.GetRawSkill(id)
	writeJSON(w, http.StatusOK, updated)
}

func extractZip(data []byte) (map[string][]byte, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	files := make(map[string][]byte)
	for _, f := range r.File {
		clean := filepath.Clean(f.Name)
		if strings.Contains(clean, "..") {
			continue
		}
		if f.FileInfo().IsDir() {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return nil, err
		}
		files[filepath.ToSlash(clean)] = content
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
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		clean := filepath.Clean(hdr.Name)
		if strings.Contains(clean, "..") {
			continue
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			return nil, err
		}
		files[filepath.ToSlash(clean)] = content
	}
	return files, nil
}

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

func parseSkillFrontmatter(text string) (name, description string) {
	trimmed := strings.TrimSpace(text)
	if !strings.HasPrefix(trimmed, "---") {
		return "", ""
	}
	after := strings.Index(trimmed[3:], "\n")
	if after == -1 {
		return "", ""
	}
	rest := trimmed[3+after+1:]
	closing := strings.Index(rest, "\n---")
	if closing == -1 {
		return "", ""
	}
	block := rest[:closing]
	for _, line := range strings.Split(block, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "name:") {
			name = strings.TrimSpace(line[5:])
		} else if strings.HasPrefix(line, "description:") {
			description = strings.TrimSpace(line[12:])
		}
	}
	return name, description
}

func archiveBaseName(filename string) string {
	base := filepath.Base(filename)
	for _, ext := range []string{".tar.gz", ".tgz", ".zip"} {
		if strings.HasSuffix(strings.ToLower(base), ext) {
			return base[:len(base)-len(ext)]
		}
	}
	return strings.TrimSuffix(base, filepath.Ext(base))
}
