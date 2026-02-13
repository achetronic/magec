package clients

// Schema is a JSON Schema object represented as a plain map so providers can
// define arbitrarily complex schemas (oneOf, if/then, etc.) without the
// framework imposing structural limits.
type Schema = map[string]interface{}

// Provider defines what every client type must implement.
// To add a new client type (e.g. Discord, Slack):
//  1. Create a new package under server/clients/<name>/
//  2. Implement the Provider interface
//  3. Call clients.Register() in an init() function
//  4. Add a blank import in main.go
type Provider interface {
	Type() string
	DisplayName() string
	ConfigSchema() Schema
}
