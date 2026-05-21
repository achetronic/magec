package mcp

import (
	"net/http"
	"time"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/achetronic/magec/server/api/admin"
	"github.com/achetronic/magec/server/store"
)

// Handler is the embedded MCP server's dependency container. It mirrors
// admin.Handler so MCP tools have access to the same store, conversation
// store, and ADK session service.
type Handler struct {
	store         *store.Store
	conversations *store.ConversationStore
	adminHandler  *admin.Handler

	server     *sdk.Server
	streamable *sdk.StreamableHTTPHandler
	toolCount  int
}

// NewHandler builds a fresh MCP handler. The admin handler is borrowed only
// for its ADK session service accessor (used by the reset-session tool); it
// can be nil in tests that do not exercise that path.
func NewHandler(s *store.Store, cs *store.ConversationStore, ah *admin.Handler) *Handler {
	h := &Handler{store: s, conversations: cs, adminHandler: ah}
	h.server = sdk.NewServer(&sdk.Implementation{
		Name:    "magec-admin",
		Version: "1.0.0",
	}, nil)
	h.registerAll()
	h.streamable = sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return h.server },
		&sdk.StreamableHTTPOptions{
			SessionTimeout: 10 * time.Minute,
		},
	)
	return h
}

// HTTPHandler returns the http.Handler that speaks the MCP Streamable HTTP
// transport. Callers wrap it with their own auth/CORS middleware.
func (h *Handler) HTTPHandler() http.Handler { return h.streamable }

// Server returns the underlying *mcp.Server. Used by tests that drive the
// server through the in-memory transport.
func (h *Handler) Server() *sdk.Server { return h.server }

// ToolCount returns the number of tools registered on the server. Used by
// the startup log line and the smoke test.
func (h *Handler) ToolCount() int { return h.toolCount }

// sessionService returns the borrowed ADK session service, or nil when the
// admin handler is not wired in.
func (h *Handler) sessionService() interface{} {
	if h.adminHandler == nil {
		return nil
	}
	return h.adminHandler.SessionService()
}
