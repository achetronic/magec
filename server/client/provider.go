package client

// FieldSpec describes a single configuration field for a client type.
// The admin UI uses this to dynamically render forms.
type FieldSpec struct {
	Key         string `json:"key"`
	Label       string `json:"label"`
	Type        string `json:"type"`
	Required    bool   `json:"required,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Default     string `json:"default,omitempty"`
	Options     string `json:"options,omitempty"`
}

// Provider defines what every client type must implement.
// To add a new client type (e.g. Discord, Slack):
//  1. Create a new package under server/client/<name>/
//  2. Implement the Provider interface
//  3. Call client.Register() in an init() function
//  4. Add a blank import in main.go
type Provider interface {
	Type() string
	DisplayName() string
	ConfigFields() []FieldSpec
}
