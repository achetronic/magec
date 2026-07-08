// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// integrityHandler builds an admin handler over a store where one agent is
// referenced by a client (membership) and by a flow node (structural).
func integrityHandler(t *testing.T) (*Handler, *store.Store, string) {
	t.Helper()
	s, err := store.New(filepath.Join(t.TempDir(), "store.json"), "")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	agent, _ := s.CreateAgent(store.AgentDefinition{Name: "Agent"})
	_, _ = s.CreateClient(store.ClientDefinition{
		Name: "UI", Type: "direct", AllowedAgents: []string{agent.ID},
	})
	_, _ = s.CreateFlow(store.FlowDefinition{
		Name:  "Flow",
		Entry: "a1",
		Nodes: []store.FlowNode{{ID: "a1", Type: store.FlowNodeAgent, AgentID: agent.ID}},
	})
	return New(s), s, agent.ID
}

// TestDeleteAgent_ReferencedReturns409 is the guard's canary: deleting a
// referenced entity without force must not delete anything and must return
// the reference breakdown the UI renders in the force-confirm dialog.
func TestDeleteAgent_ReferencedReturns409(t *testing.T) {
	h, s, agentID := integrityHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/agents/"+agentID, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("status = %d, want 409", rec.Code)
	}
	var resp ReferencesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp.References) != 2 {
		t.Fatalf("references = %+v, want client + flow", resp.References)
	}
	kinds := map[string]bool{}
	for _, r := range resp.References {
		kinds[r.Kind] = true
	}
	if !kinds[store.RefMembership] || !kinds[store.RefStructural] {
		t.Fatalf("expected one membership and one structural ref, got %+v", resp.References)
	}
	if _, ok := s.GetAgent(agentID); !ok {
		t.Fatalf("agent must survive a non-forced delete")
	}
}

// TestDeleteAgent_ForceScrubsAndDeletes: with ?force=true the delete scrubs
// the references and removes the entity in one request.
func TestDeleteAgent_ForceScrubsAndDeletes(t *testing.T) {
	h, s, agentID := integrityHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/agents/"+agentID+"?force=true", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204 (body %s)", rec.Code, rec.Body.String())
	}
	if _, ok := s.GetAgent(agentID); ok {
		t.Fatalf("agent must be gone after force delete")
	}
	if refs := s.Referrers(agentID); len(refs) != 0 {
		t.Fatalf("references must be scrubbed, got %+v", refs)
	}
	if dead := s.DeadReferences(); len(dead) != 0 {
		t.Fatalf("force delete must leave no corpses, got %+v", dead)
	}
}

// TestDeleteAgent_UnreferencedNeedsNoForce: entities without referrers keep
// the plain delete behaviour.
func TestDeleteAgent_UnreferencedNeedsNoForce(t *testing.T) {
	h, s, _ := integrityHandler(t)
	lone, _ := s.CreateAgent(store.AgentDefinition{Name: "Loner"})

	req := httptest.NewRequest(http.MethodDelete, "/agents/"+lone.ID, nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
}

// TestDeadReferenceEndpoints: the Settings button's API pair — list finds the
// corpses a pre-integrity delete left behind, clean removes them.
func TestDeadReferenceEndpoints(t *testing.T) {
	h, s, agentID := integrityHandler(t)

	// Bypass the guard, as pre-integrity Magec did.
	if err := s.DeleteAgent(agentID); err != nil {
		t.Fatalf("DeleteAgent: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/integrity/dead-references", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want 200", rec.Code)
	}
	var list DeadReferencesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("unmarshal list: %v", err)
	}
	if len(list.References) != 2 {
		t.Fatalf("dead refs = %+v, want client + flow", list.References)
	}

	req = httptest.NewRequest(http.MethodPost, "/integrity/dead-references/clean", nil)
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("clean status = %d, want 200", rec.Code)
	}
	var cleaned CleanDeadReferencesResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &cleaned); err != nil {
		t.Fatalf("unmarshal clean: %v", err)
	}
	if cleaned.Removed != 2 {
		t.Fatalf("removed = %d, want 2", cleaned.Removed)
	}
	if dead := s.DeadReferences(); len(dead) != 0 {
		t.Fatalf("corpses survived the clean: %+v", dead)
	}
}
