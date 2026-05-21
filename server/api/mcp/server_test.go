package mcp

import (
	"context"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestSmoke_ToolsRegistered drives the embedded MCP server through the
// in-memory transport and verifies the full tool catalogue is exposed.
func TestSmoke_ToolsRegistered(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	serverT, clientT := sdk.NewInMemoryTransports()
	if _, err := h.server.Connect(ctx, serverT, nil); err != nil {
		t.Fatalf("server connect: %v", err)
	}

	client := sdk.NewClient(&sdk.Implementation{Name: "test-client", Version: "0"}, nil)
	sess, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer sess.Close()

	res, err := sess.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	if len(res.Tools) < 50 {
		t.Fatalf("expected at least 50 tools registered, got %d", len(res.Tools))
	}

	// Check that the SDK-reported count matches the handler's internal count.
	if got, want := len(res.Tools), h.ToolCount(); got != want {
		t.Fatalf("tool count mismatch: ListTools=%d, ToolCount=%d", got, want)
	}

	// Spot-check that a representative tool from each major group is present.
	required := []string{
		"magec_list_backends",
		"magec_create_agent",
		"magec_list_flows",
		"magec_list_clients",
		"magec_list_commands",
		"magec_list_skills",
		"magec_get_settings",
		"magec_list_secrets",
		"magec_list_conversations",
		"magec_list_voice_types",
		"magec_list_mcp_servers",
		"magec_list_memory_providers",
	}
	have := make(map[string]bool, len(res.Tools))
	for _, tool := range res.Tools {
		have[tool.Name] = true
	}
	for _, name := range required {
		if !have[name] {
			t.Errorf("required tool missing from server: %s", name)
		}
	}
}
