// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/achetronic/magec/server/agent/runrecorder"
	"github.com/achetronic/magec/server/store"
)

// RunAudit annotates the run recorder with the caller identity of /run and
// /run_sse requests (the plugin layer cannot see HTTP) and reports SSE error
// frames back to it as run-fatal errors, which never surface as events.
// A nil recorder disables the middleware entirely.
func RunAudit(next http.Handler, recorder *runrecorder.Recorder, dataStore *store.Store) http.Handler {
	if recorder == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		isRun := (strings.HasSuffix(r.URL.Path, "/run") || strings.HasSuffix(r.URL.Path, "/run_sse")) && r.Method == "POST"
		if !isRun {
			next.ServeHTTP(w, r)
			return
		}

		bodyBytes, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

		var reqBody struct {
			SessionID string `json:"sessionId"`
		}
		json.Unmarshal(bodyBytes, &reqBody)
		if reqBody.SessionID != "" {
			clientID, source := "", "voice-ui"
			if cl, ok := getClientFromRequest(r, dataStore); ok {
				clientID, source = cl.ID, cl.Type
			}
			recorder.Annotate(reqBody.SessionID, clientID, source)
		}

		scanner := &sseErrorScanner{ResponseWriter: w, flusher: flusherOf(w)}
		next.ServeHTTP(scanner, r)
		scanner.finishLine()

		if scanner.errorMessage != "" && scanner.invocationID != "" {
			recorder.MarkRunError(scanner.invocationID, scanner.errorMessage)
		}
	})
}

// getClientFromRequest resolves the calling client from its bearer token so
// runs carry their attribution (client ID and source).
func getClientFromRequest(r *http.Request, dataStore *store.Store) (store.ClientDefinition, bool) {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return store.ClientDefinition{}, false
	}
	token := strings.TrimPrefix(auth, "Bearer ")
	if token == auth {
		return store.ClientDefinition{}, false
	}
	return dataStore.GetClientByToken(token)
}

// flusherOf returns the writer's flusher when it has one.
func flusherOf(w http.ResponseWriter) http.Flusher {
	f, _ := w.(http.Flusher)
	return f
}

// sseErrorScanner passes the response through untouched while scanning the
// SSE lines for an error frame and for the run's invocation id. It keeps only
// the current partial line, not the whole stream.
type sseErrorScanner struct {
	http.ResponseWriter
	flusher http.Flusher

	line            []byte
	nextDataIsError bool
	invocationID    string
	errorMessage    string
}

func (s *sseErrorScanner) Write(b []byte) (int, error) {
	for _, c := range b {
		if c == '\n' {
			s.finishLine()
			continue
		}
		s.line = append(s.line, c)
	}
	return s.ResponseWriter.Write(b)
}

func (s *sseErrorScanner) Flush() {
	if s.flusher != nil {
		s.flusher.Flush()
	}
}

func (s *sseErrorScanner) Unwrap() http.ResponseWriter {
	return s.ResponseWriter
}

// finishLine classifies a completed SSE line. Error frames arrive as an
// "event: error" line followed by its data line; regular frames are data
// lines whose JSON carries the invocationId used to attribute the error.
func (s *sseErrorScanner) finishLine() {
	line := strings.TrimSpace(string(s.line))
	s.line = s.line[:0]

	if line == "event: error" {
		s.nextDataIsError = true
		return
	}
	data, ok := strings.CutPrefix(line, "data: ")
	if !ok {
		return
	}
	if s.nextDataIsError {
		s.nextDataIsError = false
		s.errorMessage = errorText(data)
		return
	}
	if s.invocationID == "" {
		var frame struct {
			InvocationID string `json:"invocationId"`
		}
		if json.Unmarshal([]byte(data), &frame) == nil {
			s.invocationID = frame.InvocationID
		}
	}
}

// errorText extracts a readable message from an error data frame, which may
// be a JSON object or plain text.
func errorText(data string) string {
	var frame struct {
		Error   string `json:"error"`
		Message string `json:"message"`
	}
	if json.Unmarshal([]byte(data), &frame) == nil {
		if frame.Error != "" {
			return frame.Error
		}
		if frame.Message != "" {
			return frame.Message
		}
	}
	return data
}
