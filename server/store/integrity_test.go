//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package store

import (
	"path/filepath"
	"testing"
)

// integrityStore builds a store populated with one of everything, wired
// together so every reference slot in the graph is exercised at least once.
func integrityStore(t *testing.T) (*Store, map[string]string) {
	t.Helper()
	s, err := New(filepath.Join(t.TempDir(), "store.json"), "")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	backend, _ := s.CreateBackend(BackendDefinition{Name: "LLM Backend", Type: "openai"})
	ttsBackend, _ := s.CreateBackend(BackendDefinition{Name: "TTS Backend", Type: "openai"})
	memory, _ := s.CreateMemoryProvider(MemoryProvider{
		Name: "PG", Type: "postgres", Category: "semantic",
		Embedding: &BackendRef{Backend: backend.ID, Model: "embed"},
	})
	mcp, _ := s.CreateMCPServer(MCPServer{Name: "Tools", Endpoint: "http://x"})
	agent, _ := s.CreateAgent(AgentDefinition{
		Name:          "Agent",
		LLM:           BackendRef{Backend: backend.ID, Model: "gpt"},
		TTS:           TTSRef{Backend: ttsBackend.ID, Model: "tts-1", Voice: "alloy"},
		Transcription: BackendRef{Backend: ttsBackend.ID, Model: "whisper-1"},
		MCPServers:    []string{mcp.ID},
	})
	command, _ := s.CreateCommand(Command{Name: "Daily", Prompt: "do it"})
	client, _ := s.CreateClient(ClientDefinition{
		Name: "Bot", Type: "telegram",
		AllowedAgents: []string{agent.ID},
		Config: ClientConfig{
			Telegram: &TelegramClientConfig{BotToken: "t", DefaultAgent: agent.ID},
		},
	})
	cronClient, _ := s.CreateClient(ClientDefinition{
		Name: "Nightly", Type: "cron",
		AllowedAgents: []string{agent.ID},
		Config:        ClientConfig{Cron: &CronClientConfig{Schedule: "@daily", CommandID: command.ID}},
	})
	subflow, _ := s.CreateFlow(FlowDefinition{
		Name:  "Sub",
		Entry: "a1",
		Nodes: []FlowNode{{ID: "a1", Type: FlowNodeAgent, AgentID: agent.ID}},
	})
	flow, _ := s.CreateFlow(FlowDefinition{
		Name:  "Main",
		Entry: "a1",
		Nodes: []FlowNode{
			{ID: "a1", Type: FlowNodeAgent, AgentID: agent.ID},
			{ID: "sub1", Type: FlowNodeSubflow, FlowID: subflow.ID},
			{ID: "t1", Type: FlowNodeTemplate, Template: "hi {{ input }}"},
		},
		Edges: []FlowEdge{{From: "a1", To: "sub1"}, {From: "sub1", To: "t1"}},
	})
	if err := s.UpdateSettings(Settings{SessionProvider: memory.ID, LongTermProvider: memory.ID}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	return s, map[string]string{
		"backend": backend.ID, "ttsBackend": ttsBackend.ID, "memory": memory.ID,
		"mcp": mcp.ID, "agent": agent.ID, "command": command.ID,
		"client": client.ID, "cronClient": cronClient.ID,
		"subflow": subflow.ID, "flow": flow.ID,
	}
}

func findRefs(refs []Reference, referrerID string) []Reference {
	var out []Reference
	for _, r := range refs {
		if r.ReferrerID == referrerID {
			out = append(out, r)
		}
	}
	return out
}

func TestReferrers_Agent(t *testing.T) {
	s, ids := integrityStore(t)

	refs := s.Referrers(ids["agent"])
	// telegram client: allowedAgents + defaultAgent; cron client: allowedAgents;
	// main flow: node a1; subflow: node a1.
	if len(refs) != 5 {
		t.Fatalf("Referrers(agent) = %d refs, want 5: %+v", len(refs), refs)
	}
	if got := findRefs(refs, ids["client"]); len(got) != 2 {
		t.Fatalf("telegram client refs = %+v, want allowedAgents + defaultAgent", got)
	}
	for _, r := range findRefs(refs, ids["flow"]) {
		if r.Kind != RefStructural {
			t.Fatalf("flow node reference must be structural, got %+v", r)
		}
	}
}

func TestReferrers_KindClassification(t *testing.T) {
	s, ids := integrityStore(t)

	for _, r := range s.Referrers(ids["backend"]) {
		switch r.Field {
		case "llm.backend", "embedding.backend":
			if r.Kind != RefStructural {
				t.Errorf("%s must be structural, got %s", r.Field, r.Kind)
			}
		default:
			t.Errorf("unexpected field for backend: %+v", r)
		}
	}
	for _, r := range s.Referrers(ids["ttsBackend"]) {
		if r.Kind != RefMembership {
			t.Errorf("tts/transcription backend refs must be membership, got %+v", r)
		}
	}
	for _, r := range s.Referrers(ids["command"]) {
		if r.Kind != RefStructural {
			t.Errorf("cron commandId must be structural, got %+v", r)
		}
	}
}

func TestReferrers_NoneForUnreferenced(t *testing.T) {
	s, ids := integrityStore(t)
	if refs := s.Referrers(ids["client"]); len(refs) != 0 {
		t.Fatalf("nothing references clients, got %+v", refs)
	}
	if refs := s.Referrers(ids["flow"]); len(refs) != 0 {
		t.Fatalf("nothing references the top flow, got %+v", refs)
	}
}

func TestScrubReferences_Agent(t *testing.T) {
	s, ids := integrityStore(t)

	n, err := s.ScrubReferences(ids["agent"])
	if err != nil {
		t.Fatalf("ScrubReferences: %v", err)
	}
	if n != 5 {
		t.Fatalf("scrubbed %d refs, want 5", n)
	}

	cl, _ := s.GetClient(ids["client"])
	if len(cl.AllowedAgents) != 0 {
		t.Errorf("allowedAgents not scrubbed: %+v", cl.AllowedAgents)
	}
	if cl.Config.Telegram.DefaultAgent != "" {
		t.Errorf("defaultAgent not scrubbed: %q", cl.Config.Telegram.DefaultAgent)
	}

	f, _ := s.GetFlow(ids["flow"])
	for _, node := range f.Nodes {
		if node.AgentID == ids["agent"] {
			t.Errorf("agent node survived the scrub: %+v", node)
		}
	}
	// Node a1 was the entry and had edges: both must be gone.
	if f.Entry != "" {
		t.Errorf("entry should be cleared, got %q", f.Entry)
	}
	for _, e := range f.Edges {
		if e.From == "a1" || e.To == "a1" {
			t.Errorf("edge touching removed node survived: %+v", e)
		}
	}
	// The template node, untouched by the scrub, must survive.
	if len(f.Nodes) != 2 {
		t.Errorf("expected subflow + template nodes to survive, got %+v", f.Nodes)
	}

	if refs := s.Referrers(ids["agent"]); len(refs) != 0 {
		t.Fatalf("references survived the scrub: %+v", refs)
	}
}

func TestScrubReferences_Backend(t *testing.T) {
	s, ids := integrityStore(t)

	if _, err := s.ScrubReferences(ids["backend"]); err != nil {
		t.Fatalf("ScrubReferences: %v", err)
	}
	a, _ := s.GetAgent(ids["agent"])
	if a.LLM.Backend != "" || a.LLM.Model != "" {
		t.Errorf("llm ref not cleared: %+v", a.LLM)
	}
	if a.TTS.Backend == "" {
		t.Errorf("tts ref should be untouched, got %+v", a.TTS)
	}
	m, _ := s.GetMemoryProvider(ids["memory"])
	if m.Embedding != nil {
		t.Errorf("embedding ref not cleared: %+v", m.Embedding)
	}
}

func TestScrubReferences_PersistsToRawData(t *testing.T) {
	s, ids := integrityStore(t)

	if _, err := s.ScrubReferences(ids["agent"]); err != nil {
		t.Fatalf("ScrubReferences: %v", err)
	}
	// Raw (unexpanded) reads must reflect the scrub too: a stale rawData copy
	// would resurrect the references on the next raw round-trip.
	raw, ok := s.GetRawClient(ids["client"])
	if !ok {
		t.Fatalf("raw client missing")
	}
	if len(raw.AllowedAgents) != 0 {
		t.Fatalf("rawData not scrubbed: %+v", raw.AllowedAgents)
	}
}

func TestDeadReferences_FindsCorpses(t *testing.T) {
	s, ids := integrityStore(t)

	if dead := s.DeadReferences(); len(dead) != 0 {
		t.Fatalf("fresh store must have no dead refs, got %+v", dead)
	}

	// Delete the agent bypassing the guard, as pre-integrity Magec did.
	if err := s.DeleteAgent(ids["agent"]); err != nil {
		t.Fatalf("DeleteAgent: %v", err)
	}

	dead := s.DeadReferences()
	if len(dead) != 5 {
		t.Fatalf("dead refs = %d, want 5: %+v", len(dead), dead)
	}
	for _, d := range dead {
		if d.TargetID != ids["agent"] {
			t.Errorf("unexpected dead target %q", d.TargetID)
		}
	}
}

func TestCleanDeadReferences(t *testing.T) {
	s, ids := integrityStore(t)

	_ = s.DeleteAgent(ids["agent"])
	_ = s.DeleteBackend(ids["ttsBackend"])

	n, err := s.CleanDeadReferences()
	if err != nil {
		t.Fatalf("CleanDeadReferences: %v", err)
	}
	if n == 0 {
		t.Fatalf("expected corpses to be cleaned")
	}
	if dead := s.DeadReferences(); len(dead) != 0 {
		t.Fatalf("dead refs survived the clean: %+v", dead)
	}

	n2, err := s.CleanDeadReferences()
	if err != nil || n2 != 0 {
		t.Fatalf("second clean should be a no-op, got n=%d err=%v", n2, err)
	}
}
