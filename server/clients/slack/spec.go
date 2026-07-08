// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package slack

import (
	"github.com/achetronic/magec/server/clients"
)

type Provider struct{}

func init() {
	clients.Register(&Provider{})
}

func (p *Provider) Type() string        { return "slack" }
func (p *Provider) DisplayName() string { return "Slack" }

func (p *Provider) ConfigSchema() clients.Schema {
	return clients.Schema{
		"type": "object",
		"properties": clients.Schema{
			"botToken": clients.Schema{
				"type":          "string",
				"title":         "Bot Token",
				"minLength":     1,
				"x-format":      "password",
				"x-placeholder": "xoxb-...",
			},
			"appToken": clients.Schema{
				"type":          "string",
				"title":         "App Token (Socket Mode)",
				"minLength":     1,
				"x-format":      "password",
				"x-placeholder": "xapp-...",
			},
			"allowedUsers": clients.Schema{
				"type":          "array",
				"items":         clients.Schema{"type": "string"},
				"title":         "Allowed Users",
				"x-placeholder": "Comma-separated Slack user IDs (e.g. U01ABCDEF)",
			},
			"allowedChannels": clients.Schema{
				"type":          "array",
				"items":         clients.Schema{"type": "string"},
				"title":         "Allowed Channels",
				"x-placeholder": "Comma-separated Slack channel IDs (e.g. C01ABCDEF)",
			},
			"responseMode": clients.Schema{
				"type":    "string",
				"title":   "Response Mode",
				"default": "text",
				"enum":    []string{"text", "voice", "mirror", "both"},
			},
			"defaultAgent":       clients.DefaultAgentSchema(),
			"threadHistoryLimit": clients.ThreadHistoryLimitSchema(1000),
		},
		"required": []string{"botToken", "appToken"},
	}
}
