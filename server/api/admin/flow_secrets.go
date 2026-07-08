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

package admin

import (
	"fmt"
	"regexp"

	"github.com/achetronic/magec/server/store"
)

// Secret references per node language: {{ secret.KEY }} in templates,
// secret.KEY in CEL, secret["KEY"] in Starlark. Detection is textual and
// best-effort: it catches the typo'd key at save time instead of an empty
// value at run time.
var (
	templateSecretRE = regexp.MustCompile(`\{\{\s*secret\.([A-Za-z_][A-Za-z0-9_]*)\s*\}\}`)
	celSecretRE      = regexp.MustCompile(`\bsecret\.([A-Za-z_][A-Za-z0-9_]*)`)
	starlarkSecretRE = regexp.MustCompile(`\bsecret\[\s*["']([A-Za-z_][A-Za-z0-9_]*)["']\s*\]`)
)

// validateFlowSecrets rejects flow nodes referencing secret keys that do not
// exist in the store.
func (h *Handler) validateFlowSecrets(f *store.FlowDefinition) error {
	known := map[string]bool{}
	for _, sec := range h.store.ListSecrets() {
		known[sec.Key] = true
	}

	check := func(nodeID string, re *regexp.Regexp, text string) error {
		for _, m := range re.FindAllStringSubmatch(text, -1) {
			if !known[m[1]] {
				return fmt.Errorf("node %q references unknown secret %q", nodeID, m[1])
			}
		}
		return nil
	}

	for i := range f.Nodes {
		n := &f.Nodes[i]
		var err error
		switch n.Type {
		case store.FlowNodeTemplate:
			err = check(n.ID, templateSecretRE, n.Template)
		case store.FlowNodeExpression:
			err = check(n.ID, celSecretRE, n.Expression)
		case store.FlowNodeCode:
			err = check(n.ID, starlarkSecretRE, n.Script)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
