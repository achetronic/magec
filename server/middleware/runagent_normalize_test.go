package middleware

import (
	"encoding/json"
	"testing"
)

func TestNormalizeRunAgentJSONBody_SnakeToCamel(t *testing.T) {
	in := []byte(`{"app_name":"agent-1","user_id":"u","session_id":"sess","new_message":{"role":"user","parts":[{"text":"hi"}]}}`)
	out, changed, err := normalizeRunAgentJSONBody(in)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed")
	}
	var m map[string]interface{}
	if err := json.Unmarshal(out, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["app_name"]; ok {
		t.Fatal("app_name should be removed")
	}
	if m["appName"] != "agent-1" {
		t.Fatalf("appName: got %v", m["appName"])
	}
	if m["userId"] != "u" {
		t.Fatalf("userId: got %v", m["userId"])
	}
	if m["sessionId"] != "sess" {
		t.Fatalf("sessionId: got %v", m["sessionId"])
	}
}

func TestNormalizeRunAgentJSONBody_CamelPreferred(t *testing.T) {
	in := []byte(`{"appName":"from-camel","app_name":"from-snake","userId":"u","sessionId":"s","newMessage":{"role":"user","parts":[]}}`)
	out, changed, err := normalizeRunAgentJSONBody(in)
	if err != nil {
		t.Fatal(err)
	}
	if !changed {
		t.Fatal("expected changed (snake keys dropped)")
	}
	var m map[string]interface{}
	json.Unmarshal(out, &m)
	if m["appName"] != "from-camel" {
		t.Fatalf("want camel appName preserved, got %v", m["appName"])
	}
}

func TestNormalizeRunAgentJSONBody_AlreadyCamel(t *testing.T) {
	in := []byte(`{"appName":"a","userId":"u","sessionId":"s","newMessage":{}}`)
	out, changed, err := normalizeRunAgentJSONBody(in)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Fatalf("should not change, got %s", string(out))
	}
}
