package device

import (
	"github.com/achetronic/magec/server/client"
)

type Provider struct{}

func init() {
	client.Register(&Provider{})
}

func (p *Provider) Type() string        { return "device" }
func (p *Provider) DisplayName() string { return "Device" }

func (p *Provider) ConfigFields() []client.FieldSpec {
	return []client.FieldSpec{}
}
