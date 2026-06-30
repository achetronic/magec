// SPDX-FileCopyrightText: 2026 Alby Hernández <hola@achetronic.com>
// SPDX-License-Identifier: Apache-2.0

package slack

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"strings"
	"sync"
	"time"

	slackapi "github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"github.com/slack-go/slack/socketmode"
	"google.golang.org/adk/artifact"

	"github.com/achetronic/magec/server/clients/msgutil"
	"github.com/achetronic/magec/server/store"
)

const (
	ResponseModeText   = "text"
	ResponseModeVoice  = "voice"
	ResponseModeMirror = "mirror"
	ResponseModeBoth   = "both"
)

type AgentInfo struct {
	ID   string
	Name string
}

type Client struct {
	clientDef store.ClientDefinition
	agentURL  string
	agents    []AgentInfo
	logger    *slog.Logger
	store     interface {
		UpdateClient(id string, c store.ClientDefinition) error
	}

	api    *slackapi.Client
	socket *socketmode.Client
	cancel context.CancelFunc

	activeAgentMu sync.RWMutex
	activeAgent   map[string]string // channelID -> agentID

	responseMu           sync.RWMutex
	responseModeOverride string

	showToolsMu sync.RWMutex
	showTools   bool

	seenMu sync.Mutex
	seen   map[string]struct{}

	botUserID string

	// artifacts is optional: when set, user attachments
	// are persisted here instead of being dropped.
	// Set via SetArtifactService before Start.
	artifacts artifact.Service
}

// SetArtifactService enables offloading large user attachments to the session
// artifact store. When unset, all files are inlined regardless of size.
func (c *Client) SetArtifactService(a artifact.Service) {
	c.artifacts = a
}

func New(clientDef store.ClientDefinition, agentURL string, agents []AgentInfo, s interface {
	UpdateClient(id string, c store.ClientDefinition) error
}, logger *slog.Logger) (*Client, error) {
	if clientDef.Config.Slack == nil {
		return nil, fmt.Errorf("slack config is required")
	}
	cfg := clientDef.Config.Slack
	if cfg.BotToken == "" {
		return nil, fmt.Errorf("slack bot token is required")
	}
	if cfg.AppToken == "" {
		return nil, fmt.Errorf("slack app token is required")
	}

	api := slackapi.New(
		cfg.BotToken,
		slackapi.OptionAppLevelToken(cfg.AppToken),
	)

	return &Client{
		seen:        make(map[string]struct{}),
		api:         api,
		socket:      socketmode.New(api),
		clientDef:   clientDef,
		agentURL:    agentURL,
		agents:      agents,
		store:       s,
		activeAgent: make(map[string]string),
		logger:      logger,
	}, nil
}

func (c *Client) Start(ctx context.Context) error {
	authResp, err := c.api.AuthTestContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to authenticate slack bot: %w", err)
	}
	c.botUserID = authResp.UserID
	c.logger.Info("Slack bot started", "bot_user_id", c.botUserID, "team", authResp.Team)

	smCtx, cancel := context.WithCancel(ctx)
	c.cancel = cancel

	go func() {
		for evt := range c.socket.Events {
			switch evt.Type {
			case socketmode.EventTypeEventsAPI:
				c.handleEventsAPI(evt)
			case socketmode.EventTypeConnected:
				c.logger.Info("Slack Socket Mode connected")
			case socketmode.EventTypeConnectionError:
				c.logger.Error("Slack Socket Mode connection error", "data", evt.Data)
			}
		}
	}()

	if err := c.socket.RunContext(smCtx); err != nil {
		if smCtx.Err() != nil {
			return nil
		}
		return fmt.Errorf("slack socket mode error: %w", err)
	}
	return nil
}

func (c *Client) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.logger.Info("Slack bot stopped")
}

func (c *Client) handleEventsAPI(evt socketmode.Event) {
	eventsAPIEvent, ok := evt.Data.(slackevents.EventsAPIEvent)
	if !ok {
		c.socket.Ack(*evt.Request)
		return
	}
	c.socket.Ack(*evt.Request)

	switch eventsAPIEvent.InnerEvent.Type {
	case string(slackevents.AppMention):
		c.handleAppMention(eventsAPIEvent)
	case string(slackevents.Message):
		c.handleMessage(eventsAPIEvent)
	}
}

