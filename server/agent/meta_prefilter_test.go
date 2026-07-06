// Copyright 2025 Alberto Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0

package agent

import (
	"reflect"
	"testing"
)

// TestExtractMagecMeta covers the shapes clients actually produce: the block
// prepended before the message, appended after it, absent, and malformed.
func TestExtractMagecMeta(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		wantText string
		wantMeta map[string]any
	}{
		{
			name:     "block before the message",
			text:     "<!--MAGEC_META:{\"source\":\"telegram\",\"telegram_user_id\":42}:MAGEC_META-->\nbuy milk",
			wantText: "buy milk",
			wantMeta: map[string]any{"source": "telegram", "telegram_user_id": float64(42)},
		},
		{
			name:     "block after the message",
			text:     "hello<!--MAGEC_META:{\"source\":\"discord\"}:MAGEC_META-->",
			wantText: "hello",
			wantMeta: map[string]any{"source": "discord"},
		},
		{
			name:     "no block",
			text:     "just words",
			wantText: "just words",
			wantMeta: nil,
		},
		{
			name:     "malformed json still stripped",
			text:     "<!--MAGEC_META:not-json:MAGEC_META-->\nquestion",
			wantText: "question",
			wantMeta: map[string]any{"raw": "not-json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotText, gotMeta := extractMagecMeta(tt.text)

			if gotText != tt.wantText {
				t.Fatalf("cleaned text = %q, want %q", gotText, tt.wantText)
			}
			if !reflect.DeepEqual(gotMeta, tt.wantMeta) {
				t.Fatalf("meta = %#v, want %#v", gotMeta, tt.wantMeta)
			}
		})
	}
}
