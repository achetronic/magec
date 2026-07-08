// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package agent

import (
	"testing"

	"github.com/achetronic/magec/server/agent/flowexit"
)

// TestRenderTemplate_SecretRoot: {{ secret.KEY }} resolves, unknown keys render empty.
func TestRenderTemplate_SecretRoot(t *testing.T) {
	secret := map[string]string{"API_TOKEN": "tok-123"}

	got := renderTemplate("auth: {{ secret.API_TOKEN }}", "in", nil, secret)
	if got != "auth: tok-123" {
		t.Fatalf("render = %q", got)
	}
	if got := renderTemplate("[{{ secret.NOPE }}]", "in", nil, secret); got != "[]" {
		t.Fatalf("unknown secret should render empty, got %q", got)
	}
	if got := renderTemplate("[{{ secret.API_TOKEN.deep }}]", "in", nil, secret); got != "[]" {
		t.Fatalf("secrets have no nested fields, got %q", got)
	}
}

// TestEvaluateValue_SecretVariable: CEL expressions read secret.KEY.
func TestEvaluateValue_SecretVariable(t *testing.T) {
	prog, err := flowexit.CompileValue(`"Bearer " + secret.API_TOKEN`)
	if err != nil {
		t.Fatalf("CompileValue: %v", err)
	}
	out, err := flowexit.EvaluateValue(prog, "", "in", map[string]any{}, map[string]string{"API_TOKEN": "tok-123"})
	if err != nil {
		t.Fatalf("EvaluateValue: %v", err)
	}
	if out != "Bearer tok-123" {
		t.Fatalf("out = %v", out)
	}
}
