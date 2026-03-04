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
			"allowedChats": clients.Schema{
				"type": "array",
				"items": clients.Schema{
					"type": "object",
					"properties": clients.Schema{
						"chatId": clients.Schema{
							"type":  "integer",
							"title": "Chat ID",
						},
						"threadId": clients.Schema{
							"type":  "integer",
							"title": "Thread ID",
						},
					},
					"required": []string{"chatId"},
				},
				"title":       "Allowed Chats",
				"description": "Allowed chat IDs with optional thread IDs (topics).",
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
