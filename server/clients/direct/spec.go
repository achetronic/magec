package direct

import (
	"github.com/achetronic/magec/server/clients"
)

type Provider struct{}

func init() {
	clients.Register(&Provider{})
}

func (p *Provider) Type() string        { return "direct" }
func (p *Provider) DisplayName() string { return "Direct" }

func (p *Provider) ConfigSchema() clients.Schema {
	return clients.Schema{
		"type":       "object",
		"properties": clients.Schema{},
	}
}
