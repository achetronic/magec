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

package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Backend types
const (
	BackendTypeOpenAI    = "openai"
	BackendTypeAnthropic = "anthropic"
	BackendTypeGemini    = "gemini"

	// Default URLs
	DefaultOpenAIURL = "https://api.openai.com/v1"
)

// AgentConfig holds agent-specific configuration
type AgentConfig struct {
	// SystemPrompt overrides the default agent system prompt
	SystemPrompt string `yaml:"systemPrompt,omitempty"`
	// SystemPromptSuffix is appended to the default (or custom) system prompt
	SystemPromptSuffix string `yaml:"systemPromptSuffix,omitempty"`
}

// Config represents the full YAML configuration file structure
type Config struct {
	Server        Server         `yaml:"server"`
	Log           Log            `yaml:"log"`
	Agent         AgentConfig    `yaml:"agent"`
	Backends      []Backend      `yaml:"backends"`
	Transcription BackendRef     `yaml:"transcription"`
	LLM           BackendRef     `yaml:"llm"`
	TTS           TTSConfig      `yaml:"tts"`
	WakeWord      WakeWordConfig `yaml:"wakeWord"`
	Memory        Memory         `yaml:"memory"`
	MCPServers    []MCPServer    `yaml:"mcpServers"`
}

type Server struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// Log configures the application logger
type Log struct {
	Level  string `yaml:"level"`  // debug, info, warn, error (default: info)
	Format string `yaml:"format"` // console, json (default: console)
}

// Backend represents a reusable AI backend
type Backend struct {
	Name   string `yaml:"name"`
	Type   string `yaml:"type"` // openai, anthropic, gemini
	URL    string `yaml:"url"`
	APIKey string `yaml:"apiKey"`
}

// BackendRef holds a reference to a backend + model, with resolved backend pointer
type BackendRef struct {
	Backend  string   `yaml:"backend"`
	Model    string   `yaml:"model"`
	Resolved *Backend `yaml:"-"` // populated by resolve()
}

// TTSConfig holds TTS-specific configuration
type TTSConfig struct {
	Backend  string   `yaml:"backend"`
	Model    string   `yaml:"model"`
	Voice    string   `yaml:"voice"`
	Speed    float64  `yaml:"speed"`
	Resolved *Backend `yaml:"-"` // populated by resolve()
}

// WakeWordConfig holds wake word detection configuration
type WakeWordConfig struct {
	Enabled bool `yaml:"enabled"`
}

// WakeWordModelsConfig is loaded from models/wakewords.yaml
type WakeWordModelsConfig struct {
	Models []WakeWordModel `yaml:"models"`
}

// WakeWordModel represents a wake word model configuration
type WakeWordModel struct {
	ID        string  `yaml:"id"`
	Name      string  `yaml:"name"`
	File      string  `yaml:"file"`
	Phrase    string  `yaml:"phrase"`
	Threshold float32 `yaml:"threshold"`
}

type Memory struct {
	Session  SessionMemory  `yaml:"session"`
	LongTerm LongTermMemory `yaml:"longTerm"`
}

type SessionMemory struct {
	Redis Redis `yaml:"redis"`
}

type Redis struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	TTL      string `yaml:"ttl"`
}

type LongTermMemory struct {
	Postgres  Postgres   `yaml:"postgres"`
	Embedding BackendRef `yaml:"embedding"`
}

type Postgres struct {
	ConnectionString string `yaml:"connectionString"`
}

// MCPTransportType defines the type of MCP transport
type MCPTransportType string

const (
	MCPTransportTypeHTTP  MCPTransportType = "http"
	MCPTransportTypeStdio MCPTransportType = "stdio"
)

// MCPServer represents an MCP server configuration
type MCPServer struct {
	Name     string            `yaml:"name"`
	Type     MCPTransportType  `yaml:"type,omitempty"` // http (default) or stdio
	Endpoint string            `yaml:"endpoint,omitempty"` // For HTTP transport
	Headers  map[string]string `yaml:"headers,omitempty"`  // For HTTP transport
	Command  string            `yaml:"command,omitempty"`  // For stdio transport
	Args     []string          `yaml:"args,omitempty"`     // For stdio transport
	Env      map[string]string `yaml:"env,omitempty"`      // For stdio transport
	WorkDir  string            `yaml:"workDir,omitempty"` // For stdio transport
	// SystemPrompt is an optional prompt to add context about how to use this MCP
	SystemPrompt string `yaml:"systemPrompt,omitempty"`
}

// Load reads, parses, and resolves a config file with environment variable expansion
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := os.ExpandEnv(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults
	cfg.applyDefaults()

	// Resolve backend references
	if err := cfg.resolve(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// LoadWakeWordModels loads wake word model configurations from wakewords.yaml
func LoadWakeWordModels(modelsPath string) (*WakeWordModelsConfig, error) {
	configPath := fmt.Sprintf("%s/wakewords.yaml", modelsPath)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read wakewords.yaml: %w", err)
	}

	var cfg WakeWordModelsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse wakewords.yaml: %w", err)
	}

	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "console"
	}
	if c.Memory.Session.Redis.TTL == "" {
		c.Memory.Session.Redis.TTL = "24h"
	}
	if c.Memory.LongTerm.Embedding.Model == "" {
		c.Memory.LongTerm.Embedding.Model = "nomic-embed-text"
	}
	if c.TTS.Model == "" {
		c.TTS.Model = "tts-1"
	}
	if c.TTS.Voice == "" {
		c.TTS.Voice = "alloy"
	}
	if c.TTS.Speed == 0 {
		c.TTS.Speed = 1.0
	}

	// Apply default URLs to backends
	for i := range c.Backends {
		b := &c.Backends[i]
		if b.URL == "" && b.Type == BackendTypeOpenAI {
			b.URL = DefaultOpenAIURL
		}
	}
}

func (c *Config) resolve() error {
	// Build backend lookup map
	backends := make(map[string]*Backend)
	for i := range c.Backends {
		backends[c.Backends[i].Name] = &c.Backends[i]
	}

	// Helper to resolve a backend reference
	resolveRef := func(ref *BackendRef, name string, requiredType string) error {
		if ref.Backend == "" {
			return nil
		}
		backend, ok := backends[ref.Backend]
		if !ok {
			return fmt.Errorf("%s backend %q not found", name, ref.Backend)
		}
		if requiredType != "" && backend.Type != requiredType {
			return fmt.Errorf("%s backend %q must be type %q, got %q", name, ref.Backend, requiredType, backend.Type)
		}
		ref.Resolved = backend
		return nil
	}

	// Resolve transcription (must be OpenAI-compatible for Whisper API)
	if err := resolveRef(&c.Transcription, "transcription", BackendTypeOpenAI); err != nil {
		return err
	}

	// Resolve LLM (any LLM type allowed)
	if err := resolveRef(&c.LLM, "LLM", ""); err != nil {
		return err
	}

	// Resolve embedding
	if err := resolveRef(&c.Memory.LongTerm.Embedding, "embedding", ""); err != nil {
		return err
	}

	// Resolve TTS (must be OpenAI-compatible for TTS API)
	if c.TTS.Backend != "" {
		backend, ok := backends[c.TTS.Backend]
		if !ok {
			return fmt.Errorf("TTS backend %q not found", c.TTS.Backend)
		}
		if backend.Type != BackendTypeOpenAI {
			return fmt.Errorf("TTS backend %q must be type %q, got %q", c.TTS.Backend, BackendTypeOpenAI, backend.Type)
		}
		c.TTS.Resolved = backend
	}

	return nil
}