func (c *Client) isDuplicate(ts string) bool {
	c.seenMu.Lock()
	defer c.seenMu.Unlock()
	if _, dup := c.seen[ts]; dup {
		return true
	}
	c.seen[ts] = struct{}{}
	if len(c.seen) > 1000 {
		for k := range c.seen {
			delete(c.seen, k)
			break
		}
	}
	return false
}

func (c *Client) handleAppMention(event slackevents.EventsAPIEvent) {
	ev, ok := event.InnerEvent.Data.(*slackevents.AppMentionEvent)
	if !ok || ev == nil || ev.User == c.botUserID {
		return
	}
	if c.isDuplicate(ev.TimeStamp) {
		return
	}
	if !c.isAllowed(ev.User, ev.Channel) {
		c.logger.Debug("Unauthorized Slack mention", "user", ev.User, "channel", ev.Channel)
		return
	}

	text := c.stripBotMention(ev.Text)
	if text == "" {
		return
	}

	// If the mention is in the channel root, anchor the reply thread to this message.
	// If it's already inside a thread, keep that thread's TS.
	threadTS := ev.ThreadTimeStamp
	if threadTS == "" {
		threadTS = ev.TimeStamp
	}

	if c.handleBotCommand(ev.User, ev.Channel, text, threadTS) {
		return
	}

	c.logger.Info("Slack mention received", "user", ev.User, "channel", ev.Channel, "text", text)
	c.processMessage(ev.User, ev.Channel, "channel", text, threadTS, ev.TimeStamp, event.TeamID, false, nil)
}

func (c *Client) handleMessage(event slackevents.EventsAPIEvent) {
	ev, ok := event.InnerEvent.Data.(*slackevents.MessageEvent)
	if !ok || ev == nil {
		return
	}
	if ev.SubType != "" || ev.BotID != "" || ev.User == c.botUserID || ev.User == "" {
		return
	}
	if ev.ChannelType != "im" {
		return
	}
	if c.isDuplicate(ev.TimeStamp) {
		return
	}
	if !c.isAllowed(ev.User, ev.Channel) {
		c.logger.Debug("Unauthorized Slack DM", "user", ev.User, "channel", ev.Channel)
		return
	}

	if c.handleAudioClip(ev, event.TeamID) {
		return
	}

	files := c.collectNonAudioFiles(ev)

	text := strings.TrimSpace(ev.Text)
	if text == "" && len(files) == 0 {
		return
	}

	if text != "" && c.handleBotCommand(ev.User, ev.Channel, text, ev.ThreadTimeStamp) {
		return
	}

	c.logger.Info("Slack DM received", "user", ev.User, "channel", ev.Channel, "text", text, "attachments", len(files))
	c.processMessage(ev.User, ev.Channel, "im", text, ev.ThreadTimeStamp, ev.TimeStamp, event.TeamID, false, files)
}

func (c *Client) handleAudioClip(ev *slackevents.MessageEvent, teamID string) bool {
	if ev.Message == nil || len(ev.Message.Files) == 0 {
		return false
	}

	for _, file := range ev.Message.Files {
		if !strings.HasPrefix(file.Mimetype, "audio/") {
			continue
		}

		c.logger.Info("Slack audio clip received",
			"user", ev.User,
			"channel", ev.Channel,
			"mimetype", file.Mimetype,
			"size", file.Size,
			"fileID", file.ID,
		)

		downloadURL := c.resolveFileURL(file.ID, file.URLPrivateDownload, file.URLPrivate)
		if downloadURL == "" {
			c.logger.Error("Audio clip has no download URL")
			c.postMessage(ev.Channel, "Sorry, I couldn't download your audio clip.", "")
			return true
		}

		audioData, err := c.downloadSlackFile(downloadURL)
		if err != nil {
			c.logger.Error("Failed to download audio clip", "error", err)
			c.postMessage(ev.Channel, "Sorry, I couldn't download your audio clip.", "")
			return true
		}

		agentID := c.getActiveAgentID(ev.Channel)

		wavData, err := c.convertAudioToWav(audioData)
		if err != nil {
			c.logger.Error("Failed to convert audio clip", "error", err)
			c.postMessage(ev.Channel, "Sorry, I couldn't process your audio clip.", "")
			return true
		}

		text, err := c.transcribeAudio(wavData, agentID)
		if err != nil {
			c.logger.Error("Failed to transcribe audio clip", "error", err)
			c.postMessage(ev.Channel, "Sorry, I couldn't transcribe your audio clip.", "")
			return true
		}

		c.logger.Info("Transcribed audio clip", "text", text)
		c.processMessage(ev.User, ev.Channel, "im", text, ev.ThreadTimeStamp, ev.TimeStamp, teamID, true, nil)
		return true
	}
	return false
}

