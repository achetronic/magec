//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package store

import "fmt"

// Reference kinds. A membership reference is a list entry or an optional
// pointer whose removal leaves the referrer perfectly valid (an agent with one
// skill less, a client with one allowed agent less). A structural reference is
// load-bearing: scrubbing it leaves the referrer broken or invalid (a flow
// node without its agent, an agent without its LLM backend, a cron client
// without its command).
const (
	RefMembership = "membership"
	RefStructural = "structural"
)

// Reference describes one entity referencing another. It is the unit the
// delete guard returns in 409 responses so the UI can show a demolition
// quote before a force delete.
type Reference struct {
	// ReferrerType is the entity kind holding the reference: agent, client,
	// flow, memoryProvider or settings.
	ReferrerType string `json:"referrerType"`
	ReferrerID   string `json:"referrerId,omitempty"`
	ReferrerName string `json:"referrerName"`
	// Field names where inside the referrer the reference lives, in a
	// human-readable form ("allowedAgents", "llm.backend", "node bender_pros").
	Field string `json:"field"`
	// Kind is RefMembership or RefStructural.
	Kind string `json:"kind"`
}

// DeadReference is a Reference whose target no longer exists in the store.
type DeadReference struct {
	Reference
	TargetID string `json:"targetId"`
}

// refEntry pairs a reference with the ID it points at. collectRefs emits one
// entry per reference slot; Referrers and DeadReferences filter the list.
type refEntry struct {
	target string
	ref    Reference
}

// collectRefs walks every cross-entity reference slot in the store data and
// returns one entry per non-empty reference. It is the single source of truth
// for the reference graph: Referrers, DeadReferences and scrubData must stay
// in lockstep with it.
func collectRefs(d *StoreData) []refEntry {
	var out []refEntry
	add := func(target string, refType, refID, refName, field, kind string) {
		if target == "" {
			return
		}
		out = append(out, refEntry{target: target, ref: Reference{
			ReferrerType: refType,
			ReferrerID:   refID,
			ReferrerName: refName,
			Field:        field,
			Kind:         kind,
		}})
	}

	for i := range d.Agents {
		a := &d.Agents[i]
		add(a.LLM.Backend, "agent", a.ID, a.Name, "llm.backend", RefStructural)
		add(a.Transcription.Backend, "agent", a.ID, a.Name, "transcription.backend", RefMembership)
		add(a.TTS.Backend, "agent", a.ID, a.Name, "tts.backend", RefMembership)
		for _, mcpID := range a.MCPServers {
			add(mcpID, "agent", a.ID, a.Name, "mcpServers", RefMembership)
		}
		for _, skillID := range a.Skills {
			add(skillID, "agent", a.ID, a.Name, "skills", RefMembership)
		}
	}

	for i := range d.MemoryProviders {
		m := &d.MemoryProviders[i]
		if m.Embedding != nil {
			add(m.Embedding.Backend, "memoryProvider", m.ID, m.Name, "embedding.backend", RefStructural)
		}
	}

	for i := range d.Clients {
		c := &d.Clients[i]
		for _, agentID := range c.AllowedAgents {
			add(agentID, "client", c.ID, c.Name, "allowedAgents", RefMembership)
		}
		if c.Config.Telegram != nil {
			add(c.Config.Telegram.DefaultAgent, "client", c.ID, c.Name, "config.telegram.defaultAgent", RefMembership)
		}
		if c.Config.Discord != nil {
			add(c.Config.Discord.DefaultAgent, "client", c.ID, c.Name, "config.discord.defaultAgent", RefMembership)
		}
		if c.Config.Slack != nil {
			add(c.Config.Slack.DefaultAgent, "client", c.ID, c.Name, "config.slack.defaultAgent", RefMembership)
		}
		if c.Config.Cron != nil {
			add(c.Config.Cron.CommandID, "client", c.ID, c.Name, "config.cron.commandId", RefStructural)
		}
		if c.Config.Webhook != nil {
			add(c.Config.Webhook.CommandID, "client", c.ID, c.Name, "config.webhook.commandId", RefStructural)
		}
	}

	for i := range d.Flows {
		f := &d.Flows[i]
		for j := range f.Nodes {
			n := &f.Nodes[j]
			add(n.AgentID, "flow", f.ID, f.Name, "node "+n.ID, RefStructural)
			add(n.FlowID, "flow", f.ID, f.Name, "node "+n.ID, RefStructural)
		}
	}

	// Session/long-term providers fall back to in-memory when cleared, so
	// these references are membership despite living on global settings.
	add(d.Settings.SessionProvider, "settings", "", "Global settings", "settings.sessionProvider", RefMembership)
	add(d.Settings.LongTermProvider, "settings", "", "Global settings", "settings.longTermProvider", RefMembership)

	return out
}

// existingIDs returns the set of every entity ID present in the store data.
func existingIDs(d *StoreData) map[string]bool {
	ids := make(map[string]bool)
	for i := range d.Backends {
		ids[d.Backends[i].ID] = true
	}
	for i := range d.MemoryProviders {
		ids[d.MemoryProviders[i].ID] = true
	}
	for i := range d.MCPServers {
		ids[d.MCPServers[i].ID] = true
	}
	for i := range d.Skills {
		ids[d.Skills[i].ID] = true
	}
	for i := range d.Agents {
		ids[d.Agents[i].ID] = true
	}
	for i := range d.Clients {
		ids[d.Clients[i].ID] = true
	}
	for i := range d.Flows {
		ids[d.Flows[i].ID] = true
	}
	for i := range d.Commands {
		ids[d.Commands[i].ID] = true
	}
	return ids
}

