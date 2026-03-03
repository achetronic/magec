package admin

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/achetronic/magec/server/store"
	"github.com/mymmrac/telego"
)

type fakeTelegramSender struct {
	sent     []*telego.SendMessageParams
	failures map[string]error
}

func (f *fakeTelegramSender) SendMessage(_ context.Context, params *telego.SendMessageParams) (*telego.Message, error) {
	f.sent = append(f.sent, params)
	key := targetKey(params.ChatID.ID, params.MessageThreadID)
	if err, ok := f.failures[key]; ok {
		return nil, err
	}
	return &telego.Message{}, nil
}

func TestBuildTelegramTestTargets_MixedAndDeduplicated(t *testing.T) {
	thread := 12
	cfg := &store.TelegramClientConfig{
		AllowedUsers: []int64{1001, 1001, 1002},
		AllowedChatThreads: store.TelegramChatThreadRules{
			{ChatID: -1001234567890},
			{ChatID: -1001234567890, ThreadID: &thread},
			{ChatID: -1001234567890, ThreadID: &thread},
		},
	}

	targets := buildTelegramTestTargets(cfg)
	if len(targets) != 4 {
		t.Fatalf("expected 4 unique targets, got %d", len(targets))
	}

	got := map[string]bool{}
	for _, target := range targets {
		got[targetKey(target.ChatID, target.ThreadID)] = true
	}
	for _, key := range []string{
		targetKey(1001, 0),
		targetKey(1002, 0),
		targetKey(-1001234567890, 0),
		targetKey(-1001234567890, 12),
	} {
		if !got[key] {
			t.Fatalf("expected target %s not found", key)
		}
	}
}

func TestBuildTelegramTestTargets_NilConfig(t *testing.T) {
	targets := buildTelegramTestTargets(nil)
	if len(targets) != 0 {
		t.Fatalf("expected no targets for nil config, got %d", len(targets))
	}
}

func TestSendTelegramConfigTest_AllSuccess(t *testing.T) {
	thread := 44
	cfg := &store.TelegramClientConfig{
		AllowedUsers: []int64{7},
		AllowedChatThreads: store.TelegramChatThreadRules{
			{ChatID: -1001001001, ThreadID: &thread},
		},
	}

	sender := &fakeTelegramSender{}
	result := sendTelegramConfigTest(context.Background(), sender, cfg)

	if result.Attempted != 2 || result.Sent != 2 || result.Failed != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(sender.sent) != 2 {
		t.Fatalf("expected 2 send calls, got %d", len(sender.sent))
	}
	for _, msg := range sender.sent {
		if msg.Text != telegramConfigTestMessage {
			t.Fatalf("unexpected test message text: %q", msg.Text)
		}
	}
}

func TestSendTelegramConfigTest_PartialFailures(t *testing.T) {
	cfg := &store.TelegramClientConfig{
		AllowedUsers: []int64{10, 11},
	}

	sender := &fakeTelegramSender{
		failures: map[string]error{
			targetKey(11, 0): errors.New("forbidden"),
		},
	}
	result := sendTelegramConfigTest(context.Background(), sender, cfg)

	if result.Attempted != 2 || result.Sent != 1 || result.Failed != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if len(result.Errors) != 1 {
		t.Fatalf("expected one error detail, got %d", len(result.Errors))
	}
}

func TestSendTelegramConfigTest_NoTargets(t *testing.T) {
	result := sendTelegramConfigTest(context.Background(), &fakeTelegramSender{}, &store.TelegramClientConfig{})
	if result.Attempted != 0 || result.Sent != 0 || result.Failed != 0 {
		t.Fatalf("expected empty result for no targets, got %+v", result)
	}
}

func TestResolveTelegramTestConfig_UsesOverrideTargetsWithoutSave(t *testing.T) {
	stored := &store.TelegramClientConfig{
		BotToken:     "stored-token",
		AllowedUsers: []int64{1},
	}
	override := &store.TelegramClientConfig{
		BotToken:     "override-token",
		AllowedUsers: []int64{99},
	}

	resolved := resolveTelegramTestConfig(stored, override)
	if resolved.BotToken != "override-token" {
		t.Fatalf("expected override token, got %q", resolved.BotToken)
	}
	if len(resolved.AllowedUsers) != 1 || resolved.AllowedUsers[0] != 99 {
		t.Fatalf("expected override allowedUsers, got %+v", resolved.AllowedUsers)
	}
}