// downloadedFile groups an attachment ready for the inline-vs-artifact policy.
type downloadedFile struct {
	Name     string
	MIMEType string
	Data     []byte
}

// collectNonAudioFiles downloads every non-audio attachment of a Slack
// message and returns them in a flat list. The inline-vs-artifact decision
// happens later, in processMessage, where the session context is known.
// Audio is filtered out because it flows through handleAudioClip's STT path.
func (c *Client) collectNonAudioFiles(ev *slackevents.MessageEvent) []downloadedFile {
	if ev == nil || ev.Message == nil || len(ev.Message.Files) == 0 {
		return nil
	}
	const maxFileSize = 20 * 1024 * 1024 // Slack's own upload cap
	var out []downloadedFile
	for _, file := range ev.Message.Files {
		if strings.HasPrefix(file.Mimetype, "audio/") {
			continue
		}
		if file.Size > maxFileSize {
			c.logger.Warn("Slack file too large to process",
				"name", file.Name,
				"mimetype", file.Mimetype,
				"size", file.Size,
			)
			continue
		}
		downloadURL := c.resolveFileURL(file.ID, file.URLPrivateDownload, file.URLPrivate)
		if downloadURL == "" {
			c.logger.Debug("Slack file has no download URL", "name", file.Name)
			continue
		}
		data, err := c.downloadSlackFile(downloadURL)
		if err != nil {
			c.logger.Warn("Failed to download Slack file", "name", file.Name, "error", err)
			continue
		}
		if len(data) > maxFileSize {
			c.logger.Warn("Slack file exceeded size limit after download", "name", file.Name, "size", len(data))
			continue
		}
		name := file.Name
		if name == "" {
			name = "attachment"
		}
		out = append(out, downloadedFile{
			Name:     fmt.Sprintf("%s_%s", ev.TimeStamp, name),
			MIMEType: file.Mimetype,
			Data:     data,
		})
	}
	return out
}

// resolveFileURL returns the best available download URL for a Slack file,
// preferring the fresher URL from the API over the one embedded in the event.
func (c *Client) resolveFileURL(fileID, eventDownloadURL, eventPrivateURL string) string {
	if fullFile, _, _, err := c.api.GetFileInfo(fileID, 0, 0); err == nil && fullFile != nil {
		if fullFile.URLPrivateDownload != "" {
			return fullFile.URLPrivateDownload
		}
		if fullFile.URLPrivate != "" {
			return fullFile.URLPrivate
		}
	}
	if eventDownloadURL != "" {
		return eventDownloadURL
	}
	return eventPrivateURL
}

