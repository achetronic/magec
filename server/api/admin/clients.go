package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/achetronic/magec/server/clients"
	"github.com/achetronic/magec/server/store"
)

const telegramConfigTestMessage = "This is a test message from Magec Telegram client configuration."

type telegramMessageSender interface {
	SendMessage(ctx context.Context, params *telego.SendMessageParams) (*telego.Message, error)
}

type telegramTestTarget struct {
	ChatID   int64
	ThreadID int
}

type TelegramConfigTestResult struct {
	Attempted int      `json:"attempted"`
	Sent      int      `json:"sent"`
	Failed    int      `json:"failed"`
	Errors    []string `json:"errors,omitempty"`
}

type telegramTestRequest struct {
	Config *store.TelegramClientConfig `json:"config,omitempty"`
}

func targetKey(chatID int64, threadID int) string {
	return fmt.Sprintf("%d:%d", chatID, threadID)
}

func buildTelegramTestTargets(cfg *store.TelegramClientConfig) []telegramTestTarget {
	if cfg == nil {
		return nil
	}

	targets := make([]telegramTestTarget, 0, len(cfg.AllowedUsers)+len(cfg.AllowedChatThreads))
	seen := map[string]struct{}{}

	for _, userID := range cfg.AllowedUsers {
		key := targetKey(userID, 0)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, telegramTestTarget{ChatID: userID})
	}

	for _, rule := range cfg.AllowedChatThreads {
		threadID := 0
		if rule.ThreadID != nil {
			threadID = *rule.ThreadID
		}
		key := targetKey(rule.ChatID, threadID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		targets = append(targets, telegramTestTarget{ChatID: rule.ChatID, ThreadID: threadID})
	}

	return targets
}

func sendTelegramConfigTest(ctx context.Context, sender telegramMessageSender, cfg *store.TelegramClientConfig) TelegramConfigTestResult {
	targets := buildTelegramTestTargets(cfg)
	result := TelegramConfigTestResult{Attempted: len(targets)}

	for _, target := range targets {
		params := &telego.SendMessageParams{
			ChatID: tu.ID(target.ChatID),
			Text:   telegramConfigTestMessage,
		}
		if target.ThreadID > 0 {
			params.MessageThreadID = target.ThreadID
		}

		if _, err := sender.SendMessage(ctx, params); err != nil {
			result.Failed++
			result.Errors = append(result.Errors, fmt.Sprintf("target %s: %v", targetKey(target.ChatID, target.ThreadID), err))
			continue
		}
		result.Sent++
	}

	return result
}

func resolveTelegramTestConfig(stored, override *store.TelegramClientConfig) *store.TelegramClientConfig {
	if override == nil {
		return stored
	}
	data, err := json.Marshal(override)
	if err != nil {
		return override
	}

	expanded := os.ExpandEnv(string(data))
	var resolved store.TelegramClientConfig
	if err := json.Unmarshal([]byte(expanded), &resolved); err != nil {
		return override
	}

	return &resolved
}

// listClients returns all clients.
// @Summary      List clients
// @Description  Returns all configured clients (devices, Telegram bots, etc.)
// @Tags         clients
// @Produce      json
// @Success      200  {array}  store.ClientDefinition
// @Security     AdminAuth
// @Router       /clients [get]
func (h *Handler) listClients(w http.ResponseWriter, r *http.Request) {
	clients := h.store.ListRawClients()
	writeJSON(w, http.StatusOK, clients)
}

// getClient returns a single client by ID.
// @Summary      Get client
// @Description  Returns a client by its unique ID
// @Tags         clients
// @Produce      json
// @Param        id    path      string  true  "Client ID"
// @Success      200   {object}  store.ClientDefinition
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /clients/{id} [get]
func (h *Handler) getClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	c, ok := h.store.GetRawClient(id)
	if !ok {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	writeJSON(w, http.StatusOK, c)
}

// createClient creates a new client with an auto-generated token.
// @Summary      Create client
// @Description  Creates a new client. A unique auth token is generated automatically.
// @Tags         clients
// @Accept       json
// @Produce      json
// @Param        body  body      store.ClientDefinition  true  "Client definition (token is auto-generated)"
// @Success      201   {object}  store.ClientDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      409   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /clients [post]
func (h *Handler) createClient(w http.ResponseWriter, r *http.Request) {
	var c store.ClientDefinition
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	if c.Type == "" {
		writeError(w, http.StatusBadRequest, "type is required")
		return
	}
	if !clients.ValidType(c.Type) {
		writeError(w, http.StatusBadRequest, "unsupported client type: "+c.Type)
		return
	}
	if err := validateClientConfig(c); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	created, err := h.store.CreateClient(c)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// updateClient updates an existing client.
// @Summary      Update client
// @Description  Updates a client by ID. Token and ID are preserved.
// @Tags         clients
// @Accept       json
// @Produce      json
// @Param        id    path      string                  true  "Client ID"
// @Param        body  body      store.ClientDefinition  true  "Client definition"
// @Success      200   {object}  store.ClientDefinition
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /clients/{id} [put]
func (h *Handler) updateClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	var c store.ClientDefinition
	if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if c.Type != "" {
		if err := validateClientConfig(c); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if err := h.store.UpdateClient(id, c); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	updated, _ := h.store.GetRawClient(id)
	writeJSON(w, http.StatusOK, updated)
}

// deleteClient deletes a client.
// @Summary      Delete client
// @Description  Deletes a client by ID, revoking its access token
// @Tags         clients
// @Param        id  path  string  true  "Client ID"
// @Success      204
// @Failure      404  {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /clients/{id} [delete]
func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	if err := h.store.DeleteClient(id); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// regenerateClientToken generates a new auth token for a client.
// @Summary      Regenerate client token
// @Description  Generates a new authentication token for a client, invalidating the previous one
// @Tags         clients
// @Produce      json
// @Param        id    path      string  true  "Client ID"
// @Success      200   {object}  store.ClientDefinition
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /clients/{id}/regenerate-token [post]
func (h *Handler) regenerateClientToken(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	cl, err := h.store.RegenerateClientToken(id)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, cl)
}