func TestResolveTelegramTestConfig_NilOverrideReturnsStored(t *testing.T) {
	stored := &store.TelegramClientConfig{BotToken: "stored-token"}
	resolved := resolveTelegramTestConfig(stored, nil)
	if resolved != stored {
		t.Fatalf("expected stored config pointer when override is nil")
	}
}

func TestResolveTelegramTestConfig_UsesCurrentOverrideWithoutFallback(t *testing.T) {
	stored := &store.TelegramClientConfig{
		BotToken:     "stored-token",
		AllowedUsers: []int64{1},
	}
	override := &store.TelegramClientConfig{
		AllowedUsers: []int64{42},
	}

	resolved := resolveTelegramTestConfig(stored, override)
	if resolved.BotToken != "" {
		t.Fatalf("expected no token fallback, got %q", resolved.BotToken)
	}
	if len(resolved.AllowedUsers) != 1 || resolved.AllowedUsers[0] != 42 {
		t.Fatalf("expected override targets, got %+v", resolved.AllowedUsers)
	}
}

func TestResolveTelegramTestConfig_ExpandsEnvPlaceholderInOverrideToken(t *testing.T) {
	const key = "MAGEC_TEST_TELEGRAM_TOKEN"
	const value = "token-from-secret"
	t.Setenv(key, value)

	override := &store.TelegramClientConfig{
		BotToken: "${" + key + "}",
	}

	resolved := resolveTelegramTestConfig(nil, override)
	if resolved.BotToken != value {
		t.Fatalf("expected expanded token %q, got %q", value, resolved.BotToken)
	}
}

func TestResolveTelegramTestConfig_UsesEmptyTokenWhenEnvMissing(t *testing.T) {
	const key = "MAGEC_TEST_TELEGRAM_MISSING_TOKEN"
	_ = os.Unsetenv(key)

	override := &store.TelegramClientConfig{
		BotToken: "${" + key + "}",
	}

	resolved := resolveTelegramTestConfig(nil, override)
	if resolved.BotToken != "" {
		t.Fatalf("expected empty token for missing env var, got %q", resolved.BotToken)
	}
}

func TestValidateClientConfig_TelegramRejectsZeroChatID(t *testing.T) {
	c := store.ClientDefinition{
		Type: "telegram",
		Config: store.ClientConfig{
			Telegram: &store.TelegramClientConfig{
				BotToken: "token",
				AllowedChatThreads: store.TelegramChatThreadRules{
					{ChatID: 0},
				},
			},
		},
	}

	err := validateClientConfig(c)
	if err == nil {
		t.Fatalf("expected validation error for chatId=0")
	}
}

func TestValidateClientConfig_TelegramRejectsNonPositiveThreadID(t *testing.T) {
	threadID := 0
	c := store.ClientDefinition{
		Type: "telegram",
		Config: store.ClientConfig{
			Telegram: &store.TelegramClientConfig{
				BotToken: "token",
				AllowedChatThreads: store.TelegramChatThreadRules{
					{ChatID: -1001234567890, ThreadID: &threadID},
				},
			},
		},
	}

	err := validateClientConfig(c)
	if err == nil {
		t.Fatalf("expected validation error for threadId<=0")
	}
}

func TestValidateClientConfig_TelegramAcceptsValidChatThreadRules(t *testing.T) {
	threadID := 12
	c := store.ClientDefinition{
		Type: "telegram",
		Config: store.ClientConfig{
			Telegram: &store.TelegramClientConfig{
				BotToken: "token",
				AllowedChatThreads: store.TelegramChatThreadRules{
					{ChatID: -1001234567890, ThreadID: &threadID},
					{ChatID: -1001234567891},
				},
			},
		},
	}

	if err := validateClientConfig(c); err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
}
