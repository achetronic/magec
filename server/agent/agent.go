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

package agent

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/memory"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/server/adkrest"
	"google.golang.org/adk/session"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"google.golang.org/genai"

	genaianthro "github.com/achetronic/adk-utils-go/genai/anthropic"
	genaiopenai "github.com/achetronic/adk-utils-go/genai/openai"
	memorypostgres "github.com/achetronic/adk-utils-go/memory/postgres"
	sessionredis "github.com/achetronic/adk-utils-go/session/redis"
	toolsmemory "github.com/achetronic/adk-utils-go/tools/memory"

	"github.com/achetronic/magec/server/config"
)

const appName = "magec_agent"

const baseInstruction = `You are Magec, a helpful voice assistant that helps users with various tasks.
Keep responses concise and natural for voice interaction.
Respond in the same language as the user's input.`

const memoryInstruction = `
You have access to long-term memory tools:
- Use 'search_memory' to recall information from past conversations. IMPORTANT: When this tool returns memories, you MUST use that information in your response. The 'memories' array contains the actual data - read the 'text' field of each entry.
- Use 'save_to_memory' to remember important facts, user preferences, or anything the user asks you to remember

CRITICAL: At the START of every conversation, you MUST call search_memory with a broad query to retrieve any stored user preferences, instructions, or important information. This ensures you always have context about the user before responding.

When a user asks you to remember something or asks about past information:
1. First use search_memory to check if you have relevant information
2. If search_memory returns results (count > 0), USE the text from those memories in your answer
3. Only say you don't have information if search_memory returns count: 0

When a user shares preferences or important information, proactively save it to memory for future reference.`

// Service manages the ADK agent with memory.
type Service struct {
	handler http.Handler
}

// New creates a new agent service from config.
func New(ctx context.Context, cfg *config.Config) (*Service, error) {
	// Session service (Redis or in-memory fallback)
	sessionSvc, err := createSessionService(cfg)
	if err != nil {
		return nil, err
	}

	// Long-term memory service (optional)
	memorySvc, err := createMemoryService(ctx, cfg)
	if err != nil {
		return nil, err
	}

	// LLM model
	if cfg.LLM.Resolved == nil {
		return nil, fmt.Errorf("LLM backend not resolved")
	}
	llmModel, err := createLLM(ctx, cfg.LLM.Resolved, cfg.LLM.Model)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM: %w", err)
	}

	// Toolsets
	toolsets, err := buildToolsets(cfg, memorySvc)
	if err != nil {
		return nil, fmt.Errorf("failed to build toolsets: %w", err)
	}

	// Build instruction based on available features
	instruction := baseInstruction
	if memorySvc != nil {
		instruction += memoryInstruction
	}

	// Root agent
	rootAgent, err := llmagent.New(llmagent.Config{
		Name:        appName,
		Model:       llmModel,
		Description: "Voice assistant agent.",
		Instruction: instruction,
		Toolsets:    toolsets,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// REST handler
	launcherCfg := &launcher.Config{
		SessionService: sessionSvc,
		AgentLoader:    agent.NewSingleLoader(rootAgent),
	}
	if memorySvc != nil {
		launcherCfg.MemoryService = memorySvc
	}

	return &Service{
		handler: adkrest.NewHandler(launcherCfg, 30*time.Second),
	}, nil
}

// Handler returns the ADK REST handler.
func (s *Service) Handler() http.Handler {
	return s.handler
}

func createSessionService(cfg *config.Config) (session.Service, error) {
	if cfg.Memory.Session.Redis.Address == "" {
		return session.InMemoryService(), nil
	}

	svc, err := sessionredis.NewRedisSessionService(sessionredis.RedisSessionServiceConfig{
		Addr: cfg.Memory.Session.Redis.Address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis session service: %w", err)
	}
	return svc, nil
}

func createMemoryService(ctx context.Context, cfg *config.Config) (memory.Service, error) {
	ltm := cfg.Memory.LongTerm
	if ltm.Postgres.ConnectionString == "" || ltm.Embedding.Resolved == nil {
		return nil, nil // Memory disabled
	}

	svc, err := memorypostgres.NewPostgresMemoryService(ctx, memorypostgres.PostgresMemoryServiceConfig{
		ConnString: ltm.Postgres.ConnectionString,
		EmbeddingModel: memorypostgres.NewOpenAICompatibleEmbedding(memorypostgres.OpenAICompatibleEmbeddingConfig{
			BaseURL: ltm.Embedding.Resolved.URL,
			Model:   ltm.Embedding.Model,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Postgres memory service: %w", err)
	}
	return svc, nil
}

func createLLM(ctx context.Context, backend *config.Backend, modelName string) (model.LLM, error) {
	switch backend.Type {
	case config.BackendTypeOpenAI:
		return genaiopenai.New(genaiopenai.Config{
			APIKey:    backend.APIKey,
			BaseURL:   backend.URL,
			ModelName: modelName,
		}), nil

	case config.BackendTypeAnthropic:
		return genaianthro.New(genaianthro.Config{
			APIKey:    backend.APIKey,
			ModelName: modelName,
		}), nil

	case config.BackendTypeGemini:
		return gemini.NewModel(ctx, modelName, &genai.ClientConfig{
			APIKey: backend.APIKey,
		})

	default:
		return nil, fmt.Errorf("unsupported LLM backend type: %s", backend.Type)
	}
}

func buildToolsets(cfg *config.Config, memorySvc memory.Service) ([]tool.Toolset, error) {
	var toolsets []tool.Toolset

	// Memory toolset (if enabled)
	if memorySvc != nil {
		ts, err := toolsmemory.NewToolset(toolsmemory.ToolsetConfig{
			MemoryService: memorySvc,
			AppName:       appName,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create memory toolset: %w", err)
		}
		toolsets = append(toolsets, ts)
	}

	// MCP toolsets
	for _, srv := range cfg.MCPServers {
		ts, err := mcptoolset.New(mcptoolset.Config{
			Transport: &mcp.StreamableClientTransport{
				Endpoint:   srv.Endpoint,
				HTTPClient: httpClientWithHeaders(srv.Headers),
				MaxRetries: 5,
			},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create MCP toolset %q: %w", srv.Name, err)
		}
		toolsets = append(toolsets, ts)
	}

	return toolsets, nil
}

func httpClientWithHeaders(headers map[string]string) *http.Client {
	if len(headers) == 0 {
		return http.DefaultClient
	}
	return &http.Client{
		Transport: &headerTransport{
			base:    http.DefaultTransport,
			headers: headers,
		},
	}
}

type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}
