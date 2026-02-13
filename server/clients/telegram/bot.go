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
	"strings"
	"sync"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/achetronic/magec/server/store"
)

// Response modes control how the bot replies to messages.
// "text" sends only text, "voice" sends only audio, "mirror" matches the input
// format (voice reply to voice, text reply to text), and "both" sends both.
const (
	ResponseModeText   = "text"
	ResponseModeVoice  = "voice"
	ResponseModeMirror = "mirror"
	ResponseModeBoth   = "both"
)

// AgentInfo is a lightweight reference to an agent that the Telegram client
// is allowed to talk to. It only carries display information; all runtime
// config (TTS, transcription, LLM) is resolved server-side by the proxies.
type AgentInfo struct {
	ID   string
	Name string
}

// Client is a Telegram bot that receives messages via long-polling and
// forwards them to the magec agent API. It supports multi-agent switching
// per chat, voice transcription, TTS responses, and runtime commands.
type Client struct {
	// Config: injected at creation, read-only after New().
	clientDef store.ClientDefinition
	agentURL  string
	agents    []AgentInfo
	logger    *slog.Logger

	// Runtime: created during Start(), managed internally.
	bot     *telego.Bot
	handler *th.BotHandler
	cancel  context.CancelFunc

	// Mutable state: per-chat agent selection and response mode override.
	activeAgentMu sync.RWMutex
	activeAgent   map[int64]string // chatID -> agentID

	responseMu           sync.RWMutex
	responseModeOverride string
}

// New creates a Telegram client ready to be started. It validates the bot token
// and prepares the internal state, but does not connect to Telegram yet.
func New(clientDef store.ClientDefinition, agentURL string, agents []AgentInfo, logger *slog.Logger) (*Client, error) {
	if clientDef.Config.Telegram == nil {
		return nil, fmt.Errorf("telegram config is required")
	}
	if clientDef.Config.Telegram.BotToken == "" {
		return nil, fmt.Errorf("telegram bot token is required")
	}

	bot, err := telego.NewBot(clientDef.Config.Telegram.BotToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create telegram bot: %w", err)
	}

	return &Client{
		bot:         bot,
		clientDef:   clientDef,
		agentURL:    agentURL,
		agents:      agents,
		activeAgent: make(map[int64]string),
		logger:      logger,
	}, nil
}

// Start connects to Telegram via long-polling, registers all command and message
// handlers, and blocks until the context is cancelled or Stop is called.
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

	// Handle /start command
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleStartCommand(ctx, msg)
	}, th.CommandEqual("start"))

	// Handle /help command
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleHelpCommand(ctx, msg)
	}, th.CommandEqual("help"))

	// Handle /agent command
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleAgentCommand(ctx, msg)
	}, th.CommandEqual("agent"))

	// Handle /responsemode command
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleResponseModeCommand(ctx, msg)
	}, th.CommandEqual("responsemode"))

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

// Stop cancels the long-polling loop and shuts down the message handler.
func (c *Client) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	if c.handler != nil {
		c.handler.Stop()
	}
	c.logger.Info("Telegram bot stopped")
}

// getActiveAgentID returns the agent ID currently selected for a chat.
// If the user hasn't switched agents with /agent, it returns the default.
func (c *Client) getActiveAgentID(chatID int64) string {
	c.activeAgentMu.RLock()
	defer c.activeAgentMu.RUnlock()
	if id, ok := c.activeAgent[chatID]; ok {
		return id
	}
	return c.clientDef.AllowedAgents[0]
}

// setActiveAgentID changes which agent a specific chat is talking to.
func (c *Client) setActiveAgentID(chatID int64, agentID string) {
	c.activeAgentMu.Lock()
	defer c.activeAgentMu.Unlock()
	c.activeAgent[chatID] = agentID
}

// getAgentInfo looks up an agent by ID in the allowed agents list.
// Returns nil if the agent is not in the list.
func (c *Client) getAgentInfo(agentID string) *AgentInfo {
	for i := range c.agents {
		if c.agents[i].ID == agentID {
			return &c.agents[i]
		}
	}
	return nil
}

// getActiveAgentInfo is a shortcut that combines getActiveAgentID and getAgentInfo
// to return the full AgentInfo for the chat's currently selected agent.
func (c *Client) getActiveAgentInfo(chatID int64) *AgentInfo {
	return c.getAgentInfo(c.getActiveAgentID(chatID))
}

