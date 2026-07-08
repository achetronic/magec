// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package agent

import "testing"

func TestRenderTemplate(t *testing.T) {
	input := map[string]any{"name": "Ada", "score": 42}
	state := map[string]any{"lang": "es"}

	cases := []struct {
		name string
		tmpl string
		in   any
		want string
	}{
		{"plain input string", "Hello {{ input }}", "world", "Hello world"},
		{"input field", "Name: {{ input.name }}", input, "Name: Ada"},
		{"non-string field via %v", "Score: {{ input.score }}", input, "Score: 42"},
		{"state key", "Lang: {{ state.lang }}", "x", "Lang: es"},
		{"mix input and state", "{{ input.name }} / {{ state.lang }}", input, "Ada / es"},
		{"missing path renders empty", "[{{ input.missing }}]", input, "[]"},
		{"unknown root renders empty", "[{{ foo.bar }}]", input, "[]"},
		{"whitespace tolerated", "{{input}}!", "hi", "hi!"},
		{"no placeholders", "static text", "x", "static text"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := renderTemplate(tc.tmpl, tc.in, state, nil)
			if got != tc.want {
				t.Fatalf("renderTemplate(%q) = %q, want %q", tc.tmpl, got, tc.want)
			}
		})
	}
}

func TestStringify(t *testing.T) {
	if got := stringify(nil); got != "" {
		t.Fatalf("nil -> %q, want empty", got)
	}
	if got := stringify("hi"); got != "hi" {
		t.Fatalf("string -> %q", got)
	}
	if got := stringify([]any{"a", "b"}); got != `["a","b"]` {
		t.Fatalf("slice -> %q, want JSON", got)
	}
	if got := stringify(map[string]any{"k": "v"}); got != `{"k":"v"}` {
		t.Fatalf("map -> %q, want JSON", got)
	}
}
