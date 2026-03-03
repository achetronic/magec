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
	"encoding/base64"
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

	"github.com/achetronic/magec/server/clients/msgutil"
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
	store     interface {
		UpdateClient(id string, c store.ClientDefinition) error
	}

	// Runtime: created during Start(), managed internally.
	bot     *telego.Bot
	handler *th.BotHandler
	cancel  context.CancelFunc

	// Mutable state: per-chat agent selection and response mode override.
	activeAgentMu sync.RWMutex
	activeAgent   map[int64]string // chatID -> agentID

	responseMu           sync.RWMutex
	responseModeOverride string

	showToolsMu sync.RWMutex
	showTools   bool
}

// New creates a Telegram client ready to be started. It validates the bot token
// and prepares the internal state, but does not connect to Telegram yet.
func New(clientDef store.ClientDefinition, agentURL string, agents []AgentInfo, s interface {
	UpdateClient(id string, c store.ClientDefinition) error
}, logger *slog.Logger) (*Client, error) {
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
		store:       s,
		activeAgent: make(map[int64]string),
		logger:      logger,
	}, nil
}

// Start connects to Telegram via long-polling, registers all command and message
// handlers, and blocks until the context is cancelled or Stop is called.
func (c *Client) Start(ctx context.Context) error {
	botUser, err := c.bot.GetMe(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bot info: %w", err)
	}
	c.logger.Info("Telegram bot started", "username", botUser.Username)

	pollCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	updates, err := c.bot.UpdatesViaLongPolling(pollCtx, nil)
	if err != nil {
		return fmt.Errorf("failed to start long polling: %w", err)
	}

	handler, err := th.NewBotHandler(c.bot, updates)
	if err != nil {
		return fmt.Errorf("failed to create bot handler: %w", err)
	}
	c.handler = handler

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleStartCommand(ctx, msg)
	}, th.CommandEqual("start"))

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleHelpCommand(ctx, msg)
	}, th.CommandEqual("help"))

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleAgentCommand(ctx, msg)
	}, th.CommandEqual("agent"))

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleResetCommand(ctx, msg)
	}, th.CommandEqual("reset"))

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleResponseModeCommand(ctx, msg)
	}, th.CommandEqual("responsemode"))

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		return c.handleShowToolsCommand(ctx, msg)
	}, th.CommandEqual("showtools"))

	// Voice handler must be registered before the general text handler.
	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		c.logger.Info("Voice handler triggered", "chat_id", msg.Chat.ID, "user_id", msg.From.ID)
		return c.handleVoice(ctx, msg)
	}, func(_ context.Context, update telego.Update) bool {
		return update.Message != nil && update.Message.Voice != nil
	})

	handler.HandleMessage(func(ctx *th.Context, msg telego.Message) error {
		c.logger.Info("Text handler triggered", "chat_id", msg.Chat.ID, "user_id", msg.From.ID, "text", msg.Text)
		return c.handleMessage(ctx, msg)
	}, func(_ context.Context, update telego.Update) bool {
		return update.Message != nil && update.Message.Voice == nil && update.Message.Text != ""
	})

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
	if def := c.clientDef.Config.Telegram.DefaultAgent; def != "" {
		return def
	}
	return c.clientDef.AllowedAgents[0]
}

// setActiveAgentID changes which agent a specific chat is talking to.
func (c *Client) setActiveAgentID(chatID int64, agentID string) {
	c.activeAgentMu.Lock()
	c.activeAgent[chatID] = agentID
	c.activeAgentMu.Unlock()

	def := c.clientDef
	def.Config.Telegram.DefaultAgent = agentID
	c.clientDef = def
	if err := c.store.UpdateClient(def.ID, def); err != nil {
		c.logger.Warn("Failed to persist default agent", "error", err)
	}
}

// getAgentInfo looks up an agent by ID in the allowed agents list.
// Returns nil if the agent is not in the list.
func (c *Client) getAgentInfo(agentID string) *AgentInfo {
	i := slices.IndexFunc(c.agents, func(a AgentInfo) bool { return a.ID == agentID })
	if i < 0 {
		return nil
	}
	return &c.agents[i]
}

