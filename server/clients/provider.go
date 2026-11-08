// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package clients

import "github.com/achetronic/magec/server/store"

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

// DefaultAgentSchema is the shared JSON Schema fragment for the defaultAgent
// field, common to all bot clients (Discord, Slack, Telegram).
func DefaultAgentSchema() Schema {
	return Schema{
		"type":        "string",
		"title":       "Default Agent",
		"description": "Agent ID to use on startup. Automatically updated when a user runs !agent <id>.",
	}
}

// ThreadHistoryLimitSchema returns the shared JSON Schema fragment for the
// threadHistoryLimit field. maxItems is platform-specific (100 for Discord,
// 1000 for Slack).
func ThreadHistoryLimitSchema(max int) Schema {
	return Schema{
		"type":        "integer",
		"title":       "Thread History Messages",
		"description": "Number of previous thread messages passed to the agent as context. Use lower values for smaller models.",
		"default":     50,
		"minimum":     1,
		"maximum":     max,
	}
}

// HideThinkingSchema returns the shared JSON Schema fragment for the
// hideThinking field, common to all bot clients.
func HideThinkingSchema() Schema {
	return Schema{
		"type":        "boolean",
		"title":       "Hide Thinking Messages",
		"description": "Hide the model's thinking messages in chat. Change it at runtime with the thinking command.",
		"default":     false,
	}
}

// SSEClientTimeoutSchema returns the shared JSON Schema fragment for the
// sseClientTimeoutMinutes field.
func SSEClientTimeoutSchema() Schema {
	return Schema{
		"type":        "integer",
		"title":       "Response Timeout (minutes)",
		"description": "How long to wait for the agent's reply before the request is cancelled. Raise it for slow models or long tasks.",
		"default":     store.DefaultSSETimeoutMinutes,
		"minimum":     1,
	}
}