// handleStartCommand responds to /start with a welcome message showing the
// active agent name and a hint to use /help.
func (c *Client) handleStartCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID) {
		return nil
	}

	agent := c.getActiveAgentInfo(msg.Chat.ID)
	agentName := c.getActiveAgentID(msg.Chat.ID)
	if agent != nil {
		agentName = agent.Name
	}

	text := fmt.Sprintf("👋 *Welcome!* You are now talking to *%s*.\n\nType /help to see available commands.", agentName)
	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    tu.ID(msg.Chat.ID),
		Text:      text,
		ParseMode: "Markdown",
	})
	return nil
}

// handleHelpCommand responds to /help with a list of all supported bot commands.
func (c *Client) handleHelpCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID) {
		return nil
	}

	text := "*Available commands:*\n\n" +
		"/help — Show this help message\n" +
		"/agent — Show or switch the active agent\n" +
		"/responsemode — Show or change the response mode\n" +
		"/start — Show the welcome message"

	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    tu.ID(msg.Chat.ID),
		Text:      text,
		ParseMode: "Markdown",
	})
	return nil
}

// handleAgentCommand responds to /agent. Without arguments it shows the active
// agent and lists all available ones. With an agent ID it switches the chat to
// that agent, creating a new session for it.
func (c *Client) handleAgentCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID) {
		return nil
	}

	args := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/agent"))

	if args == "" {
		currentID := c.getActiveAgentID(msg.Chat.ID)
		current := c.getAgentInfo(currentID)
		currentLabel := currentID
		if current != nil && current.Name != "" {
			currentLabel = fmt.Sprintf("%s (`%s`)", current.Name, currentID)
		}

		var agentList string
		for _, a := range c.agents {
			marker := "  "
			if a.ID == currentID {
				marker = "▸ "
			}
			label := a.ID
			if a.Name != "" {
				label = fmt.Sprintf("%s (`%s`)", a.Name, a.ID)
			}
			agentList += fmt.Sprintf("%s%s\n", marker, label)
		}

		text := fmt.Sprintf("*Active agent:* %s\n\n*Available agents:*\n%s\nUsage: `/agent <id>`", currentLabel, agentList)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:    tu.ID(msg.Chat.ID),
			Text:      text,
			ParseMode: "Markdown",
		})
		return nil
	}

	found := false
	for _, a := range c.agents {
		if a.ID == args {
			found = true
			break
		}
	}
	if !found {
		var ids []string
		for _, a := range c.agents {
			ids = append(ids, "`"+a.ID+"`")
		}
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:    tu.ID(msg.Chat.ID),
			Text:      fmt.Sprintf("Unknown agent `%s`. Available: %s", args, strings.Join(ids, ", ")),
			ParseMode: "Markdown",
		})
		return nil
	}

	c.setActiveAgentID(msg.Chat.ID, args)
	agent := c.getAgentInfo(args)
	label := args
	if agent != nil && agent.Name != "" {
		label = agent.Name
	}

	c.logger.Info("Agent switched",
		"chat_id", msg.Chat.ID,
		"user_id", msg.From.ID,
		"agent", args,
	)

	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    tu.ID(msg.Chat.ID),
		Text:      fmt.Sprintf("Switched to agent *%s* (`%s`)", label, args),
		ParseMode: "Markdown",
	})
	return nil
}

// handleMessage processes a regular text message: checks permissions, sends a
// typing indicator, forwards the text to the active agent, and sends the response.
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
	response, err := c.callAgent(msg, msg.Text)
	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: tu.ID(msg.Chat.ID),
			Text:   "Sorry, I encountered an error processing your request.",
		})
		return nil
	}

	c.sendResponse(ctx, msg.Chat.ID, response, false)

	return nil
}

// handleVoice processes a voice message: downloads the audio from Telegram,
// transcribes it via the magec transcription proxy, sends the resulting text
// to the active agent, and replies with the response.
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
	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.clientDef.Config.Telegram.BotToken, file.FilePath)
	audioData, err := c.downloadFile(fileURL)
	if err != nil {
		c.logger.Error("Failed to download voice file", "error", err)
		return nil
	}

	// Transcribe audio
	agentID := c.getActiveAgentID(msg.Chat.ID)
	text, err := c.transcribeAudio(audioData, file.FilePath, agentID)
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
	response, err := c.callAgent(msg, text)
	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: tu.ID(msg.Chat.ID),
			Text:   "Sorry, I encountered an error processing your request.",
		})
		return nil
	}

	c.sendResponse(ctx, msg.Chat.ID, response, true)

	return nil
}

