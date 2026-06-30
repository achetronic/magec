// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package discord

import (
	"github.com/achetronic/magec/server/clients"
)

type Provider struct{}

func init() {
	clients.Register(&Provider{})
}

func (p *Provider) Type() string        { return "discord" }
func (p *Provider) DisplayName() string { return "Discord" }

func (p *Provider) ConfigSchema() clients.Schema {
	return clients.Schema{
		"type": "object",
		"properties": clients.Schema{
			"botToken": clients.Schema{
				"type":          "string",
				"title":         "Bot Token",
				"minLength":     1,
				"x-format":      "password",
				"x-placeholder": "MTIz...",
			},
			"allowedUsers": clients.Schema{
				"type":          "array",
				"items":         clients.Schema{"type": "string"},
				"title":         "Allowed Users",
				"x-placeholder": "Comma-separated Discord user IDs",
			},
			"allowedChannels": clients.Schema{
				"type":          "array",
				"items":         clients.Schema{"type": "string"},
				"title":         "Allowed Channels",
				"x-placeholder": "Comma-separated Discord channel IDs",
			},
			"responseMode": clients.Schema{
				"type":    "string",
				"title":   "Response Mode",
				"default": "text",
				"enum":    []string{"text", "voice", "mirror", "both"},
			},
			"defaultAgent":       clients.DefaultAgentSchema(),
			"threadHistoryLimit": clients.ThreadHistoryLimitSchema(100),
		},
		"required": []string{"botToken"},
	}
}
