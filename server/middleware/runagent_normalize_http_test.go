package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestRunAgentJSONNormalize_ChainToNextHandler verifies GitHub #26 end-to-end on the wire:
// clients send snake_case; the handler that receives the request after this middleware
// must see camelCase keys that ADK unmarshaling expects.
func TestRunAgentJSONNormalize_ChainToNextHandler(t *testing.T) {
	snakePayload := `{"app_name":"agent-uuid-1","user_id":"u1","session_id":"sess-gh26","new_message":{"role":"user","parts":[{"text":"hi"}]}}`

	for _, path := range []string{"/api/v1/agent/run", "/api/v1/agent/run_sse"} {
		t.Run(path, func(t *testing.T) {
			var got []byte
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var err error
				got, err = io.ReadAll(r.Body)
				if err != nil {
					t.Fatal(err)
				}
				w.WriteHeader(http.StatusOK)
			})

			h := RunAgentJSONNormalize(next)
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(snakePayload))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("unexpected status %d", rec.Code)
			}

			var m map[string]json.RawMessage
			if err := json.Unmarshal(got, &m); err != nil {
				t.Fatalf("downstream body not JSON: %v\n%s", err, string(got))
			}
			if _, ok := m["app_name"]; ok {
				t.Fatal("downstream must not contain app_name (GitHub #26)")
			}
			if string(m["appName"]) != `"agent-uuid-1"` {
				t.Fatalf("appName: got %s", string(m["appName"]))
			}
			if string(m["userId"]) != `"u1"` {
				t.Fatalf("userId: got %s", string(m["userId"]))
			}
			if string(m["sessionId"]) != `"sess-gh26"` {
				t.Fatalf("sessionId: got %s", string(m["sessionId"]))
			}
			if _, ok := m["new_message"]; ok {
				t.Fatal("downstream must not contain new_message")
			}
			if _, ok := m["newMessage"]; !ok {
				t.Fatal("downstream must contain newMessage")
			}
		})
	}
}

// Other POST paths must not rewrite JSON (only /run and /run_sse).
func TestRunAgentJSONNormalize_OtherPathsPassThrough(t *testing.T) {
	body := `{"app_name":"x","user_id":"y"}`
	var got []byte
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	h := RunAgentJSONNormalize(next)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/other", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !bytes.Equal(got, []byte(body)) {
		t.Fatalf("expected pass-through, got %s", string(got))
	}
}

func TestRunAgentJSONNormalize_GETPassesThrough(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	h := RunAgentJSONNormalize(next)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent/run", nil)
	h.ServeHTTP(httptest.NewRecorder(), req)
}
