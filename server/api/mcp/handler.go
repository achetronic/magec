package mcp

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/achetronic/openapi2tools/mcptools"
	"github.com/achetronic/openapi2tools/openapi"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/swaggo/swag"

	// Importing the admin docs package registers the spec with swaggo's
	// global registry so swag.ReadDoc("swagger") returns the rendered JSON.
	_ "github.com/achetronic/magec/server/api/admin/docs"
)

// swaggerInstanceName is the swag instance identifier set in the admin
// docs package. Anything that loads the spec uses the same constant.
const swaggerInstanceName = "swagger"

// toolNamePrefix is prepended to every tool generated from the admin spec.
// Keeps the catalogue easy to spot when the client connects to several MCP
// servers at once.
const toolNamePrefix = "magec_"

// HandlerConfig wires the MCP handler to the admin REST API. AdminBaseURL is
// the absolute base for the admin endpoints (host + /api/v1/admin), and
// AdminPassword is forwarded as the bearer token used by AdminAuth so the
// loopback request authenticates the same way an external client would.
type HandlerConfig struct {
	AdminBaseURL  string
	AdminPassword string
}

// Handler is the embedded MCP server's dependency container. It loads the
// admin OpenAPI spec at construction time, registers one tool per filtered
// operation, and exposes a *mcp.StreamableHTTPHandler for the main server to
// mount behind the bearer middleware.
type Handler struct {
	server     *sdk.Server
	streamable *sdk.StreamableHTTPHandler
	toolCount  int
}

// NewHandler builds a fresh MCP handler from the admin spec. The spec is
// embedded in the binary via swaggo; no network I/O happens here.
func NewHandler(cfg HandlerConfig) (*Handler, error) {
	spec, err := loadAdminSpec()
	if err != nil {
		return nil, fmt.Errorf("load admin spec: %w", err)
	}

	filters, err := openapi.CompileRouteFilters(defaultRouteFilters())
	if err != nil {
		return nil, fmt.Errorf("compile route filters: %w", err)
	}
	routes := openapi.FilterRoutes(spec, filters)

	exec := &mcptools.HTTPExecutor{
		BaseURL: cfg.AdminBaseURL,
		Client:  &http.Client{Timeout: 30 * time.Second},
	}
	if cfg.AdminPassword != "" {
		exec.DefaultHeaders = map[string]string{
			"Authorization": "Bearer " + cfg.AdminPassword,
		}
	}

	descriptors := mcptools.Describe(routes, mcptools.DescribeOptions{
		NamePrefix:           toolNamePrefix,
		CustomHandlerFactory: mcptools.HTTPHandlerFactory(exec),
	})

	server := sdk.NewServer(&sdk.Implementation{
		Name:    "magec-admin",
		Version: "1.0.0",
	}, nil)

	count, err := registerDescriptors(server, descriptors)
	if err != nil {
		return nil, fmt.Errorf("register descriptors: %w", err)
	}

	streamable := sdk.NewStreamableHTTPHandler(
		func(*http.Request) *sdk.Server { return server },
		&sdk.StreamableHTTPOptions{SessionTimeout: 10 * time.Minute},
	)

	return &Handler{
		server:     server,
		streamable: streamable,
		toolCount:  count,
	}, nil
}

// HTTPHandler returns the http.Handler that speaks the MCP Streamable HTTP
// transport. Callers wrap it with their own auth and CORS middleware.
func (h *Handler) HTTPHandler() http.Handler { return h.streamable }

// Server returns the underlying *mcp.Server. Used by tests that drive the
// server through the in-memory transport.
func (h *Handler) Server() *sdk.Server { return h.server }

// ToolCount returns the number of tools registered on the server. Surfaced
// in the startup log so operators can sanity-check the catalogue.
func (h *Handler) ToolCount() int { return h.toolCount }

// loadAdminSpec reads the admin swagger 2.0 document registered by swaggo,
// converts it to OpenAPI 3.0, and lets openapi2tools do the heavy lifting
// (ref resolution, example stripping, flexible numeric parameters).
func loadAdminSpec() (*openapi.Spec, error) {
	raw, err := swag.ReadDoc(swaggerInstanceName)
	if err != nil {
		return nil, fmt.Errorf("read swagger doc: %w", err)
	}
	var v2 openapi2.T
	if err := json.Unmarshal([]byte(raw), &v2); err != nil {
		return nil, fmt.Errorf("unmarshal swagger 2.0: %w", err)
	}
	v3, err := openapi2conv.ToV3(&v2)
	if err != nil {
		return nil, fmt.Errorf("convert swagger 2.0 to openapi 3.0: %w", err)
	}
	v3JSON, err := json.Marshal(v3)
	if err != nil {
		return nil, fmt.Errorf("marshal openapi 3.0: %w", err)
	}
	loader := openapi.NewLoader(openapi.LoadOptions{
		ResolveRefs:        true,
		RemoveExamples:     true,
		FlexibleParameters: true,
	})
	return loader.LoadBytes(v3JSON)
}

// defaultRouteFilters trims the admin surface down to the operations that
// make sense as MCP tools. Binary streams (skills upload/download, backup
// and restore) and the admin-UI helper endpoints (auth/check, overview) are
// out — the first three because MCP tool inputs and outputs cannot carry
// large archives cleanly, and the last two because they are not operator
// actions.
func defaultRouteFilters() []openapi.RouteFilterConfig {
	return []openapi.RouteFilterConfig{
		{Paths: `^/skills/upload$`, Exclude: true},
		{Paths: `^/skills/[^/]+/download$`, Exclude: true},
		{Paths: `^/settings/backup$`, Exclude: true},
		{Paths: `^/settings/restore$`, Exclude: true},
		{Paths: `^/auth/check$`, Exclude: true},
		{Paths: `^/overview$`, Exclude: true},
		{Paths: `.*`},
	}
}
