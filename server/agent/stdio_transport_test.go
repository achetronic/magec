// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package agent

import (
	"context"
	"strings"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// TestStdioCommandTransportReconnects is the canary for issue #70: the
// mcptoolset reconnects when a session dies, and the transport must survive
// more than one Connect instead of failing with "exec: Stdout already set".
// The command is "cat" because it exits on stdin EOF, which is exactly how
// the SDK's stdio connection shuts down: a clean exit, no SIGTERM involved.
func TestStdioCommandTransportReconnects(t *testing.T) {
	srv := &store.MCPServer{
		Name:    "echo",
		Type:    "stdio",
		Command: "cat",
	}

	transport := &stdioCommandTransport{srv: srv}

	for i := range 3 {
		conn, err := transport.Connect(context.Background())
		if err != nil {
			t.Fatalf("Connect #%d: %v", i+1, err)
		}
		if err := conn.Close(); err != nil {
			t.Fatalf("Close on connection #%d: %v", i+1, err)
		}
	}
}

// TestStdioCommandHonoursServerFields ensures every spawned subprocess
// carries the working directory and environment from the server definition.
func TestStdioCommandHonoursServerFields(t *testing.T) {
	srv := &store.MCPServer{
		Name:    "probe",
		Type:    "stdio",
		Command: "server",
		Args:    []string{"--verbose"},
		WorkDir: "/srv/mcp",
		Env:     map[string]string{"MAGEC_MARK": "fresh"},
	}

	cmd := stdioCommand(srv)

	if cmd.Path != "server" {
		t.Fatalf("cmd.Path = %q, want %q", cmd.Path, "server")
	}
	if got := strings.Join(cmd.Args, " "); got != "server --verbose" {
		t.Fatalf("cmd.Args = %q, want %q", got, "server --verbose")
	}
	if cmd.Dir != "/srv/mcp" {
		t.Fatalf("cmd.Dir = %q, want %q", cmd.Dir, "/srv/mcp")
	}
	if !envContains(cmd.Env, "MAGEC_MARK=fresh") {
		t.Fatalf("cmd.Env does not contain %q: %v", "MAGEC_MARK=fresh", cmd.Env)
	}
}

func envContains(env []string, want string) bool {
	for _, kv := range env {
		if kv == want {
			return true
		}
	}
	return false
}
