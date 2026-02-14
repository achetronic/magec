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

// Backend types (used by store and agent packages)
const (
	BackendTypeOpenAI    = "openai"
	BackendTypeAnthropic = "anthropic"
	BackendTypeGemini    = "gemini"

	DefaultOpenAIURL = "https://api.openai.com/v1"
)

// Config represents the YAML configuration file.
// Only server infrastructure settings live here.
// All resources (agents, backends, MCPs, memory, etc.) are managed via the admin API and persisted in the store.
type Config struct {
	Server Server `yaml:"server"`
	Voice  Voice  `yaml:"voice"`
	Log    Log    `yaml:"log"`
}

// Server holds network and runtime settings for the HTTP servers.
type Server struct {
	Host      string `yaml:"host"`
	Port      int    `yaml:"port"`
	AdminPort int    `yaml:"adminPort"`
}

// Voice holds voice-related configuration (UI, ONNX runtime, etc.).
type Voice struct {
	UI              VoiceUI `yaml:"ui"`
	OnnxLibraryPath string  `yaml:"onnxLibraryPath"`
}

// VoiceUI controls whether the Voice UI frontend and voice routes are enabled.
type VoiceUI struct {
	Enabled *bool `yaml:"enabled"`
}

// Log configures the application logger
type Log struct {
	Level  string `yaml:"level"`  // debug, info, warn, error (default: info)
	Format string `yaml:"format"` // console, json (default: console)
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

	cfg.applyDefaults()

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

// applyDefaults fills in zero-value fields with sensible defaults.
func (c *Config) applyDefaults() {
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.AdminPort == 0 {
		c.Server.AdminPort = 8081
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "console"
	}
	if c.Voice.UI.Enabled == nil {
		enabled := true
		c.Voice.UI.Enabled = &enabled
	}
}