// testTelegramClient sends a fixed test message to all configured Telegram test targets
// (allowed users and allowed chat/thread rules) of the given client.
// @Summary      Test Telegram client delivery
// @Description  Sends a fixed English test message to all configured Telegram user/chat targets
// @Tags         clients
// @Produce      json
// @Param        id    path      string  true  "Client ID"
// @Success      200   {object}  TelegramConfigTestResult
// @Failure      400   {object}  ErrorResponse
// @Failure      404   {object}  ErrorResponse
// @Security     AdminAuth
// @Router       /clients/{id}/telegram-test [post]
func (h *Handler) testTelegramClient(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	client, ok := h.store.GetClient(id)
	if !ok {
		writeError(w, http.StatusNotFound, "client not found")
		return
	}
	if client.Type != "telegram" {
		writeError(w, http.StatusBadRequest, "client is not telegram")
		return
	}
	if client.Config.Telegram == nil {
		writeError(w, http.StatusBadRequest, "telegram config is missing")
		return
	}

	testCfg := client.Config.Telegram
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read request body: "+err.Error())
			return
		}
		if strings.TrimSpace(string(body)) != "" {
			var req telegramTestRequest
			if err := json.Unmarshal(body, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
				return
			}
			testCfg = resolveTelegramTestConfig(client.Config.Telegram, req.Config)
		}
	}

	if strings.TrimSpace(testCfg.BotToken) == "" {
		writeError(w, http.StatusBadRequest, "telegram bot token is required")
		return
	}

	targets := buildTelegramTestTargets(testCfg)
	if len(targets) == 0 {
		writeError(w, http.StatusBadRequest, "no Telegram targets configured (allowedUsers or allowedChatThreads)")
		return
	}

	bot, err := telego.NewBot(testCfg.BotToken)
	if err != nil {
		writeError(w, http.StatusBadRequest, "failed to initialize telegram bot: "+err.Error())
		return
	}

	result := sendTelegramConfigTest(r.Context(), bot, testCfg)
	writeJSON(w, http.StatusOK, result)
}

// ClientTypeInfo represents a registered client type with its JSON Schema.
type ClientTypeInfo struct {
	Type         string         `json:"type" example:"telegram"`
	DisplayName  string         `json:"displayName" example:"Telegram"`
	ConfigSchema clients.Schema `json:"configSchema"`
}

// listClientTypes returns all registered client types with field specs.
// @Summary      List client types
// @Description  Returns registered client types with config field specifications for dynamic form rendering
// @Tags         clients
// @Produce      json
// @Success      200  {array}  ClientTypeInfo
// @Security     AdminAuth
// @Router       /clients/types [get]
func (h *Handler) listClientTypes(w http.ResponseWriter, r *http.Request) {
	var types []ClientTypeInfo
	for _, p := range clients.All() {
		types = append(types, ClientTypeInfo{
			Type:         p.Type(),
			DisplayName:  p.DisplayName(),
			ConfigSchema: p.ConfigSchema(),
		})
	}
	writeJSON(w, http.StatusOK, types)
}

func validateClientConfig(c store.ClientDefinition) error {
	raw, err := json.Marshal(c.Config)
	if err != nil {
		return nil
	}
	var full map[string]map[string]interface{}
	if err := json.Unmarshal(raw, &full); err != nil {
		return nil
	}
	if err := clients.ValidateConfig(c.Type, full[c.Type]); err != nil {
		return err
	}

	if c.Type == "telegram" && c.Config.Telegram != nil {
		for i, rule := range c.Config.Telegram.AllowedChatThreads {
			if rule.ChatID == 0 {
				return fmt.Errorf("allowedChatThreads[%d].chatId must be a non-zero integer", i)
			}
			if rule.ThreadID != nil && *rule.ThreadID <= 0 {
				return fmt.Errorf("allowedChatThreads[%d].threadId must be a positive integer", i)
			}
		}
	}

	return nil
}
