package direct

import (
	"github.com/achetronic/magec/server/client"
)

type Provider struct{}

func init() {
	client.Register(&Provider{})
}

func (p *Provider) Type() string        { return "direct" }
func (p *Provider) DisplayName() string { return "Direct" }

func (p *Provider) ConfigSchema() client.Schema {
	return client.Schema{
		"type":       "object",
		"properties": client.Schema{},
	}
}