func (c *Client) handleBotCommand(userID, channelID, text, threadTS string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))

	if !strings.HasPrefix(lower, "!") {
		return false
	}
	lower = lower[1:]
	text = strings.TrimSpace(text[1:])

	if lower == "help" {
		helpText := "*Available commands:*\n" +
			"• `!help` — Show this help message\n" +
			"• `!agent` — Show or switch the active agent\n" +
			"• `!agent <id>` — Switch to a specific agent\n" +
			"• `!reset` — Reset the conversation session\n" +
			"• `!responsemode` — Show or change the response mode\n" +
			"• `!responsemode <mode>` — Set response mode (`text`, `voice`, `mirror`, `both`, `reset`)\n" +
			"• `!showtools` — Toggle tool call visibility"
		c.postMessage(channelID, helpText, threadTS)
		return true
	}

	if lower == "agent" {
		currentID := c.getActiveAgentID(channelID)
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
		c.postMessage(channelID, fmt.Sprintf("*Active agent:* %s\n\n*Available agents:*\n%s\nUsage: `!agent <id>`", currentLabel, agentList), threadTS)
		return true
	}

	if strings.HasPrefix(lower, "agent ") {
		agentID := strings.TrimSpace(text[6:])
		found := slices.ContainsFunc(c.agents, func(a AgentInfo) bool { return a.ID == agentID })
		if !found {
			var ids []string
			for _, a := range c.agents {
				ids = append(ids, "`"+a.ID+"`")
			}
			c.postMessage(channelID, fmt.Sprintf("Unknown agent `%s`. Available: %s", agentID, strings.Join(ids, ", ")), threadTS)
			return true
		}
		c.setActiveAgentID(channelID, agentID)
		agent := c.getAgentInfo(agentID)
		label := agentID
		if agent != nil && agent.Name != "" {
			label = agent.Name
		}
		c.logger.Info("Slack agent switched", "channel", channelID, "agent", agentID)
		c.postMessage(channelID, fmt.Sprintf("Switched to agent *%s* (`%s`)", label, agentID), threadTS)
		return true
	}

	if lower == "reset" {
		agentID := c.getActiveAgentID(channelID)
		sessionID := c.buildSessionID(channelID, threadTS, agentID)
		if err := c.deleteSession(agentID, sessionID); err != nil {
			c.logger.Error("Failed to delete session", "error", err)
			c.postMessage(channelID, "Failed to reset session.", threadTS)
			return true
		}
		c.logger.Info("Session reset", "channel", channelID, "agent", agentID, "session", sessionID)
		agent := c.getAgentInfo(agentID)
		label := agentID
		if agent != nil && agent.Name != "" {
			label = agent.Name
		}
		c.postMessage(channelID, fmt.Sprintf("Session reset for *%s*. Next message starts a fresh conversation.", label), threadTS)
		return true
	}

	if lower == "responsemode" {
		current := c.getResponseMode()
		c.responseMu.RLock()
		overridden := c.responseModeOverride != ""
		c.responseMu.RUnlock()
		status := fmt.Sprintf("*Response mode:* `%s`", current)
		if overridden {
			status += fmt.Sprintf(" (override, config: `%s`)", c.clientDef.Config.Slack.ResponseMode)
		}
		status += "\n*Options:* `text`, `voice`, `mirror`, `both`, `reset`"
		c.postMessage(channelID, status, threadTS)
		return true
	}

	if strings.HasPrefix(lower, "responsemode ") {
		arg := strings.TrimSpace(text[13:])
		return c.handleResponseModeCommand(channelID, arg, threadTS)
	}

	if lower == "showtools" {
		c.showToolsMu.Lock()
		c.showTools = !c.showTools
		state := c.showTools
		c.showToolsMu.Unlock()
		label := "OFF"
		if state {
			label = "ON"
		}
		c.postMessage(channelID, fmt.Sprintf("🔧 Tool call visibility: *%s*", label), threadTS)
		return true
	}

	return false
}

func (c *Client) handleResponseModeCommand(channelID, arg, threadTS string) bool {
	validModes := []string{ResponseModeText, ResponseModeVoice, ResponseModeMirror, ResponseModeBoth}

	if arg == "reset" {
		c.responseMu.Lock()
		c.responseModeOverride = ""
		c.responseMu.Unlock()
		c.logger.Info("Response mode override cleared", "config_mode", c.clientDef.Config.Slack.ResponseMode)
		c.postMessage(channelID, fmt.Sprintf("Response mode reset to config default: `%s`", c.clientDef.Config.Slack.ResponseMode), threadTS)
		return true
	}

	if !slices.Contains(validModes, arg) {
		c.postMessage(channelID, fmt.Sprintf("Invalid mode `%s`. Valid options: `text`, `voice`, `mirror`, `both`, `reset`", arg), threadTS)
		return true
	}

	c.responseMu.Lock()
	c.responseModeOverride = arg
	c.responseMu.Unlock()

	c.logger.Info("Response mode overridden", "new_mode", arg)
	c.postMessage(channelID, fmt.Sprintf("Response mode set to `%s` (until restart)", arg), threadTS)
	return true
}

