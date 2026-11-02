// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package clients_test

import (
	"encoding/json"
	"testing"

	"github.com/achetronic/magec/server/clients"
	_ "github.com/achetronic/magec/server/clients/cron"
	_ "github.com/achetronic/magec/server/clients/discord"
	_ "github.com/achetronic/magec/server/clients/slack"
	_ "github.com/achetronic/magec/server/clients/telegram"
	_ "github.com/achetronic/magec/server/clients/webhook"
	"github.com/achetronic/magec/server/store"
)

// TestSSEClientTimeoutSchemaMatchesStore is the canary for the per-client SSE
// timeout: the property the Admin UI renders from each provider's schema must
// be the same key the store struct decodes, carry the shared default, and
// reject values below one minute. A renamed JSON tag or a provider that drops
// the shared fragment would otherwise let the UI save a timeout nobody reads.
func TestSSEClientTimeoutSchemaMatchesStore(t *testing.T) {
	cases := []struct {
		clientType string
		config     any
		required   map[string]any
	}{
		{"telegram", store.TelegramClientConfig{SSEClientTimeoutMinutes: 30}, map[string]any{"botToken": "t"}},
		{"discord", store.DiscordClientConfig{SSEClientTimeoutMinutes: 30}, map[string]any{"botToken": "t"}},
		{"slack", store.SlackClientConfig{SSEClientTimeoutMinutes: 30}, map[string]any{"botToken": "t", "appToken": "a"}},
		{"cron", store.CronClientConfig{SSEClientTimeoutMinutes: 30}, map[string]any{"schedule": "0 9 * * *", "commandId": "c"}},
		{"webhook", store.WebhookClientConfig{SSEClientTimeoutMinutes: 30}, map[string]any{"passthrough": true}},
	}
	for _, tc := range cases {
		t.Run(tc.clientType, func(t *testing.T) {
			p := clients.Get(tc.clientType)
			if p == nil {
				t.Fatalf("provider %q not registered", tc.clientType)
			}

			// Derive the key from the struct's JSON tag, not from a literal.
			raw, err := json.Marshal(tc.config)
			if err != nil {
				t.Fatal(err)
			}
			var fromStruct map[string]any
			if err := json.Unmarshal(raw, &fromStruct); err != nil {
				t.Fatal(err)
			}
			var key string
			for k, v := range fromStruct {
				if v == float64(30) {
					key = k
				}
			}
			if key == "" {
				t.Fatalf("store config did not serialise SSEClientTimeoutMinutes: %s", raw)
			}

			props, _ := p.ConfigSchema()["properties"].(clients.Schema)
			field, ok := props[key].(clients.Schema)
			if !ok {
				t.Fatalf("schema has no %q property; the UI would save a key the store ignores", key)
			}
			if field["default"] != store.DefaultSSETimeoutMinutes {
				t.Fatalf("schema default = %v, want %d", field["default"], store.DefaultSSETimeoutMinutes)
			}

			valid := map[string]any{key: 30}
			for k, v := range tc.required {
				valid[k] = v
			}
			if err := clients.ValidateConfig(tc.clientType, valid); err != nil {
				t.Fatalf("valid timeout rejected: %v", err)
			}
			valid[key] = 0
			if err := clients.ValidateConfig(tc.clientType, valid); err == nil {
				t.Fatalf("timeout 0 accepted; schema minimum must be 1")
			}
		})
	}
}
