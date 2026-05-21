package mcp

import (
	"context"
	"path/filepath"
	"testing"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/store"
)

func newTestHandler(t *testing.T) *Handler {
	t.Helper()
	s, err := store.New(filepath.Join(t.TempDir(), "store.json"), "")
	if err != nil {
		t.Fatalf("store.New: %v", err)
	}
	return NewHandler(s, nil, nil)
}

func TestListBackends_Empty(t *testing.T) {
	h := newTestHandler(t)
	_, out, err := h.listBackends(context.Background(), &sdk.CallToolRequest{}, struct{}{})
	if err != nil {
		t.Fatalf("listBackends: %v", err)
	}
	if len(out.Backends) != 0 {
		t.Fatalf("got %d backends, want 0", len(out.Backends))
	}
}

func TestCreateBackend(t *testing.T) {
	cases := []struct {
		name      string
		in        store.BackendDefinition
		wantErr   bool
		wantStore int
	}{
		{"valid", store.BackendDefinition{Name: "openai-1", Type: "openai"}, false, 1},
		{"missing name", store.BackendDefinition{Type: "openai"}, true, 0},
		{"missing type", store.BackendDefinition{Name: "x"}, true, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newTestHandler(t)
			_, out, err := h.createBackend(context.Background(), &sdk.CallToolRequest{}, createBackendInput{Definition: tc.in})
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
			if !tc.wantErr {
				if out.ID == "" {
					t.Fatal("expected non-empty id")
				}
				if err != nil && !IsValidation(err) {
					t.Fatalf("expected validation error, got %v", err)
				}
			}
			if got := len(h.store.ListRawBackends()); got != tc.wantStore {
				t.Fatalf("store size: got %d want %d", got, tc.wantStore)
			}
		})
	}
}

func TestUpdateAndDeleteBackend(t *testing.T) {
	h := newTestHandler(t)
	_, created, err := h.createBackend(context.Background(), &sdk.CallToolRequest{}, createBackendInput{
		Definition: store.BackendDefinition{Name: "openai-1", Type: "openai"},
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	_, updated, err := h.updateBackend(context.Background(), &sdk.CallToolRequest{}, updateBackendInput{
		ID:         created.ID,
		Definition: store.BackendDefinition{Name: "openai-renamed", Type: "openai"},
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "openai-renamed" {
		t.Fatalf("update did not persist; got name=%q", updated.Name)
	}

	if _, _, err := h.deleteBackend(context.Background(), &sdk.CallToolRequest{}, idInput{ID: created.ID}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if got := len(h.store.ListRawBackends()); got != 0 {
		t.Fatalf("after delete: %d backends remain", got)
	}

	if _, _, err := h.getBackend(context.Background(), &sdk.CallToolRequest{}, idInput{ID: created.ID}); err == nil {
		t.Fatal("expected error fetching deleted backend")
	}
}
