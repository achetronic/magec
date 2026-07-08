// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package admin

import (
	"net/http"

	"github.com/achetronic/magec/server/voice"
)

// VoiceProviderInfo represents a registered voice provider with its JSON Schemas.
type VoiceProviderInfo struct {
	Type            string       `json:"type" example:"openai"`
	DisplayName     string       `json:"displayName" example:"OpenAI Compatible"`
	SupportsTTS     bool         `json:"supportsTts"`
	SupportsSTT     bool         `json:"supportsStt"`
	TTSConfigSchema voice.Schema `json:"ttsConfigSchema"`
	STTConfigSchema voice.Schema `json:"sttConfigSchema"`
}

// listVoiceTypes returns all registered voice providers with their config schemas.
// @Summary      List voice provider types
// @Description  Returns registered voice providers with config schemas for dynamic form rendering. The backend type determines which voice provider handles TTS/STT requests.
// @Tags         voice
// @Produce      json
// @Success      200  {array}  VoiceProviderInfo
// @Security     AdminAuth
// @Router       /voice/types [get]
func (h *Handler) listVoiceTypes(w http.ResponseWriter, r *http.Request) {
	var types []VoiceProviderInfo
	for _, p := range voice.All() {
		types = append(types, VoiceProviderInfo{
			Type:            p.Type(),
			DisplayName:     p.DisplayName(),
			SupportsTTS:     p.SupportsTTS(),
			SupportsSTT:     p.SupportsSTT(),
			TTSConfigSchema: p.TTSConfigSchema(),
			STTConfigSchema: p.STTConfigSchema(),
		})
	}
	writeJSON(w, http.StatusOK, types)
}