// getActiveAgentInfo returns the full AgentInfo for the chat's currently selected agent.
func (c *Client) getActiveAgentInfo(chatID int64) *AgentInfo {
	return c.getAgentInfo(c.getActiveAgentID(chatID))
}

// agentLabel returns the display name for an agent, falling back to the ID.
func agentLabel(a *AgentInfo, fallback string) string {
	if a != nil && a.Name != "" {
		return a.Name
	}
	return fallback
}

// handleStartCommand responds to /start with a welcome message.
func (c *Client) handleStartCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		return nil
	}

	agentID := c.getActiveAgentID(msg.Chat.ID)
	name := agentLabel(c.getAgentInfo(agentID), agentID)
	text := fmt.Sprintf("👋 *Welcome!* You are now talking to *%s*.\n\nType /help to see available commands.", name)
	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Text:            text,
		ParseMode:       "Markdown",
	})
	return nil
}

// handleHelpCommand responds to /help with a list of all supported bot commands.
func (c *Client) handleHelpCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		return nil
	}

	text := "*Available commands:*\n\n" +
		"/help — Show this help message\n" +
		"/agent — Show or switch the active agent\n" +
		"/reset — Reset the conversation session\n" +
		"/responsemode — Show or change the response mode\n" +
		"/showtools — Toggle tool call visibility\n" +
		"/start — Show the welcome message"

	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Text:            text,
		ParseMode:       "Markdown",
	})
	return nil
}

// handleAgentCommand responds to /agent. Without arguments it shows the active
// agent and lists all available ones. With an agent ID it switches the chat to
// that agent.
func (c *Client) handleAgentCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
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

		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            fmt.Sprintf("*Active agent:* %s\n\n*Available agents:*\n%s\nUsage: `/agent <id>`", currentLabel, agentList),
			ParseMode:       "Markdown",
		})
		return nil
	}

	if c.getAgentInfo(args) == nil {
		ids := make([]string, 0, len(c.agents))
		for _, a := range c.agents {
			ids = append(ids, "`"+a.ID+"`")
		}
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            fmt.Sprintf("Unknown agent `%s`. Available: %s", args, strings.Join(ids, ", ")),
			ParseMode:       "Markdown",
		})
		return nil
	}

	c.setActiveAgentID(msg.Chat.ID, args)
	c.logger.Info("Agent switched", "chat_id", msg.Chat.ID, "user_id", msg.From.ID, "agent", args)

	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Text:            fmt.Sprintf("Switched to agent *%s* (`%s`)", agentLabel(c.getAgentInfo(args), args), args),
		ParseMode:       "Markdown",
	})
	return nil
}

// handleResetCommand deletes the ADK session for the current chat and thread.
func (c *Client) handleResetCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		return nil
	}

	agentID := c.getActiveAgentID(msg.Chat.ID)
	sessionID := c.buildSessionID(msg.Chat.ID, msg.MessageThreadID, agentID)
	if err := c.deleteSession(agentID, sessionID); err != nil {
		c.logger.Error("Failed to delete session", "error", err)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            "Failed to reset session.",
		})
		return nil
	}

	c.logger.Info("Session reset",
		"chat_id", msg.Chat.ID,
		"user_id", msg.From.ID,
		"agent", agentID,
		"session", sessionID,
	)

	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Text:            fmt.Sprintf("Session reset for *%s*. Next message starts a fresh conversation.", agentLabel(c.getAgentInfo(agentID), agentID)),
		ParseMode:       "Markdown",
	})
	return nil
}

