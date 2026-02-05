// Copyright 2025 Alby Hernández
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os/exec"
	"slices"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/achetronic/magec/server/config"
)

const appName = "magec_agent"

// Client represents a Telegram bot client that connects to the ADK agent
type Client struct {
	bot       *telego.Bot
	handler   *th.BotHandler
	cfg       *config.TelegramConfig
	agentURL  string
	ttsURL    string
	ttsConfig *config.TTSConfig
	logger    *slog.Logger
	cancel    context.CancelFunc
}

// New creates a new Telegram client
func New(cfg *config.TelegramConfig, agentURL string, ttsURL string, ttsConfig *config.TTSConfig, logger *slog.Logger) (*Client, error) {
	if cfg.Token == "" {
		return nil, fmt.Errorf("telegram token is required")
	}

	bot, err := telego.NewBot(cfg.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &Client{
		bot:       bot,
		cfg:       cfg,
		agentURL:  agentURL,
		ttsURL:    ttsURL,
		ttsConfig: ttsConfig,
		logger:    logger,
	}, nil
}

// Start begins the long polling loop
func (c *Client) Start(ctx context.Context) error {
	// Get bot info
	botUser, err := c.bot.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	c.logger.Info("Telegram bot started", "username", botUser.Username)

	// Create cancellable context for long polling
	pollCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	// Create update channel with long polling
	updates, err := c.bot.UpdatesViaLongPolling(pollCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to start long polling: %w", err)
	}

	// Create handler
	handler, err := th.NewBotHandler(c.bot, updates)
	if err != nil {
		return fmt.Errorf("failed to create bot handler: %w", err)
	}
	c.handler = handler

	// Handle voice messages (must be registered before general message handler)
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		c.logger.Info("Voice handler triggered", "chat_id", msg.Chat.ID, "user_id", msg.From.ID)
		return c.handleVoice(ctx, msg)
	}, func(_ context.Context, update telego.Update) bool {
		return update.Message != nil && update.Message.Voice != nil
	})

	// Handle text messages (exclude voice messages)
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		c.logger.Info("Text handler triggered", "chat_id", msg.Chat.ID, "user_id", msg.From.ID, "text", msg.Text)
		return c.handleMessage(ctx, msg)
	}, func(_ context.Context, update telego.Update) bool {
		match := update.Message != nil && update.Message.Voice == nil && update.Message.Text != ""
		if update.Message != nil {
			c.logger.Debug("Text predicate check",
				"chat_id", update.Message.Chat.ID,
				"user_id", update.Message.From.ID,
				"text", update.Message.Text,
				"has_voice", update.Message.Voice != nil,
				"match", match,
			)
		}
		return match
	})

	// Start handling (blocks until stopped)
	c.handler.Start()

	return nil
}

// Stop gracefully stops the bot
func (c *Client) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.handler != nil {
		c.handler.Stop()
	}
	c.logger.Info("Telegram bot stopped")
}

// handleMessage processes incoming text messages
func (c *Client) handleMessage(ctx *th.Context, msg telego.Message) error {
	if msg.Text == "" {
		return nil
	}

	// Check permissions
	if !c.isAllowed(msg.From.ID, msg.Chat.ID) {
		c.logger.Debug("Unauthorized access attempt",
			"user_id", msg.From.ID,
			"chat_id", msg.Chat.ID,
		)
		return nil
	}

	c.logger.Info("Received message",
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID,
		"text", msg.Text,
	)

	// Send typing indicator
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID: tu.ID(msg.Chat.ID),
		Action: telego.ChatActionTyping,
	})

	// Call agent
	response, err := c.callAgent(msg.From.ID, msg.Chat.ID, msg.Text)
	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: tu.ID(msg.Chat.ID),
			Text:   "Sorry, I encountered an error processing your request.",
		})
		return nil
	}

	// Send text response
	_, err = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID: tu.ID(msg.Chat.ID),
		Text:   response,
	})
	if err != nil {
		c.logger.Error("Failed to send message", "error", err)
	}

	// Send voice response if enabled
	if c.cfg.VoiceResponses && c.ttsURL != "" {
		c.sendVoiceResponse(ctx, msg.Chat.ID, response)
	}

	return nil
}

// handleVoice processes incoming voice messages
func (c *Client) handleVoice(ctx *th.Context, msg telego.Message) error {
	if msg.Voice == nil {
		return nil
	}

	// Check permissions
	if !c.isAllowed(msg.From.ID, msg.Chat.ID) {
		return nil
	}

	c.logger.Info("Received voice message",
		"user_id", msg.From.ID,
		"chat_id", msg.Chat.ID,
		"duration", msg.Voice.Duration,
	)

	// Send typing indicator
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID: tu.ID(msg.Chat.ID),
		Action: telego.ChatActionTyping,
	})

	// Download voice file
	file, err := ctx.Bot().GetFile(ctx, &telego.GetFileParams{FileID: msg.Voice.FileID})
	if err != nil {
		c.logger.Error("Failed to get voice file", "error", err)
		return nil
	}

	// Download file content
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.cfg.Token, file.FilePath)
	audioData, err := c.downloadFile(fileURL)
	if err != nil {
		c.logger.Error("Failed to download voice file", "error", err)
		return nil
	}

	// Transcribe audio
	text, err := c.transcribeAudio(audioData, file.FilePath)
	if err != nil {
		c.logger.Error("Failed to transcribe audio", "error", err)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: tu.ID(msg.Chat.ID),
			Text:   "Sorry, I couldn't transcribe your voice message.",
		})
		return nil
	}

	c.logger.Info("Transcribed voice", "text", text)

	// Call agent with transcribed text
	response, err := c.callAgent(msg.From.ID, msg.Chat.ID, text)
	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: tu.ID(msg.Chat.ID),
			Text:   "Sorry, I encountered an error processing your request.",
		})
		return nil
	}

	// Send text response
	_, err = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID: tu.ID(msg.Chat.ID),
		Text:   response,
	})
	if err != nil {
		c.logger.Error("Failed to send message", "error", err)
	}

	// Send voice response if enabled
	if c.cfg.VoiceResponses && c.ttsURL != "" {
		c.sendVoiceResponse(ctx, msg.Chat.ID, response)
	}

	return nil
}

