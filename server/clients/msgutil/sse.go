package msgutil

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"sort"
	"strings"
)

// SSEEventType classifies the type of event received from the ADK /run_sse endpoint.
type SSEEventType int

const (
	SSEEventText                SSEEventType = iota // Agent produced text content
	SSEEventToolCall                                // Agent called a tool (functionCall)
	SSEEventToolResult                              // Tool returned a result (functionResponse)
	SSEEventExecutableCode                          // Model generated code for execution
	SSEEventCodeExecutionResult                     // Result of code execution
	SSEEventInlineData                              // Raw binary data (image, audio, etc.)
	SSEEventFileData                                // File reference by URI
	SSEEventError                                   // Event-level error from ADK/LLM
	SSEEventUnknown                                 // Unrecognized event
)

// SSEEvent represents a single parsed event from the ADK /run_sse stream.
type SSEEvent struct {
	Type       SSEEventType
	Author     string
	Text       string
	ToolName   string
	ToolArgs   interface{}
	ToolResult interface{}
	Raw        map[string]interface{}

	Code         string // ExecutableCode: code content
	CodeLanguage string // ExecutableCode: language
	CodeOutcome  string // CodeExecutionResult: outcome
	CodeOutput   string // CodeExecutionResult: output
	MIMEType     string // InlineData/FileData: MIME type
	FileURI      string // FileData: file URI
	DataBytes    string // InlineData: base64-encoded data

	ErrorCode    string // Event-level error code
	ErrorMessage string // Event-level error message
	FinishReason string // Why the model stopped (STOP, MAX_TOKENS, SAFETY, etc.)
	TurnComplete bool   // Whether the model's turn is finished
	Partial      bool   // Whether this is a partial streaming chunk

	UsageMetadata *UsageMetadata // Token usage info
}

// UsageMetadata holds token count information from the LLM.
type UsageMetadata struct {
	PromptTokens    int
	CandidateTokens int
	TotalTokens     int
	CachedTokens    int
}

// ParseSSEStream reads a /run_sse response body and calls handler for each
// meaningful event as it arrives. The handler receives events one at a time.
// This is a blocking call that returns when the stream ends or an error occurs.
func ParseSSEStream(reader io.Reader, handler func(SSEEvent)) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 0, 256*1024), 1024*1024)

	const adkErrorPrefix = "Error while running agent: "

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, adkErrorPrefix) {
			errMsg := strings.TrimPrefix(line, adkErrorPrefix)
			slog.Error("SSE stream: ADK agent error received as plain text",
				"error", errMsg,
			)
			handler(SSEEvent{
				Type:         SSEEventError,
				ErrorMessage: errMsg,
			})
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "" {
			continue
		}

		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(data), &raw); err != nil {
			continue
		}

		events := classifyEvent(raw)
		if len(events) == 0 {
			author, _ := raw["author"].(string)
			slog.Debug("SSE event dropped (no classifiable parts)",
				"author", author,
				"raw_keys", mapKeys(raw),
				"data_len", len(data),
			)
		}
		for _, evt := range events {
			handler(evt)
		}
	}

	return scanner.Err()
}

