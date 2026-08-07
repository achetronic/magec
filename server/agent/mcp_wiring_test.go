// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	adkagent "google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/tool"

	"github.com/achetronic/magec/server/store"
)

// fetchParams mirrors the shape of the tool the reproduction server exposes.
type fetchParams struct {
	URL string `json:"url"`
}

// startMCPServer serves a real streamable-http MCP server with a single
// "fetch" tool. It stands in for any FastMCP-Python server an operator
// might register.
func startMCPServer(t *testing.T) *httptest.Server {
	t.Helper()

	server := mcp.NewServer(&mcp.Implementation{
		Name:    "repro-server",
		Version: "1.0.0",
	}, nil)
	mcp.AddTool(server, &mcp.Tool{
		Name:        "fetch",
		Description: "Fetches a URL",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params fetchParams) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{}, "ok: " + params.URL, nil
	})

	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return server }, nil)
	srv := httptest.NewServer(handler)
	// No t.Cleanup(srv.Close): the streamable transport holds a long-lived
	// SSE stream and Close blocks until it ends. The test binary exits
	// right after, so letting the server leak is the honest trade.
	return srv
}

// TestNewMCPServerBecomesInvocable reproduces the registration flow from the
// field: register an http MCP server via the store, attach its ID to an
// agent, simulate a full process restart, and assert the agent's toolsets
// actually surface the remote tool. The canary guards the whole path: map
// building, buildToolsets, transport creation and the lazy tool listing.
func TestNewMCPServerBecomesInvocable(t *testing.T) {
	// MAGEC_MCP_ENDPOINT points the test at an external streamable-http
	// server (e.g. a FastMCP-Python instance) instead of the in-process
	// go-sdk one, so the exact third-party stack from the field can be
	// exercised without vendoring a Python runtime.
	if endpoint := os.Getenv("MAGEC_MCP_ENDPOINT"); endpoint != "" {
		runRegistrationRepro(t, endpoint)
		return
	}
	httpSrv := startMCPServer(t)

	runRegistrationRepro(t, httpSrv.URL+"/mcp")
}

// runRegistrationRepro drives the reported sequence against one endpoint:
// register the server, attach it to an agent, restart, rebuild the toolsets
// and list the tools.
func runRegistrationRepro(t *testing.T, endpoint string) {
	t.Helper()

	// Register the MCP server, then attach its ID to an agent: the two
	// admin API calls from the report.
	s := newStore(t)
	mcpDef, err := s.CreateMCPServer(store.MCPServer{
		Name:     "repro",
		Type:     "http",
		Endpoint: endpoint,
	})
	if err != nil {
		t.Fatalf("CreateMCPServer: %v", err)
	}
	agentDef, err := s.CreateAgent(store.AgentDefinition{
		Name:       "repro-agent",
		MCPServers: []string{mcpDef.ID},
	})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}

	// Simulate systemctl restart: the store is re-read from disk and the
	// toolsets rebuilt from scratch.
	s2 := reloadStore(t, s)

	mcpServerMap := make(map[string]store.MCPServer)
	for _, m := range s2.Data().MCPServers {
		mcpServerMap[m.ID] = m
	}

	toolsets, err := buildToolsets(agentDef, mcpServerMap, nil)
	if err != nil {
		t.Fatalf("buildToolsets: %v", err)
	}
	if len(toolsets) != 1 {
		t.Fatalf("buildToolsets returned %d toolsets, want 1 (memory off, one MCP attached)", len(toolsets))
	}

	ctx := adkagent.NewStrictContextMock(context.Background())
	tools, err := toolsets[0].Tools(&ctx)
	if err != nil {
		t.Fatalf("Tools(): %v", err)
	}
	if len(tools) != 1 || tools[0].Name() != "fetch" {
		t.Fatalf("Tools() = %v, want exactly one tool named %q", toolNames(tools), "fetch")
	}
}

func toolNames(ts []tool.Tool) []string {
	names := make([]string, 0, len(ts))
	for _, tl := range ts {
		names = append(names, tl.Name())
	}
	return names
}

// newStore creates a persisted store rooted in the test's temp dir.
func newStore(t *testing.T) *store.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "store.json")
	s, err := store.New(path, "")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

// reloadStore re-reads the same file the first store persisted to, the way
// a fresh process does at startup.
func reloadStore(t *testing.T, s *store.Store) *store.Store {
	t.Helper()
	path := filepath.Join(s.DataDir(), "store.json")
	s2, err := store.New(path, "")
	if err != nil {
		t.Fatalf("store.New on reload: %v", err)
	}
	return s2
}

// TestBuildToolsetsWarnsOnUnresolvedMCP is the observability canary: an
// agent referencing a nonexistent MCP server is still skipped tolerantly
// (a dead reference must not take the agent down), but the skip now leaves
// a warn log an operator can act on instead of failing in total silence.
func TestBuildToolsetsWarnsOnUnresolvedMCP(t *testing.T) {
	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(prev) })

	agentDef := store.AgentDefinition{ID: "agent-1", MCPServers: []string{"ghost"}}
	toolsets, err := buildToolsets(agentDef, map[string]store.MCPServer{}, nil)
	if err != nil {
		t.Fatalf("buildToolsets: %v", err)
	}
	if len(toolsets) != 0 {
		t.Fatalf("buildToolsets returned %d toolsets, want 0", len(toolsets))
	}
	if !strings.Contains(buf.String(), "does not exist") {
		t.Fatalf("expected a warn log about the unresolved MCP reference, got %q", buf.String())
	}
}