// sendResponse delivers the agent's reply to the chat. Depending on the active
// response mode it sends text, voice (TTS), or both. In mirror mode it matches
// the format of the original message (voice in → voice out, text in → text out).
func (c *Client) sendResponse(ctx *th.Context, chatID int64, text string, inputWasVoice bool) {
	mode := c.getResponseMode()

	sendText := false
	sendVoice := false

	switch mode {
	case ResponseModeVoice:
		sendVoice = true
	case ResponseModeMirror:
		if inputWasVoice {
			sendVoice = true
		} else {
			sendText = true
		}
	case ResponseModeBoth:
		sendText = true
		sendVoice = true
	default:
		sendText = true
	}

	if sendText {
		_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID: tu.ID(chatID),
			Text:   text,
		})
		if err != nil {
			c.logger.Error("Failed to send message", "error", err)
		}
	}

	if sendVoice {
		agentID := c.getActiveAgentID(chatID)
		c.sendVoiceResponse(ctx, chatID, text, agentID)
	}
}

// getResponseMode returns the current response mode. If the user has set a
// runtime override via /responsemode it takes precedence over the config default.
func (c *Client) getResponseMode() string {
	c.responseMu.RLock()
	defer c.responseMu.RUnlock()
	if c.responseModeOverride != "" {
		return c.responseModeOverride
	}
	return c.clientDef.Config.Telegram.ResponseMode
}

// handleResponseModeCommand responds to /responsemode. Without arguments it shows
// the current mode. With a mode name it overrides the config default for the
// rest of the session. "reset" clears the override back to the config value.
func (c *Client) handleResponseModeCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID) {
		return nil
	}

	validModes := []string{
		ResponseModeText,
		ResponseModeVoice,
		ResponseModeMirror,
		ResponseModeBoth,
	}

	args := strings.TrimSpace(strings.TrimPrefix(msg.Text, "/responsemode"))
	if args == "" {
		current := c.getResponseMode()
		c.responseMu.RLock()
		overridden := c.responseModeOverride != ""
		c.responseMu.RUnlock()

		status := fmt.Sprintf("*Response mode:* `%s`", current)
		if overridden {
			status += fmt.Sprintf(" (override, config: `%s`)", c.clientDef.Config.Telegram.ResponseMode)
		}
		status += fmt.Sprintf("\n*Options:* `%s`, `reset`", strings.Join(validModes, "`, `"))

		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:    tu.ID(msg.Chat.ID),
			Text:      status,
			ParseMode: "Markdown",
		})
		return nil
	}

	if args == "reset" {
		c.responseMu.Lock()
		c.responseModeOverride = ""
		c.responseMu.Unlock()
		c.logger.Info("Response mode override cleared, back to config default",
			"user_id", msg.From.ID,
			"config_mode", c.clientDef.Config.Telegram.ResponseMode,
		)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:    tu.ID(msg.Chat.ID),
			Text:      fmt.Sprintf("Response mode reset to config default: `%s`", c.clientDef.Config.Telegram.ResponseMode),
			ParseMode: "Markdown",
		})
		return nil
	}

	if !slices.Contains(validModes, args) {
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:    tu.ID(msg.Chat.ID),
			Text:      fmt.Sprintf("Invalid mode `%s`. Valid options: `%s`, `reset`", args, strings.Join(validModes, "`, `")),
			ParseMode: "Markdown",
		})
		return nil
	}

	c.responseMu.Lock()
	c.responseModeOverride = args
	c.responseMu.Unlock()

	c.logger.Info("Response mode overridden",
		"user_id", msg.From.ID,
		"new_mode", args,
	)
	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:    tu.ID(msg.Chat.ID),
		Text:      fmt.Sprintf("Response mode set to `%s` (until restart)", args),
		ParseMode: "Markdown",
	})
	return nil
}

// setAuthHeader adds the Bearer token to requests sent to the magec API.
// This authenticates the Telegram client against the clientAuthMiddleware.
func (c *Client) setAuthHeader(req *http.Request) {
	if c.clientDef.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.clientDef.Token)
	}
}

// isAllowed checks whether a Telegram user or chat is permitted to interact
// with this bot. If no allowlists are configured, all users are allowed.
func (c *Client) isAllowed(userID, chatID int64) bool {
	// If no restrictions, allow all
	if len(c.clientDef.Config.Telegram.AllowedUsers) == 0 && len(c.clientDef.Config.Telegram.AllowedChats) == 0 {
		return true
	}

	// Check user allowlist
	if len(c.clientDef.Config.Telegram.AllowedUsers) > 0 && slices.Contains(c.clientDef.Config.Telegram.AllowedUsers, userID) {
		return true
	}

	// Check chat allowlist
	if len(c.clientDef.Config.Telegram.AllowedChats) > 0 && slices.Contains(c.clientDef.Config.Telegram.AllowedChats, chatID) {
		return true
	}

	return false
}

