package webhook

import (
	"github.com/achetronic/magec/server/client"
)

type Provider struct{}

func init() {
	client.Register(&Provider{})
}

func (p *Provider) Type() string        { return "webhook" }
func (p *Provider) DisplayName() string { return "Webhook" }

func (p *Provider) ConfigSchema() client.Schema {
	return client.Schema{
		"type": "object",
		"properties": client.Schema{
			"passthrough": client.Schema{
				"type":    "boolean",
				"title":   "Passthrough",
				"default": false,
				"description": "When enabled, the prompt comes from the webhook request body instead of a command.",
			},
			"commandId": client.Schema{
				"type":     "string",
				"title":    "Command",
				"x-entity": "commands",
			},
		},
		"oneOf": []client.Schema{
			{
				"properties": client.Schema{
					"passthrough": client.Schema{"const": false},
				},
				"required": []string{"commandId"},
			},
			{
				"properties": client.Schema{
					"passthrough": client.Schema{"const": true},
				},
			},
		},
	}
}
