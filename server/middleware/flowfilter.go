package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/achetronic/magec/server/store"
)

// bufferedResponseWriter captures the full response without writing to the
// client. The caller writes the (possibly filtered) body manually afterwards.
type bufferedResponseWriter struct {
	http.ResponseWriter
	body       bytes.Buffer
	statusCode int
	header     http.Header
}

func newBufferedResponseWriter(w http.ResponseWriter) *bufferedResponseWriter {
	return &bufferedResponseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		header:         http.Header{},
	}
}

func (b *bufferedResponseWriter) Header() http.Header {
	return b.header
}

func (b *bufferedResponseWriter) WriteHeader(code int) {
	b.statusCode = code
}

func (b *bufferedResponseWriter) Write(data []byte) (int, error) {
	return b.body.Write(data)
}

// filteringSSEWriter intercepts SSE writes, parses each data: line, and only
// forwards events whose author is in the allowed set. Non-data lines (comments,
// blank lines) pass through.
type filteringSSEWriter struct {
	http.ResponseWriter
	flusher    http.Flusher
	allowedSet map[string]bool
	buf        bytes.Buffer
}

func (f *filteringSSEWriter) Write(data []byte) (int, error) {
	f.buf.Write(data)
	for {
		line, err := f.buf.ReadString('\n')
		if err != nil {
			f.buf.WriteString(line)
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			f.ResponseWriter.Write([]byte(line))
			continue
		}
		if !strings.HasPrefix(trimmed, "data: ") {
			f.ResponseWriter.Write([]byte(line))
			continue
		}
		jsonData := strings.TrimPrefix(trimmed, "data: ")
		var event struct {
			Author string `json:"author"`
		}
		if json.Unmarshal([]byte(jsonData), &event) != nil || f.allowedSet[event.Author] {
			f.ResponseWriter.Write([]byte(line))
		}
	}
	return len(data), nil
}

func (f *filteringSSEWriter) Flush() {
	if f.flusher != nil {
		f.flusher.Flush()
	}
}

func (f *filteringSSEWriter) Unwrap() http.ResponseWriter {
	return f.ResponseWriter
}

// sessionGetRe matches GET /…/apps/{appName}/users/{userId}/sessions/{sessionId}
var sessionGetRe = regexp.MustCompile(`/apps/([^/]+)/users/[^/]+/sessions/[^/]+$`)

// FlowResponseFilter filters ADK response events for flow executions. When the
// requested appName is a flow, only events from agents marked as responseAgent
// are forwarded to the client. Applies to /run, /run_sse, and session GET.
func FlowResponseFilter(next http.Handler, dataStore *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isRun := strings.HasSuffix(path, "/run") && r.Method == "POST"
		isRunSSE := strings.HasSuffix(path, "/run_sse") && r.Method == "POST"

		var isSessionGet bool
		var appNameFromPath string
		if r.Method == "GET" {
			if m := sessionGetRe.FindStringSubmatch(path); m != nil {
				isSessionGet = true
				appNameFromPath = m[1]
			}
		}

		if !isRun && !isRunSSE && !isSessionGet {
			next.ServeHTTP(w, r)
			return
		}

		var appName string
		if isRun || isRunSSE {
			bodyBytes, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))

			var reqBody struct {
				AppName string `json:"appName"`
			}
			json.Unmarshal(bodyBytes, &reqBody)
			appName = reqBody.AppName
		} else {
			appName = appNameFromPath
		}

		flow, isFlow := dataStore.GetFlow(appName)
		if !isFlow {
			next.ServeHTTP(w, r)
			return
		}

		responseIDs := flow.ResponseAgentIDs()
		if len(responseIDs) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		allowedSet := make(map[string]bool, len(responseIDs))
		for _, id := range responseIDs {
			allowedSet[id] = true
		}

		switch {
		case isRun:
			filterRunResponse(next, w, r, allowedSet)
		case isRunSSE:
			filterRunSSEResponse(next, w, r, allowedSet)
		case isSessionGet:
			filterSessionResponse(next, w, r, allowedSet)
		}
	})
}

func filterRunResponse(next http.Handler, w http.ResponseWriter, r *http.Request, allowedSet map[string]bool) {
	buf := newBufferedResponseWriter(w)
	next.ServeHTTP(buf, r)

	for k, vals := range buf.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	if buf.statusCode != http.StatusOK {
		w.WriteHeader(buf.statusCode)
		w.Write(buf.body.Bytes())
		return
	}

	var events []json.RawMessage
	if err := json.Unmarshal(buf.body.Bytes(), &events); err != nil {
		w.WriteHeader(buf.statusCode)
		w.Write(buf.body.Bytes())
		return
	}

	filtered := make([]json.RawMessage, 0, len(events))
	for _, raw := range events {
		var evt struct {
			Author string `json:"author"`
		}
		if json.Unmarshal(raw, &evt) != nil || allowedSet[evt.Author] {
			filtered = append(filtered, raw)
		}
	}

	out, err := json.Marshal(filtered)
	if err != nil {
		w.WriteHeader(buf.statusCode)
		w.Write(buf.body.Bytes())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}

func filterRunSSEResponse(next http.Handler, w http.ResponseWriter, r *http.Request, allowedSet map[string]bool) {
	flusher, _ := w.(http.Flusher)
	fw := &filteringSSEWriter{
		ResponseWriter: w,
		flusher:        flusher,
		allowedSet:     allowedSet,
	}
	next.ServeHTTP(fw, r)
}

func filterSessionResponse(next http.Handler, w http.ResponseWriter, r *http.Request, allowedSet map[string]bool) {
	buf := newBufferedResponseWriter(w)
	next.ServeHTTP(buf, r)

	for k, vals := range buf.header {
		for _, v := range vals {
			w.Header().Add(k, v)
		}
	}

	if buf.statusCode != http.StatusOK {
		w.WriteHeader(buf.statusCode)
		w.Write(buf.body.Bytes())
		return
	}

	raw := buf.body.Bytes()
	var full map[string]json.RawMessage
	if err := json.Unmarshal(raw, &full); err != nil {
		w.WriteHeader(buf.statusCode)
		w.Write(raw)
		return
	}

	var events []json.RawMessage
	if evRaw, ok := full["events"]; ok {
		json.Unmarshal(evRaw, &events)
	}

	filtered := make([]json.RawMessage, 0, len(events))
	for _, evRaw := range events {
		var evt struct {
			Author  string `json:"author"`
			Content struct {
				Role string `json:"role"`
			} `json:"content"`
		}
		if json.Unmarshal(evRaw, &evt) != nil {
			filtered = append(filtered, evRaw)
			continue
		}
		if evt.Content.Role == "user" || evt.Author == "user" || evt.Author == "" || allowedSet[evt.Author] {
			filtered = append(filtered, evRaw)
		}
	}

	filteredJSON, _ := json.Marshal(filtered)
	full["events"] = filteredJSON

	out, err := json.Marshal(full)
	if err != nil {
		w.WriteHeader(buf.statusCode)
		w.Write(raw)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(out)
}
