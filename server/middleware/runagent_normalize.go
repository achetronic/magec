package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// RunAgentJSONNormalize rewrites POST /agent/run and /agent/run_sse JSON bodies so that
// snake_case keys used in older clients and bug reports (GitHub #26) are accepted:
// app_name→appName, user_id→userId, session_id→sessionId, new_message→newMessage,
// state_delta→stateDelta. CamelCase keys win if both are present.
// The ADK REST layer only unmarshals camelCase; without this, session lookup fails with
// "session not found" even when the session exists.
func RunAgentJSONNormalize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		path := r.URL.Path
		isRun := strings.HasSuffix(path, "/run")
		isRunSSE := strings.HasSuffix(path, "/run_sse")
		if !isRun && !isRunSSE {
			next.ServeHTTP(w, r)
			return
		}

		body, err := io.ReadAll(r.Body)
		r.Body.Close()
		if err != nil || len(body) == 0 {
			if len(body) == 0 {
				r.Body = io.NopCloser(bytes.NewReader(nil))
			}
			next.ServeHTTP(w, r)
			return
		}

		out, changed, normErr := normalizeRunAgentJSONBody(body)
		if normErr != nil {
			r.Body = io.NopCloser(bytes.NewReader(body))
			r.ContentLength = int64(len(body))
			next.ServeHTTP(w, r)
			return
		}
		if !changed {
			r.Body = io.NopCloser(bytes.NewReader(body))
			next.ServeHTTP(w, r)
			return
		}

		r.Body = io.NopCloser(bytes.NewReader(out))
		r.ContentLength = int64(len(out))
		r.Header.Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func normalizeRunAgentJSONBody(body []byte) (out []byte, changed bool, err error) {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, false, err
	}
	if len(m) == 0 {
		return body, false, nil
	}

	aliases := []struct {
		snake, camel string
	}{
		{"app_name", "appName"},
		{"user_id", "userId"},
		{"session_id", "sessionId"},
		{"new_message", "newMessage"},
		{"state_delta", "stateDelta"},
	}

	for _, a := range aliases {
		snakeRaw, hasSnake := m[a.snake]
		if !hasSnake {
			continue
		}
		if _, hasCamel := m[a.camel]; !hasCamel {
			m[a.camel] = snakeRaw
			changed = true
		}
		delete(m, a.snake)
		changed = true
	}

	if !changed {
		return body, false, nil
	}

	out, err = json.Marshal(m)
	if err != nil {
		return nil, false, err
	}
	return out, true, nil
}
