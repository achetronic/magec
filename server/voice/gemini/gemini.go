package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
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

const defaultGeminiURL = "https://generativelanguage.googleapis.com"

type provider struct{}

func (p *provider) Type() string        { return "gemini" }
func (p *provider) DisplayName() string { return "Google Gemini" }
func (p *provider) SupportsTTS() bool   { return true }
func (p *provider) SupportsSTT() bool   { return true }

func resolveURL(backend store.BackendDefinition) string {
	if backend.URL != "" {
		return backend.URL
	}
	return defaultGeminiURL
}

func (p *provider) TTSConfigSchema() voice.Schema {
	return voice.Schema{
		"type": "object",
		"propertyOrder": []string{"languageCode", "temperature", "stylePrompt"},
		"properties": map[string]interface{}{
			"languageCode": map[string]interface{}{
				"type":          "string",
				"title":         "Language",
				"default":       "en-US",
				"x-placeholder": "en-US",
				"x-size":        "half",
				"x-advanced":    true,
				"x-link":        "https://cloud.google.com/text-to-speech/docs/gemini-tts#available_languages",
			},
			"temperature": map[string]interface{}{
				"type":          "number",
				"title":         "Temperature",
				"description":   "Voice expressiveness (0–2). Higher = more varied.",
				"minimum":       0,
				"maximum":       2.0,
				"x-placeholder": "1.0",
				"x-size":        "half",
				"x-advanced":    true,
			},
			"stylePrompt": map[string]interface{}{
				"type":          "string",
				"title":         "Style Prompt",
				"description":   "Tone instructions for speech. E.g. \"Speak in a calm, friendly tone\".",
				"x-placeholder": "Speak in a warm, friendly tone",
				"x-format":      "textarea",
				"x-advanced":    true,
			},
		},
	}
}

func (p *provider) STTConfigSchema() voice.Schema { return nil }

func (p *provider) SynthesizeSpeech(ctx context.Context, req voice.TTSRequest, backend store.BackendDefinition) (*voice.TTSResponse, error) {
	target, err := url.Parse(resolveURL(backend))
	if err != nil {
		return nil, fmt.Errorf("invalid backend URL: %w", err)
	}

	text := req.Input
	gcfg := req.Config.Gemini
	if gcfg != nil && gcfg.StylePrompt != "" {
		text = gcfg.StylePrompt + ": " + text
	}

	voiceConfig := map[string]interface{}{
		"prebuiltVoiceConfig": map[string]interface{}{
			"voiceName": req.Voice,
		},
	}

	speechConfig := map[string]interface{}{
		"voiceConfig": voiceConfig,
	}
	if gcfg != nil && gcfg.LanguageCode != "" {
		speechConfig["languageCode"] = gcfg.LanguageCode
	}

	genConfig := map[string]interface{}{
		"responseModalities": []string{"AUDIO"},
		"speechConfig":       speechConfig,
	}
	if gcfg != nil && gcfg.Temperature != 0 {
		genConfig["temperature"] = gcfg.Temperature
	}

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{"text": text},
				},
			},
		},
		"generationConfig": genConfig,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	proxyURL := *target
	proxyURL.Path = fmt.Sprintf("/v1beta/models/%s:generateContent", req.Model)
	if backend.APIKey != "" {
		q := proxyURL.Query()
		q.Set("key", backend.APIKey)
		proxyURL.RawQuery = q.Encode()
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, proxyURL.String(), bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("Gemini TTS request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read Gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini TTS returned %d: %s", resp.StatusCode, string(respBody))
	}

	audioData, mimeType, err := extractAudioFromResponse(respBody)
	if err != nil {
		return nil, err
	}

	if strings.HasPrefix(mimeType, "audio/L16") || strings.Contains(mimeType, "pcm") {
		audioData = wrapPCMAsWAV(audioData, 24000, 1, 16)
		mimeType = "audio/wav"
	}

	return &voice.TTSResponse{
		Audio:       io.NopCloser(bytes.NewReader(audioData)),
		ContentType: mimeType,
	}, nil
}

