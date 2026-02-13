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
				"type":          "string",
				"title":         "Allowed Users",
				"x-placeholder": "Comma-separated Telegram user IDs",
			},
			"allowedChats": clients.Schema{
				"type":          "string",
				"title":         "Allowed Chats",
				"x-placeholder": "Comma-separated Telegram chat IDs",
			},
			"responseMode": clients.Schema{
				"type":    "string",
				"title":   "Response Mode",
				"default": "text",
				"enum":    []string{"text", "voice", "mirror", "both"},
			},
		},
		"required": []string{"botToken"},
	}
}
