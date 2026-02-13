package cron

import (
	"github.com/achetronic/magec/server/client"
)

type Provider struct{}

func init() {
	client.Register(&Provider{})
}

func (p *Provider) Type() string        { return "cron" }
func (p *Provider) DisplayName() string { return "Cron" }

func (p *Provider) ConfigSchema() client.Schema {
	return client.Schema{
		"type": "object",
		"properties": client.Schema{
			"schedule": client.Schema{
				"type":          "string",
				"title":         "Schedule",
				"x-placeholder": "0 9 * * *",
				"description":   "Standard cron expression (min hour day month weekday)",
			},
			"commandId": client.Schema{
				"type":     "string",
				"title":    "Command",
				"x-entity": "commands",
			},
		},
		"required": []string{"schedule", "commandId"},
	}
}