// isAllowed checks if the user/chat is authorized
func (c *Client) isAllowed(userID, chatID int64) bool {
	// If no restrictions, allow all
	if len(c.cfg.AllowedUsers) == 0 && len(c.cfg.AllowedChats) == 0 {
		return true
	}

	// Check user allowlist
	if len(c.cfg.AllowedUsers) > 0 && slices.Contains(c.cfg.AllowedUsers, userID) {
		return true
	}

	// Check chat allowlist
	if len(c.cfg.AllowedChats) > 0 && slices.Contains(c.cfg.AllowedChats, chatID) {
		return true
	}

	return false
}

// callAgent sends a message to the ADK agent and returns the response
func (c *Client) callAgent(userID, chatID int64, message string) (string, error) {
	// Use chat ID as session ID for conversation continuity
	// User is always default_user until multiuser support is added
	sessionID := fmt.Sprintf("telegram_%d", chatID)
	userIDStr := "default_user"

	// Ensure session exists
	if err := c.ensureSession(userIDStr, sessionID); err != nil {
		c.logger.Warn("Failed to ensure session, continuing anyway", "error", err)
	}

	// Build request
	reqBody := map[string]interface{}{
		"appName":   appName,
		"userId":    userIDStr,
		"sessionId": sessionID,
		"newMessage": map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": message},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Call agent
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.agentURL+"/run", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract text from response
	return c.extractResponseText(events), nil
}

// ensureSession creates a session if it doesn't exist
func (c *Client) ensureSession(userID, sessionID string) error {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions/%s", c.agentURL, appName, userID, sessionID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	// 200 = created, 409 = already exists, both are fine
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusConflict {
		return fmt.Errorf("failed to create session: status %d", resp.StatusCode)
	}

	return nil
}

// extractResponseText extracts the text content from ADK response events
func (c *Client) extractResponseText(events []map[string]interface{}) string {
	var result string

	for _, event := range events {
		content, ok := event["content"].(map[string]interface{})
		if !ok {
			continue
		}

		parts, ok := content["parts"].([]interface{})
		if !ok {
			continue
		}

		for _, part := range parts {
			partMap, ok := part.(map[string]interface{})
			if !ok {
				continue
			}

			if text, ok := partMap["text"].(string); ok {
				result += text
			}
		}
	}

	if result == "" {
		return "I couldn't generate a response."
	}

	return result
}

// downloadFile downloads a file from the given URL
func (c *Client) downloadFile(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// convertOggToWav converts OGG audio to WAV format using ffmpeg
func (c *Client) convertOggToWav(oggData []byte) ([]byte, error) {
	cmd := exec.Command("ffmpeg",
		"-i", "pipe:0",
		"-ar", "16000",
		"-ac", "1",
		"-f", "wav",
		"pipe:1",
	)

	cmd.Stdin = bytes.NewReader(oggData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %w, stderr: %s", err, stderr.String())
	}

	return stdout.Bytes(), nil
}

// transcribeAudio sends audio to the transcription service
func (c *Client) transcribeAudio(audioData []byte, filePath string) (string, error) {
	// Convert OGG to WAV (Telegram sends voice as OGG/Opus, but transcription expects WAV)
	wavData, err := c.convertOggToWav(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}

	// Create multipart form
	var buf bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"

	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"audio.wav\"\r\n")
	buf.WriteString("Content-Type: audio/wav\r\n\r\n")
	buf.Write(wavData)
	buf.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
	buf.WriteString("Content-Disposition: form-data; name=\"model\"\r\n\r\n")
	buf.WriteString("whisper-1")
	buf.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// Call transcription service
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Use the internal transcription endpoint
	transcriptionURL := c.agentURL[:len(c.agentURL)-len("/agent")] + "/transcription/v1/audio/transcriptions"

	req, err := http.NewRequestWithContext(ctx, "POST", transcriptionURL, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("transcription failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Text, nil
}

// sendVoiceResponse generates TTS audio and sends it as a voice message
func (c *Client) sendVoiceResponse(ctx *th.Context, chatID int64, text string) {
	// Send recording indicator
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID: tu.ID(chatID),
		Action: telego.ChatActionRecordVoice,
	})

	// Generate TTS
	audioData, err := c.generateTTS(text)
	if err != nil {
		c.logger.Error("Failed to generate TTS", "error", err)
		return
	}

	// Send voice message
	_, err = ctx.Bot().SendVoice(ctx, &telego.SendVoiceParams{
		ChatID: tu.ID(chatID),
		Voice:  tu.FileFromReader(bytes.NewReader(audioData), "voice.ogg"),
	})
	if err != nil {
		c.logger.Error("Failed to send voice message", "error", err)
	}
}

// generateTTS calls the TTS service to generate audio
func (c *Client) generateTTS(text string) ([]byte, error) {
	reqBody := map[string]interface{}{
		"input":           text,
		"model":           c.ttsConfig.Model,
		"voice":           c.ttsConfig.Voice,
		"speed":           c.ttsConfig.Speed,
		"response_format": "opus", // Telegram supports opus in ogg container
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.ttsURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("TTS failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}