// classifyEvent examines a raw ADK event JSON and returns one or more typed SSEEvents.
func classifyEvent(raw map[string]interface{}) []SSEEvent {
	author, _ := raw["author"].(string)

	base := SSEEvent{
		Author: author,
		Raw:    raw,
	}

	if v, ok := raw["turn_complete"].(bool); ok {
		base.TurnComplete = v
	}
	if v, ok := raw["partial"].(bool); ok {
		base.Partial = v
	}
	if v, ok := raw["finish_reason"].(string); ok {
		base.FinishReason = v
	}
	if v, ok := raw["error_code"].(string); ok {
		base.ErrorCode = v
	}
	if v, ok := raw["error_message"].(string); ok {
		base.ErrorMessage = v
	}
	base.UsageMetadata = parseUsageMetadata(raw)

	if base.ErrorCode != "" || base.ErrorMessage != "" {
		evt := base
		evt.Type = SSEEventError
		return []SSEEvent{evt}
	}

	content, ok := raw["content"].(map[string]interface{})
	if !ok {
		if base.FinishReason != "" || base.UsageMetadata != nil || base.TurnComplete {
			evt := base
			evt.Type = SSEEventUnknown
			return []SSEEvent{evt}
		}
		return nil
	}

	parts, ok := content["parts"].([]interface{})
	if !ok {
		if base.FinishReason != "" || base.UsageMetadata != nil || base.TurnComplete {
			evt := base
			evt.Type = SSEEventUnknown
			return []SSEEvent{evt}
		}
		return nil
	}

	var events []SSEEvent

	for _, part := range parts {
		partMap, ok := part.(map[string]interface{})
		if !ok {
			continue
		}

		if text, ok := partMap["text"].(string); ok && text != "" {
			evt := base
			evt.Type = SSEEventText
			evt.Text = text
			events = append(events, evt)
		}

		if fc, ok := partMap["functionCall"].(map[string]interface{}); ok {
			name, _ := fc["name"].(string)
			evt := base
			evt.Type = SSEEventToolCall
			evt.ToolName = name
			evt.ToolArgs = fc["args"]
			events = append(events, evt)
		}

		if fr, ok := partMap["functionResponse"].(map[string]interface{}); ok {
			name, _ := fr["name"].(string)
			evt := base
			evt.Type = SSEEventToolResult
			evt.ToolName = name
			evt.ToolResult = fr["response"]
			events = append(events, evt)
		}

		if ec, ok := partMap["executableCode"].(map[string]interface{}); ok {
			evt := base
			evt.Type = SSEEventExecutableCode
			evt.Code, _ = ec["code"].(string)
			evt.CodeLanguage, _ = ec["language"].(string)
			events = append(events, evt)
		}

		if cr, ok := partMap["codeExecutionResult"].(map[string]interface{}); ok {
			evt := base
			evt.Type = SSEEventCodeExecutionResult
			evt.CodeOutcome, _ = cr["outcome"].(string)
			evt.CodeOutput, _ = cr["output"].(string)
			events = append(events, evt)
		}

		if id, ok := partMap["inlineData"].(map[string]interface{}); ok {
			evt := base
			evt.Type = SSEEventInlineData
			evt.MIMEType, _ = id["mimeType"].(string)
			evt.DataBytes, _ = id["data"].(string)
			events = append(events, evt)
		}

		if fd, ok := partMap["fileData"].(map[string]interface{}); ok {
			evt := base
			evt.Type = SSEEventFileData
			evt.MIMEType, _ = fd["mimeType"].(string)
			evt.FileURI, _ = fd["fileUri"].(string)
			events = append(events, evt)
		}
	}

	return events
}

func parseUsageMetadata(raw map[string]interface{}) *UsageMetadata {
	um, ok := raw["usage_metadata"].(map[string]interface{})
	if !ok {
		return nil
	}
	toInt := func(key string) int {
		if v, ok := um[key].(float64); ok {
			return int(v)
		}
		return 0
	}
	return &UsageMetadata{
		PromptTokens:    toInt("prompt_token_count"),
		CandidateTokens: toInt("candidates_token_count"),
		TotalTokens:     toInt("total_token_count"),
		CachedTokens:    toInt("cached_content_token_count"),
	}
}

// CollectSSEEvents is a convenience wrapper that reads the full SSE stream
// and returns all events and the aggregated text response.
func CollectSSEEvents(reader io.Reader) ([]SSEEvent, string, error) {
	var events []SSEEvent
	var textParts []string

	err := ParseSSEStream(reader, func(evt SSEEvent) {
		events = append(events, evt)
		if evt.Type == SSEEventText {
			textParts = append(textParts, evt.Text)
		}
	})

	fullText := strings.Join(textParts, "")
	if fullText == "" {
		fullText = "(no response)"
	}

	return events, fullText, err
}

// FormatToolCallTelegram formats a single tool call as a Telegram expandable
// blockquote. The tool name is always visible; args are human-readable inside.
func FormatToolCallTelegram(evt SSEEvent) string {
	lines := humanArgLines(evt.ToolArgs)
	if len(lines) == 0 {
		return fmt.Sprintf("<blockquote>🔧 <b>%s</b></blockquote>", escapeHTML(evt.ToolName))
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<blockquote expandable>🔧 <b>%s</b>\n", escapeHTML(evt.ToolName)))
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("<b>%s</b>: %s\n", escapeHTML(l.Key), escapeHTML(l.Value)))
	}
	b.WriteString("</blockquote>")
	return b.String()
}

