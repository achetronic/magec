// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package msgutil

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestParseSSEStream_LargeToolResponse: a multi-megabyte data: line parses instead of dropping the stream.
func TestParseSSEStream_LargeToolResponse(t *testing.T) {
	payload := map[string]any{
		"author": "agent_1",
		"content": map[string]any{
			"parts": []any{
				map[string]any{"functionResponse": map[string]any{
					"name":     "get_logs",
					"response": strings.Repeat("x", 5*1024*1024),
				}},
			},
		},
	}
	line, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	body := "data: " + string(line) + "\n\ndata: {\"author\":\"agent_1\",\"turn_complete\":true,\"finish_reason\":\"STOP\"}\n"

	var got []SSEEvent
	if err := ParseSSEStream(strings.NewReader(body), func(e SSEEvent) { got = append(got, e) }); err != nil {
		t.Fatalf("ParseSSEStream: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("events = %d, want tool result + turn_complete", len(got))
	}
	if got[0].Type != SSEEventToolResult || got[0].ToolName != "get_logs" {
		t.Fatalf("first event = %+v, want the large tool result", got[0])
	}
	if s, ok := got[0].ToolResult.(string); !ok || len(s) != 5*1024*1024 {
		t.Fatalf("tool result truncated or mistyped")
	}
	if !got[1].TurnComplete {
		t.Fatalf("events after the large line were lost")
	}
}

// TestParseSSEStream_NoTrailingNewline: the last event survives a stream with no final newline.
func TestParseSSEStream_NoTrailingNewline(t *testing.T) {
	body := `data: {"author":"agent_1","content":{"parts":[{"text":"hi"}]}}`

	var got []SSEEvent
	if err := ParseSSEStream(strings.NewReader(body), func(e SSEEvent) { got = append(got, e) }); err != nil {
		t.Fatalf("ParseSSEStream: %v", err)
	}
	if len(got) != 1 || got[0].Type != SSEEventText || got[0].Text != "hi" {
		t.Fatalf("events = %+v, want the single text event", got)
	}
}

// TestParseSSEStream_HappyPath: text, comments, blanks, CRLF and ADK plain-text errors dispatch correctly.
func TestParseSSEStream_HappyPath(t *testing.T) {
	body := strings.Join([]string{
		`data: {"author":"agent_1","content":{"parts":[{"text":"hello "}]}}`,
		``,
		`: comment line`,
		"data: {\"author\":\"agent_1\",\"content\":{\"parts\":[{\"text\":\"world\"}]}}\r",
		`data: not json`,
		`Error while running agent: boom`,
		`data: {"author":"agent_1","turn_complete":true,"finish_reason":"STOP"}`,
	}, "\n") + "\n"

	events, text, err := CollectSSEEvents(strings.NewReader(body))
	if err != nil {
		t.Fatalf("CollectSSEEvents: %v", err)
	}
	if text != "hello world" {
		t.Fatalf("text = %q", text)
	}
	var kinds []SSEEventType
	for _, e := range events {
		kinds = append(kinds, e.Type)
	}
	want := []SSEEventType{SSEEventText, SSEEventText, SSEEventError, SSEEventUnknown}
	if len(kinds) != len(want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	for i := range want {
		if kinds[i] != want[i] {
			t.Fatalf("kinds = %v, want %v", kinds, want)
		}
	}
	if events[2].ErrorMessage != "boom" {
		t.Fatalf("ADK error frame lost: %+v", events[2])
	}
}
