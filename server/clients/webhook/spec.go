package webhook

import (
	"github.com/achetronic/magec/server/clients"
)

type Provider struct{}

func init() {
	clients.Register(&Provider{})
}

func (p *Provider) Type() string        { return "webhook" }
func (p *Provider) DisplayName() string { return "Webhook" }

func (p *Provider) ConfigSchema() clients.Schema {
	return clients.Schema{
		"type": "object",
		"properties": clients.Schema{
			"passthrough": clients.Schema{
				"type":    "boolean",
				"title":   "Passthrough",
				"default": false,
				"description": "When enabled, the prompt comes from the webhook request body instead of a command.",
			},
			"commandId": clients.Schema{
				"type":      "string",
				"title":     "Command",
				"minLength": 1,
				"x-entity":  "commands",
			},
		},
		"oneOf": []clients.Schema{
			{
				"properties": clients.Schema{
					"passthrough": clients.Schema{"const": false},
				},
				"required": []string{"commandId"},
			},
			{
				"properties": clients.Schema{
					"passthrough": clients.Schema{"const": true},
				},
			},
		},
	}
}
