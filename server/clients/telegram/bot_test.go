package telegram

import (
	"testing"

	"github.com/achetronic/magec/server/store"
)

func TestIsAllowed_NoRulesConfigured_Allows(t *testing.T) {
	c := &Client{clientDef: store.ClientDefinition{Config: store.ClientConfig{Telegram: &store.TelegramClientConfig{}}}}
	if !c.isAllowed(1001, -1001234567890, 0) {
		t.Fatalf("expected access allowed when no rules are configured")
	}
}

func TestIsAllowed_AllowedUsers_Allows(t *testing.T) {
	c := &Client{clientDef: store.ClientDefinition{Config: store.ClientConfig{Telegram: &store.TelegramClientConfig{AllowedUsers: []int64{42}}}}}
	if !c.isAllowed(42, -1001234567890, 99) {
		t.Fatalf("expected access allowed for allowed user")
	}
}

func TestIsAllowed_UsersConfigured_DeniesUnknownUserWithoutChatRules(t *testing.T) {
	c := &Client{clientDef: store.ClientDefinition{Config: store.ClientConfig{Telegram: &store.TelegramClientConfig{AllowedUsers: []int64{42}}}}}
	if c.isAllowed(99, -1001234567890, 0) {
		t.Fatalf("expected access denied for user not in allowedUsers")
	}
}

func TestIsAllowed_ChatOnlyRule_AllowsAnyThread(t *testing.T) {
	c := &Client{clientDef: store.ClientDefinition{Config: store.ClientConfig{Telegram: &store.TelegramClientConfig{AllowedChatThreads: store.TelegramChatThreadRules{{ChatID: -1001234567890}}}}}}
	if !c.isAllowed(5, -1001234567890, 0) {
		t.Fatalf("expected access allowed for matching chat without thread")
	}
	if !c.isAllowed(5, -1001234567890, 77) {
		t.Fatalf("expected access allowed for matching chat with any thread")
	}
}

func TestIsAllowed_ChatAndThreadRule_RestrictsByThread(t *testing.T) {
	thread := 12
	c := &Client{clientDef: store.ClientDefinition{Config: store.ClientConfig{Telegram: &store.TelegramClientConfig{AllowedChatThreads: store.TelegramChatThreadRules{{ChatID: -1001234567890, ThreadID: &thread}}}}}}
	if !c.isAllowed(5, -1001234567890, 12) {
		t.Fatalf("expected access allowed for matching chat and thread")
	}
	if c.isAllowed(5, -1001234567890, 13) {
		t.Fatalf("expected access denied for non-matching thread")
	}
	if c.isAllowed(5, -1001234567890, 0) {
		t.Fatalf("expected access denied when required thread is missing")
	}
}

func TestIsAllowed_NonMatchingRules_Denies(t *testing.T) {
	c := &Client{clientDef: store.ClientDefinition{Config: store.ClientConfig{Telegram: &store.TelegramClientConfig{AllowedChatThreads: store.TelegramChatThreadRules{{ChatID: -1001234567890}}}}}}
	if c.isAllowed(5, -1000000000000, 0) {
		t.Fatalf("expected access denied when chat does not match")
	}
}
