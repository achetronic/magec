package cron

import (
	"github.com/achetronic/magec/server/clients"
)

type Provider struct{}

func init() {
	clients.Register(&Provider{})
}

func (p *Provider) Type() string        { return "cron" }
func (p *Provider) DisplayName() string { return "Cron" }

func (p *Provider) ConfigSchema() clients.Schema {
	return clients.Schema{
		"type": "object",
		"properties": clients.Schema{
			"schedule": clients.Schema{
				"type":          "string",
				"title":         "Schedule",
				"minLength":     1,
				"x-placeholder": "0 9 * * *",
				"description":   "Standard cron expression (min hour day month weekday)",
			},
			"commandId": clients.Schema{
				"type":      "string",
				"title":     "Command",
				"minLength": 1,
				"x-entity":  "commands",
			},
		},
		"required": []string{"schedule", "commandId"},
	}
}
