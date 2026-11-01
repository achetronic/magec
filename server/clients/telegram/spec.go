// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package telegram

import (
	"github.com/achetronic/magec/server/clients"
)

type Provider struct{}

func init() {
	clients.Register(&Provider{})
}

func (p *Provider) Type() string        { return "telegram" }
func (p *Provider) DisplayName() string { return "Telegram" }

func (p *Provider) ConfigSchema() clients.Schema {
	return clients.Schema{
		"type": "object",
		"properties": clients.Schema{
			"botToken": clients.Schema{
				"type":          "string",
				"title":         "Bot Token",
				"minLength":     1,
				"x-format":      "password",
				"x-placeholder": "123456:ABC-DEF...",
			},
			"allowedUsers": clients.Schema{
				"type":          "array",
				"items":         clients.Schema{"type": "integer"},
				"title":         "Allowed Users",
				"x-placeholder": "Comma-separated Telegram user IDs",
			},
			"groupPolicy": clients.Schema{
				"type":    "string",
				"title":   "Group Policy",
				"enum":    []string{"mention", "all", "observe"},
				"default": "mention",
				"description": "When the bot answers in a group or supergroup. " +
					"'mention': only when it is named (the historical behaviour). " +
					"'all': every message. " +
					"'observe': never answers, but still receives the messages " +
					"and keeps them as context, so it can be asked later what " +
					"was said. Has no effect on private chats.",
			},
			"allowedChats": clients.Schema{
				"type":          "array",
				"items":         clients.Schema{"type": "integer"},
				"title":         "Allowed Chats",
				"x-placeholder": "Comma-separated Telegram chat IDs",
			},
			"responseMode": clients.Schema{
				"type":    "string",
				"title":   "Response Mode",
				"default": "text",
				"enum":    []string{"text", "voice", "mirror", "both"},
			},
			"defaultAgent": clients.DefaultAgentSchema(),
		},
		"required": []string{"botToken"},
	}
}
