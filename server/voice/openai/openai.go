package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/achetronic/magec/server/store"
	"github.com/achetronic/magec/server/voice"
)

func init() {
	voice.Register(&provider{})
}

type provider struct{}

func (p *provider) Type() string        { return "openai" }
func (p *provider) DisplayName() string { return "OpenAI Compatible" }
func (p *provider) SupportsTTS() bool   { return true }
func (p *provider) SupportsSTT() bool   { return true }

func (p *provider) TTSConfigSchema() voice.Schema {
	return voice.Schema{
		"type": "object",
		"properties": map[string]interface{}{
			"speed": map[string]interface{}{
				"type":          "number",
				"title":         "Speed",
				"description":   "Speaking rate (0.25–4.0). Default 1.0.",
				"minimum":       0.25,
				"maximum":       4.0,
				"x-placeholder": "1.0",
				"x-size":        "half",
			},
		},
	}
}
func (p *provider) STTConfigSchema() voice.Schema { return nil }

func (p *provider) SynthesizeSpeech(ctx context.Context, req voice.TTSRequest, backend store.BackendDefinition) (*voice.TTSResponse, error) {
	target, err := url.Parse(backend.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid backend URL: %w", err)
	}

	body := map[string]interface{}{
		"input": req.Input,
		"model": req.Model,
		"voice": req.Voice,
	}
	if req.Config.OpenAI != nil && req.Config.OpenAI.Speed != 0 {
		body["speed"] = req.Config.OpenAI.Speed
	}
	if req.ResponseFormat != "" {
		body["response_format"] = req.ResponseFormat
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	proxyURL := *target
	proxyURL.Path = "/v1/audio/speech"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if backend.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+backend.APIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("TTS request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		errBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS backend returned %d: %s", resp.StatusCode, string(errBody))
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "audio/mpeg"
	}

	return &voice.TTSResponse{
		Audio:       resp.Body,
		ContentType: contentType,
	}, nil
}

func (p *provider) TranscribeAudio(ctx context.Context, req voice.STTRequest, backend store.BackendDefinition) (string, error) {
	target, err := url.Parse(backend.URL)
	if err != nil {
		return "", fmt.Errorf("invalid backend URL: %w", err)
	}

	body, err := io.ReadAll(req.Audio)
	if err != nil {
		return "", fmt.Errorf("failed to read audio: %w", err)
	}

	var proxyBody bytes.Buffer
	if req.Model != "" && strings.Contains(req.ContentType, "multipart/form-data") {
		boundary := extractBoundary(req.ContentType)
		if boundary == "" {
			return "", fmt.Errorf("missing multipart boundary")
		}

		closingBoundary := fmt.Sprintf("\r\n--%s--", boundary)
		trimmed := bytes.TrimSuffix(body, []byte(closingBoundary))
		trimmed = bytes.TrimSuffix(trimmed, []byte(fmt.Sprintf("--%s--\r\n", boundary)))
		trimmed = bytes.TrimSuffix(trimmed, []byte(fmt.Sprintf("--%s--", boundary)))

		proxyBody.Write(trimmed)
		proxyBody.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		proxyBody.WriteString("Content-Disposition: form-data; name=\"model\"\r\n\r\n")
		proxyBody.WriteString(req.Model)
		proxyBody.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else {
		proxyBody.Write(body)
	}

	proxyURL := *target
	proxyURL.Path = "/v1/audio/transcriptions"

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, proxyURL.String(), &proxyBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", req.ContentType)
	if backend.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+backend.APIKey)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("STT request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("STT backend returned %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return string(respBody), nil
	}
	return result.Text, nil
}

func extractBoundary(contentType string) string {
	for _, param := range strings.Split(contentType, ";") {
		param = strings.TrimSpace(param)
		if strings.HasPrefix(param, "boundary=") {
			return strings.TrimPrefix(param, "boundary=")
		}
	}
	return ""
}
