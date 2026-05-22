// Package mcp exposes the Magec admin API as MCP tools over Streamable HTTP.
//
// The catalogue is built at startup by reading the admin swagger spec
// (swag.ReadDoc), converting it from swagger 2.0 to OpenAPI 3.0, and feeding
// it to github.com/achetronic/openapi2tools. Each operation becomes one tool
// whose handler proxies the call back to the admin REST API as an HTTP
// request — same validation, same redaction, same conversation logging as a
// direct admin client.
//
// Adding a new admin endpoint (a new @Router annotation) automatically grows
// the MCP catalogue on the next build; the MCP package itself only contains
// the wiring, never per-resource tool definitions.
//
// See decision #30 in .agents/DECISIONS.md and the public docs at
// website/content/docs/admin-mcp-server.md.
package mcp
