package mcp

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/achetronic/magec/server/middleware"
)

// initFrame is a minimal MCP initialize request used to coerce the SDK into
// a response. We only check that auth is enforced; whether the SDK accepts
// the payload is incidental.
const initFrame = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"curl","version":"0"}}}`

func TestHTTP_BearerRequired(t *testing.T) {
	h := newTestHandler(t)
	srv := httptest.NewServer(middleware.BearerAuth(h.HTTPHandler(), "secret"))
	defer srv.Close()

	// No bearer -> 401.
	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(initFrame))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no auth: got %d, want 401", resp.StatusCode)
	}

	// Wrong bearer -> 401.
	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader(initFrame))
	req.Header.Set("Authorization", "Bearer wrong")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong auth: got %d, want 401", resp.StatusCode)
	}

	// Correct bearer -> anything but 401.
	req, _ = http.NewRequest("POST", srv.URL, strings.NewReader(initFrame))
	req.Header.Set("Authorization", "Bearer secret")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatalf("authenticated request rejected: status %d", resp.StatusCode)
	}
}

func TestHTTP_OpenModeBypass(t *testing.T) {
	h := newTestHandler(t)
	// Empty password -> middleware short-circuits, requests pass through.
	srv := httptest.NewServer(middleware.BearerAuth(h.HTTPHandler(), ""))
	defer srv.Close()

	req, _ := http.NewRequest("POST", srv.URL, strings.NewReader(initFrame))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		t.Fatal("open mode unexpectedly returned 401")
	}
}
