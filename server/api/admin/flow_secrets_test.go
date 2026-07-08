// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"path/filepath"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// TestValidateFlowSecrets: unknown keys rejected per language, known ones pass.
func TestValidateFlowSecrets(t *testing.T) {
	s, err := store.New(filepath.Join(t.TempDir(), "store.json"), "")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	if _, err := s.CreateSecret(store.Secret{Name: "token", Key: "API_TOKEN", Value: "v"}); err != nil {
		t.Fatalf("CreateSecret: %v", err)
	}
	h := New(s)

	flow := func(n store.FlowNode) *store.FlowDefinition {
		return &store.FlowDefinition{Name: "f", Entry: n.ID, Nodes: []store.FlowNode{n}}
	}

	cases := []struct {
		name    string
		node    store.FlowNode
		wantErr bool
	}{
		{"template known", store.FlowNode{ID: "t", Type: store.FlowNodeTemplate, Template: "x {{ secret.API_TOKEN }}"}, false},
		{"template unknown", store.FlowNode{ID: "t", Type: store.FlowNodeTemplate, Template: "{{ secret.NOPE }}"}, true},
		{"cel known", store.FlowNode{ID: "e", Type: store.FlowNodeExpression, Expression: `"B " + secret.API_TOKEN`}, false},
		{"cel unknown", store.FlowNode{ID: "e", Type: store.FlowNodeExpression, Expression: `secret.NOPE`}, true},
		{"starlark known", store.FlowNode{ID: "c", Type: store.FlowNodeCode, Script: `output = secret["API_TOKEN"]`}, false},
		{"starlark unknown", store.FlowNode{ID: "c", Type: store.FlowNodeCode, Script: `output = secret["NOPE"]`}, true},
		{"no refs", store.FlowNode{ID: "t", Type: store.FlowNodeTemplate, Template: "plain {{ input }}"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := h.validateFlowSecrets(flow(tc.node))
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr = %v", err, tc.wantErr)
			}
		})
	}
}
