// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package secrets

import (
	"google.golang.org/adk/v2/agent"
	"google.golang.org/adk/v2/model"
	"google.golang.org/adk/v2/plugin"
	"google.golang.org/genai"
)

// Plugin returns the model-boundary guard: it redacts every known secret
// value from the request right before it reaches the LLM, whatever path the
// value took to get there (flow node outputs, session history, tool echoes).
// Contents are rewritten in place: parts about to be sent must never carry
// raw secrets anywhere else either.
func (s *Snapshot) Plugin() (*plugin.Plugin, error) {
	return plugin.New(plugin.Config{
		Name: "secretguard",
		BeforeModelCallback: func(ctx agent.Context, req *model.LLMRequest) (*model.LLMResponse, error) {
			for _, content := range req.Contents {
				s.redactContent(content)
			}
			if req.Config != nil && req.Config.SystemInstruction != nil {
				s.redactContent(req.Config.SystemInstruction)
			}
			return nil, nil
		},
	})
}

// redactContent rewrites the text, function call arguments and function
// response payloads of one content block.
func (s *Snapshot) redactContent(content *genai.Content) {
	if content == nil {
		return
	}
	for _, part := range content.Parts {
		if part == nil {
			continue
		}
		if part.Text != "" {
			part.Text = s.RedactString(part.Text)
		}
		if part.FunctionCall != nil && part.FunctionCall.Args != nil {
			part.FunctionCall.Args, _ = s.RedactValue(part.FunctionCall.Args).(map[string]any)
		}
		if part.FunctionResponse != nil && part.FunctionResponse.Response != nil {
			part.FunctionResponse.Response, _ = s.RedactValue(part.FunctionResponse.Response).(map[string]any)
		}
	}
}
