// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// backupDownload streams a .tar.gz of the entire data/ directory.
// @Summary      Download backup
// @Description  Streams a .tar.gz archive containing the entire data directory (store.json, conversations.json, skills files).
// @Tags         backup
// @Produce      application/gzip
// @Success      200  {file}    binary  "Backup archive"
// @Failure      500  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /settings/backup [get]
func (h *Handler) backupDownload(w http.ResponseWriter, r *http.Request) {
	dataDir := h.store.DataDir()
	if dataDir == "" {
		writeError(w, http.StatusInternalServerError, "store has no data directory")
		return
	}

	ts := time.Now().UTC().Format("20060102-150405")
	filename := fmt.Sprintf("magec-backup-%s.tar.gz", ts)

	w.Header().Set("Content-Type", "application/gzip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))

	gz := gzip.NewWriter(w)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	filepath.Walk(dataDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(dataDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = io.Copy(tw, f)
		return err
	})
}

// backupRestore accepts a .tar.gz upload and replaces the data/ directory.
// @Summary      Restore from backup
// @Description  Accepts a .tar.gz archive (max 500MB) and atomically replaces the entire data directory. The archive must contain a valid store.json at the root level. After extraction, both the main store and conversation store are reloaded in memory.
// @Tags         backup
// @Accept       application/gzip
// @Produce      json
// @Param        body  body      []byte  true  "Backup .tar.gz archive"
// @Success      200   {object}  map[string]string  "status: restored"
// @Failure      400   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /settings/restore [post]
func (h *Handler) backupRestore(w http.ResponseWriter, r *http.Request) {
	dataDir := h.store.DataDir()
	if dataDir == "" {
		writeError(w, http.StatusInternalServerError, "store has no data directory")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 500<<20) // 500MB limit

	gz, err := gzip.NewReader(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid gzip: "+err.Error())
		return
	}
	defer gz.Close()

	tr := tar.NewReader(gz)

	tmpDir := dataDir + ".restore-tmp"
	os.RemoveAll(tmpDir)
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create temp dir: "+err.Error())
		return
	}
	defer os.RemoveAll(tmpDir)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid tar: "+err.Error())
			return
		}

		clean := filepath.Clean(hdr.Name)
		if strings.Contains(clean, "..") {
			writeError(w, http.StatusBadRequest, "path traversal detected: "+hdr.Name)
			return
		}

		target := filepath.Join(tmpDir, clean)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0o755); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create dir: "+err.Error())
				return
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create parent dir: "+err.Error())
				return
			}
			f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				writeError(w, http.StatusInternalServerError, "failed to create file: "+err.Error())
				return
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				writeError(w, http.StatusInternalServerError, "failed to write file: "+err.Error())
				return
			}
			f.Close()
		}
	}

	if _, err := os.Stat(filepath.Join(tmpDir, "store.json")); os.IsNotExist(err) {
		writeError(w, http.StatusBadRequest, "invalid backup: missing store.json")
		return
	}

	backupDir := dataDir + ".backup"
	os.RemoveAll(backupDir)
	if err := os.Rename(dataDir, backupDir); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to move current data: "+err.Error())
		return
	}

	if err := os.Rename(tmpDir, dataDir); err != nil {
		os.Rename(backupDir, dataDir)
		writeError(w, http.StatusInternalServerError, "failed to install backup: "+err.Error())
		return
	}

	os.RemoveAll(backupDir)

	if err := h.store.Reload(); err != nil {
		writeError(w, http.StatusInternalServerError, "data restored but store reload failed: "+err.Error())
		return
	}
	if h.conversations != nil {
		if err := h.conversations.Reload(); err != nil {
			writeError(w, http.StatusInternalServerError, "data restored but conversations reload failed: "+err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "restored"})
}