// handleMessage processes a regular text message: checks permissions, sends a
// typing indicator, forwards the text to the active agent via SSE, and delivers
// each response event incrementally.
func (c *Client) handleMessage(ctx *th.Context, msg telego.Message) error {
	if msg.Text == "" {
		return nil
	}
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		c.logger.Debug("Unauthorized access attempt", "user_id", msg.From.ID, "chat_id", msg.Chat.ID)
		return nil
	}

	c.logger.Info("Received message", "user_id", msg.From.ID, "chat_id", msg.Chat.ID, "text", msg.Text)
	c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👀")
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Action:          telego.ChatActionTyping,
	})
	c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "🧠")

	typingDone := c.startTypingLoop(ctx, msg.Chat.ID, msg.MessageThreadID)

	inputText, truncated := msgutil.ValidateInputLength(msg.Text, msgutil.DefaultMaxInputLength)
	if truncated {
		c.logger.Warn("Inbound message truncated", "chat_id", msg.Chat.ID, "original_len", len([]rune(msg.Text)))
	}

	agentID := c.getActiveAgentID(msg.Chat.ID)
	sessionID := c.buildSessionID(msg.Chat.ID, msg.MessageThreadID, agentID)
	userIDStr := "default_user"
	artifactsBefore := c.listArtifacts(agentID, userIDStr, sessionID)

	hasText := false
	hasToolActivity := false
	eventCount := 0
	var lastFinishReason, lastErrorMessage string
	var lastUsage *msgutil.UsageMetadata
	toolCount := 0
	var toolCounterMsgID int

	err := c.callAgentSSE(msg, agentID, sessionID, inputText, func(evt msgutil.SSEEvent) {
		eventCount++
		if evt.FinishReason != "" {
			lastFinishReason = evt.FinishReason
		}
		if evt.ErrorMessage != "" {
			lastErrorMessage = evt.ErrorMessage
		}
		if evt.UsageMetadata != nil {
			lastUsage = evt.UsageMetadata
		}
		switch evt.Type {
		case msgutil.SSEEventText:
			hasText = true
			toolCount = 0
			toolCounterMsgID = 0
			c.sendTextResponse(ctx, msg.Chat.ID, msg.MessageThreadID, evt.Text, false)
		case msgutil.SSEEventToolCall:
			hasToolActivity = true
			toolCounterMsgID = c.sendToolCounter(ctx, msg.Chat.ID, msg.MessageThreadID, toolCounterMsgID, &toolCount, evt)
		case msgutil.SSEEventToolResult:
			hasToolActivity = true
			if c.getShowTools() {
				_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
					ChatID:          tu.ID(msg.Chat.ID),
					MessageThreadID: msg.MessageThreadID,
					Text:            msgutil.FormatToolResultTelegram(evt),
					ParseMode:       "HTML",
				})
			}
		case msgutil.SSEEventError:
			c.logger.Error("Agent stream error",
				"chat_id", msg.Chat.ID,
				"error_code", evt.ErrorCode,
				"error_message", evt.ErrorMessage,
			)
		}
	})
	close(typingDone)

	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👎")
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            fmt.Sprintf("Failed to process your request: %s", sanitizeError(err)),
		})
		return nil
	}

	if !hasText && !hasToolActivity {
		logFields := []any{
			"chat_id", msg.Chat.ID,
			"agent", agentID,
			"session", sessionID,
			"events_received", eventCount,
			"finish_reason", lastFinishReason,
		}
		if lastUsage != nil {
			logFields = append(logFields, "prompt_tokens", lastUsage.PromptTokens, "total_tokens", lastUsage.TotalTokens)
		}
		c.logger.Warn("No text in agent response", logFields...)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            msgutil.ExplainNoResponse(lastFinishReason, lastErrorMessage),
		})
	}

	c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👍")
	c.sendNewArtifacts(ctx, msg.Chat.ID, msg.MessageThreadID, agentID, userIDStr, sessionID, artifactsBefore)

	return nil
}