func (c *Client) processMessage(userID, channelID, channelType, text, threadTS, messageTS, teamID string, inputWasVoice bool, files []downloadedFile) {
	msgRef := slackapi.NewRefToMessage(channelID, messageTS)
	c.addReaction("eyes", msgRef)

	agentID := c.getActiveAgentID(channelID)
	sessionID := c.buildSessionID(channelID, threadTS, agentID)
	userID0 := "default_user"

	// Persist attached files as session artifacts so the LLM can call
	// load_artifact on demand. See decision #24 in .agents/DECISIONS.md.
	var artifactLines []string
	ctx := context.Background()
	for _, f := range files {
		if c.artifacts == nil {
			c.logger.Warn("Artifact service not available, dropping attachment", "file", f.Name)
			continue
		}
		line, err := msgutil.StoreAsArtifact(ctx, c.artifacts, agentID, userID0, sessionID, f.Name, f.MIMEType, f.Data)
		if err != nil {
			c.logger.Warn("Failed to save artifact, dropping attachment", "file", f.Name, "error", err)
			continue
		}
		artifactLines = append(artifactLines, line)
	}

	artifactsBefore := c.listArtifacts(agentID, userID0, sessionID)

	validatedText, truncated := msgutil.ValidateInputLength(text, msgutil.DefaultMaxInputLength)
	if truncated {
		c.logger.Warn("Inbound message truncated", "channel", channelID, "original_len", len([]rune(text)))
	}

	fullMessage := c.buildMessageContext(userID, channelID, channelType, threadTS, teamID) +
		c.fetchThreadContext(channelID, threadTS, messageTS) +
		validatedText +
		msgutil.AttachedArtifactsBlock(artifactLines)

	c.addReaction("brain", msgRef)

	var lastTextResponse string
	hasText := false
	hasToolActivity := false
	var lastFinishReason string
	var lastErrorMessage string
	toolCount := 0
	var toolCounterTS string

	err := c.callAgentSSE(agentID, sessionID, fullMessage, func(evt msgutil.SSEEvent) {
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
			toolCounterTS = ""
			c.sendTextMessage(channelID, evt.Text, threadTS, inputWasVoice)
		case msgutil.SSEEventToolCall:
			hasToolActivity = true
			toolCounterTS = c.sendToolCounter(channelID, threadTS, toolCounterTS, &toolCount, evt)
		case msgutil.SSEEventToolResult:
			hasToolActivity = true
			if c.getShowTools() {
				c.postMessage(channelID, msgutil.FormatToolResultSlack(evt), threadTS)
			}
		}
	})

	if err != nil {
		c.logger.Error("Failed to call agent", "error", err)
		c.addReaction("x", msgRef)
		c.postMessage(channelID, fmt.Sprintf("Failed to reach the agent: %s", sanitizeError(err)), threadTS)
		return
	}

	if !hasText && !hasToolActivity {
		c.postMessage(channelID, msgutil.ExplainNoResponse(lastFinishReason, lastErrorMessage), threadTS)
	}

	mode := c.getResponseMode()
	if (mode == ResponseModeVoice || mode == ResponseModeBoth || (mode == ResponseModeMirror && inputWasVoice)) && lastTextResponse != "" {
		c.sendVoiceResponse(channelID, lastTextResponse, threadTS, agentID)
	}

	c.addReaction("white_check_mark", msgRef)
	c.sendNewArtifacts(channelID, threadTS, agentID, "default_user", sessionID, artifactsBefore)
}

