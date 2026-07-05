// Copyright 2025 Alby Hernández
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