// handleVoice processes a voice message: downloads the audio from Telegram,
// transcribes it via the magec transcription proxy, forwards to the agent via
// SSE, and delivers each response event incrementally.
func (c *Client) handleVoice(ctx *th.Context, msg telego.Message) error {
	if msg.Voice == nil {
		return nil
	}
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		return nil
	}

	c.logger.Info("Received voice message", "user_id", msg.From.ID, "chat_id", msg.Chat.ID, "duration", msg.Voice.Duration)
	c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👀")
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Action:          telego.ChatActionTyping,
	})

	file, err := ctx.Bot().GetFile(ctx, &telego.GetFileParams{FileID: msg.Voice.FileID})
	if err != nil {
		c.logger.Error("Failed to get voice file", "error", err)
		c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👎")
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            "Failed to download your voice message. Please try again.",
		})
		return nil
	}

	fileURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.clientDef.Config.Telegram.BotToken, file.FilePath)
	audioData, err := c.downloadFile(fileURL)
	if err != nil {
		c.logger.Error("Failed to download voice file", "error", err)
		c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👎")
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            "Failed to download your voice message. Please try again.",
		})
		return nil
	}

	agentID := c.getActiveAgentID(msg.Chat.ID)
	text, err := c.transcribeAudio(audioData, file.FilePath, agentID)
	if err != nil {
		c.logger.Error("Failed to transcribe audio", "error", err)
		c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👎")
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            "Sorry, I couldn't transcribe your voice message.",
		})
		return nil
	}

	c.logger.Info("Transcribed voice", "text", text)
	c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "🧠")

	typingDone := c.startTypingLoop(ctx, msg.Chat.ID, msg.MessageThreadID)

	voiceInput, truncated := msgutil.ValidateInputLength(text, msgutil.DefaultMaxInputLength)
	if truncated {
		c.logger.Warn("Transcribed voice message truncated", "chat_id", msg.Chat.ID, "original_len", len([]rune(text)))
	}

	sessionID := c.buildSessionID(msg.Chat.ID, msg.MessageThreadID, agentID)
	userIDStr := "default_user"
	artifactsBefore := c.listArtifacts(agentID, userIDStr, sessionID)

	var lastTextResponse string
	hasText := false
	hasToolActivity := false
	var lastFinishReason, lastErrorMessage string
	toolCount := 0
	var toolCounterMsgID int

	err = c.callAgentSSE(msg, agentID, sessionID, voiceInput, func(evt msgutil.SSEEvent) {
		if evt.FinishReason != "" {
			lastFinishReason = evt.FinishReason
		}
		if evt.ErrorMessage != "" {
			lastErrorMessage = evt.ErrorMessage
		}
		switch evt.Type {
		case msgutil.SSEEventText:
			hasText = true
			lastTextResponse = evt.Text
			toolCount = 0
			toolCounterMsgID = 0
			c.sendTextResponse(ctx, msg.Chat.ID, msg.MessageThreadID, evt.Text, true)
		case msgutil.SSEEventToolCall:
			hasToolActivity = true
			toolCounterMsgID = c.sendToolCounter(ctx, msg.Chat.ID, msg.MessageThreadID, toolCounterMsgID, &toolCount, evt)
		case msgutil.SSEEventToolResult:
			hasToolActivity = true
			if c.getShowTools() {
				_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
					ChatID:          tu.ID(msg.Chat.ID),
					MessageThreadID: msg.MessageThreadID,
					Text:            msgutil.FormatToolResultTelegram(evt),
					ParseMode:       "HTML",
				})
			}
		case msgutil.SSEEventError:
			c.logger.Error("Agent stream error",
				"chat_id", msg.Chat.ID,
				"error_code", evt.ErrorCode,
				"error_message", evt.ErrorMessage,
			)
		}
	})
	close(typingDone)

	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👎")
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            fmt.Sprintf("Failed to process your request: %s", sanitizeError(err)),
		})
		return nil
	}

	if !hasText && !hasToolActivity {
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            msgutil.ExplainNoResponse(lastFinishReason, lastErrorMessage),
		})
	}

	mode := c.getResponseMode()
	if (mode == ResponseModeVoice || mode == ResponseModeBoth || mode == ResponseModeMirror) && lastTextResponse != "" {
		c.sendVoiceResponse(ctx, msg.Chat.ID, msg.MessageThreadID, lastTextResponse, agentID)
	}

	c.setReaction(ctx, msg.Chat.ID, msg.MessageID, "👍")
	c.sendNewArtifacts(ctx, msg.Chat.ID, msg.MessageThreadID, agentID, userIDStr, sessionID, artifactsBefore)

	return nil
}

