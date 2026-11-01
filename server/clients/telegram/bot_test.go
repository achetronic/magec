// SPDX-License-Identifier: Apache-2.0

package telegram

import (
	"io"
	"log/slog"
	"testing"

	"github.com/achetronic/magec/server/store"
)

// newGateClient builds the smallest Client that groupGate needs: the
// Telegram config it reads, a bot username to match mentions against,
// and a logger that discards output so an unknown-policy warning does
// not pollute the test run.
func newGateClient(policy string) *Client {
	return &Client{
		botUsername: "mybot",
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		clientDef: store.ClientDefinition{
			Config: store.ClientConfig{
				Telegram: &store.TelegramClientConfig{GroupPolicy: policy},
			},
		},
	}
}

func TestGroupGate(t *testing.T) {
	cases := []struct {
		name       string
		policy     string
		chatType   string
		text       string
		wantText   string
		wantAnswer bool
	}{
		// A private chat is never gated, whatever the policy says.
		// Someone writing to the bot directly is addressing it.
		{"private chat, mention policy", "mention", "private",
			"what is the deadline?", "what is the deadline?", true},
		{"private chat, observe policy", "observe", "private",
			"what is the deadline?", "what is the deadline?", true},

		// The default, and the behaviour of every version before this
		// setting existed. An empty policy must behave identically to
		// an explicit "mention", or upgrading would silently change how
		// an existing bot behaves.
		{"mention: named", "mention", "group",
			"@mybot what is the deadline?", "what is the deadline?", true},
		{"mention: not named", "mention", "group",
			"we should meet on Tuesday", "we should meet on Tuesday", false},
		{"empty policy behaves as mention", "", "group",
			"we should meet on Tuesday", "we should meet on Tuesday", false},
		{"empty policy, named", "", "supergroup",
			"@mybot hello", "hello", true},

		{"all: answers without being named", "all", "group",
			"we should meet on Tuesday", "we should meet on Tuesday", true},
		{"all: still strips the mention", "all", "group",
			"@mybot hello", "hello", true},

		// The point of "observe": the message is received and returned
		// as context, but nothing is said. Receiving is not answering.
		{"observe: stays silent even when named", "observe", "group",
			"@mybot hello", "hello", false},
		{"observe: stays silent on chatter", "observe", "supergroup",
			"we should meet on Tuesday", "we should meet on Tuesday", false},

		// A bot that speaks because its configuration was not understood
		// is worse than one that stays quiet.
		{"unknown policy falls back to mention, not named", "nonsense",
			"group", "we should meet", "we should meet", false},
		{"unknown policy falls back to mention, named", "nonsense",
			"group", "@mybot ping", "ping", true},

		// Supergroups are gated like groups; forgetting one of the two
		// is the easy mistake here.
		{"supergroup is gated like a group", "mention", "supergroup",
			"we should meet", "we should meet", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := newGateClient(tc.policy)
			text, answer := c.groupGate(tc.chatType, tc.text)
			if answer != tc.wantAnswer {
				t.Errorf("answer = %v, want %v", answer, tc.wantAnswer)
			}
			if text != tc.wantText {
				t.Errorf("text = %q, want %q", text, tc.wantText)
			}
		})
	}
}

// TestGroupGateNilConfig guards the case where the Telegram config is
// missing entirely. groupGate must not panic: a nil pointer here would
// take down the whole bot on the first group message.
func TestGroupGateNilConfig(t *testing.T) {
	c := &Client{
		botUsername: "mybot",
		logger:      slog.New(slog.NewTextHandler(io.Discard, nil)),
		clientDef:   store.ClientDefinition{Config: store.ClientConfig{}},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("groupGate panicked with a nil Telegram config: %v", r)
		}
	}()

	if _, answer := c.groupGate("group", "hello"); answer {
		t.Error("with no config, an unnamed group message must not be answered")
	}
}
