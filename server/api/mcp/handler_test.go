package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/middleware"
)

// initFrame is the minimal MCP initialise envelope used by HTTP-level tests
// where we only care about the transport, not the protocol response.
const initFrame = `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"0"}}}`

// newTestHandler builds an MCP handler wired to a stub admin server, so the
// unit suite never depends on a real magec instance. The stub records every
// request it sees and replies with a canned payload — enough to assert the
// dispatcher serialised arguments correctly and propagated the bearer.
func newTestHandler(t *testing.T) (*Handler, *adminStub) {
	t.Helper()
	stub := newAdminStub()
	srv := httptest.NewServer(stub)
	t.Cleanup(srv.Close)

	h, err := NewHandler(HandlerConfig{
		AdminBaseURL:  srv.URL + "/api/v1/admin",
		AdminPassword: "test-pwd",
	})
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	return h, stub
}

// connectInMemory spins up the MCP server over the in-memory transport and
// returns an open client session. Centralised so individual tests don't
// re-implement the four-line dance.
func connectInMemory(t *testing.T, h *Handler) *sdk.ClientSession {
	t.Helper()
	ctx := context.Background()
	serverT, clientT := sdk.NewInMemoryTransports()
	if _, err := h.Server().Connect(ctx, serverT, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}
	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "0"}, nil)
	sess, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	t.Cleanup(func() { _ = sess.Close() })
	return sess
}

func TestSmoke_ToolsRegistered(t *testing.T) {
	h, _ := newTestHandler(t)
	sess := connectInMemory(t, h)

	res, err := sess.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(res.Tools) == 0 {
		t.Fatal("no tools registered")
	}
	if h.ToolCount() != len(res.Tools) {
		t.Fatalf("tool count mismatch: ToolCount=%d ListTools=%d", h.ToolCount(), len(res.Tools))
	}

	have := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		have[tool.Name] = true
		if !strings.HasPrefix(tool.Name, toolNamePrefix) {
			t.Errorf("tool %q missing %q prefix", tool.Name, toolNamePrefix)
		}
	}
	for _, want := range []string{"magec_get_backends", "magec_post_agents"} {
		if !have[want] {
			t.Errorf("expected representative tool %q to be registered", want)
		}
	}
}

func TestDispatcher_ForwardsArgumentsAndBearer(t *testing.T) {
	h, stub := newTestHandler(t)
	sess := connectInMemory(t, h)

	stub.respond("POST", "/api/v1/admin/backends", http.StatusCreated,
		`{"id":"abc","name":"OpenAI","type":"openai"}`)

	res, err := sess.CallTool(context.Background(), &sdk.CallToolParams{
		Name: "magec_post_backends",
		Arguments: map[string]any{
			"name": "OpenAI",
			"type": "openai",
		},
	})
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("tool returned error: %+v", res.Content)
	}

	if got, want := stub.lastAuth, "Bearer test-pwd"; got != want {
		t.Errorf("bearer not forwarded: got %q want %q", got, want)
	}
	if got, want := stub.lastMethod, http.MethodPost; got != want {
		t.Errorf("method: got %q want %q", got, want)
	}
	if got, want := stub.lastPath, "/api/v1/admin/backends"; got != want {
		t.Errorf("path: got %q want %q", got, want)
	}
	var body map[string]any
	if err := json.Unmarshal(stub.lastBody, &body); err != nil {
		t.Fatalf("body parse: %v", err)
	}
	if body["name"] != "OpenAI" || body["type"] != "openai" {
		t.Errorf("body mismatch: %v", body)
	}
}

func TestHTTP_BearerRequired(t *testing.T) {
	h, _ := newTestHandler(t)
	srv := httptest.NewServer(middleware.BearerAuth(h.HTTPHandler(), "secret"))
	defer srv.Close()

	// No bearer → 401.
	resp, err := http.Post(srv.URL, "application/json", strings.NewReader(initFrame))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no auth: got %d want 401", resp.StatusCode)
	}

	// Correct bearer → anything but 401.
	req, _ := http.NewRequest(http.MethodPost, srv.URL, strings.NewReader(initFrame))
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

// adminStub is a tiny HTTP server that records the most recent request it
// received and replies with whatever the test rigged via respond.
type adminStub struct {
	lastAuth   string
	lastMethod string
	lastPath   string
	lastBody   []byte

	responses map[string]stubResponse
}

type stubResponse struct {
	status int
	body   string
}

func newAdminStub() *adminStub {
	return &adminStub{responses: map[string]stubResponse{}}
}

func (s *adminStub) respond(method, path string, status int, body string) {
	s.responses[method+" "+path] = stubResponse{status: status, body: body}
}

func (s *adminStub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.lastAuth = r.Header.Get("Authorization")
	s.lastMethod = r.Method
	s.lastPath = r.URL.Path
	if r.Body != nil {
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		s.lastBody = body
	}

	resp, ok := s.responses[r.Method+" "+r.URL.Path]
	if !ok {
		resp = stubResponse{status: http.StatusOK, body: "{}"}
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)
	_, _ = w.Write([]byte(resp.body))
}
