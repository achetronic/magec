---
title: "Binary Installation"
---

Run Magec as a native binary on your machine. This is the best option when you want to use **local MCP tools** (filesystem access, git, shell commands), need full control over the process, or are developing on Magec itself.

Dependencies like Redis, PostgreSQL, or Ollama can be installed natively or run as Docker containers; the installer guides you through both options. Docker is the recommended way to manage dependencies: spinning up a database or an Ollama instance takes seconds and keeps your system clean.

The Magec binary is self-contained: it includes the Admin UI, Voice UI, wake word models, and all auxiliary ONNX models baked in. You just need the binary, a `config.yaml`, and the external services you want to use.

## One-line install

The interactive installer can also set up Magec as a binary: select **binary** mode when prompted and it will download the right binary for your platform, set up external dependencies, and generate the configuration:

```bash
curl -fsSL https://magec.dev/install | bash
```

## Manual setup

If you prefer to set things up yourself:

### Download

Download the latest release for your platform from the [GitHub Releases](https://github.com/achetronic/magec/releases) page:

| Platform | File |
|----------|------|
| Linux (x86_64) | `magec-linux-amd64.tar.gz` |
| Linux (ARM64) | `magec-linux-arm64.tar.gz` |
| macOS (Apple Silicon) | `magec-darwin-arm64.tar.gz` |
| Windows (x86_64) | `magec-windows-amd64.zip` |

Extract the archive:

```bash
tar xzf magec-linux-amd64.tar.gz
# Contains: magec (binary) + config.example.yaml
```

## Configuration

Copy the example config and adjust if needed:

```bash
cp config.example.yaml config.yaml
```

The defaults are fine for most setups:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081

voice:
  ui:
    enabled: true

log:
  level: info
  format: console
```

## Run Magec

```bash
./magec --config config.yaml
```

Magec starts and serves:

| URL | What |
|-----|------|
| `http://localhost:8081` | Admin UI + Admin API |
| `http://localhost:8080` | Voice UI + User API |

## External dependencies

The binary is self-contained for the core functionality, but some features need external programs installed on your system:

### Required: An LLM provider

Magec needs at least one AI backend. You configure it through the Admin UI after starting, but the service needs to be reachable. Common options:

**Ollama (local LLM):**

```bash
# Install Ollama — https://ollama.com
ollama serve
ollama pull qwen3:8b
ollama pull nomic-embed-text
```

Then create a backend in the Admin UI:
- Type: `openai`, URL: `http://localhost:11434/v1`

**OpenAI (cloud):**

No installation needed. Create a backend in the Admin UI:
- Type: `openai`, API Key: `sk-...`

**Anthropic / Gemini:**

Create the appropriate backend type in the Admin UI with your API key.

### Optional: ffmpeg (for voice messages)

Required for Telegram voice message processing (OGG → WAV conversion). Not needed if you only use the Voice UI or API.

```bash
# macOS
brew install ffmpeg

# Ubuntu / Debian
sudo apt install ffmpeg

# Arch
sudo pacman -S ffmpeg
```

Magec checks for `ffmpeg` in your PATH at startup and logs a warning if it's missing.

### Optional: ONNX Runtime (for wake word detection)

The wake word ("Oye Magec") and voice activity detection use ONNX models. The ONNX Runtime shared library is needed for these features. Without it, voice still works: you just use push-to-talk instead of wake word.

```bash
# Download ONNX Runtime 1.23.2
# Linux x86_64:
curl -LO https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-x64-1.23.2.tgz
tar xzf onnxruntime-linux-x64-1.23.2.tgz
sudo cp onnxruntime-linux-x64-1.23.2/lib/libonnxruntime.so* /usr/lib/

# Linux ARM64:
curl -LO https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-linux-aarch64-1.23.2.tgz
tar xzf onnxruntime-linux-aarch64-1.23.2.tgz
sudo cp onnxruntime-linux-aarch64-1.23.2/lib/libonnxruntime.so* /usr/lib/

# macOS (both architectures):
curl -LO https://github.com/microsoft/onnxruntime/releases/download/v1.23.2/onnxruntime-osx-universal2-1.23.2.tgz
tar xzf onnxruntime-osx-universal2-1.23.2.tgz
sudo cp onnxruntime-osx-universal2-1.23.2/lib/libonnxruntime.dylib /usr/local/lib/
```

If the library is in a non-standard location, point Magec at it in `config.yaml`:

```yaml
voice:
  onnxLibraryPath: /path/to/libonnxruntime.so
```

Magec logs a warning at startup if ONNX Runtime is not found. If the Voice UI is enabled, **ONNX Runtime is required:** the voice system needs it for both wake word detection and push-to-talk. Without it, the Voice UI will fail to initialize.

### Optional: Local STT / TTS

For voice with local speech processing (no cloud), you need a speech-to-text and text-to-speech service:

**Parakeet (STT):**

```bash
docker run -d -p 5092:5092 ghcr.io/achetronic/parakeet:latest
```

Backend: Type `openai`, URL `http://localhost:5092`

**OpenAI Edge TTS (local TTS):**

```bash
docker run -d -p 5050:5050 -e REQUIRE_API_KEY=False travisvn/openai-edge-tts:latest
```

Backend: Type `openai`, URL `http://localhost:5050`

These are small containers that don't need GPU. You can also run them natively if you prefer. Any service implementing the OpenAI-compatible `/v1/audio/transcriptions` (STT) or `/v1/audio/speech` (TTS) endpoints works.

## Optional: Memory (Redis + PostgreSQL)

For conversation history and long-term memory:

```bash
# Session memory
docker run -d -p 6379:6379 redis:alpine

# Long-term memory (requires pgvector)
docker run -d -p 5432:5432 \
  -e POSTGRES_USER=magec \
  -e POSTGRES_PASSWORD=magec \
  -e POSTGRES_DB=magec \
  pgvector/pgvector:pg17
```

Then configure in the Admin UI under **Memory**:
- Session: `redis://localhost:6379`
- Long-term: `postgres://magec:magec@localhost:5432/magec?sslmode=disable` with an embedding backend + model

Or use existing Redis/PostgreSQL instances on your network.

## Set up your first agent

1. Open `http://localhost:8081`
2. Create a **backend** (e.g., Ollama at `http://localhost:11434/v1`)
3. Create an **agent**: give it a name, system prompt, and select the backend + model
4. Optionally add voice (STT/TTS backends + models)
5. Create a **client** (Voice UI type) and copy the pairing token
6. Open `http://localhost:8080`, paste the token, and start talking

## MCP tools

The main reason to run Magec as a binary is **direct access to local MCP tools**. When Magec runs as a binary on your machine, stdio-based MCP servers can access your local filesystem, run shell commands, interact with git repositories, and do anything your user account can do.

Example MCP tools you can connect through the Admin UI:

```
# Filesystem access
npx -y @modelcontextprotocol/server-filesystem /home/user/projects

# GitHub
GITHUB_TOKEN=ghp_... npx -y @modelcontextprotocol/server-github

# Shell commands
npx -y @anthropic/mcp-shell

# SQLite
npx -y @anthropic/mcp-sqlite --db /path/to/database.db

# Any stdio MCP server
/path/to/your/custom-mcp-server
```

With the binary install, your agents can interact with your entire local environment through MCP tools. This is harder to set up with Docker (you'd need to mount volumes, expose ports, or run MCP servers on the host).

## Running as a service

### systemd (Linux)

Create `/etc/systemd/system/magec.service`:

```ini
[Unit]
Description=Magec AI Platform
After=network.target

[Service]
Type=simple
User=magec
WorkingDirectory=/opt/magec
ExecStart=/opt/magec/magec --config /opt/magec/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now magec
sudo journalctl -u magec -f    # follow logs
```

### launchd (macOS)

Create `~/Library/LaunchAgents/com.magec.server.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>com.magec.server</string>
  <key>ProgramArguments</key>
  <array>
    <string>/usr/local/bin/magec</string>
    <string>--config</string>
    <string>/usr/local/etc/magec/config.yaml</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>KeepAlive</key>
  <true/>
  <key>StandardOutPath</key>
  <string>/usr/local/var/log/magec.log</string>
  <key>StandardErrorPath</key>
  <string>/usr/local/var/log/magec.log</string>
</dict>
</plist>
```

```bash
launchctl load ~/Library/LaunchAgents/com.magec.server.plist
```

## Data

Magec stores its data in a `data/` directory relative to the working directory:

```
data/
├── store.json           # All configuration (agents, backends, clients, etc.)
└── conversations.json   # Conversation history
```

Back up `store.json` to preserve your entire configuration. To restore, copy it back and restart Magec.

## Next steps

- **[Configuration](/docs/configuration/).** Understand `config.yaml` vs. Admin UI resources
- **[Agents](/docs/agents/).** Create agents with custom prompts and tools
- **[MCP Tools](/docs/mcp/).** Connect local and remote tools to your agents
- **[Flows](/docs/flows/).** Chain agents into multi-step workflows
- **[Voice UI](/docs/voice-ui/).** Wake word, push-to-talk, PWA installation