// sendToolCounter posts or edits a compact tool activity counter in the channel/thread.
// Returns the updated message TS for subsequent edits.
func (c *Client) sendToolCounter(channelID, threadTS, counterTS string, toolCount *int, evt msgutil.SSEEvent) string {
	if c.getShowTools() {
		c.postMessage(channelID, msgutil.FormatToolCallSlack(evt), threadTS)
		return counterTS
	}

	*toolCount++
	counterText := fmt.Sprintf("⚙️ x%d", *toolCount)

	if counterTS == "" {
		opts := []slackapi.MsgOption{slackapi.MsgOptionText(counterText, false)}
		if threadTS != "" {
			opts = append(opts, slackapi.MsgOptionTS(threadTS))
		}
		if _, ts, err := c.api.PostMessage(channelID, opts...); err == nil {
			return ts
		}
		return ""
	}

	c.api.UpdateMessage(channelID, counterTS, slackapi.MsgOptionText(counterText, false))
	return counterTS
}

// sendTextMessage sends text respecting the response mode.
func (c *Client) sendTextMessage(channelID, text, threadTS string, inputWasVoice bool) {
	mode := c.getResponseMode()
	switch mode {
	case ResponseModeVoice:
		return
	case ResponseModeMirror:
		if inputWasVoice {
			return
		}
	}
	c.postMessage(channelID, text, threadTS)
}

