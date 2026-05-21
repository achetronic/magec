// Package mcp exposes the Magec admin API as MCP tools over Streamable HTTP.
//
// The server is embedded inside magec-server and runs on its own port (default
// 8082) when server.mcp.enabled is true. Authentication reuses
// server.adminPassword as a bearer token, validated with constant-time
// comparison and a per-IP rate limiter (5 failures per minute), the same as
// the admin REST API.
//
// Tools call the data store directly. There is no HTTP roundtrip back to the
// admin port. The package mirrors the layout of server/api/admin:
//
//	handler.go            Handler container and registration aggregator
//	server.go             *mcp.Server construction and HTTP wiring
//	schemas.go            shared input/output structs
//	errors.go             small error helpers
//	tools_<resource>.go   one file per admin resource group
//
// See decision #30 in .agents/DECISIONS.md and the public docs at
// website/content/docs/admin-mcp-server.md.
package mcp