// buildMessageContext prepends invisible metadata to the user's message so the
// LLM knows who is writing and from which chat. The metadata is wrapped in
// <!--MAGEC_META:{...}:MAGEC_META--> delimiters.
func (c *Client) buildMessageContext(msg telego.Message) string {
	meta := map[string]interface{}{
		"source":             "telegram",
		"telegram_user_id":   msg.From.ID,
		"telegram_chat_id":   msg.Chat.ID,
		"telegram_chat_type": string(msg.Chat.Type),
	}

	if msg.From.Username != "" {
		meta["telegram_username"] = "@" + msg.From.Username
	}

	name := strings.TrimSpace(msg.From.FirstName + " " + msg.From.LastName)
	if name != "" {
		meta["telegram_name"] = name
	}

	if msg.Chat.Title != "" {
		meta["telegram_chat_title"] = msg.Chat.Title
	}

	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		c.logger.Warn("Failed to marshal message context metadata", "error", err)
		return ""
	}

	return fmt.Sprintf("<!--MAGEC_META:%s:MAGEC_META-->\n", string(jsonBytes))
}

// callAgent sends a user message to the active agent via the magec API and
// returns the text response. It ensures a session exists for the chat+agent
// pair before making the request.
func (c *Client) callAgent(msg telego.Message, message string) (string, error) {
	agentID := c.getActiveAgentID(msg.Chat.ID)
	sessionID := fmt.Sprintf("telegram_%d_%s", msg.Chat.ID, agentID)
	userIDStr := "default_user"

	// Ensure session exists
	if err := c.ensureSession(agentID, userIDStr, sessionID); err != nil {
		c.logger.Warn("Failed to ensure session, continuing anyway", "error", err)
	}

	fullMessage := c.buildMessageContext(msg) + message

	// Build request
	reqBody := map[string]interface{}{
		"appName":   agentID,
		"userId":    userIDStr,
		"sessionId": sessionID,
		"newMessage": map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": fullMessage},
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
	c.setAuthHeader(req)

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

// ensureSession calls the ADK session endpoint to create a session for the
// given agent, user, and session ID. If the session already exists (409) it
// silently succeeds.
func (c *Client) ensureSession(agentID, userID, sessionID string) error {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions/%s", c.agentURL, agentID, userID, sessionID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

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

// extractResponseText walks the ADK response events and concatenates all text
// parts into a single string. Returns a fallback message if no text was found.
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

// downloadFile fetches a file by URL. Used to download voice messages from
// the Telegram file API (not from magec, so no auth header is added).
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

// convertOggToWav shells out to ffmpeg to convert Telegram's OGG/Opus voice
// files into 16 kHz mono WAV, which is the format expected by transcription backends.
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

// transcribeAudio converts a voice message to text. It first re-encodes the
// audio to WAV, then sends it to the magec transcription proxy which routes
// it to the backend configured for the active agent.
func (c *Client) transcribeAudio(audioData []byte, filePath string, agentID string) (string, error) {
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
	buf.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	// Call transcription service
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Use the internal transcription endpoint
	transcriptionURL := strings.TrimSuffix(c.agentURL, "/agent") + "/voice/" + agentID + "/transcription"

	req, err := http.NewRequestWithContext(ctx, "POST", transcriptionURL, &buf)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", fmt.Sprintf("multipart/form-data; boundary=%s", boundary))
	c.setAuthHeader(req)

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

// sendVoiceResponse generates speech audio for the given text via the magec
// TTS proxy and sends it back to the chat as a Telegram voice message.
func (c *Client) sendVoiceResponse(ctx *th.Context, chatID int64, text string, agentID string) {
	// Send recording indicator
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID: tu.ID(chatID),
		Action: telego.ChatActionRecordVoice,
	})

	// Generate TTS
	audioData, err := c.generateTTS(text, agentID)
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

// generateTTS calls the magec TTS proxy for the given agent. It only sends the
// text and desired format; the proxy injects model, voice, and speed from the
// agent's store config.
func (c *Client) generateTTS(text string, agentID string) ([]byte, error) {
	reqBody := map[string]interface{}{
		"input":           text,
		"response_format": "opus",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	ttsURL := strings.TrimSuffix(c.agentURL, "/agent") + "/voice/" + agentID + "/speech"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", ttsURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	c.setAuthHeader(req)

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
