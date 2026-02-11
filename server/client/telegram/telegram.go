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

func (p *Provider) ConfigFields() []client.FieldSpec {
	return []client.FieldSpec{
		{Key: "botToken", Label: "Bot Token", Type: "password", Required: true, Placeholder: "123456:ABC-DEF..."},
		{Key: "allowedUsers", Label: "Allowed Users", Type: "text", Placeholder: "Comma-separated Telegram user IDs"},
		{Key: "allowedChats", Label: "Allowed Chats", Type: "text", Placeholder: "Comma-separated Telegram chat IDs"},
		{Key: "responseMode", Label: "Response Mode", Type: "select", Default: "text", Options: "text,voice,mirror,both"},
	}
}