// startTypingLoop sends a typing indicator every 4 seconds until the returned
// channel is closed. Call close(done) when the agent response is complete.
func (c *Client) startTypingLoop(ctx *th.Context, chatID int64, threadID int) chan struct{} {
	done := make(chan struct{})
	go func() {
		ticker := time.NewTicker(4 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
					ChatID:          tu.ID(chatID),
					MessageThreadID: threadID,
					Action:          telego.ChatActionTyping,
				})
			}
		}
	}()
	return done
}

// sendToolCounter posts or edits a compact tool activity counter in the chat.
// When showTools is enabled it posts the full tool call instead.
// Returns the updated message ID for subsequent edits.
func (c *Client) sendToolCounter(ctx *th.Context, chatID int64, threadID int, counterMsgID int, toolCount *int, evt msgutil.SSEEvent) int {
	if c.getShowTools() {
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(chatID),
			MessageThreadID: threadID,
			Text:            msgutil.FormatToolCallTelegram(evt),
			ParseMode:       "HTML",
		})
		return counterMsgID
	}

	*toolCount++
	counterText := fmt.Sprintf("⚙️ x%d", *toolCount)

	if counterMsgID == 0 {
		sent, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(chatID),
			MessageThreadID: threadID,
			Text:            counterText,
		})
		if err == nil {
			return sent.MessageID
		}
		return 0
	}

	_, _ = ctx.Bot().EditMessageText(ctx, &telego.EditMessageTextParams{
		ChatID:    tu.ID(chatID),
		MessageID: counterMsgID,
		Text:      counterText,
	})
	return counterMsgID
}

// sendTextResponse delivers a text message to the chat, splitting if needed.
// If inputWasVoice is true and the mode is voice-only or mirror, text is suppressed.
func (c *Client) sendTextResponse(ctx *th.Context, chatID int64, threadID int, text string, inputWasVoice bool) {
	mode := c.getResponseMode()
	switch mode {
	case ResponseModeVoice:
		return
	case ResponseModeMirror:
		if inputWasVoice {
			return
		}
	}

	chunks := msgutil.SplitMessage(text, msgutil.TelegramMaxMessageLength)
	for _, chunk := range chunks {
		_, err := ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(chatID),
			MessageThreadID: threadID,
			Text:            chunk,
		})
		if err != nil {
			c.logger.Error("Failed to send message", "error", err)
			break
		}
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
// the current mode. With a mode name it overrides the config default. "reset"
// clears the override back to the config value.
func (c *Client) handleResponseModeCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		return nil
	}

	validModes := []string{ResponseModeText, ResponseModeVoice, ResponseModeMirror, ResponseModeBoth}
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
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            status,
			ParseMode:       "Markdown",
		})
		return nil
	}

	if args == "reset" {
		c.responseMu.Lock()
		c.responseModeOverride = ""
		c.responseMu.Unlock()
		c.logger.Info("Response mode override cleared", "user_id", msg.From.ID, "config_mode", c.clientDef.Config.Telegram.ResponseMode)
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            fmt.Sprintf("Response mode reset to config default: `%s`", c.clientDef.Config.Telegram.ResponseMode),
			ParseMode:       "Markdown",
		})
		return nil
	}

	if !slices.Contains(validModes, args) {
		_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
			ChatID:          tu.ID(msg.Chat.ID),
			MessageThreadID: msg.MessageThreadID,
			Text:            fmt.Sprintf("Invalid mode `%s`. Valid options: `%s`, `reset`", args, strings.Join(validModes, "`, `")),
			ParseMode:       "Markdown",
		})
		return nil
	}

	c.responseMu.Lock()
	c.responseModeOverride = args
	c.responseMu.Unlock()
	c.logger.Info("Response mode overridden", "user_id", msg.From.ID, "new_mode", args)
	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Text:            fmt.Sprintf("Response mode set to `%s` (until restart)", args),
		ParseMode:       "Markdown",
	})
	return nil
}

func (c *Client) getShowTools() bool {
	c.showToolsMu.RLock()
	defer c.showToolsMu.RUnlock()
	return c.showTools
}

