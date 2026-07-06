//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package user

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// newTestStore builds an empty file-backed store in a temp dir.
func newTestStore(t *testing.T) *store.Store {
	t.Helper()
	s, err := store.New(filepath.Join(t.TempDir(), "store.json"), "")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return s
}

// TestClientInfo_DefaultAgentSkipsStaleReferences is a canary for a real bug:
// a client whose allowedAgents list began with the ID of a deleted agent
// advertised that ghost as defaultAgent, so UIs following it aimed every
// voice/run call at a 404. The default must come from the validated list.
func TestClientInfo_DefaultAgentSkipsStaleReferences(t *testing.T) {
	s := newTestStore(t)

	agent, err := s.CreateAgent(store.AgentDefinition{Name: "Live Agent"})
	if err != nil {
		t.Fatalf("CreateAgent: %v", err)
	}
	client, err := s.CreateClient(store.ClientDefinition{
		Name: "ui",
		Type: "direct",
		// First entry references an agent that no longer exists.
		AllowedAgents: []string{"deadbeef-0000-0000-0000-000000000000", agent.ID},
	})
	if err != nil {
		t.Fatalf("CreateClient: %v", err)
	}

	h := New(s)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/client/info", nil)
	req.Header.Set("X-Client-ID", client.ID)
	rec := httptest.NewRecorder()
	h.ClientInfo(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var resp ClientInfoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.DefaultAgent != agent.ID {
		t.Fatalf("defaultAgent = %q, want live agent %q (stale reference must be skipped)", resp.DefaultAgent, agent.ID)
	}
	if len(resp.AllowedAgents) != 1 || resp.AllowedAgents[0].ID != agent.ID {
		t.Fatalf("allowedAgents = %+v, want only the live agent", resp.AllowedAgents)
	}
}

// TestClientInfo_DefaultAgentEmptyWhenNothingResolves: a client whose every
// reference is stale must advertise no default at all rather than a ghost.
func TestClientInfo_DefaultAgentEmptyWhenNothingResolves(t *testing.T) {
	s := newTestStore(t)

	client, err := s.CreateClient(store.ClientDefinition{
		Name:          "ui",
		Type:          "direct",
		AllowedAgents: []string{"deadbeef-0000-0000-0000-000000000000"},
	})
	if err != nil {
		t.Fatalf("CreateClient: %v", err)
	}

	h := New(s)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/client/info", nil)
	req.Header.Set("X-Client-ID", client.ID)
	rec := httptest.NewRecorder()
	h.ClientInfo(rec, req)

	var resp ClientInfoResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.DefaultAgent != "" {
		t.Fatalf("defaultAgent = %q, want empty when nothing resolves", resp.DefaultAgent)
	}
}
