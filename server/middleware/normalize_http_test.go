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

func TestSnakeCaseNormalize_RunEndpoint(t *testing.T) {
	for _, path := range []string{"/api/v1/agent/run", "/api/v1/agent/run_sse"} {
		t.Run(path, func(t *testing.T) {
			input := `{"app_name":"agent1","user_id":"u1","session_id":"s1","new_message":{"role":"user","parts":[{"text":"hi"}]}}`

			var downstream map[string]interface{}
			handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, _ := io.ReadAll(r.Body)
				json.Unmarshal(body, &downstream)
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(input))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if _, ok := downstream["appName"]; !ok {
				t.Error("expected appName in downstream body")
			}
			if _, ok := downstream["app_name"]; ok {
				t.Error("unexpected app_name in downstream body")
			}
			if _, ok := downstream["userId"]; !ok {
				t.Error("expected userId in downstream body")
			}
			if _, ok := downstream["sessionId"]; !ok {
				t.Error("expected sessionId in downstream body")
			}
		})
	}
}

func TestSnakeCaseNormalize_NestedConversion(t *testing.T) {
	input := `{"app_name":"a","user_id":"u","session_id":"s","new_message":{"role":"user","parts":[{"inline_data":{"mime_type":"image/png","data":"abc"}}]}}`

	var downstream []byte
	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstream, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/run_sse", strings.NewReader(input))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	s := string(downstream)
	if strings.Contains(s, "inline_data") {
		t.Error("expected inline_data to be converted to inlineData")
	}
	if strings.Contains(s, "mime_type") {
		t.Error("expected mime_type to be converted to mimeType")
	}
	if !strings.Contains(s, "inlineData") {
		t.Error("expected inlineData in output")
	}
	if !strings.Contains(s, "mimeType") {
		t.Error("expected mimeType in output")
	}
}

func TestSnakeCaseNormalize_OtherPathsPassThrough(t *testing.T) {
	input := `{"app_name":"agent1"}`
	var downstream []byte

	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstream, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/apps/a/users/u/sessions", strings.NewReader(input))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !strings.Contains(string(downstream), "app_name") {
		t.Error("non-run paths should pass through unchanged")
	}
}

func TestSnakeCaseNormalize_GETPassesThrough(t *testing.T) {
	called := false
	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/v1/agent/run", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("GET should pass through to next handler")
	}
}

func TestSnakeCaseNormalize_AlreadyCamel(t *testing.T) {
	input := `{"appName":"a","userId":"u","sessionId":"s","newMessage":{"role":"user","parts":[{"text":"hi"}]}}`

	var downstream []byte
	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstream, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/run", strings.NewReader(input))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if string(downstream) != input {
		t.Errorf("already-camelCase body should pass through unchanged\ngot:  %s\nwant: %s", downstream, input)
	}
}

func TestSnakeCaseNormalize_ContentLengthUpdated(t *testing.T) {
	input := `{"app_name":"a","user_id":"u","session_id":"s"}`

	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if r.ContentLength != int64(len(body)) {
			t.Errorf("ContentLength = %d, body length = %d", r.ContentLength, len(body))
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/run", strings.NewReader(input))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSnakeCaseNormalize_EmptyBody(t *testing.T) {
	called := false
	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		body, _ := io.ReadAll(r.Body)
		if len(body) != 0 {
			t.Errorf("expected empty body, got %q", body)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/run", bytes.NewReader(nil))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("handler should be called even with empty body")
	}
}

func TestSnakeCaseNormalize_InvalidJSON(t *testing.T) {
	input := `not json at all`

	var downstream []byte
	handler := SnakeCaseNormalize(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		downstream, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/v1/agent/run", strings.NewReader(input))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if string(downstream) != input {
		t.Errorf("invalid JSON should pass through unchanged")
	}
}
