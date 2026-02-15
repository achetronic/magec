package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/store"
)

// bodyCapture wraps an http.ResponseWriter to tee the response body into a
// buffer. This is intentionally separate from the AccessLog middleware's
// httpsnoop approach — we need the full body to parse ADK events, not just
// status code and byte count.
type bodyCapture struct {
	http.ResponseWriter
	body       bytes.Buffer
	statusCode int
}

func (r *bodyCapture) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *bodyCapture) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

// ConversationRecorder wraps an HTTP handler to intercept /run calls and log
// conversations to the conversation store. The perspective parameter determines
// whether this logs the "admin" (all events) or "user" (filtered) view.
func ConversationRecorder(next http.Handler, executor *clients.Executor, dataStore *store.Store, perspective string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isRun := strings.HasSuffix(path, "/run") && r.Method == "POST"

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
			AppName    string `json:"appName"`
			UserID     string `json:"userId"`
			SessionID  string `json:"sessionId"`
			NewMessage struct {
				Role  string `json:"role"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"newMessage"`
		}
		json.Unmarshal(bodyBytes, &reqBody)

		rec := &bodyCapture{ResponseWriter: w, statusCode: 200}
		next.ServeHTTP(rec, r)

		if rec.statusCode == http.StatusOK && reqBody.AppName != "" {
			go func() {
				var prompt string
				for _, p := range reqBody.NewMessage.Parts {
					prompt += p.Text
				}

				var events []map[string]interface{}
				if err := json.Unmarshal(rec.body.Bytes(), &events); err != nil {
					return
				}

				clientID := ""
				source := "voice-ui"
				if cl, ok := getClientFromRequest(r, dataStore); ok {
					clientID = cl.ID
					source = cl.Type
				}

				executor.LogExternalConversation(
					reqBody.AppName,
					reqBody.UserID,
					reqBody.SessionID,
					source,
					clientID,
					prompt,
					perspective,
					events,
				)
			}()
		}
	})
}

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

// ConversationRecorderSSE wraps an HTTP handler to intercept /run_sse calls
// and log conversations. The perspective parameter determines whether this logs
// the "admin" (all events) or "user" (filtered) view.
func ConversationRecorderSSE(next http.Handler, executor *clients.Executor, dataStore *store.Store, perspective string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		isRunSSE := strings.HasSuffix(path, "/run_sse") && r.Method == "POST"

		if !isRunSSE {
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
			AppName    string `json:"appName"`
			UserID     string `json:"userId"`
			SessionID  string `json:"sessionId"`
			NewMessage struct {
				Role  string `json:"role"`
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"newMessage"`
		}
		json.Unmarshal(bodyBytes, &reqBody)

		rec := &sseResponseRecorder{ResponseWriter: w, flusher: w.(http.Flusher)}
		next.ServeHTTP(rec, r)

		if reqBody.AppName != "" {
			go func() {
				var prompt string
				for _, p := range reqBody.NewMessage.Parts {
					prompt += p.Text
				}

				events := parseSSEEvents(rec.body.String())
				if len(events) == 0 {
					return
				}

				clientID := ""
				source := "voice-ui"
				if cl, ok := getClientFromRequest(r, dataStore); ok {
					clientID = cl.ID
					source = cl.Type
				}

				executor.LogExternalConversation(
					reqBody.AppName,
					reqBody.UserID,
					reqBody.SessionID,
					source,
					clientID,
					prompt,
					perspective,
					events,
				)
			}()
		}
	})
}

type sseResponseRecorder struct {
	http.ResponseWriter
	body    bytes.Buffer
	flusher http.Flusher
}

func (r *sseResponseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *sseResponseRecorder) Flush() {
	if r.flusher != nil {
		r.flusher.Flush()
	}
}

func parseSSEEvents(raw string) []map[string]interface{} {
	var events []map[string]interface{}
	lines := strings.Split(raw, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		var event map[string]interface{}
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			continue
		}
		events = append(events, event)
	}
	return events
}
