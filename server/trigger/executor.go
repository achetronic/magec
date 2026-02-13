package trigger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/achetronic/magec/server/store"
	"github.com/google/uuid"
)

// Executor runs commands against agents through the internal ADK API.
type Executor struct {
	store    *store.Store
	agentURL string
	logger   *slog.Logger
}

// NewExecutor creates a trigger executor. agentURL is the base URL for the
// agent API (e.g. "http://127.0.0.1:8080/api/v1/agent").
func NewExecutor(s *store.Store, agentURL string, logger *slog.Logger) *Executor {
	return &Executor{
		store:    s,
		agentURL: agentURL,
		logger:   logger,
	}
}

// RunClient resolves the client's command and agents, then calls the agent API
// for each allowed agent. For passthrough webhooks, prompt is provided directly.
func (e *Executor) RunClient(ctx context.Context, cl store.ClientDefinition, passthroughPrompt string) (string, error) {
	var prompt string
	var commandID string

	switch cl.Type {
	case "cron":
		if cl.Config.Cron == nil {
			return "", fmt.Errorf("client %q: missing cron config", cl.Name)
		}
		commandID = cl.Config.Cron.CommandID
	case "webhook":
		if cl.Config.Webhook == nil {
			return "", fmt.Errorf("client %q: missing webhook config", cl.Name)
		}
		if cl.Config.Webhook.Passthrough {
			prompt = passthroughPrompt
			if prompt == "" {
				return "", fmt.Errorf("passthrough webhook requires a prompt in the request body")
			}
		} else {
			commandID = cl.Config.Webhook.CommandID
		}
	default:
		return "", fmt.Errorf("client %q: unsupported type %q for execution", cl.Name, cl.Type)
	}

	if commandID != "" {
		cmd, ok := e.store.GetCommand(commandID)
		if !ok {
			return "", fmt.Errorf("command %q not found", commandID)
		}
		prompt = cmd.Prompt
	}

	if len(cl.AllowedAgents) == 0 {
		return "", fmt.Errorf("client %q: no allowed agents configured", cl.Name)
	}

	var allResults string
	for _, agentID := range cl.AllowedAgents {
		result, err := e.callAgent(ctx, agentID, prompt, cl.Token)
		if err != nil {
			e.logger.Error("Failed to run agent", "client", cl.Name, "agent", agentID, "error", err)
			continue
		}
		if allResults != "" {
			allResults += "\n---\n"
		}
		allResults += result
	}

	if allResults == "" {
		return "", fmt.Errorf("all agents failed for client %q", cl.Name)
	}
	return allResults, nil
}

// callAgent sends a prompt to the agent API and returns the response text.
func (e *Executor) callAgent(ctx context.Context, agentID, prompt, token string) (string, error) {
	userID := "trigger"
	sessionID := uuid.New().String()

	if err := e.ensureSession(ctx, agentID, userID, sessionID, token); err != nil {
		e.logger.Warn("Failed to ensure session, continuing anyway", "error", err)
	}

	reqBody := map[string]interface{}{
		"appName":   agentID,
		"userId":    userID,
		"sessionId": sessionID,
		"newMessage": map[string]interface{}{
			"role": "user",
			"parts": []map[string]string{
				{"text": prompt},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	callCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(callCtx, "POST", e.agentURL+"/run", bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call agent: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("agent returned status %d: %s", resp.StatusCode, string(body))
	}

	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return extractResponseText(events), nil
}

func (e *Executor) ensureSession(ctx context.Context, agentID, userID, sessionID, token string) error {
	url := fmt.Sprintf("%s/apps/%s/users/%s/sessions/%s", e.agentURL, agentID, userID, sessionID)

	sessCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(sessCtx, "POST", url, bytes.NewReader([]byte("{}")))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

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

func extractResponseText(events []map[string]interface{}) string {
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
		return "(no response)"
	}
	return result
}