func (p *provider) TranscribeAudio(ctx context.Context, req voice.STTRequest, backend store.BackendDefinition) (string, error) {
	target, err := url.Parse(resolveURL(backend))
	if err != nil {
		return "", fmt.Errorf("invalid backend URL: %w", err)
	}

	audioBytes, err := readAudioFromRequest(req)
	if err != nil {
		return "", err
	}

	mimeType := "audio/wav"

	prompt := "Transcribe this audio."

	body := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"role": "user",
				"parts": []map[string]interface{}{
					{"text": prompt},
					{
						"inlineData": map[string]interface{}{
							"mimeType": mimeType,
							"data":     base64.StdEncoding.EncodeToString(audioBytes),
						},
					},
				},
			},
		},
	}

	model := req.Model
	if model == "" {
		model = "gemini-2.0-flash"
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	proxyURL := *target
	proxyURL.Path = fmt.Sprintf("/v1beta/models/%s:generateContent", model)
	if backend.APIKey != "" {
		q := proxyURL.Query()
		q.Set("key", backend.APIKey)
		proxyURL.RawQuery = q.Encode()
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(reqCtx, http.MethodPost, proxyURL.String(), bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := (&http.Client{}).Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("Gemini STT request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read Gemini response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Gemini STT returned %d: %s", resp.StatusCode, string(respBody))
	}

	return extractTextFromResponse(respBody)
}

func extractAudioFromResponse(data []byte) ([]byte, string, error) {
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					InlineData *struct {
						MIMEType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			if p.InlineData != nil && p.InlineData.Data != "" {
				decoded, err := base64.StdEncoding.DecodeString(p.InlineData.Data)
				if err != nil {
					return nil, "", fmt.Errorf("failed to decode audio: %w", err)
				}
				return decoded, p.InlineData.MIMEType, nil
			}
		}
	}
	return nil, "", fmt.Errorf("no audio data in Gemini response")
}

func extractTextFromResponse(data []byte) (string, error) {
	var resp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return "", fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	var texts []string
	for _, c := range resp.Candidates {
		for _, p := range c.Content.Parts {
			if p.Text != "" {
				texts = append(texts, p.Text)
			}
		}
	}
	if len(texts) == 0 {
		return "", fmt.Errorf("no text in Gemini response")
	}
	return strings.Join(texts, " "), nil
}

func readAudioFromRequest(req voice.STTRequest) ([]byte, error) {
	data, err := io.ReadAll(req.Audio)
	if err != nil {
		return nil, fmt.Errorf("failed to read audio: %w", err)
	}

	if !strings.Contains(req.ContentType, "multipart/form-data") {
		return data, nil
	}

	boundary := ""
	for _, param := range strings.Split(req.ContentType, ";") {
		param = strings.TrimSpace(param)
		if strings.HasPrefix(param, "boundary=") {
			boundary = strings.TrimPrefix(param, "boundary=")
			break
		}
	}
	if boundary == "" {
		return data, nil
	}

	parts := bytes.Split(data, []byte("--"+boundary))
	for _, part := range parts {
		if bytes.Contains(part, []byte("Content-Disposition")) && bytes.Contains(part, []byte("name=\"file\"")) {
			idx := bytes.Index(part, []byte("\r\n\r\n"))
			if idx >= 0 {
				content := part[idx+4:]
				content = bytes.TrimSuffix(content, []byte("\r\n"))
				return content, nil
			}
		}
	}

	return data, nil
}

func wrapPCMAsWAV(pcm []byte, sampleRate, channels, bitsPerSample int) []byte {
	dataSize := len(pcm)
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	var buf bytes.Buffer
	buf.Write([]byte("RIFF"))
	binary.Write(&buf, binary.LittleEndian, int32(36+dataSize))
	buf.Write([]byte("WAVE"))

	buf.Write([]byte("fmt "))
	binary.Write(&buf, binary.LittleEndian, int32(16))
	binary.Write(&buf, binary.LittleEndian, int16(1))
	binary.Write(&buf, binary.LittleEndian, int16(channels))
	binary.Write(&buf, binary.LittleEndian, int32(sampleRate))
	binary.Write(&buf, binary.LittleEndian, int32(byteRate))
	binary.Write(&buf, binary.LittleEndian, int16(blockAlign))
	binary.Write(&buf, binary.LittleEndian, int16(bitsPerSample))

	buf.Write([]byte("data"))
	binary.Write(&buf, binary.LittleEndian, int32(dataSize))
	buf.Write(pcm)

	return buf.Bytes()
}
