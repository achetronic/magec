package mcp

import (
	"context"
	"encoding/json"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSecretTools_ValueRedacted(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()

	_, created, err := h.createSecret(ctx, &sdk.CallToolRequest{}, createSecretInput{
		Name:  "OpenAI Key",
		Key:   "OPENAI_API_KEY",
		Value: "sk-secret-do-not-leak",
	})
	if err != nil {
		t.Fatalf("create secret: %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty id")
	}
	// SecretResponse intentionally has no Value field; JSON encoding must
	// never include it.
	raw, _ := json.Marshal(created)
	if got := string(raw); contains(got, "sk-secret") {
		t.Fatalf("secret value leaked in create response: %s", got)
	}

	_, got, err := h.getSecret(ctx, &sdk.CallToolRequest{}, idInput{ID: created.ID})
	if err != nil {
		t.Fatalf("get secret: %v", err)
	}
	raw, _ = json.Marshal(got)
	if s := string(raw); contains(s, "sk-secret") {
		t.Fatalf("secret value leaked in get response: %s", s)
	}

	_, list, err := h.listSecrets(ctx, &sdk.CallToolRequest{}, struct{}{})
	if err != nil {
		t.Fatalf("list secrets: %v", err)
	}
	raw, _ = json.Marshal(list)
	if s := string(raw); contains(s, "sk-secret") {
		t.Fatalf("secret value leaked in list response: %s", s)
	}
}

func TestSecret_ValidationErrors(t *testing.T) {
	h := newTestHandler(t)
	ctx := context.Background()
	cases := []createSecretInput{
		{Name: "", Key: "K", Value: "v"},
		{Name: "n", Key: "", Value: "v"},
		{Name: "n", Key: "K", Value: ""},
	}
	for i, in := range cases {
		if _, _, err := h.createSecret(ctx, &sdk.CallToolRequest{}, in); err == nil || !IsValidation(err) {
			t.Fatalf("case %d: expected validation error, got %v", i, err)
		}
	}
}

func contains(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