func (c *Client) handleShowToolsCommand(ctx *th.Context, msg telego.Message) error {
	if !c.isAllowed(msg.From.ID, msg.Chat.ID, msg.MessageThreadID) {
		return nil
	}

	c.showToolsMu.Lock()
	c.showTools = !c.showTools
	state := c.showTools
	c.showToolsMu.Unlock()

	label := "OFF"
	if state {
		label = "ON"
	}
	_, _ = ctx.Bot().SendMessage(ctx, &telego.SendMessageParams{
		ChatID:          tu.ID(msg.Chat.ID),
		MessageThreadID: msg.MessageThreadID,
		Text:            fmt.Sprintf("🔧 Tool call visibility: *%s*", label),
		ParseMode:       "Markdown",
	})
	return nil
}

// isAllowed checks whether a Telegram user or chat is permitted to interact
// with this bot. If no allowlists are configured, all users are allowed.
func (c *Client) isAllowed(userID, chatID int64, threadID int) bool {
	if len(c.clientDef.Config.Telegram.AllowedUsers) == 0 && len(c.clientDef.Config.Telegram.AllowedChatThreads) == 0 {
		return true
	}
	if len(c.clientDef.Config.Telegram.AllowedUsers) > 0 && slices.Contains(c.clientDef.Config.Telegram.AllowedUsers, userID) {
		return true
	}
	for _, pair := range c.clientDef.Config.Telegram.AllowedChatThreads {
		if pair.ChatID != chatID {
			continue
		}
		if pair.ThreadID == nil || *pair.ThreadID == threadID {
			return true
		}
	}
	return false
}

// buildMessageContext prepends invisible metadata to the user's message so the
// LLM knows who is writing and from which chat.
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

	if name := strings.TrimSpace(msg.From.FirstName + " " + msg.From.LastName); name != "" {
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

// buildSessionID builds a stable session ID scoped to a chat/thread + agent.
// When threadID is set, the session is scoped to that thread; otherwise to the chat.
func (c *Client) buildSessionID(chatID int64, threadID int, agentID string) string {
	if threadID != 0 {
		return fmt.Sprintf("telegram_%d_%d_%s", chatID, threadID, agentID)
	}
	return fmt.Sprintf("telegram_%d_%s", chatID, agentID)
}

// callAgentSSE sends a user message to the active agent via the /run_sse endpoint
// and calls handler for each event as it arrives from the SSE stream.
func (c *Client) callAgentSSE(msg telego.Message, agentID, sessionID, message string, handler func(msgutil.SSEEvent)) error {
	userIDStr := "default_user"

	if err := c.ensureSession(agentID, userIDStr, sessionID); err != nil {
		c.logger.Warn("Failed to ensure session, continuing anyway", "error", err)
	}

	fullMessage := c.buildMessageContext(msg) + message

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
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", c.agentURL+"/run_sse", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")
	c.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to call agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	return msgutil.ParseSSEStream(resp.Body, handler)
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

func (c *Client) deleteSession(agentID, sessionID string) error {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions/%s", c.agentURL, agentID, "default_user", sessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return err
	}
	c.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		return fmt.Errorf("failed to delete session: status %d", resp.StatusCode)
	}
	return nil
}

// setAuthHeader adds the Bearer token to requests sent to the magec API.
func (c *Client) setAuthHeader(req *http.Request) {
	if c.clientDef.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.clientDef.Token)
	}
}

// downloadFile fetches a file by URL. Used to download voice messages from
// the Telegram file API (no auth header added — not a magec endpoint).
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

