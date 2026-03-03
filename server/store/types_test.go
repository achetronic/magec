package store

import (
	"encoding/json"
	"testing"
)

func TestTelegramChatThreadRules_UnmarshalStructured(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChatThreads":[{"chatId":-1001234567890,"threadId":12},{"chatId":-1001234567891}]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChatThreads) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.AllowedChatThreads))
	}
	if cfg.AllowedChatThreads[0].ChatID != -1001234567890 {
		t.Fatalf("unexpected chatId at index 0: %d", cfg.AllowedChatThreads[0].ChatID)
	}
	if cfg.AllowedChatThreads[0].ThreadID == nil || *cfg.AllowedChatThreads[0].ThreadID != 12 {
		t.Fatalf("expected threadId 12 at index 0")
	}
	if cfg.AllowedChatThreads[1].ChatID != -1001234567891 {
		t.Fatalf("unexpected chatId at index 1: %d", cfg.AllowedChatThreads[1].ChatID)
	}
	if cfg.AllowedChatThreads[1].ThreadID != nil {
		t.Fatalf("expected nil threadId at index 1")
	}
}

func TestTelegramChatThreadRules_UnmarshalLegacyStrings(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChatThreads":["-1001234567890-12","-1001234567891"]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChatThreads) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.AllowedChatThreads))
	}
	if cfg.AllowedChatThreads[0].ChatID != -1001234567890 {
		t.Fatalf("unexpected chatId at index 0: %d", cfg.AllowedChatThreads[0].ChatID)
	}
	if cfg.AllowedChatThreads[0].ThreadID == nil || *cfg.AllowedChatThreads[0].ThreadID != 12 {
		t.Fatalf("expected threadId 12 at index 0")
	}
	if cfg.AllowedChatThreads[1].ChatID != -1001234567891 {
		t.Fatalf("unexpected chatId at index 1: %d", cfg.AllowedChatThreads[1].ChatID)
	}
	if cfg.AllowedChatThreads[1].ThreadID != nil {
		t.Fatalf("expected nil threadId at index 1")
	}
}

func TestTelegramChatThreadRules_UnmarshalLegacyWithWhitespace(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChatThreads":[" -1001234567890 - 12 "," -1001234567891 "]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChatThreads) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(cfg.AllowedChatThreads))
	}
	if cfg.AllowedChatThreads[0].ThreadID == nil || *cfg.AllowedChatThreads[0].ThreadID != 12 {
		t.Fatalf("expected threadId 12 at index 0")
	}
	if cfg.AllowedChatThreads[1].ThreadID != nil {
		t.Fatalf("expected nil threadId at index 1")
	}
}

func TestTelegramChatThreadRules_UnmarshalInvalidLegacyValue(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChatThreads":["invalid"]}`), &cfg)
	if err == nil {
		t.Fatalf("expected unmarshal error for invalid legacy value")
	}
}

func TestTelegramChatThreadRules_UnmarshalInvalidLegacyNegativeThread(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChatThreads":["-1001234567890--1"]}`), &cfg)
	if err == nil {
		t.Fatalf("expected unmarshal error for negative threadId")
	}
}

func TestTelegramClientConfig_UnmarshalLegacyAllowedChats(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChats":[-1001234567890,-1001234567891]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChatThreads) != 2 {
		t.Fatalf("expected 2 migrated rules, got %d", len(cfg.AllowedChatThreads))
	}
	if cfg.AllowedChatThreads[0].ChatID != -1001234567890 || cfg.AllowedChatThreads[0].ThreadID != nil {
		t.Fatalf("unexpected first migrated rule: %+v", cfg.AllowedChatThreads[0])
	}
	if cfg.AllowedChatThreads[1].ChatID != -1001234567891 || cfg.AllowedChatThreads[1].ThreadID != nil {
		t.Fatalf("unexpected second migrated rule: %+v", cfg.AllowedChatThreads[1])
	}
}

func TestTelegramClientConfig_UnmarshalPrefersAllowedChatThreadsOverLegacyAllowedChats(t *testing.T) {
	var cfg TelegramClientConfig
	err := json.Unmarshal([]byte(`{"allowedChats":[-1001234567890],"allowedChatThreads":[{"chatId":-1009999999999,"threadId":55}]}`), &cfg)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if len(cfg.AllowedChatThreads) != 1 {
		t.Fatalf("expected 1 explicit rule, got %d", len(cfg.AllowedChatThreads))
	}
	if cfg.AllowedChatThreads[0].ChatID != -1009999999999 {
		t.Fatalf("expected explicit chatId to win, got %d", cfg.AllowedChatThreads[0].ChatID)
	}
	if cfg.AllowedChatThreads[0].ThreadID == nil || *cfg.AllowedChatThreads[0].ThreadID != 55 {
		t.Fatalf("expected explicit threadId to win")
	}
}