// FormatToolCallDiscord formats a single tool call for Discord.
func FormatToolCallDiscord(evt SSEEvent) string {
	lines := humanArgLines(evt.ToolArgs)
	if len(lines) == 0 {
		return fmt.Sprintf("> 🔧 **%s**", evt.ToolName)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("> 🔧 **%s**\n", evt.ToolName))
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("> **%s**: %s\n", l.Key, l.Value))
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatToolCallSlack formats a single tool call for Slack.
func FormatToolCallSlack(evt SSEEvent) string {
	lines := humanArgLines(evt.ToolArgs)
	if len(lines) == 0 {
		return fmt.Sprintf("> 🔧 *%s*", evt.ToolName)
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("> 🔧 *%s*\n", evt.ToolName))
	for _, l := range lines {
		b.WriteString(fmt.Sprintf("> *%s*: %s\n", l.Key, l.Value))
	}
	return strings.TrimRight(b.String(), "\n")
}

// FormatToolResultTelegram formats a tool result as a Telegram expandable blockquote.
func FormatToolResultTelegram(evt SSEEvent) string {
	result := prettyResult(evt.ToolResult)
	if result == "" {
		return fmt.Sprintf("<blockquote>📎 <b>%s</b> → (empty)</blockquote>", escapeHTML(evt.ToolName))
	}
	return fmt.Sprintf("<blockquote expandable>📎 <b>%s</b>\n%s</blockquote>", escapeHTML(evt.ToolName), escapeHTML(result))
}

// FormatToolResultDiscord formats a tool result for Discord.
func FormatToolResultDiscord(evt SSEEvent) string {
	result := prettyResult(evt.ToolResult)
	if result == "" {
		return fmt.Sprintf("> 📎 **%s** → (empty)", evt.ToolName)
	}
	return fmt.Sprintf("📎 **%s**\n```\n%s\n```", evt.ToolName, result)
}

// FormatToolResultSlack formats a tool result for Slack.
func FormatToolResultSlack(evt SSEEvent) string {
	result := prettyResult(evt.ToolResult)
	if result == "" {
		return fmt.Sprintf("> 📎 *%s* → (empty)", evt.ToolName)
	}
	return fmt.Sprintf("📎 *%s*\n```\n%s\n```", evt.ToolName, result)
}

// argLine holds a key-value pair from tool arguments.
type argLine struct {
	Key   string
	Value string
}

// prettyResult formats a tool result for display inside a code block.
func prettyResult(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}

// humanArgLines converts tool arguments into human-readable key-value pairs.
func humanArgLines(args interface{}) []argLine {
	if args == nil {
		return nil
	}
	m, ok := args.(map[string]interface{})
	if !ok {
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var lines []argLine
	for _, k := range keys {
		v := m[k]
		s := humanValue(v)
		if s == "" {
			continue
		}
		lines = append(lines, argLine{Key: k, Value: s})
	}
	return lines
}

// humanValue formats a single value for human reading.
func humanValue(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%g", val)
	case bool:
		if val {
			return "true"
		}
		return "false"
	case nil:
		return ""
	case []interface{}:
		if len(val) == 0 {
			return "[]"
		}
		var items []string
		for _, item := range val {
			items = append(items, humanValue(item))
			if len(items) >= 5 {
				items = append(items, "…")
				break
			}
		}
		return "[" + strings.Join(items, ", ") + "]"
	case map[string]interface{}:
		b, _ := json.Marshal(val)
		return string(b)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func escapeHTML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func mapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ExplainNoResponse returns a user-facing message explaining why the agent
// did not produce text, based on the finish_reason and error_message from the
// SSE stream.
func ExplainNoResponse(finishReason, errorMessage string) string {
	if errorMessage != "" {
		return fmt.Sprintf("⚠️ The agent returned an error: %s", errorMessage)
	}
	switch strings.ToUpper(finishReason) {
	case "MAX_TOKENS":
		return "⚠️ The response was cut short because the context window is full. Try /reset to start a fresh session."
	case "SAFETY":
		return "⚠️ The response was blocked by the model's safety filters."
	case "RECITATION":
		return "⚠️ The response was blocked due to recitation/copyright concerns."
	case "STOP":
		return "The agent completed without generating a text response."
	default:
		if finishReason != "" {
			return fmt.Sprintf("⚠️ The agent stopped unexpectedly (reason: %s).", finishReason)
		}
		return "I couldn't generate a response."
	}
}