func (c *Client) callAgentSSE(agentID, sessionID, message string, handler func(msgutil.SSEEvent)) error {
	if err := c.ensureSession(agentID, "default_user", sessionID); err != nil {
		c.logger.Warn("Failed to ensure session, continuing anyway", "error", err)
	}

	parts := []interface{}{
		map[string]string{"text": message},
	}

	reqBody := map[string]interface{}{
		"appName":   agentID,
		"userId":    "default_user",
		"sessionId": sessionID,
		"newMessage": map[string]interface{}{
			"role":  "user",
			"parts": parts,
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

func (c *Client) sendVoiceResponse(channelID, text, threadTS, agentID string) {
	audioData, err := c.generateTTS(text, agentID)
	if err != nil {
		c.logger.Error("Failed to generate TTS", "error", err)
		c.postMessage(channelID, text, threadTS)
		return
	}

	params := slackapi.UploadFileV2Parameters{
		Channel:  channelID,
		Filename: "voice.ogg",
		FileSize: len(audioData),
		Reader:   bytes.NewReader(audioData),
		Title:    "Voice response",
		AltTxt:   text,
	}
	if threadTS != "" {
		params.ThreadTimestamp = threadTS
	}

	if _, err = c.api.UploadFileV2(params); err != nil {
		c.logger.Error("Failed to upload voice file", "channel", channelID, "error", err)
		c.postMessage(channelID, text, threadTS)
	}
}

func (c *Client) getResponseMode() string {
	c.responseMu.RLock()
	defer c.responseMu.RUnlock()
	if c.responseModeOverride != "" {
		return c.responseModeOverride
	}
	if mode := c.clientDef.Config.Slack.ResponseMode; mode != "" {
		return mode
	}
	return ResponseModeText
}

func (c *Client) getShowTools() bool {
	c.showToolsMu.RLock()
	defer c.showToolsMu.RUnlock()
	return c.showTools
}

// buildSessionID builds a stable session ID scoped to a channel/thread + agent.
// When threadTS is set, the session is scoped to that thread; otherwise to the channel.
func (c *Client) buildSessionID(channelID, threadTS, agentID string) string {
	if threadTS != "" {
		return fmt.Sprintf("slack_%s_%s_%s", channelID, threadTS, agentID)
	}
	return fmt.Sprintf("slack_%s_%s", channelID, agentID)
}

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

// fetchThreadContext returns prior messages in a thread as context for the agent.
// Returns empty string when threadTS is empty (DM or first message in channel).
// Excludes the current message (identified by currentMsgTS) to avoid duplication.
func (c *Client) fetchThreadContext(channelID, threadTS, currentMsgTS string) string {
	if threadTS == "" {
		return ""
	}

	limit := c.clientDef.Config.Slack.ThreadHistoryLimit
	if limit <= 0 {
		limit = 50
	}
	msgs, _, _, err := c.api.GetConversationReplies(&slackapi.GetConversationRepliesParameters{
		ChannelID: channelID,
		Timestamp: threadTS,
		Limit:     limit,
	})
	if err != nil {
		c.logger.Debug("Failed to fetch thread context", "error", err)
		return ""
	}
	if len(msgs) <= 1 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("<!--MAGEC_THREAD_HISTORY:\n")
	for _, msg := range msgs {
		if msg.Timestamp == currentMsgTS {
			continue
		}
		text := strings.TrimSpace(msg.Text)
		if text == "" {
			continue
		}
		name := msg.Username
		if name == "" {
			if info, err := c.api.GetUserInfo(msg.User); err == nil && info != nil {
				if info.RealName != "" {
					name = info.RealName
				} else {
					name = info.Name
				}
			} else {
				name = msg.User
			}
		}
		if msg.BotID != "" && name == "" {
			name = "assistant"
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", name, text))
	}
	sb.WriteString(":MAGEC_THREAD_HISTORY-->\n")
	return sb.String()
}

func (c *Client) buildMessageContext(userID, channelID, channelType, threadTS, teamID string) string {
	meta := map[string]interface{}{
		"source":           "slack",
		"slack_user_id":    userID,
		"slack_channel_id": channelID,
	}
	if channelType != "" {
		meta["slack_channel_type"] = channelType
	}
	if teamID != "" {
		meta["slack_team_id"] = teamID
	}
	if threadTS != "" {
		meta["slack_thread_ts"] = threadTS
	}

	if userInfo, err := c.api.GetUserInfo(userID); err == nil && userInfo != nil {
		if userInfo.Name != "" {
			meta["slack_username"] = userInfo.Name
		}
		if userInfo.RealName != "" {
			meta["slack_name"] = userInfo.RealName
		}
		if userInfo.Profile.Email != "" {
			meta["slack_email"] = userInfo.Profile.Email
		}
	}

	jsonBytes, err := json.Marshal(meta)
	if err != nil {
		c.logger.Warn("Failed to marshal message context metadata", "error", err)
		return ""
	}
	return fmt.Sprintf("<!--MAGEC_META:%s:MAGEC_META-->\n", string(jsonBytes))
}

func (c *Client) postMessage(channelID, text, threadTS string) {
	for _, chunk := range msgutil.SplitMessage(text, msgutil.SlackMaxMessageLength) {
		opts := []slackapi.MsgOption{slackapi.MsgOptionText(chunk, false)}
		if threadTS != "" {
			opts = append(opts, slackapi.MsgOptionTS(threadTS))
		}
		if _, _, err := c.api.PostMessage(channelID, opts...); err != nil {
			c.logger.Error("Failed to send Slack message", "channel", channelID, "error", err)
			break
		}
	}
}

func (c *Client) addReaction(emoji string, ref slackapi.ItemRef) {
	if err := c.api.AddReaction(emoji, ref); err != nil {
		c.logger.Debug("Failed to add reaction", "emoji", emoji, "error", err)
	}
}

func (c *Client) generateTTS(text string, agentID string) ([]byte, error) {
	reqBody := map[string]interface{}{
		"input":           text,
		"response_format": "opus",
	}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		strings.TrimSuffix(c.agentURL, "/agent")+"/voice/"+agentID+"/speech",
		bytes.NewReader(jsonBody),
	)
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

func (c *Client) downloadSlackFile(fileURL string) ([]byte, error) {
	token := c.clientDef.Config.Slack.BotToken
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			req.Header.Set("Authorization", "Bearer "+token)
			return nil
		},
	}

	req, err := http.NewRequest(http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download failed with status %d: %s", resp.StatusCode, string(body[:min(len(body), 200)]))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "text/html") {
		return nil, fmt.Errorf("download returned HTML instead of file data (Content-Type: %s, size: %d)", ct, len(data))
	}

	c.logger.Info("Downloaded slack file", "url", fileURL, "size", len(data))
	return data, nil
}

func (c *Client) convertAudioToWav(audioData []byte) ([]byte, error) {
	tmpIn, err := os.CreateTemp("", "magec-audio-in-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp input file: %w", err)
	}
	defer os.Remove(tmpIn.Name())

	if _, err := tmpIn.Write(audioData); err != nil {
		tmpIn.Close()
		return nil, fmt.Errorf("failed to write temp input file: %w", err)
	}
	tmpIn.Close()

	cmd := exec.Command("ffmpeg", "-i", tmpIn.Name(), "-ar", "16000", "-ac", "1", "-f", "wav", "pipe:1")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg conversion failed: %w, stderr: %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

func (c *Client) transcribeAudio(wavData []byte, agentID string) (string, error) {
	var buf bytes.Buffer
	boundary := "----WebKitFormBoundary7MA4YWxkTrZu0gW"
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Disposition: form-data; name=\"file\"; filename=\"audio.wav\"\r\n")
	buf.WriteString("Content-Type: audio/wav\r\n\r\n")
	buf.Write(wavData)
	buf.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST",
		strings.TrimSuffix(c.agentURL, "/agent")+"/voice/"+agentID+"/transcription",
		&buf,
	)
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

	var payload struct {
		Text       string `json:"text,omitempty"`
		InlineData *struct {
			MIMEType string `json:"mimeType"`
			Data     string `json:"data"`
		} `json:"inlineData,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", fmt.Errorf("failed to decode artifact: %w", err)
	}

	if payload.InlineData != nil {
		data, err := base64.StdEncoding.DecodeString(payload.InlineData.Data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to decode artifact binary data: %w", err)
		}
		return data, payload.InlineData.MIMEType, nil
	}
	return []byte(payload.Text), "text/plain", nil
}

func (c *Client) sendNewArtifacts(channelID, threadTS, agentID, userID, sessionID string, before []string) {
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
		params := slackapi.UploadFileV2Parameters{
			Channel:  channelID,
			Filename: name,
			FileSize: len(data),
			Reader:   bytes.NewReader(data),
			Title:    name,
		}
		if threadTS != "" {
			params.ThreadTimestamp = threadTS
		}
		if _, err = c.api.UploadFileV2(params); err != nil {
			c.logger.Error("Failed to upload artifact", "name", name, "error", err)
		}
	}
}

func (c *Client) setAuthHeader(req *http.Request) {
	if c.clientDef.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.clientDef.Token)
	}
}

func (c *Client) isAllowed(userID, channelID string) bool {
	cfg := c.clientDef.Config.Slack
	if len(cfg.AllowedUsers) == 0 && len(cfg.AllowedChannels) == 0 {
		return true
	}
	if len(cfg.AllowedUsers) > 0 && slices.Contains(cfg.AllowedUsers, userID) {
		return true
	}
	if len(cfg.AllowedChannels) > 0 && slices.Contains(cfg.AllowedChannels, channelID) {
		return true
	}
	return false
}

func (c *Client) getActiveAgentID(channelID string) string {
	c.activeAgentMu.RLock()
	defer c.activeAgentMu.RUnlock()
	if id, ok := c.activeAgent[channelID]; ok {
		return id
	}
	if def := c.clientDef.Config.Slack.DefaultAgent; def != "" {
		return def
	}
	return c.clientDef.AllowedAgents[0]
}

func (c *Client) setActiveAgentID(channelID, agentID string) {
	c.activeAgentMu.Lock()
	c.activeAgent[channelID] = agentID
	c.activeAgentMu.Unlock()

	def := c.clientDef
	def.Config.Slack.DefaultAgent = agentID
	c.clientDef = def
	if err := c.store.UpdateClient(def.ID, def); err != nil {
		c.logger.Warn("Failed to persist default agent", "error", err)
	}
}

func (c *Client) getAgentInfo(agentID string) *AgentInfo {
	for i := range c.agents {
		if c.agents[i].ID == agentID {
			return &c.agents[i]
		}
	}
	return nil
}

func (c *Client) stripBotMention(text string) string {
	return strings.TrimSpace(strings.ReplaceAll(text, fmt.Sprintf("<@%s>", c.botUserID), ""))
}

func sanitizeError(err error) string {
	msg := err.Error()
	if len(msg) > 200 {
		msg = msg[:200] + "..."
	}
	for _, secret := range []string{"Bearer ", "xoxb-", "xapp-", "bot", "token"} {
		if strings.Contains(strings.ToLower(msg), strings.ToLower(secret)) {
			return "an internal error occurred"
		}
	}
	return msg
}
