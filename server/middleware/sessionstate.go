package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/achetronic/magec/server/store"
)

// SessionStateSeed intercepts session creation requests (POST to
// /apps/{app}/users/{user}/sessions/{session}) and injects empty outputKey
// values into the session state. This ensures that flow agents referencing
// {{agent.output:outputKey}} template variables in their system prompts don't fail when
// sessions are created by the Voice UI or other API clients that don't
// pre-seed the state themselves.
func SessionStateSeed(next http.Handler, dataStore *store.Store) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || !isSessionCreatePath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		outputKeys := collectOutputKeys(dataStore)
		if len(outputKeys) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		var body map[string]interface{}
		if r.Body != nil {
			data, err := io.ReadAll(r.Body)
			r.Body.Close()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if len(data) > 0 {
				if err := json.Unmarshal(data, &body); err != nil {
					r.Body = io.NopCloser(bytes.NewReader(data))
					next.ServeHTTP(w, r)
					return
				}
			}
		}
		if body == nil {
			body = map[string]interface{}{}
		}

		state, _ := body["state"].(map[string]interface{})
		if state == nil {
			state = map[string]interface{}{}
		}

		for key, val := range outputKeys {
			if _, exists := state[key]; !exists {
				state[key] = val
			}
		}
		body["state"] = state

		newBody, _ := json.Marshal(body)
		r.Body = io.NopCloser(bytes.NewReader(newBody))
		r.ContentLength = int64(len(newBody))

		next.ServeHTTP(w, r)
	})
}

// isSessionCreatePath matches paths like /api/v1/agent/apps/{app}/users/{user}/sessions/{session}
// which is the ADK session creation endpoint.
func isSessionCreatePath(path string) bool {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	// Expected: api/v1/agent/apps/{app}/users/{user}/sessions/{session}
	// That's 9 segments: api, v1, agent, apps, {app}, users, {user}, sessions, {session}
	if len(parts) != 9 {
		return false
	}
	return parts[3] == "apps" && parts[5] == "users" && parts[7] == "sessions"
}

func collectOutputKeys(dataStore *store.Store) map[string]interface{} {
	state := map[string]interface{}{}
	for _, a := range dataStore.ListAgents() {
		if a.OutputKey != "" {
			state[a.OutputKey] = ""
		}
	}
	return state
}
