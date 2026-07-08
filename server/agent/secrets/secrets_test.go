// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package secrets

import (
	"testing"

	"github.com/achetronic/magec/server/store"
)

func snapshot() *Snapshot {
	return NewSnapshot([]store.Secret{
		{Key: "API_TOKEN", Value: "tok-supersecret-123"},
		{Key: "DB_PASS", Value: "hunter2-longenough"},
		{Key: "SHORTY", Value: "1234"},
	})
}

// TestExpandString: allowed placeholders expand, unknown and disallowed stay.
func TestExpandString(t *testing.T) {
	s := snapshot()
	allowed := s.Map([]string{"API_TOKEN"}, false)

	got := s.ExpandString("header {{secret:API_TOKEN}} and {{ secret:API_TOKEN }}", allowed)
	if got != "header tok-supersecret-123 and tok-supersecret-123" {
		t.Fatalf("expand = %q", got)
	}
	if got := s.ExpandString("{{secret:DB_PASS}}", allowed); got != "{{secret:DB_PASS}}" {
		t.Fatalf("disallowed key expanded: %q", got)
	}
	if got := s.ExpandString("{{secret:NOPE}}", allowed); got != "{{secret:NOPE}}" {
		t.Fatalf("unknown key mangled: %q", got)
	}
}

// TestExpandValue: nested maps and slices expand into copies, input untouched.
func TestExpandValue(t *testing.T) {
	s := snapshot()
	allowed := s.Map(nil, true)

	in := map[string]any{
		"url":  "https://x",
		"auth": "Bearer {{secret:API_TOKEN}}",
		"nested": []any{
			map[string]any{"pass": "{{secret:DB_PASS}}"},
		},
	}
	out := s.ExpandValue(in, allowed).(map[string]any)
	if out["auth"] != "Bearer tok-supersecret-123" {
		t.Fatalf("auth = %v", out["auth"])
	}
	if in["auth"] != "Bearer {{secret:API_TOKEN}}" {
		t.Fatalf("input mutated: %v", in["auth"])
	}
	nested := out["nested"].([]any)[0].(map[string]any)
	if nested["pass"] != "hunter2-longenough" {
		t.Fatalf("nested = %v", nested["pass"])
	}
}

// TestRedact: known values become placeholders, short values never redact.
func TestRedact(t *testing.T) {
	s := snapshot()

	got := s.RedactString("log: tok-supersecret-123 seen with hunter2-longenough")
	if got != "log: {{secret:API_TOKEN}} seen with {{secret:DB_PASS}}" {
		t.Fatalf("redact = %q", got)
	}
	if got := s.RedactString("pin 1234 stays"); got != "pin 1234 stays" {
		t.Fatalf("short value redacted: %q", got)
	}

	v := s.RedactValue(map[string]any{"out": []any{"key=tok-supersecret-123"}}).(map[string]any)
	if v["out"].([]any)[0] != "key={{secret:API_TOKEN}}" {
		t.Fatalf("nested redact = %v", v)
	}
}

// TestRoundTrip: expansion followed by redaction restores the placeholder.
func TestRoundTrip(t *testing.T) {
	s := snapshot()
	allowed := s.Map(nil, true)

	expanded := s.ExpandString("x {{secret:API_TOKEN}} y", allowed)
	if got := s.RedactString(expanded); got != "x {{secret:API_TOKEN}} y" {
		t.Fatalf("round trip = %q", got)
	}
}