// transcribeAudio converts a voice message to text. It re-encodes the audio
// to WAV, then sends it to the magec transcription proxy which routes it to
// the backend configured for the active agent.
func (c *Client) transcribeAudio(audioData []byte, filePath string, agentID string) (string, error) {
	wavData, err := c.convertOggToWav(audioData)
	if err != nil {
		return "", fmt.Errorf("failed to convert audio: %w", err)
	}

	var buf bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"audio.wav\"\r\n")
	buf.WriteString("Content-Type: audio/wav\r\n\r\n")
	buf.Write(wavData)
	buf.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	transcriptionURL := strings.TrimSuffix(c.agentURL, "/agent") + "/voice/" + agentID + "/transcription"

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

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
func (c *Client) sendVoiceResponse(ctx *th.Context, chatID int64, threadID int, text string, agentID string) {
	_ = ctx.Bot().SendChatAction(ctx, &telego.SendChatActionParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		Action:          telego.ChatActionRecordVoice,
	})

	audioData, err := c.generateTTS(text, agentID)
	if err != nil {
		c.logger.Error("Failed to generate TTS", "error", err)
		return
	}

	_, err = ctx.Bot().SendVoice(ctx, &telego.SendVoiceParams{
		ChatID:          tu.ID(chatID),
		MessageThreadID: threadID,
		Voice:           tu.FileFromReader(bytes.NewReader(audioData), "voice.ogg"),
	})
	if err != nil {
		c.logger.Error("Failed to send voice message", "error", err)
	}
}

// generateTTS calls the magec TTS proxy for the given agent.
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

func (c *Client) setReaction(ctx context.Context, chatID int64, messageID int, emoji string) {
	var reaction []telego.ReactionType
	if emoji != "" {
		reaction = []telego.ReactionType{
			&telego.ReactionTypeEmoji{Type: "emoji", Emoji: emoji},
		}
	}
	err := c.bot.SetMessageReaction(ctx, &telego.SetMessageReactionParams{
		ChatID:    tu.ID(chatID),
		MessageID: messageID,
		Reaction:  reaction,
	})
	if err != nil {
		c.logger.Debug("Failed to set reaction", "emoji", emoji, "error", err)
	}
}

func sanitizeError(err error) string {
	msg := err.Error()
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	for _, secret := range []string{"Bearer ", "bot", "token"} {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(secret)) {
			return "an internal error occurred"
		}
	}
	return msg
}

func (c *Client) listArtifacts(agentID, userID, sessionID string) []string {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions/%s/artifacts", c.agentURL, agentID, userID, sessionID)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		c.logger.Debug("Failed to create artifact list request", "error", err)
		return nil
	}
	c.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c.logger.Debug("Failed to list artifacts", "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var result []string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		c.logger.Debug("Failed to decode artifact list", "error", err)
		return nil
	}
	return result
}

func (c *Client) downloadArtifact(agentID, userID, sessionID, name string) ([]byte, string, error) {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions/%s/artifacts/%s", c.agentURL, agentID, userID, sessionID, name)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, "", err
	}
	c.setAuthHeader(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("artifact download returned status %d", resp.StatusCode)
	}

	var artifact struct {
		Text       string `json:"text,omitempty"`
		InlineData *struct {
			MIMEType string `json:"mimeType"`
			Data     string `json:"data"`
		} `json:"inlineData,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&artifact); err != nil {
		return nil, "", fmt.Errorf("failed to decode artifact: %w", err)
	}

	if artifact.InlineData != nil {
		data, err := base64.StdEncoding.DecodeString(artifact.InlineData.Data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode artifact binary data: %w", err)
		}
		return data, artifact.InlineData.MIMEType, nil
	}

	return []byte(artifact.Text), "text/plain", nil
}

func (c *Client) sendNewArtifacts(ctx *th.Context, chatID int64, threadID int, agentID, userID, sessionID string, before []string) {
	after := c.listArtifacts(agentID, userID, sessionID)
	if len(after) == 0 {
		return
	}

	beforeSet := make(map[string]bool, len(before))
	for _, name := range before {
		beforeSet[name] = true
	}

	for _, name := range after {
		if beforeSet[name] {
			continue
		}

		data, _, err := c.downloadArtifact(agentID, userID, sessionID, name)
		if err != nil {
			c.logger.Error("Failed to download artifact", "name", name, "error", err)
			continue
		}

		_, err = ctx.Bot().SendDocument(ctx, &telego.SendDocumentParams{
			ChatID:          tu.ID(chatID),
			MessageThreadID: threadID,
			Document:        tu.FileFromReader(bytes.NewReader(data), name),
		})
		if err != nil {
			c.logger.Error("Failed to send artifact as document", "name", name, "error", err)
		}
	}
}
