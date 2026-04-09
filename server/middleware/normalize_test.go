package middleware

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSnakeToCamel(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"app_name", "appName"},
		{"user_id", "userId"},
		{"session_id", "sessionId"},
		{"new_message", "newMessage"},
		{"state_delta", "stateDelta"},
		{"inline_data", "inlineData"},
		{"function_call", "functionCall"},
		{"function_response", "functionResponse"},
		{"mime_type", "mimeType"},
		{"display_name", "displayName"},
		{"file_uri", "fileUri"},
		{"code_execution_result", "codeExecutionResult"},
		{"text", "text"},
		{"role", "role"},
		{"parts", "parts"},
		{"data", "data"},
		{"", ""},
	}
	for _, tc := range tests {
		got := snakeToCamel(tc.in)
		if got != tc.want {
			t.Errorf("snakeToCamel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestNormalizeJSON_TopLevel(t *testing.T) {
	input := `{"app_name":"agent1","user_id":"u1","session_id":"s1","new_message":{"role":"user","parts":[{"text":"hi"}]}}`
	out, changed := normalizeJSON([]byte(input))
	if !changed {
		t.Fatal("expected changed=true")
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid output JSON: %v", err)
	}
	for _, key := range []string{"appName", "userId", "sessionId", "newMessage"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing key %q in output", key)
		}
	}
	for _, key := range []string{"app_name", "user_id", "session_id", "new_message"} {
		if _, ok := m[key]; ok {
			t.Errorf("snake_case key %q should not be present", key)
		}
	}
}

func TestNormalizeJSON_Nested(t *testing.T) {
	input := `{"appName":"a","userId":"u","sessionId":"s","newMessage":{"role":"user","parts":[{"inline_data":{"mime_type":"image/png","data":"abc"}}]}}`
	out, changed := normalizeJSON([]byte(input))
	if !changed {
		t.Fatal("expected changed=true for nested snake_case")
	}
	s := string(out)
	assertContains(t, s, `"inlineData"`)
	assertContains(t, s, `"mimeType"`)
	assertNotContains(t, s, `"inline_data"`)
	assertNotContains(t, s, `"mime_type"`)
}

func TestNormalizeJSON_AlreadyCamel(t *testing.T) {
	input := `{"appName":"a","userId":"u","sessionId":"s","newMessage":{"role":"user","parts":[{"text":"hi"}]}}`
	_, changed := normalizeJSON([]byte(input))
	if changed {
		t.Fatal("expected changed=false for already-camelCase body")
	}
}

func TestNormalizeJSON_CamelTakesPrecedence(t *testing.T) {
	input := `{"app_name":"snake","appName":"camel","user_id":"u1"}`
	out, changed := normalizeJSON([]byte(input))
	if !changed {
		t.Fatal("expected changed=true")
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatalf("invalid output JSON: %v", err)
	}
	if v, _ := m["appName"].(string); v != "camel" {
		t.Errorf("appName = %q, want %q", v, "camel")
	}
	if v, _ := m["userId"].(string); v != "u1" {
		t.Errorf("userId = %q, want %q", v, "u1")
	}
}

func TestNormalizeJSON_InvalidJSON(t *testing.T) {
	input := `not json`
	out, changed := normalizeJSON([]byte(input))
	if changed {
		t.Fatal("expected changed=false for invalid JSON")
	}
	if string(out) != input {
		t.Errorf("expected original bytes returned")
	}
}

func TestNormalizeJSON_EmptyObject(t *testing.T) {
	_, changed := normalizeJSON([]byte(`{}`))
	if changed {
		t.Fatal("expected changed=false for empty object")
	}
}

func TestNormalizeJSON_DeepNesting(t *testing.T) {
	input := `{"new_message":{"role":"user","parts":[{"function_call":{"name":"test","partial_args":[{"will_continue":true}]}}]}}`
	out, changed := normalizeJSON([]byte(input))
	if !changed {
		t.Fatal("expected changed=true")
	}
	s := string(out)
	assertContains(t, s, `"functionCall"`)
	assertContains(t, s, `"partialArgs"`)
	assertContains(t, s, `"willContinue"`)
	assertNotContains(t, s, `"function_call"`)
	assertNotContains(t, s, `"partial_args"`)
	assertNotContains(t, s, `"will_continue"`)
}

func TestNormalizeJSON_StateDeltaKeysConverted(t *testing.T) {
	input := `{"app_name":"a","state_delta":{"my_custom_key":"val"}}`
	out, changed := normalizeJSON([]byte(input))
	if !changed {
		t.Fatal("expected changed=true")
	}
	s := string(out)
	assertContains(t, s, `"stateDelta"`)
	assertContains(t, s, `"myCustomKey"`)
}

func TestNormalizeJSON_MixedSnakeAndCamelNested(t *testing.T) {
	input := `{"appName":"a","new_message":{"role":"user","parts":[{"inline_data":{"mimeType":"image/png","data":"abc"}}]}}`
	out, changed := normalizeJSON([]byte(input))
	if !changed {
		t.Fatal("expected changed=true")
	}
	s := string(out)
	assertContains(t, s, `"newMessage"`)
	assertContains(t, s, `"inlineData"`)
	assertNotContains(t, s, `"new_message"`)
	assertNotContains(t, s, `"inline_data"`)
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected output to contain %q, got:\n%s", substr, s)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected output NOT to contain %q, got:\n%s", substr, s)
	}
}
