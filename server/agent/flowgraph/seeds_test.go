// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package flowgraph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// TestSeedFlowsValidate guards the shipped example store: every flow in
// data/seeds/examples.json must pass graph validation and reference only
// agents defined in the same file, so a fresh install never boots with
// broken demo flows.
func TestSeedFlowsValidate(t *testing.T) {
	path := filepath.Join("..", "..", "..", "data", "seeds", "examples.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("seed file not present: %v", err)
	}

	var seed struct {
		Agents []store.AgentDefinition `json:"agents"`
		Flows  []store.FlowDefinition  `json:"flows"`
	}
	if err := json.Unmarshal(raw, &seed); err != nil {
		t.Fatalf("seed file does not parse: %v", err)
	}
	if len(seed.Flows) == 0 {
		t.Fatal("seed file defines no flows")
	}

	agentIDs := make(map[string]bool, len(seed.Agents))
	for _, a := range seed.Agents {
		agentIDs[a.ID] = true
	}

	for i := range seed.Flows {
		f := &seed.Flows[i]
		if err := Validate(f); err != nil {
			t.Errorf("seed flow %q fails validation: %v", f.Name, err)
		}
		for _, n := range f.Nodes {
			if n.AgentID != "" && !agentIDs[n.AgentID] {
				t.Errorf("seed flow %q node %q references agent %q, which the seed file does not define", f.Name, n.ID, n.AgentID)
			}
		}
	}
}
