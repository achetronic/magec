// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package agent

import "testing"

// dialectFor maps each known provider host to its adapter dialect; unknown
// hosts and OpenAI-pure servers get nil.
func TestDialectForHostMapping(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		wantName string // empty means nil dialect
	}{
		{name: "reflection", baseURL: "https://api.reflection.ai/openai/v1", wantName: "reflection"},
		{name: "deepseek", baseURL: "https://api.deepseek.com/v1", wantName: "deepseek"},
		{name: "openrouter", baseURL: "https://openrouter.ai/api/v1", wantName: "openrouter"},
		{name: "openai native", baseURL: "https://api.openai.com/v1", wantName: ""},
		{name: "ollama", baseURL: "http://ollama.ollama.svc:11434/v1", wantName: ""},
		{name: "malformed", baseURL: "://not a url", wantName: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dialectFor(tt.baseURL)
			if tt.wantName == "" {
				if d != nil {
					t.Fatalf("dialectFor(%q) = %v, want nil", tt.baseURL, d)
				}
				return
			}
			if d == nil {
				t.Fatalf("dialectFor(%q) = nil, want %q dialect", tt.baseURL, tt.wantName)
			}
			if got := d.Name(); got != tt.wantName {
				t.Fatalf("dialectFor(%q).Name() = %q, want %q", tt.baseURL, got, tt.wantName)
			}
		})
	}
}