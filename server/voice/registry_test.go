package voice

import (
	"context"
	"testing"

	"github.com/achetronic/magec/server/store"
)

type mockProvider struct {
	typ         string
	displayName string
	tts         bool
	stt         bool
}

func (m *mockProvider) Type() string            { return m.typ }
func (m *mockProvider) DisplayName() string     { return m.displayName }
func (m *mockProvider) SupportsTTS() bool       { return m.tts }
func (m *mockProvider) SupportsSTT() bool       { return m.stt }
func (m *mockProvider) TTSConfigSchema() Schema { return nil }
func (m *mockProvider) STTConfigSchema() Schema { return nil }
func (m *mockProvider) SynthesizeSpeech(_ context.Context, _ TTSRequest, _ store.BackendDefinition) (*TTSResponse, error) {
	return nil, nil
}
func (m *mockProvider) TranscribeAudio(_ context.Context, _ STTRequest, _ store.BackendDefinition) (string, error) {
	return "", nil
}

func TestRegisterAndGet(t *testing.T) {
	mu.Lock()
	saved := providers
	providers = map[string]Provider{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		providers = saved
		mu.Unlock()
	}()

	p := &mockProvider{typ: "test", displayName: "Test", tts: true, stt: true}
	Register(p)

	got := Get("test")
	if got == nil {
		t.Fatal("expected provider, got nil")
	}
	if got.Type() != "test" {
		t.Errorf("Type() = %q, want %q", got.Type(), "test")
	}
	if got.DisplayName() != "Test" {
		t.Errorf("DisplayName() = %q, want %q", got.DisplayName(), "Test")
	}
	if !got.SupportsTTS() {
		t.Error("expected SupportsTTS() = true")
	}
	if !got.SupportsSTT() {
		t.Error("expected SupportsSTT() = true")
	}
}

func TestGetUnknown(t *testing.T) {
	got := Get("nonexistent_type_xyz")
	if got != nil {
		t.Errorf("expected nil for unknown type, got %v", got)
	}
}

func TestAllSorted(t *testing.T) {
	mu.Lock()
	saved := providers
	providers = map[string]Provider{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		providers = saved
		mu.Unlock()
	}()

	Register(&mockProvider{typ: "zebra", displayName: "Zebra"})
	Register(&mockProvider{typ: "alpha", displayName: "Alpha"})

	all := All()
	if len(all) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(all))
	}
	if all[0].Type() != "alpha" {
		t.Errorf("first provider = %q, want %q", all[0].Type(), "alpha")
	}
	if all[1].Type() != "zebra" {
		t.Errorf("second provider = %q, want %q", all[1].Type(), "zebra")
	}
}

func TestRegisterDuplicatePanics(t *testing.T) {
	mu.Lock()
	saved := providers
	providers = map[string]Provider{}
	mu.Unlock()
	defer func() {
		mu.Lock()
		providers = saved
		mu.Unlock()
	}()

	Register(&mockProvider{typ: "dup", displayName: "First"})

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic on duplicate registration")
		}
	}()
	Register(&mockProvider{typ: "dup", displayName: "Second"})
}
