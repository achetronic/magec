package store

import (
	"encoding/json"
	"testing"
)

func TestTelegramAllowedChatRules_UnmarshalStructured(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChats":[{"chatId":-1001234567890,"threadId":12},{"chatId":-1001234567891}]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChats) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.AllowedChats))
	}
	if cfg.AllowedChats[0].ChatID != -1001234567890 {
		t.Fatalf("unexpected chatId at index 0: %d", cfg.AllowedChats[0].ChatID)
	}
	if cfg.AllowedChats[0].ThreadID == nil || *cfg.AllowedChats[0].ThreadID != 12 {
		t.Fatalf("expected threadId 12 at index 0")
	}
	if cfg.AllowedChats[1].ChatID != -1001234567891 {
		t.Fatalf("unexpected chatId at index 1: %d", cfg.AllowedChats[1].ChatID)
	}
	if cfg.AllowedChats[1].ThreadID != nil {
		t.Fatalf("expected nil threadId at index 1")
	}
}

func TestTelegramAllowedChatRules_UnmarshalLegacyIntArray(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChats":[-1001234567890,-1001234567891]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChats) != 2 {
		t.Fatalf("expected 2 migrated rules, got %d", len(cfg.AllowedChats))
	}
	if cfg.AllowedChats[0].ChatID != -1001234567890 || cfg.AllowedChats[0].ThreadID != nil {
		t.Fatalf("unexpected first migrated rule: %+v", cfg.AllowedChats[0])
	}
	if cfg.AllowedChats[1].ChatID != -1001234567891 || cfg.AllowedChats[1].ThreadID != nil {
		t.Fatalf("unexpected second migrated rule: %+v", cfg.AllowedChats[1])
	}
}

func TestTelegramAllowedChatRules_UnmarshalInvalidArrayType(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChats":["invalid"]}`), &cfg)
	if err == nil {
		t.Fatalf("expected unmarshal error for invalid allowedChats payload")
	}
}
