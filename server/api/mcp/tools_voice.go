package mcp

import (
	"context"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/voice"
)

type voiceProviderInfo struct {
	Type            string       `json:"type"`
	DisplayName     string       `json:"displayName"`
	SupportsTTS     bool         `json:"supportsTts"`
	SupportsSTT     bool         `json:"supportsStt"`
	TTSConfigSchema voice.Schema `json:"ttsConfigSchema"`
	STTConfigSchema voice.Schema `json:"sttConfigSchema"`
}

type listVoiceTypesOutput struct {
	Types []voiceProviderInfo `json:"types"`
}

func (h *Handler) listVoiceTypes(_ context.Context, _ *sdk.CallToolRequest, _ struct{}) (*sdk.CallToolResult, listVoiceTypesOutput, error) {
	var types []voiceProviderInfo
	for _, p := range voice.All() {
		types = append(types, voiceProviderInfo{
			Type:            p.Type(),
			DisplayName:     p.DisplayName(),
			SupportsTTS:     p.SupportsTTS(),
			SupportsSTT:     p.SupportsSTT(),
			TTSConfigSchema: p.TTSConfigSchema(),
			STTConfigSchema: p.STTConfigSchema(),
		})
	}
	return nil, listVoiceTypesOutput{Types: types}, nil
}

func (h *Handler) registerVoiceTools() {
	sdk.AddTool(h.server, &sdk.Tool{
		Name: "magec_list_voice_types", Title: "List voice provider types",
		Description: "List registered voice providers (TTS/STT) with their JSON schemas.",
		Annotations: &sdk.ToolAnnotations{ReadOnlyHint: true},
	}, h.listVoiceTypes)
	h.toolCount++
}
