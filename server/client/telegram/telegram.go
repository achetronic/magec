package telegram

import (
	"github.com/achetronic/magec/server/client"
)

type Provider struct{}

func init() {
	client.Register(&Provider{})
}

func (p *Provider) Type() string        { return "telegram" }
func (p *Provider) DisplayName() string { return "Telegram" }

func (p *Provider) ConfigSchema() client.Schema {
	return client.Schema{
		"type": "object",
		"properties": client.Schema{
			"botToken": client.Schema{
				"type":          "string",
				"title":         "Bot Token",
				"x-format":      "password",
				"x-placeholder": "123456:ABC-DEF...",
			},
			"allowedUsers": client.Schema{
				"type":          "string",
				"title":         "Allowed Users",
				"x-placeholder": "Comma-separated Telegram user IDs",
			},
			"allowedChats": client.Schema{
				"type":          "string",
				"title":         "Allowed Chats",
				"x-placeholder": "Comma-separated Telegram chat IDs",
			},
			"responseMode": client.Schema{
				"type":    "string",
				"title":   "Response Mode",
				"default": "text",
				"enum":    []string{"text", "voice", "mirror", "both"},
			},
		},
		"required": []string{"botToken"},
	}
}
