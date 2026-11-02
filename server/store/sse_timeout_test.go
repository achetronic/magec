// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package store

import (
	"testing"
	"time"
)

// TestClientConfigSSETimeout verifies the single fallback point for the SSE
// timeout: each client type reads its own field, and an unset (0), invalid
// (negative) or absent config resolves to the default.
func TestClientConfigSSETimeout(t *testing.T) {
	def := time.Duration(DefaultSSETimeoutMinutes) * time.Minute
	cases := []struct {
		name string
		cfg  ClientConfig
		want time.Duration
	}{
		{"no config (direct client, open mode)", ClientConfig{}, def},
		{"telegram unset", ClientConfig{Telegram: &TelegramClientConfig{}}, def},
		{"telegram negative", ClientConfig{Telegram: &TelegramClientConfig{SSEClientTimeoutMinutes: -5}}, def},
		{"telegram set", ClientConfig{Telegram: &TelegramClientConfig{SSEClientTimeoutMinutes: 120}}, 120 * time.Minute},
		{"discord set", ClientConfig{Discord: &DiscordClientConfig{SSEClientTimeoutMinutes: 30}}, 30 * time.Minute},
		{"slack set", ClientConfig{Slack: &SlackClientConfig{SSEClientTimeoutMinutes: 45}}, 45 * time.Minute},
		{"cron set", ClientConfig{Cron: &CronClientConfig{SSEClientTimeoutMinutes: 60}}, 60 * time.Minute},
		{"webhook set", ClientConfig{Webhook: &WebhookClientConfig{SSEClientTimeoutMinutes: 1}}, time.Minute},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.cfg.SSETimeout(); got != tc.want {
				t.Fatalf("SSETimeout() = %v, want %v", got, tc.want)
			}
		})
	}
}