// Referrers returns every reference pointing at the given entity ID. An empty
// result means the entity can be deleted without touching anything else.
func (s *Store) Referrers(id string) []Reference {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var refs []Reference
	for _, e := range collectRefs(&s.data) {
		if e.target == id {
			refs = append(refs, e.ref)
		}
	}
	return refs
}

// DeadReferences returns every reference whose target no longer exists in the
// store: the corpses left behind by deletes performed before referential
// integrity existed (or by force deletes interrupted halfway).
func (s *Store) DeadReferences() []DeadReference {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := existingIDs(&s.data)
	var dead []DeadReference
	for _, e := range collectRefs(&s.data) {
		if !ids[e.target] {
			dead = append(dead, DeadReference{Reference: e.ref, TargetID: e.target})
		}
	}
	return dead
}

// ScrubReferences removes every reference to the given entity ID from both
// the expanded and raw store copies in one transaction: list entries are
// filtered out, optional backend/default fields are cleared, and flow nodes
// pointing at the entity are removed together with their edges (the entry
// pointer is cleared when its node goes). Returns how many references were
// scrubbed. Flows left invalid stay saved: editor validation and the builder
// report them loudly.
func (s *Store) ScrubReferences(id string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := 0
	for _, e := range collectRefs(&s.data) {
		if e.target == id {
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	scrubData(&s.data, id)
	scrubData(&s.rawData, id)
	return count, s.persist()
}

// CleanDeadReferences scrubs every dead reference found in the store in one
// transaction and returns how many were removed.
func (s *Store) CleanDeadReferences() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	ids := existingIDs(&s.data)
	targets := map[string]bool{}
	count := 0
	for _, e := range collectRefs(&s.data) {
		if !ids[e.target] {
			targets[e.target] = true
			count++
		}
	}
	if count == 0 {
		return 0, nil
	}
	for target := range targets {
		scrubData(&s.data, target)
		scrubData(&s.rawData, target)
	}
	return count, s.persist()
}

// scrubData removes every reference to id from one copy of the store data.
// It must cover exactly the slots collectRefs walks; both are the reference
// graph's two halves.
func scrubData(d *StoreData, id string) {
	for i := range d.Agents {
		a := &d.Agents[i]
		if a.LLM.Backend == id {
			a.LLM = BackendRef{}
		}
		if a.Transcription.Backend == id {
			a.Transcription = BackendRef{}
		}
		if a.TTS.Backend == id {
			a.TTS = TTSRef{}
		}
		a.MCPServers = filterOut(a.MCPServers, id)
		a.Skills = filterOut(a.Skills, id)
	}

	for i := range d.MemoryProviders {
		m := &d.MemoryProviders[i]
		if m.Embedding != nil && m.Embedding.Backend == id {
			m.Embedding = nil
		}
	}

	for i := range d.Clients {
		c := &d.Clients[i]
		c.AllowedAgents = filterOut(c.AllowedAgents, id)
		if c.Config.Telegram != nil && c.Config.Telegram.DefaultAgent == id {
			c.Config.Telegram.DefaultAgent = ""
		}
		if c.Config.Discord != nil && c.Config.Discord.DefaultAgent == id {
			c.Config.Discord.DefaultAgent = ""
		}
		if c.Config.Slack != nil && c.Config.Slack.DefaultAgent == id {
			c.Config.Slack.DefaultAgent = ""
		}
		if c.Config.Cron != nil && c.Config.Cron.CommandID == id {
			c.Config.Cron.CommandID = ""
		}
		if c.Config.Webhook != nil && c.Config.Webhook.CommandID == id {
			c.Config.Webhook.CommandID = ""
		}
	}

	for i := range d.Flows {
		scrubFlow(&d.Flows[i], id)
	}

	if d.Settings.SessionProvider == id {
		d.Settings.SessionProvider = ""
	}
	if d.Settings.LongTermProvider == id {
		d.Settings.LongTermProvider = ""
	}
}

// scrubFlow removes from one flow every node referencing id (as agent or
// subflow), the edges touching those nodes, and the entry pointer when its
// node was removed.
func scrubFlow(f *FlowDefinition, id string) {
	removed := map[string]bool{}
	kept := f.Nodes[:0]
	for i := range f.Nodes {
		n := f.Nodes[i]
		if n.AgentID == id || n.FlowID == id {
			removed[n.ID] = true
			continue
		}
		kept = append(kept, n)
	}
	if len(removed) == 0 {
		return
	}
	f.Nodes = kept

	keptEdges := f.Edges[:0]
	for _, e := range f.Edges {
		if removed[e.From] || removed[e.To] {
			continue
		}
		keptEdges = append(keptEdges, e)
	}
	f.Edges = keptEdges

	if removed[f.Entry] {
		f.Entry = ""
	}
}

// filterOut returns the slice without every occurrence of id, preserving
// order. Returns the original slice untouched when id is absent.
func filterOut(list []string, id string) []string {
	found := false
	for _, v := range list {
		if v == id {
			found = true
			break
		}
	}
	if !found {
		return list
	}
	out := make([]string, 0, len(list)-1)
	for _, v := range list {
		if v != id {
			out = append(out, v)
		}
	}
	return out
}

// String implements fmt.Stringer for logs and error messages.
func (r Reference) String() string {
	return fmt.Sprintf("%s %q (%s, %s)", r.ReferrerType, r.ReferrerName, r.Field, r.Kind)
}
