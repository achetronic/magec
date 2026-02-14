#!/usr/bin/env bash
# Magec installer — https://github.com/achetronic/magec
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/achetronic/magec/main/scripts/install.sh | bash
#   curl -fsSL .../install.sh | bash -s -- --openai
#   curl -fsSL .../install.sh | bash -s -- --anthropic
#   curl -fsSL .../install.sh | bash -s -- --gemini
#   curl -fsSL .../install.sh | bash -s -- --gpu

set -euo pipefail

# ── Config ──────────────────────────────────────────────────────────────────

REPO="achetronic/magec"
BRANCH="main"
BASE_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}"
INSTALL_DIR="${MAGEC_DIR:-magec}"

# ── Colors ──────────────────────────────────────────────────────────────────

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

info()  { printf "${CYAN}▸${NC} %s\n" "$*"; }
ok()    { printf "${GREEN}✓${NC} %s\n" "$*"; }
warn()  { printf "${YELLOW}!${NC} %s\n" "$*"; }
err()   { printf "${RED}✗${NC} %s\n" "$*" >&2; }
die()   { err "$@"; exit 1; }

# ── Parse args ──────────────────────────────────────────────────────────────

MODE="local"
GPU=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --local)      MODE="local"; shift ;;
    --openai)     MODE="openai"; shift ;;
    --anthropic)  MODE="anthropic"; shift ;;
    --gemini)     MODE="gemini"; shift ;;
    --cloud)      MODE="openai"; shift ;;
    --gpu)        GPU=true; shift ;;
    --dir)        INSTALL_DIR="$2"; shift 2 ;;
    --help|-h)
      cat <<'EOF'

  Magec Installer

  Usage:
    curl -fsSL .../install.sh | bash
    curl -fsSL .../install.sh | bash -s -- [OPTIONS]

  Modes:
    --local       Fully local deployment (default). No API keys needed.
    --openai      Cloud APIs via OpenAI. Requires OPENAI_API_KEY.
    --anthropic   Cloud LLM via Anthropic. Requires ANTHROPIC_API_KEY.
    --gemini      Cloud LLM via Google Gemini. Requires GEMINI_API_KEY.

  Options:
    --gpu         Enable NVIDIA GPU support for Ollama (local mode only).
    --dir NAME    Installation directory (default: magec).
    -h, --help    Show this help.

  Environment:
    MAGEC_DIR           Installation directory (alternative to --dir).
    OPENAI_API_KEY      Required for --openai mode.
    ANTHROPIC_API_KEY   Required for --anthropic mode.
    GEMINI_API_KEY      Required for --gemini mode.

  Examples:
    # Fully local — nothing else needed
    curl -fsSL .../install.sh | bash

    # Local with NVIDIA GPU
    curl -fsSL .../install.sh | bash -s -- --gpu

    # OpenAI
    export OPENAI_API_KEY=sk-...
    curl -fsSL .../install.sh | bash -s -- --openai

    # Anthropic
    export ANTHROPIC_API_KEY=sk-ant-...
    curl -fsSL .../install.sh | bash -s -- --anthropic

    # Gemini
    export GEMINI_API_KEY=AI...
    curl -fsSL .../install.sh | bash -s -- --gemini

EOF
      exit 0
      ;;
    *) die "Unknown option: $1. Use --help for usage." ;;
  esac
done

# ── Banner ──────────────────────────────────────────────────────────────────

printf "\n${BOLD}"
cat <<'EOF'
  __  __
 |  \/  | __ _  __ _  ___  ___
 | |\/| |/ _` |/ _` |/ _ \/ __|
 | |  | | (_| | (_| |  __/ (__
 |_|  |_|\__,_|\__, |\___|\___|
               |___/
EOF
printf "${NC}\n"

case "$MODE" in
  local)     info "Mode: ${BOLD}fully local${NC} — no API keys needed" ;;
  openai)    info "Mode: ${BOLD}OpenAI${NC}" ;;
  anthropic) info "Mode: ${BOLD}Anthropic${NC}" ;;
  gemini)    info "Mode: ${BOLD}Google Gemini${NC}" ;;
esac
$GPU && info "GPU: ${BOLD}NVIDIA${NC} enabled"
echo

# ── Preflight checks ───────────────────────────────────────────────────────

check_cmd() {
  if ! command -v "$1" &>/dev/null; then
    die "$1 is required but not installed. $2"
  fi
}

check_cmd docker "Install it from https://docs.docker.com/get-docker/"
check_cmd curl   "Install it with your package manager."

if ! docker compose version &>/dev/null && ! docker-compose version &>/dev/null; then
  die "Docker Compose is required. Install it from https://docs.docker.com/compose/install/"
fi

if ! docker info &>/dev/null; then
  die "Docker daemon is not running. Start it and try again."
fi

require_env() {
  local var_name="$1" flag="$2" example="$3"
  if [[ -z "${!var_name:-}" ]]; then
    echo
    die "${var_name} is required for ${flag} mode. Run it like this:

  ${BOLD}export ${var_name}=${example}${NC}
  curl -fsSL .../install.sh | bash -s -- ${flag}

  Or inline:
  ${BOLD}curl -fsSL .../install.sh | ${var_name}=${example} bash -s -- ${flag}${NC}"
  fi
}

case "$MODE" in
  openai)    require_env "OPENAI_API_KEY"    "--openai"    "sk-..." ;;
  anthropic) require_env "ANTHROPIC_API_KEY" "--anthropic" "sk-ant-..." ;;
  gemini)    require_env "GEMINI_API_KEY"    "--gemini"    "AI..." ;;
esac

ok "All dependencies met"

# ── Resolve compose command ─────────────────────────────────────────────────

if docker compose version &>/dev/null; then
  COMPOSE="docker compose"
else
  COMPOSE="docker-compose"
fi

# ── Download files ──────────────────────────────────────────────────────────

COMPOSE_DIR="docker/compose"

case "$MODE" in
  local)     OVERRIDE_FILE="docker-compose.local.yaml" ;;
  openai)    OVERRIDE_FILE="docker-compose.openai.yaml" ;;
  anthropic) OVERRIDE_FILE="docker-compose.anthropic.yaml" ;;
  gemini)    OVERRIDE_FILE="docker-compose.gemini.yaml" ;;
esac

info "Creating ${INSTALL_DIR}/"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

info "Downloading configuration..."

curl -fsSL "${BASE_URL}/${COMPOSE_DIR}/docker-compose.yaml" -o docker-compose.yaml
curl -fsSL "${BASE_URL}/${COMPOSE_DIR}/${OVERRIDE_FILE}" -o docker-compose.override.yaml
curl -fsSL "${BASE_URL}/${COMPOSE_DIR}/config.yaml" -o config.yaml

ok "Files downloaded"

# ── GPU support ─────────────────────────────────────────────────────────────

if $GPU && [[ "$MODE" == "local" ]]; then
  info "Enabling NVIDIA GPU for Ollama..."
  if command -v sed &>/dev/null; then
    sed -i 's/^    # \(deploy:\)/    \1/' docker-compose.override.yaml
    sed -i 's/^    #   \(resources:\)/      \1/' docker-compose.override.yaml
    sed -i 's/^    #     \(reservations:\)/        \1/' docker-compose.override.yaml
    sed -i 's/^    #       \(devices:\)/          \1/' docker-compose.override.yaml
    sed -i 's/^    #         \(- driver: nvidia\)/            \1/' docker-compose.override.yaml
    sed -i 's/^    #           \(count: all\)/              \1/' docker-compose.override.yaml
    sed -i 's/^    #           \(capabilities: \[gpu\]\)/              \1/' docker-compose.override.yaml
    ok "GPU support enabled"
  else
    warn "Could not enable GPU automatically. Uncomment the 'deploy' section in docker-compose.override.yaml manually."
  fi
elif $GPU && [[ "$MODE" != "local" ]]; then
  warn "--gpu has no effect in cloud mode (no local Ollama)"
fi

# ── Launch ──────────────────────────────────────────────────────────────────

echo
info "Starting Magec..."

ENV_ARGS=()
case "$MODE" in
  openai)    ENV_ARGS+=(OPENAI_API_KEY="${OPENAI_API_KEY}") ;;
  anthropic) ENV_ARGS+=(ANTHROPIC_API_KEY="${ANTHROPIC_API_KEY}") ;;
  gemini)    ENV_ARGS+=(GEMINI_API_KEY="${GEMINI_API_KEY}") ;;
esac

if [[ ${#ENV_ARGS[@]} -gt 0 ]]; then
  env "${ENV_ARGS[@]}" $COMPOSE up -d
else
  $COMPOSE up -d
fi

# ── Done ────────────────────────────────────────────────────────────────────

echo
printf "${GREEN}${BOLD}"
cat <<'EOF'
  ┌──────────────────────────────────────────┐
  │           Magec is running! ☀            │
  ├──────────────────────────────────────────┤
  │                                          │
  │   Voice UI  →  http://localhost:8080     │
  │   Admin UI  →  http://localhost:8081     │
  │                                          │
  └──────────────────────────────────────────┘
EOF
printf "${NC}\n"

if [[ "$MODE" == "local" ]]; then
  info "First start downloads ~5GB of models. This may take a few minutes."
  info "Track progress: ${BOLD}${COMPOSE} logs -f ollama-setup${NC}"
fi

info "Manage: ${BOLD}cd ${INSTALL_DIR} && ${COMPOSE} [up -d | down | logs]${NC}"
echo
