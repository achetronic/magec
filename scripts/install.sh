#!/usr/bin/env bash
# Magec interactive installer — https://github.com/achetronic/magec
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/achetronic/magec/master/scripts/install.sh | bash

set -euo pipefail

# ── Config ──────────────────────────────────────────────────────────────────

REPO="achetronic/magec"
BRANCH="master"
BASE_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}"
API_URL="https://api.github.com/repos/${REPO}"
INSTALL_DIR="${MAGEC_DIR:-magec}"
ONNX_VERSION="1.23.2"

# ── Terminal setup ──────────────────────────────────────────────────────────

TERM_WIDTH="$(tput cols 2>/dev/null || echo 60)"
(( TERM_WIDTH > 72 )) && TERM_WIDTH=72
(( TERM_WIDTH < 40 )) && TERM_WIDTH=40
BOX_W=$(( TERM_WIDTH - 4 ))

# ── Colors & styles ─────────────────────────────────────────────────────────

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
CYAN=$'\033[0;36m'
BLUE=$'\033[0;34m'
MAGENTA=$'\033[0;35m'
BOLD=$'\033[1m'
DIM=$'\033[2m'
ITALIC=$'\033[3m'
NC=$'\033[0m'

BG_CYAN=$'\033[46m'
BG_GREEN=$'\033[42m'
BG_YELLOW=$'\033[43m'
BG_BLUE=$'\033[44m'
BG_MAGENTA=$'\033[45m'
FG_BLACK=$'\033[30m'
FG_WHITE=$'\033[97m'

# ── Drawing primitives ─────────────────────────────────────────────────────

hline() {
  local ch="${1:-─}" len="${2:-$BOX_W}" i
  for (( i=0; i<len; i++ )); do printf '%s' "$ch"; done
}

box_top()    { printf "  ${DIM}╭$(hline)╮${NC}\n"; }
box_bottom() { printf "  ${DIM}╰$(hline)╯${NC}\n"; }
box_sep()    { printf "  ${DIM}├$(hline)┤${NC}\n"; }
box_empty()  { printf "  ${DIM}│${NC}%*s${DIM}│${NC}\n" "$BOX_W" ""; }

box_line() {
  local text="$1" color="${2:-}" align="${3:-left}"
  local stripped
  stripped="$(printf '%s' "$text" | sed $'s/\x1b\\[[0-9;]*m//g')"
  local text_len="${#stripped}"
  local pad=$(( BOX_W - text_len ))
  (( pad < 0 )) && pad=0

  if [[ "$align" == "center" ]]; then
    local left_pad=$(( pad / 2 ))
    local right_pad=$(( pad - left_pad ))
    printf "  ${DIM}│${NC}%*s${color}%s${NC}%*s${DIM}│${NC}\n" "$left_pad" "" "$text" "$right_pad" ""
  else
    printf "  ${DIM}│${NC} ${color}%s${NC}%*s${DIM}│${NC}\n" "$text" "$(( pad - 1 ))" ""
  fi
}

badge() {
  local text="$1" bg="${2:-$BG_CYAN}" fg="${3:-$FG_BLACK}"
  printf "${bg}${fg}${BOLD} %s ${NC}" "$text"
}

info()    { printf "  ${CYAN}▸${NC} %s\n" "$*"; }
ok()      { printf "  ${GREEN}✓${NC} %s\n" "$*"; }
warn()    { printf "  ${YELLOW}⚠${NC} %s\n" "$*"; }
err()     { printf "  ${RED}✗${NC} %s\n" "$*" >&2; }
die()     { err "$@"; exit 1; }
cls()     { printf '\033[H\033[2J\033[3J'; }

step_header() {
  local num="$1" title="$2"
  echo
  printf "  $(badge "STEP ${num}" "$BG_CYAN" "$FG_BLACK")  ${BOLD}%s${NC}\n" "$title"
  printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
  echo
}

# ── YAML helpers ─────────────────────────────────────────────────────────────

yaml_quote() {
  printf "'%s'" "$(printf '%s' "$1" | sed "s/'/''/g")"
}

# ── Input helpers ───────────────────────────────────────────────────────────

ask() {
  local prompt="$1" default="${2:-}"
  if [[ -n "$default" ]]; then
    printf "  ${CYAN}▸${NC} %s ${DIM}[%s]${NC}: " "$prompt" "$default"
  else
    printf "  ${CYAN}▸${NC} %s: " "$prompt"
  fi
  read -r REPLY < /dev/tty
  REPLY="${REPLY:-$default}"
}

ask_yn() {
  local prompt="$1" default="${2:-y}"
  local hint="Y/n"
  [[ "$default" == "n" ]] && hint="y/N"
  printf "  ${CYAN}▸${NC} %s ${DIM}[%s]${NC}: " "$prompt" "$hint"
  read -r REPLY < /dev/tty
  REPLY="${REPLY:-$default}"
  [[ "$REPLY" =~ ^[Yy]$ ]]
}

choose() {
  local i=1
  for opt in "$@"; do
    printf "  ${BOLD}${CYAN}%d)${NC} %s\n" "$i" "$opt"
    ((i++))
  done
  echo
  printf "  ${DIM}Your choice${NC}: "
  read -r REPLY < /dev/tty
  while [[ ! "$REPLY" =~ ^[0-9]+$ ]] || (( REPLY < 1 || REPLY > $# )); do
    printf "  ${RED}Pick a number between 1 and %d${NC}: " "$#"
    read -r REPLY < /dev/tty
  done
}

# ── Progress indicator ──────────────────────────────────────────────────────

progress_dots() {
  local msg="$1" pid="$2"
  printf "  ${CYAN}▸${NC} %s " "$msg"
  while kill -0 "$pid" 2>/dev/null; do
    printf "${DIM}.${NC}"
    sleep 0.4
  done
  wait "$pid" 2>/dev/null
  echo
}

# ═══════════════════════════════════════════════════════════════════════════
#  WELCOME
# ═══════════════════════════════════════════════════════════════════════════

clear 2>/dev/null || true

# Logo: SVG gradient runs from coral (#f87171) top-right → amber (#f59e0b) bottom-left
C=$'\033[38;5;203m'   # coral/lava (top-right)
M=$'\033[38;5;209m'   # mid orange
A=$'\033[38;5;214m'   # amber (bottom-left)
W=$'\033[38;5;255m'   # white purpurina

echo
echo
printf "                    ${M}▄▄████${C}████▄▄${NC}\n"
printf "               ${M}▄████████${C}████████████▄${NC}\n"
printf "            ${A}▄██████${M}██████████${C}██████████▄${NC}\n"
printf "          ${A}▄████████████${M}██████${C}██${W}██${C}████████▄${NC}\n"
printf "         ${A}█████████${W}██${A}██${M}████████${C}█████████████${NC}\n"
printf "         ${A}██████████████${M}████${W}███${M}████${C}█████████${NC}\n"
printf "         ${A}██████████████████${M}████${W}█${C}███████████${NC}\n"
printf "          ${A}▀██████${W}█${A}██████${M}████████${C}█████████▀${NC}\n"
printf "            ${A}▀████████████${W}██${M}████████████▀${NC}\n"
printf "               ${A}▀████████████${M}████████▀${NC}\n"
printf "                    ${A}▀▀████${M}████▀▀${NC}\n"
printf "${NC}\n"
printf "${A}${BOLD}"
cat <<'EOF'
          __  __
         |  \/  | __ _  __ _  ___  ___
         | |\/| |/ _` |/ _` |/ _ \/ __|
         | |  | | (_| | (_| |  __/ (__
         |_|  |_|\__,_|\__, |\___|\___|
                       |___/
EOF
printf "${NC}\n"

box_top
box_empty
box_line "Welcome to Magec" "$BOLD" "center"
box_line "Multi-agent AI platform" "$DIM" "center"
box_empty
box_sep
box_empty
box_line "This installer will guide you step by step."
box_line "No technical knowledge required — just answer"
box_line "a few simple questions and we'll handle the rest."
box_empty
box_line "At the end you'll have Magec up and running with:"
box_empty
box_line "  •  AI agents you can talk to"
box_line "  •  A web interface to manage everything"
box_line "  •  Optional voice, memory, and more"
box_empty
box_bottom

echo
printf "  ${DIM}Press Enter to begin...${NC}"
read -r < /dev/tty

# ═══════════════════════════════════════════════════════════════════════════
#  DETECT PLATFORM
# ═══════════════════════════════════════════════════════════════════════════

detect_platform() {
  OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
  ARCH="$(uname -m)"

  case "$OS" in
    linux*)  OS="linux" ;;
    darwin*) OS="darwin" ;;
    mingw*|msys*|cygwin*) OS="windows" ;;
    *) die "Unsupported operating system: $OS" ;;
  esac

  case "$ARCH" in
    x86_64|amd64)  ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) die "Unsupported architecture: $ARCH" ;;
  esac
}

detect_platform

os_label="$OS"
[[ "$OS" == "darwin" ]] && os_label="macOS"
[[ "$OS" == "linux" ]]  && os_label="Linux"
[[ "$OS" == "windows" ]] && os_label="Windows"
arch_label="$ARCH"
[[ "$ARCH" == "amd64" ]] && arch_label="64-bit (x86)"
[[ "$ARCH" == "arm64" ]] && arch_label="ARM (Apple Silicon / ARM64)"

echo
box_top
box_line "  Detected system:" "$BOLD"
box_line "  ${os_label}, ${arch_label}"
box_bottom
echo

# ═══════════════════════════════════════════════════════════════════════════
#  STEP 1 — INSTALLATION METHOD
# ═══════════════════════════════════════════════════════════════════════════

cls
step_header 1 "How should we install Magec?"

box_top
box_empty
box_line "There are two ways to run Magec. Pick whichever" "$BOLD"
box_line "sounds easier for you:" "$BOLD"
box_empty
box_sep
box_empty
box_line "$(badge " 1 " "$BG_CYAN" "$FG_BLACK")  Download the program directly"
box_empty
box_line "  We download a single file to your computer."
box_line "  You run it like any other app. Best if you want"
box_line "  Magec to access local files, scripts, and tools"
box_line "  on your machine."
box_empty
box_sep
box_empty
box_line "$(badge " 2 " "$BG_MAGENTA" "$FG_WHITE")  Use containers (Docker)"
box_empty
box_line "  Everything runs inside isolated containers."
box_line "  Nothing gets installed system-wide. Easiest"
box_line "  to set up — one command starts everything."
box_line "  Requires Docker to be installed."
box_empty
box_bottom
echo

choose \
  "Download the program directly (binary)" \
  "Use containers (Docker)"

INSTALL_METHOD="$REPLY"

# ═══════════════════════════════════════════════════════════════════════════
#  STEP 2 — AI MODELS
# ═══════════════════════════════════════════════════════════════════════════

cls
step_header 2 "Where should your AI brain live?"

box_top
box_empty
box_line "Magec needs an AI model to think and respond." "$BOLD"
box_line "You can run one locally on this machine, use" "$BOLD"
box_line "cloud services, or both:" "$BOLD"
box_empty
box_sep
box_empty
box_line "$(badge " 1 " "$BG_GREEN" "$FG_BLACK")  Local  — runs on this machine"
box_empty
box_line "  Uses Ollama to run AI models privately."
box_line "  No API keys, no cloud, your data stays here."
box_line "  Needs ~5 GB of disk space for the AI model."
box_empty
box_sep
box_empty
box_line "$(badge " 2 " "$BG_BLUE" "$FG_WHITE")  Cloud  — OpenAI, Anthropic, Gemini"
box_empty
box_line "  Uses cloud AI services (faster, no local GPU"
box_line "  needed). You'll need API keys from at least"
box_line "  one provider — you can add them after install."
box_empty
box_sep
box_empty
box_line "$(badge " 3 " "$BG_YELLOW" "$FG_BLACK")  Both   — local + cloud"
box_empty
box_line "  Best of both worlds. Use local models for"
box_line "  privacy, cloud models for speed. Switch"
box_line "  between them per agent in the admin panel."
box_empty
box_bottom
echo

choose \
  "Local — private, runs on this machine" \
  "Cloud — OpenAI, Anthropic, Gemini" \
  "Both — local + cloud"

LLM_CHOICE="$REPLY"

# ── CLIProxyAPI (use Claude subscription as API) ─────────────────────────

WANT_CLIPROXYAPI=false

if [[ "$LLM_CHOICE" == "2" || "$LLM_CHOICE" == "3" ]]; then
  echo
  box_top
  box_empty
  box_line "  Use Claude Without an API Key?" "$BOLD" "center"
  box_empty
  box_sep
  box_empty
  box_line "  If you have a Claude Max or Pro subscription,"
  box_line "  you can use your existing account instead of"
  box_line "  paying separately for API access."
  box_empty
  box_line "  This uses CLIProxyAPI — a local proxy that"
  box_line "  connects your Claude subscription to Magec."
  box_line "  You'll log in with your Anthropic account once,"
  box_line "  and Magec can use Claude models from there."
  box_empty
  box_line "  ${DIM}You can still add a regular Anthropic API key${NC}"
  box_line "  ${DIM}later if you prefer.${NC}"
  box_empty
  box_bottom
  echo

  if ask_yn "Enable Claude subscription proxy (CLIProxyAPI)?" "n"; then
    WANT_CLIPROXYAPI=true
    ok "CLIProxyAPI will be configured"
  else
    info "Skipped — you can set it up later"
  fi
fi

# ═══════════════════════════════════════════════════════════════════════════
#  STEP 3 — MEMORY
# ═══════════════════════════════════════════════════════════════════════════

cls
step_header 3 "Should Magec remember your conversations?"

WANT_REDIS=false
WANT_POSTGRES=false

box_top
box_empty
box_line "  Conversation Memory" "$BOLD" "center"
box_empty
box_sep
box_empty
box_line "  By default, Magec already remembers your"
box_line "  conversations while it's running. But if Magec"
box_line "  restarts or crashes, those conversations are lost."
box_empty
box_line "  With this option, conversations are saved to disk"
box_line "  so they survive restarts. You can close the"
box_line "  browser, reboot the machine, and pick up right"
box_line "  where you left off."
box_empty
if [[ "$INSTALL_METHOD" == "2" ]]; then
  box_line "  ${DIM}A Redis service will be added to your containers.${NC}"
else
  box_line "  ${DIM}Needs Redis running on this machine (via Docker${NC}"
  box_line "  ${DIM}or installed natively).${NC}"
fi
box_empty
box_bottom
echo

if ask_yn "Enable conversation memory?"; then
  WANT_REDIS=true
  ok "Conversation memory enabled"
else
  info "No worries — you can always enable this later"
fi

# ── Long-term memory ─────────────────────────────────────────────────────

cls
step_header 3 "Long-term memory"

box_top
box_empty
box_line "  Long-Term Memory" "$BOLD" "center"
box_empty
box_sep
box_empty
box_line "  This is like giving Magec a notebook. It can"
box_line "  remember facts, preferences, and context across"
box_line "  all your conversations — forever."
box_empty
box_line "  For example, if you tell it \"I prefer Python\","
box_line "  it will remember that next time you ask for code."
box_empty
box_line "  This uses semantic search — Magec doesn't just"
box_line "  store text, it understands the meaning behind it."
box_empty
if [[ "$INSTALL_METHOD" == "2" ]]; then
  box_line "  ${DIM}A PostgreSQL database will be added to your containers.${NC}"
else
  box_line "  ${DIM}Needs PostgreSQL running on this machine (via Docker${NC}"
  box_line "  ${DIM}or installed natively).${NC}"
fi
box_empty
box_bottom
echo

if ask_yn "Enable long-term memory?"; then
  WANT_POSTGRES=true
  ok "Long-term memory enabled"
  echo
  if [[ "$LLM_CHOICE" == "2" ]]; then
    box_top
    box_empty
    box_line "  ${YELLOW}Note${NC}" "" "center"
    box_empty
    box_line "  Long-term memory needs a way to understand the"
    box_line "  meaning of text (an \"embedding model\"). Since"
    box_line "  you chose cloud-only, you'll need to configure"
    box_line "  a cloud embedding model after install (like"
    box_line "  OpenAI's text-embedding-ada-002)."
    box_empty
    box_bottom
  else
    info "Will use a local embedding model for semantic search"
  fi
else
  info "No worries — you can always enable this later"
fi

# ═══════════════════════════════════════════════════════════════════════════
#  STEP 4 — VOICE
# ═══════════════════════════════════════════════════════════════════════════

cls
step_header 4 "Do you want to talk to Magec?"

WANT_VOICE=true
WANT_ONNX=false

box_top
box_empty
box_line "  Voice Interface" "$BOLD" "center"
box_empty
box_sep
box_empty
box_line "  Magec comes with a voice-enabled web interface."
box_line "  You can talk to your AI agents like a voice"
box_line "  assistant — just say \"Oye Magec\" or press a"
box_line "  button to start speaking."
box_empty
box_line "  It handles speech-to-text, text-to-speech, and"
box_line "  wake word detection, all running privately."
box_empty
box_bottom
echo

if ask_yn "Enable the voice interface?"; then
  WANT_VOICE=true
  ok "Voice interface enabled"

  if [[ "$INSTALL_METHOD" == "1" ]]; then
    echo
    box_top
    box_empty
    box_line "  Voice Engine" "$BOLD" "center"
    box_empty
    box_sep
    box_empty
    box_line "  The voice features (wake word detection and"
    box_line "  understanding when you stop speaking) need a"
    box_line "  small component called a \"voice engine\"."
    box_empty
    box_line "  We can download it automatically (~25 MB)."
    box_line "  It's a library that runs AI voice models"
    box_line "  efficiently on your machine."
    box_empty
    box_line "  ${YELLOW}Without it, the voice interface won't work.${NC}"
    box_line "  (You can still use text chat.)"
    box_empty
    box_bottom
    echo

    if ask_yn "Download the voice engine?"; then
      WANT_ONNX=true
      ok "Voice engine will be downloaded"
    else
      warn "Voice may not work without the engine."
      warn "You can install it manually later."
      info "See: https://github.com/microsoft/onnxruntime/releases"
    fi
  fi
else
  WANT_VOICE=false
  info "Voice disabled — you can still chat via text"
fi

# ═══════════════════════════════════════════════════════════════════════════
#  STEP 5 — GPU (containers with local models only)
# ═══════════════════════════════════════════════════════════════════════════

GPU=false

if [[ "$INSTALL_METHOD" == "2" ]] && [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]]; then
  cls
  step_header 5 "GPU acceleration"

  box_top
  box_empty
  box_line "  Speed Up with Your Graphics Card" "$BOLD" "center"
  box_empty
  box_sep
  box_empty
  box_line "  If you have an NVIDIA graphics card, Magec can"
  box_line "  use it to run AI models much faster."
  box_empty
  box_line "  This is optional — everything works without it,"
  box_line "  just a bit slower for local AI models."
  box_empty
  box_line "  ${DIM}Requires: NVIDIA GPU + Container Toolkit${NC}"
  box_line "  ${DIM}https://docs.nvidia.com/datacenter/cloud-native/${NC}"
  box_line "  ${DIM}container-toolkit/install-guide.html${NC}"
  box_empty
  box_bottom
  echo

  if ask_yn "Enable NVIDIA GPU acceleration?" "n"; then
    echo
    if ask_yn "Is the NVIDIA Container Toolkit already installed?"; then
      GPU=true
      ok "GPU acceleration enabled"
    else
      warn "GPU skipped — install the toolkit first, then edit docker-compose.yaml"
    fi
  else
    info "No GPU — that's fine, Magec works great on CPU too"
  fi
fi

# ═══════════════════════════════════════════════════════════════════════════
#  STEP 6 — INSTALL DIRECTORY
# ═══════════════════════════════════════════════════════════════════════════

local_step=5
[[ "$INSTALL_METHOD" == "2" ]] && [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]] && local_step=6

cls
step_header "$local_step" "Where should we install?"

box_top
box_empty
box_line "  All Magec files (configuration, data, and the"
box_line "  program itself) will go into a single folder."
box_empty
box_line "  You can move it later if you change your mind."
box_empty
box_bottom
echo

ask "Folder name" "$INSTALL_DIR"
INSTALL_DIR="$REPLY"

# ═══════════════════════════════════════════════════════════════════════════
#  DETECT EXISTING INSTALLATION
# ═══════════════════════════════════════════════════════════════════════════

SKIP_DATA=false
SKIP_CONFIG=false

if [[ -d "${INSTALL_DIR}/data" ]] || [[ -f "${INSTALL_DIR}/config.yaml" ]]; then
  cls

  (( local_step++ ))
  step_header "$local_step" "Existing installation detected"

  box_top
  box_empty
  box_line "  ${YELLOW}Existing Magec data found${NC}" "" "center"
  box_empty
  box_sep
  box_empty
  if [[ -d "${INSTALL_DIR}/data" ]]; then
    box_line "  A ${BOLD}data/${NC} folder exists in ${CYAN}${INSTALL_DIR}/${NC}."
    box_line "  It contains your agents, backends, secrets,"
    box_line "  conversations, and other configuration."
    box_empty
  fi
  if [[ -f "${INSTALL_DIR}/config.yaml" ]]; then
    box_line "  A ${BOLD}config.yaml${NC} exists with your admin password,"
    box_line "  encryption key, and other settings."
    box_empty
  fi
  box_sep
  box_empty
  box_line "  ${RED}Overwriting will permanently delete this data.${NC}"
  box_empty
  box_bottom
  echo

  choose \
    "Keep my data and config (recommended)" \
    "Overwrite everything (data will be lost)"

  if [[ "$REPLY" == "1" ]]; then
    SKIP_DATA=true
    SKIP_CONFIG=true
    ok "Your existing installation will be preserved"
  else
    warn "Existing data will be overwritten"
  fi
fi

# ═══════════════════════════════════════════════════════════════════════════
#  STEP — ADMIN PASSWORD
# ═══════════════════════════════════════════════════════════════════════════

ADMIN_PASSWORD=""
ENCRYPTION_KEY=""

if ! $SKIP_CONFIG; then

  (( local_step++ ))
  cls
  step_header "$local_step" "Protect the Admin Panel"

  box_top
  box_empty
  box_line "  The Admin Panel lets you manage agents, backends,"
  box_line "  API keys, and everything else in Magec."
  box_empty
  box_sep
  box_empty
  box_line "  Setting a password will protect the Admin API"
  box_line "  and the web panel — no one can access them"
  box_line "  without logging in first."
  box_empty
  box_line "  ${DIM}Without a password, anyone on the network can${NC}"
  box_line "  ${DIM}change your configuration.${NC}"
  box_empty
  box_sep
  box_empty
  box_line "  ${DIM}Leave blank to skip (not recommended for servers).${NC}"
  box_line "  ${DIM}You can set it later in config.yaml or via:${NC}"
  box_line "  ${CYAN}MAGEC_ADMIN_PASSWORD${NC}"
  box_empty
  box_bottom
  echo

  printf "  ${CYAN}▸${NC} Admin password (input hidden): "
  read -rs REPLY < /dev/tty
  echo
  ADMIN_PASSWORD="$REPLY"

  if [[ -n "$ADMIN_PASSWORD" ]]; then
    ok "Password set"
  else
    info "No password — Admin Panel will be open"
  fi

  # ═══════════════════════════════════════════════════════════════════════════
  #  STEP — ENCRYPTION KEY
  # ═══════════════════════════════════════════════════════════════════════════

  (( local_step++ ))
  cls
  step_header "$local_step" "Encrypt your secrets"

  box_top
  box_empty
  box_line "  Magec can store sensitive data like API keys"
  box_line "  for cloud providers (OpenAI, Anthropic, etc.)."
  box_empty
  box_sep
  box_empty
  box_line "  An encryption key protects these secrets on disk."
  box_line "  Without it, they are stored as plain text in"
  box_line "  ${CYAN}data/store.json${NC}."
  box_empty
  box_line "  ${DIM}This key is independent from the admin password,${NC}"
  box_line "  ${DIM}so you can rotate one without affecting the other.${NC}"
  box_empty
  box_sep
  box_empty
  box_line "  ${DIM}Leave blank to skip (secrets stored unencrypted).${NC}"
  box_line "  ${DIM}You can set it later in config.yaml or via:${NC}"
  box_line "  ${CYAN}MAGEC_ENCRYPTION_KEY${NC}"
  box_empty
  box_bottom
  echo

  printf "  ${CYAN}▸${NC} Encryption key (input hidden): "
  read -rs REPLY < /dev/tty
  echo
  ENCRYPTION_KEY="$REPLY"

  if [[ -n "$ENCRYPTION_KEY" ]]; then
    ok "Encryption key set"
  else
    info "No encryption — secrets will be stored in plain text"
  fi

fi

# ═══════════════════════════════════════════════════════════════════════════
#  SUMMARY
# ═══════════════════════════════════════════════════════════════════════════

cls
echo
printf "  $(badge " SUMMARY " "$BG_GREEN" "$FG_BLACK")\n"
printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
echo

method_label="Direct download (binary)"
[[ "$INSTALL_METHOD" == "2" ]] && method_label="Containers (Docker)"

llm_label="Local (Ollama — private)"
[[ "$LLM_CHOICE" == "2" ]] && llm_label="Cloud (OpenAI, Anthropic, Gemini)"
[[ "$LLM_CHOICE" == "3" ]] && llm_label="Both (local + cloud)"

voice_label="Enabled"
[[ "$WANT_VOICE" != true ]] && voice_label="Disabled"

box_top
box_empty
box_line "  ${BOLD}Install method:${NC}    $method_label"
box_line "  ${BOLD}System:${NC}            ${os_label} (${arch_label})"
box_line "  ${BOLD}AI models:${NC}         $llm_label"
if $WANT_CLIPROXYAPI; then
  box_line "  ${BOLD}Claude proxy:${NC}      Enabled (CLIProxyAPI)"
fi
box_line "  ${BOLD}Conversation memory:${NC} $($WANT_REDIS && echo "Yes" || echo "No")"
box_line "  ${BOLD}Long-term memory:${NC}  $($WANT_POSTGRES && echo "Yes" || echo "No")"
box_line "  ${BOLD}Voice:${NC}             $voice_label"
if [[ "$INSTALL_METHOD" == "1" && "$WANT_VOICE" == true ]]; then
  box_line "  ${BOLD}Voice engine:${NC}      $($WANT_ONNX && echo "Will download" || echo "Manual")"
fi
if [[ "$INSTALL_METHOD" == "2" ]] && $GPU; then
  box_line "  ${BOLD}GPU:${NC}               NVIDIA"
fi
box_line "  ${BOLD}Folder:${NC}            $INSTALL_DIR/"
if $SKIP_CONFIG; then
  box_line "  ${BOLD}Admin password:${NC}    ${GREEN}Kept existing${NC}"
  box_line "  ${BOLD}Encryption key:${NC}    ${GREEN}Kept existing${NC}"
else
  pass_label="${GREEN}Set${NC}"
  [[ -z "$ADMIN_PASSWORD" ]] && pass_label="${YELLOW}None (open)${NC}"
  box_line "  ${BOLD}Admin password:${NC}    $pass_label"
  enc_label="${GREEN}Set${NC}"
  [[ -z "$ENCRYPTION_KEY" ]] && enc_label="${YELLOW}None (plain text)${NC}"
  box_line "  ${BOLD}Encryption key:${NC}    $enc_label"
fi
if $SKIP_DATA; then
  box_line "  ${BOLD}Data:${NC}              ${GREEN}Kept existing${NC}"
fi
box_empty
box_bottom
echo

if ! ask_yn "Everything look good? Start installation?"; then
  echo
  info "Installation cancelled. Run the script again anytime."
  exit 0
fi

# ═══════════════════════════════════════════════════════════════════════════
#  HELPERS
# ═══════════════════════════════════════════════════════════════════════════

check_cmd() {
  command -v "$1" &>/dev/null
}

require_cmd() {
  if ! check_cmd "$1"; then
    die "$1 is required but not installed. $2"
  fi
}

gen_uuid() {
  if check_cmd uuidgen; then
    uuidgen | tr '[:upper:]' '[:lower:]'
  elif [[ -f /proc/sys/kernel/random/uuid ]]; then
    cat /proc/sys/kernel/random/uuid
  else
    printf '%04x%04x-%04x-%04x-%04x-%04x%04x%04x' \
      $RANDOM $RANDOM $RANDOM $(( ($RANDOM & 0x0fff) | 0x4000 )) \
      $(( ($RANDOM & 0x3fff) | 0x8000 )) $RANDOM $RANDOM $RANDOM
  fi
}

# ═══════════════════════════════════════════════════════════════════════════
#  SERVICE SETUP (binary install)
# ═══════════════════════════════════════════════════════════════════════════

test_redis() {
  if check_cmd redis-cli; then
    redis-cli -h localhost -p 6379 ping 2>/dev/null | grep -qi pong
  else
    bash -c 'exec 3<>/dev/tcp/localhost/6379 && echo PING >&3 && read -t2 reply <&3 && [[ "$reply" == *PONG* ]]' 2>/dev/null
  fi
}

test_postgres() {
  if check_cmd pg_isready; then
    pg_isready -h localhost -p 5432 -U magec -d magec &>/dev/null
  elif check_cmd psql; then
    PGPASSWORD=magec psql -h localhost -p 5432 -U magec -d magec -c "SELECT 1" &>/dev/null
  else
    return 1
  fi
}

test_ollama() {
  if check_cmd ollama; then
    ollama list &>/dev/null
  else
    curl -sf http://localhost:11434/api/tags &>/dev/null
  fi
}

setup_redis() {
  local has_docker=false
  check_cmd docker && docker info &>/dev/null 2>&1 && has_docker=true

  local retry=false
  while true; do
    cls
    echo
    printf "  $(badge " SETUP " "$BG_YELLOW" "$FG_BLACK")  ${BOLD}Conversation memory (Redis)${NC}\n"
    printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
    echo

    box_top
    box_empty
    box_line "  You chose to enable conversation memory, which"
    box_line "  needs Redis running on this machine."
    box_empty
    box_sep
    box_empty
    box_line "  ${BOLD}Connection:${NC}  ${CYAN}localhost:6379${NC}"
    box_line "  ${BOLD}Password:${NC}    none (default)"
    box_line "  ${BOLD}Database:${NC}    0 (default)"
    box_empty
    box_sep
    box_empty

    if $has_docker; then
      box_line "  ${BOLD}Quickest way — run with Docker:${NC}"
      box_empty
      box_line "  ${CYAN}docker run -d -p 6379:6379 --name magec-redis \\${NC}"
      box_line "  ${CYAN}  redis:alpine${NC}"
    else
      box_line "  ${BOLD}Install Redis:${NC}"
      box_empty
      case "$OS" in
        linux)
          box_line "  ${CYAN}sudo apt install redis-server${NC}"
          box_line "  ${CYAN}sudo systemctl start redis${NC}"
          ;;
        darwin)
          box_line "  ${CYAN}brew install redis${NC}"
          box_line "  ${CYAN}brew services start redis${NC}"
          ;;
        *)
          box_line "  ${CYAN}https://redis.io/download${NC}"
          ;;
      esac
      box_empty
      box_sep
      box_empty
      box_line "  ${DIM}Or install Docker and run:${NC}"
      box_empty
      box_line "  ${DIM}docker run -d -p 6379:6379 --name magec-redis redis:alpine${NC}"
    fi
    box_empty
    box_bottom
    echo

    info "Open another terminal, run the commands above,"
    info "then come back here."
    echo

    if $retry; then
      local msg="Could not connect to Redis on localhost:6379"
      local msg2="Make sure Redis is running and try again."
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" ""
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg2"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" ""
      echo
    fi

    echo
    choose \
      "Test connection" \
      "Skip — I'll set it up later"

    if [[ "$REPLY" == "2" ]]; then
      info "Skipped — remember to start Redis before running Magec"
      return
    fi

    echo
    info "Testing Redis on localhost:6379..."
    if test_redis; then
      ok "Redis is reachable"
      sleep 1
      return
    fi
    retry=true
  done
}

setup_postgres() {
  local has_docker=false
  check_cmd docker && docker info &>/dev/null 2>&1 && has_docker=true

  local retry=false
  while true; do
    cls
    echo
    printf "  $(badge " SETUP " "$BG_YELLOW" "$FG_BLACK")  ${BOLD}Long-term memory (PostgreSQL)${NC}\n"
    printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
    echo

    box_top
    box_empty
    box_line "  You chose to enable long-term memory, which needs"
    box_line "  PostgreSQL with the pgvector extension."
    box_empty
    box_sep
    box_empty
    box_line "  ${BOLD}Connection details:${NC}"
    box_empty
    box_line "  Host:      ${CYAN}localhost:5432${NC}"
    box_line "  User:      ${CYAN}magec${NC}"
    box_line "  Password:  ${CYAN}magec${NC}"
    box_line "  Database:  ${CYAN}magec${NC}"
    box_line "  Extension: ${CYAN}pgvector${NC}"
    box_empty
    box_sep
    box_empty

    if $has_docker; then
      box_line "  ${BOLD}Quickest way — run with Docker:${NC}"
      box_empty
      box_line "  ${CYAN}docker run -d -p 5432:5432 --name magec-postgres \\${NC}"
      box_line "  ${CYAN}  -e POSTGRES_USER=magec \\${NC}"
      box_line "  ${CYAN}  -e POSTGRES_PASSWORD=magec \\${NC}"
      box_line "  ${CYAN}  -e POSTGRES_DB=magec \\${NC}"
      box_line "  ${CYAN}  pgvector/pgvector:pg17${NC}"
      box_empty
      box_line "  Then enable the pgvector extension:"
      box_empty
      box_line "  ${CYAN}docker exec magec-postgres psql -U magec -d magec \\${NC}"
      box_line "  ${CYAN}  -c \"CREATE EXTENSION IF NOT EXISTS vector;\"${NC}"
    else
      box_line "  ${BOLD}Option A: Install natively${NC}"
      box_empty
      case "$OS" in
        linux)
          box_line "  ${CYAN}sudo apt install postgresql postgresql-17-pgvector${NC}"
          box_line "  ${CYAN}sudo systemctl start postgresql${NC}"
          ;;
        darwin)
          box_line "  ${CYAN}brew install postgresql@17 pgvector${NC}"
          box_line "  ${CYAN}brew services start postgresql@17${NC}"
          ;;
        *)
          box_line "  ${CYAN}https://www.postgresql.org/download/${NC}"
          box_line "  ${CYAN}https://github.com/pgvector/pgvector#installation${NC}"
          ;;
      esac
      box_empty
      box_line "  Then create the user, database, and extension:"
      box_empty
      box_line "  ${CYAN}sudo -u postgres createuser magec${NC}"
      box_line "  ${CYAN}sudo -u postgres createdb -O magec magec${NC}"
      box_line "  ${CYAN}sudo -u postgres psql -d magec \\${NC}"
      box_line "  ${CYAN}  -c \"ALTER USER magec PASSWORD 'magec';\"${NC}"
      box_line "  ${CYAN}sudo -u postgres psql -d magec \\${NC}"
      box_line "  ${CYAN}  -c \"CREATE EXTENSION IF NOT EXISTS vector;\"${NC}"
      box_empty
      box_sep
      box_empty
      box_line "  ${BOLD}Option B: Use Docker${NC}"
      box_empty
      box_line "  Install Docker, then:"
      box_empty
      box_line "  ${DIM}docker run -d -p 5432:5432 --name magec-postgres \\${NC}"
      box_line "  ${DIM}  -e POSTGRES_USER=magec -e POSTGRES_PASSWORD=magec \\${NC}"
      box_line "  ${DIM}  -e POSTGRES_DB=magec pgvector/pgvector:pg17${NC}"
      box_empty
      box_line "  ${DIM}docker exec magec-postgres psql -U magec -d magec \\${NC}"
      box_line "  ${DIM}  -c \"CREATE EXTENSION IF NOT EXISTS vector;\"${NC}"
    fi
    box_empty
    box_bottom
    echo

    info "Open another terminal, run the commands above,"
    info "then come back here."
    echo

    if $retry; then
      local msg="Could not connect to PostgreSQL on localhost:5432"
      local msg2="Check that PostgreSQL is running with the right"
      local msg3="user, password, database, and extension."
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" ""
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg2"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg3"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" ""
      echo
    fi

    echo
    choose \
      "Test connection" \
      "Skip — I'll set it up later"

    if [[ "$REPLY" == "2" ]]; then
      info "Skipped — remember to start PostgreSQL before running Magec"
      return
    fi

    echo
    info "Testing PostgreSQL on localhost:5432..."
    if test_postgres; then
      ok "PostgreSQL is reachable (user: magec, db: magec)"
      sleep 1
      return
    fi
    retry=true
  done
}

setup_ollama() {
  local has_docker=false
  check_cmd docker && docker info &>/dev/null 2>&1 && has_docker=true

  local need_llm=false
  [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]] && need_llm=true

  local retry=false
  while true; do
    cls
    echo
    printf "  $(badge " SETUP " "$BG_YELLOW" "$FG_BLACK")  ${BOLD}Local AI engine (Ollama)${NC}\n"
    printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
    echo

    box_top
    box_empty
    if $need_llm; then
      box_line "  Magec uses Ollama to run AI models locally."
    else
      box_line "  Long-term memory needs Ollama for embeddings."
    fi
    box_line "  Ollama must be running on this machine."
    box_empty
    box_sep
    box_empty
    box_line "  ${BOLD}Connection:${NC}  ${CYAN}http://localhost:11434${NC}"
    box_empty
    box_sep
    box_empty

    if $has_docker; then
      box_line "  ${BOLD}Option A — run with Docker:${NC}"
      box_empty
      box_line "  ${CYAN}docker run -d -p 11434:11434 --name magec-ollama \\${NC}"
      box_line "  ${CYAN}  -v ollama_data:/root/.ollama \\${NC}"
      box_line "  ${CYAN}  ollama/ollama:latest${NC}"
      box_empty
      box_sep
      box_empty
      box_line "  ${BOLD}Option B — install natively:${NC}"
      box_empty
      box_line "  ${CYAN}https://ollama.com${NC}"
    else
      box_line "  ${BOLD}Install Ollama:${NC}"
      box_empty
      box_line "  ${CYAN}https://ollama.com${NC}"
      box_empty
      box_sep
      box_empty
      box_line "  ${DIM}Or install Docker and run:${NC}"
      box_empty
      box_line "  ${DIM}docker run -d -p 11434:11434 --name magec-ollama \\${NC}"
      box_line "  ${DIM}  -v ollama_data:/root/.ollama ollama/ollama:latest${NC}"
    fi
    box_empty
    box_bottom
    echo

    info "Open another terminal, install/start Ollama,"
    info "then come back here."
    echo

    if $retry; then
      local msg="Could not connect to Ollama on localhost:11434"
      local msg2="Make sure Ollama is running and try again."
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" ""
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" "$msg2"
      printf "  ${BG_YELLOW}${FG_BLACK}${BOLD}  %-$(( BOX_W - 2 ))s  ${NC}\n" ""
      echo
    fi

    echo
    choose \
      "Test connection" \
      "Skip — I'll set it up later"

    if [[ "$REPLY" == "2" ]]; then
      info "Skipped — remember to start Ollama before running Magec"
      return 0
    fi

    echo
    info "Testing Ollama on localhost:11434..."
    if test_ollama; then
      ok "Ollama is reachable"
      sleep 1

      echo
      info "Pulling required models (this may take a while)..."
      echo

      if $need_llm; then
        info "Downloading qwen3:8b (LLM)..."
        if check_cmd ollama; then
          ollama pull qwen3:8b
        else
          curl -fsSL http://localhost:11434/api/pull -d '{"name":"qwen3:8b"}' -o /dev/null
        fi
        ok "qwen3:8b ready"
      fi

      if $WANT_POSTGRES; then
        info "Downloading nomic-embed-text (embeddings)..."
        if check_cmd ollama; then
          ollama pull nomic-embed-text
        else
          curl -fsSL http://localhost:11434/api/pull -d '{"name":"nomic-embed-text"}' -o /dev/null
        fi
        ok "nomic-embed-text ready"
      fi

      echo
      ok "All models ready"
      sleep 1
      return 0
    fi
    retry=true
  done
}

# ── Create install directory ────────────────────────────────────────────────

mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# ═══════════════════════════════════════════════════════════════════════════
#  BINARY INSTALLATION
# ═══════════════════════════════════════════════════════════════════════════

install_binary() {

  # ── Services setup ──────────────────────────────────────────────────

  local has_docker=false
  check_cmd docker && docker info &>/dev/null 2>&1 && has_docker=true

  if $WANT_REDIS; then
    setup_redis
  fi

  if $WANT_POSTGRES; then
    setup_postgres
  fi

  local need_ollama=false
  [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]] && need_ollama=true
  $WANT_POSTGRES && need_ollama=true

  if $need_ollama; then
    setup_ollama
  fi

  if [[ "$WANT_VOICE" == true ]]; then
    cls
    echo
    printf "  $(badge " SETUP " "$BG_YELLOW" "$FG_BLACK")  ${BOLD}Voice services${NC}\n"
    printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
    echo

    box_top
    box_empty
    box_line "  Voice needs two services: speech-to-text (STT)"
    box_line "  and text-to-speech (TTS). Both are available as"
    box_line "  Docker containers."
    box_empty

    if $has_docker; then
      box_sep
      box_empty
      box_line "  Run these commands:"
      box_empty
      box_line "  ${CYAN}docker run -d -p 5092:5092 --name magec-stt \\${NC}"
      box_line "  ${CYAN}  ghcr.io/achetronic/parakeet:latest${NC}"
      box_empty
      box_line "  ${CYAN}docker run -d -p 5050:5050 --name magec-tts \\${NC}"
      box_line "  ${CYAN}  -e REQUIRE_API_KEY=False \\${NC}"
      box_line "  ${CYAN}  travisvn/openai-edge-tts:latest${NC}"
    else
      box_sep
      box_empty
      box_line "  ${YELLOW}Docker is not installed.${NC} Install it first:"
      box_empty
      case "$OS" in
        linux)  box_line "  ${CYAN}https://docs.docker.com/engine/install/${NC}" ;;
        darwin) box_line "  ${CYAN}https://docs.docker.com/desktop/install/mac-install/${NC}" ;;
        windows) box_line "  ${CYAN}https://docs.docker.com/desktop/install/windows-install/${NC}" ;;
      esac
      box_empty
      box_line "  Then run:"
      box_empty
      box_line "  ${CYAN}docker run -d -p 5092:5092 ghcr.io/achetronic/parakeet:latest${NC}"
      box_line "  ${CYAN}docker run -d -p 5050:5050 -e REQUIRE_API_KEY=False \\${NC}"
      box_line "  ${CYAN}  travisvn/openai-edge-tts:latest${NC}"
    fi
    box_empty
    box_bottom
    echo

    printf "  ${DIM}Press Enter to continue...${NC}"
    read -r < /dev/tty
  fi

  # ── Download binary ────────────────────────────────────────────────

  cls
  echo
  printf "  $(badge " INSTALLING " "$BG_CYAN" "$FG_BLACK")\n"
  printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
  echo

  require_cmd curl "Install it with your package manager."

  local ext="tar.gz"
  [[ "$OS" == "windows" ]] && ext="zip"
  local asset_name="magec-${OS}-${ARCH}.${ext}"

  local supported=false
  case "${OS}-${ARCH}" in
    linux-amd64|linux-arm64|darwin-arm64|windows-amd64) supported=true ;;
  esac

  if [[ "$supported" != true ]]; then
    echo
    box_top
    box_empty
    box_line "  ${RED}No pre-built version for ${OS}/${ARCH}${NC}" "" "center"
    box_empty
    box_line "  Available platforms:"
    box_line "    Linux (x86_64, ARM64)"
    box_line "    macOS (Apple Silicon)"
    box_line "    Windows (x86_64)"
    box_empty
    box_line "  You can build from source instead:"
    box_line "  ${DIM}https://github.com/${REPO}#building-from-source${NC}"
    box_empty
    box_bottom
    exit 1
  fi

  # ── Find latest release ──────────────────────────────────────────────

  info "Checking for the latest version..."
  local release_json
  release_json="$(curl -fsSL "${API_URL}/releases/latest")" || die "Could not reach GitHub. Check your internet connection."

  local tag
  tag="$(echo "$release_json" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"\([^"]*\)".*/\1/')"
  [[ -z "$tag" ]] && die "Could not determine latest version."
  ok "Latest version: ${BOLD}${tag}${NC}"

  local download_url
  download_url="$(echo "$release_json" | grep '"browser_download_url"' | grep "$asset_name" | head -1 | sed 's/.*"browser_download_url": *"\([^"]*\)".*/\1/')"
  [[ -z "$download_url" ]] && die "Download not found for ${asset_name} in release ${tag}"

  # ── Download ─────────────────────────────────────────────────────────

  info "Downloading Magec ${tag} for ${os_label}..."
  curl -fsSL "$download_url" -o "$asset_name" || die "Download failed. Check your internet connection."

  info "Unpacking..."
  if [[ "$ext" == "zip" ]]; then
    require_cmd unzip "Install it with your package manager."
    unzip -qo "$asset_name"
  else
    tar xzf "$asset_name"
  fi
  rm -f "$asset_name"

  [[ "$OS" != "windows" ]] && chmod +x magec

  ok "Magec downloaded to ${BOLD}$(pwd)/magec${NC}"

  # ── ONNX Runtime (voice engine) ─────────────────────────────────────

  if [[ "$WANT_ONNX" == true ]]; then
    install_onnx_runtime
  fi

  # ── CLIProxyAPI setup (binary) ───────────────────────────────────────

  if $WANT_CLIPROXYAPI; then
    cls
    echo
    printf "  $(badge " SETUP " "$BG_YELLOW" "$FG_BLACK")  ${BOLD}Claude subscription proxy (CLIProxyAPI)${NC}\n"
    printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
    echo

    box_top
    box_empty
    box_line "  CLIProxyAPI runs as a separate service."
    box_line "  You can run it via Docker or download the"
    box_line "  binary from GitHub."
    box_empty
    box_sep
    box_empty
    box_line "  ${BOLD}Option A — Docker (recommended):${NC}"
    box_empty
    box_line "  ${CYAN}docker run -d -p 8317:8317 -p 54545:54545 \\${NC}"
    box_line "  ${CYAN}  -v \$(pwd)/cliproxyapi/config.yaml:\\${NC}"
    box_line "  ${CYAN}  /CLIProxyAPI/config/config.yaml \\${NC}"
    box_line "  ${CYAN}  -v cliproxyapi_auth:/CLIProxyAPI/auth \\${NC}"
    box_line "  ${CYAN}  --name magec-cliproxyapi \\${NC}"
    box_line "  ${CYAN}  eceasy/cli-proxy-api:latest${NC}"
    box_empty
    box_sep
    box_empty
    box_line "  ${BOLD}Option B — Binary:${NC}"
    box_empty
    box_line "  ${CYAN}https://github.com/router-for-me/CLIProxyAPI/releases${NC}"
    box_empty
    box_sep
    box_empty
    box_line "  ${BOLD}After starting, log in with your Claude account:${NC}"
    box_empty
    box_line "  ${CYAN}cliproxyapi --claude-login${NC}"
    box_line "  ${DIM}(or: docker exec magec-cliproxyapi \\${NC}"
    box_line "  ${DIM}  /CLIProxyAPI/CLIProxyAPI --no-browser --claude-login)${NC}"
    box_empty
    box_bottom
    echo

    printf "  ${DIM}Press Enter to continue...${NC}"
    read -r < /dev/tty

    generate_cliproxyapi_config
  fi

  # ── Generate config ─────────────────────────────────────────────────

  echo
  info "Creating configuration files..."
  if ! $SKIP_CONFIG; then
    generate_config_yaml
  else
    ok "config.yaml (kept existing)"
  fi
  if ! $SKIP_DATA; then
    generate_store_json
  else
    ok "data/store.json (kept existing)"
  fi

  print_success
}

install_onnx_runtime() {
  echo
  info "Downloading the voice engine..."

  local onnx_url="" onnx_archive="" onnx_dir="" lib_name=""

  case "${OS}-${ARCH}" in
    linux-amd64)
      onnx_archive="onnxruntime-linux-x64-${ONNX_VERSION}.tgz"
      onnx_url="https://github.com/microsoft/onnxruntime/releases/download/v${ONNX_VERSION}/${onnx_archive}"
      onnx_dir="onnxruntime-linux-x64-${ONNX_VERSION}"
      lib_name="libonnxruntime.so"
      ;;
    linux-arm64)
      onnx_archive="onnxruntime-linux-aarch64-${ONNX_VERSION}.tgz"
      onnx_url="https://github.com/microsoft/onnxruntime/releases/download/v${ONNX_VERSION}/${onnx_archive}"
      onnx_dir="onnxruntime-linux-aarch64-${ONNX_VERSION}"
      lib_name="libonnxruntime.so"
      ;;
    darwin-arm64)
      onnx_archive="onnxruntime-osx-universal2-${ONNX_VERSION}.tgz"
      onnx_url="https://github.com/microsoft/onnxruntime/releases/download/v${ONNX_VERSION}/${onnx_archive}"
      onnx_dir="onnxruntime-osx-universal2-${ONNX_VERSION}"
      lib_name="libonnxruntime.dylib"
      ;;
    windows-amd64)
      warn "Automatic voice engine download is not available on Windows."
      info "Download manually: https://github.com/microsoft/onnxruntime/releases/tag/v${ONNX_VERSION}"
      return
      ;;
    *)
      warn "No voice engine available for ${os_label} (${arch_label})."
      return
      ;;
  esac

  curl -fsSL "$onnx_url" -o "$onnx_archive" || { warn "Download failed. You can install the voice engine manually later."; return; }

  tar xzf "$onnx_archive"
  rm -f "$onnx_archive"

  ONNX_LIB_PATH="$(pwd)/${onnx_dir}/lib/${lib_name}"

  if [[ -f "$ONNX_LIB_PATH" ]]; then
    ok "Voice engine ready"
  else
    warn "Could not find the voice engine library. You may need to configure it manually in config.yaml."
  fi
}

# ═══════════════════════════════════════════════════════════════════════════
#  CONTAINER INSTALLATION
# ═══════════════════════════════════════════════════════════════════════════

install_containers() {
  cls
  echo
  printf "  $(badge " INSTALLING " "$BG_CYAN" "$FG_BLACK")\n"
  printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
  echo

  # ── Check Docker ────────────────────────────────────────────────────

  info "Checking for Docker..."

  if ! check_cmd docker; then
    echo
    box_top
    box_empty
    box_line "  ${RED}Docker is not installed${NC}" "" "center"
    box_empty
    box_line "  Docker is needed to run Magec in containers."
    box_line "  Install it and then run this script again:"
    box_empty
    case "$OS" in
      linux)  box_line "  ${CYAN}https://docs.docker.com/engine/install/${NC}" ;;
      darwin) box_line "  ${CYAN}https://docs.docker.com/desktop/install/mac-install/${NC}" ;;
      windows) box_line "  ${CYAN}https://docs.docker.com/desktop/install/windows-install/${NC}" ;;
    esac
    box_empty
    box_bottom
    exit 1
  fi

  if ! docker info &>/dev/null; then
    die "Docker is installed but not running. Start Docker and try again."
  fi

  if ! docker compose version &>/dev/null && ! docker-compose version &>/dev/null; then
    die "Docker Compose is required. Install it from https://docs.docker.com/compose/install/"
  fi

  ok "Docker is ready"

  if $GPU; then
    if ! docker info 2>/dev/null | grep -qi 'nvidia'; then
      echo
      box_top
      box_empty
      box_line "  ${YELLOW}NVIDIA toolkit not detected${NC}" "" "center"
      box_empty
      box_line "  GPU acceleration was requested but the NVIDIA"
      box_line "  Container Toolkit doesn't seem to be installed."
      box_empty
      box_line "  Install it from:"
      box_line "  ${CYAN}https://docs.nvidia.com/datacenter/cloud-native/${NC}"
      box_line "  ${CYAN}container-toolkit/install-guide.html${NC}"
      box_empty
      box_line "  Then restart Docker and run this script again."
      box_empty
      box_bottom
      exit 1
    fi
    ok "NVIDIA GPU detected"
  fi

  if docker compose version &>/dev/null; then
    COMPOSE="docker compose"
  else
    COMPOSE="docker-compose"
  fi

  # ── Generate files ──────────────────────────────────────────────────

  info "Creating configuration files..."
  generate_docker_compose
  if $WANT_CLIPROXYAPI; then
    generate_cliproxyapi_config
  fi
  if ! $SKIP_CONFIG; then
    generate_config_yaml
  else
    ok "config.yaml (kept existing)"
  fi
  if ! $SKIP_DATA; then
    generate_store_json
  else
    ok "data/store.json (kept existing)"
  fi

  ok "All files created"

  # ── Launch ──────────────────────────────────────────────────────────

  echo
  info "Starting Magec (this may take a moment)..."
  echo
  $COMPOSE up -d

  print_success
}

# ═══════════════════════════════════════════════════════════════════════════
#  CONFIGURATION GENERATORS
# ═══════════════════════════════════════════════════════════════════════════

generate_cliproxyapi_config() {
  mkdir -p cliproxyapi
  cat > cliproxyapi/config.yaml <<'CLIPROXY'
host: "0.0.0.0"
port: 8317
auth-dir: "/CLIProxyAPI/auth"
api-keys:
  - "sk-magec-local"
CLIPROXY
  ok "cliproxyapi/config.yaml"
}

generate_config_yaml() {
  local voice_enabled="true"
  [[ "$WANT_VOICE" != true ]] && voice_enabled="false"

  local onnx_line=""
  if [[ "$INSTALL_METHOD" == "1" && "$WANT_ONNX" == true && -n "${ONNX_LIB_PATH:-}" ]]; then
    onnx_line="  onnxLibraryPath: ${ONNX_LIB_PATH}"
  fi

  local admin_pass_value="$(yaml_quote '${MAGEC_ADMIN_PASSWORD}')"
  [[ -n "$ADMIN_PASSWORD" ]] && admin_pass_value="$(yaml_quote "$ADMIN_PASSWORD")"

  local enc_key_value="$(yaml_quote '${MAGEC_ENCRYPTION_KEY}')"
  [[ -n "$ENCRYPTION_KEY" ]] && enc_key_value="$(yaml_quote "$ENCRYPTION_KEY")"

  cat > config.yaml <<YAML
server:
  host: 0.0.0.0
  port: 8080
  adminPort: 8081
  adminPassword: ${admin_pass_value}
  encryptionKey: ${enc_key_value}

voice:
  ui:
    enabled: ${voice_enabled}
${onnx_line:+${onnx_line}
}
log:
  level: info
  format: console
YAML

  ok "config.yaml"
}

generate_store_json() {
  mkdir -p data

  local backends="[]"
  local memory_providers="[]"
  local settings='{}'

  local backend_entries=()
  local ollama_backend_id="$(gen_uuid)"
  local openai_backend_id="$(gen_uuid)"
  local anthropic_backend_id="$(gen_uuid)"
  local gemini_backend_id="$(gen_uuid)"
  local cliproxyapi_backend_id="$(gen_uuid)"
  local parakeet_backend_id="$(gen_uuid)"
  local tts_backend_id="$(gen_uuid)"

  if [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]] || $WANT_POSTGRES; then
    local ollama_url="http://localhost:11434/v1"
    [[ "$INSTALL_METHOD" == "2" ]] && ollama_url="http://ollama:11434/v1"
    backend_entries+=("{\"id\":\"${ollama_backend_id}\",\"name\":\"Ollama\",\"type\":\"openai\",\"url\":\"${ollama_url}\",\"apiKey\":\"\"}")
  fi

  if [[ "$LLM_CHOICE" == "2" || "$LLM_CHOICE" == "3" ]]; then
    backend_entries+=("{\"id\":\"${openai_backend_id}\",\"name\":\"OpenAI\",\"type\":\"openai\",\"url\":\"https://api.openai.com/v1\",\"apiKey\":\"\"}")
    backend_entries+=("{\"id\":\"${anthropic_backend_id}\",\"name\":\"Anthropic\",\"type\":\"anthropic\",\"url\":\"\",\"apiKey\":\"\"}")
    backend_entries+=("{\"id\":\"${gemini_backend_id}\",\"name\":\"Gemini\",\"type\":\"gemini\",\"url\":\"\",\"apiKey\":\"\"}")
  fi

  if $WANT_CLIPROXYAPI; then
    local cliproxyapi_url="http://localhost:8317"
    [[ "$INSTALL_METHOD" == "2" ]] && cliproxyapi_url="http://cliproxyapi:8317"
    backend_entries+=("{\"id\":\"${cliproxyapi_backend_id}\",\"name\":\"Claude (Subscription)\",\"type\":\"anthropic\",\"url\":\"${cliproxyapi_url}\",\"apiKey\":\"sk-magec-local\"}")
  fi

  if [[ "$WANT_VOICE" == true ]]; then
    local parakeet_url="http://localhost:5092"
    local tts_url="http://localhost:5050"
    if [[ "$INSTALL_METHOD" == "2" ]]; then
      parakeet_url="http://parakeet:5092"
      tts_url="http://tts:5050"
    fi
    backend_entries+=("{\"id\":\"${parakeet_backend_id}\",\"name\":\"Parakeet (STT)\",\"type\":\"openai\",\"url\":\"${parakeet_url}\",\"apiKey\":\"\"}")
    backend_entries+=("{\"id\":\"${tts_backend_id}\",\"name\":\"Edge TTS\",\"type\":\"openai\",\"url\":\"${tts_url}\",\"apiKey\":\"\"}")
  fi

  if [[ ${#backend_entries[@]} -gt 0 ]]; then
    backends="[$(IFS=,; echo "${backend_entries[*]}")]"
  fi

  local memory_entries=()
  local redis_id="$(gen_uuid)"
  local postgres_id="$(gen_uuid)"

  if $WANT_REDIS; then
    local redis_url="redis://localhost:6379/0"
    [[ "$INSTALL_METHOD" == "2" ]] && redis_url="redis://redis:6379/0"
    memory_entries+=("{\"id\":\"${redis_id}\",\"name\":\"Redis\",\"type\":\"redis\",\"category\":\"session\",\"config\":{\"connectionString\":\"${redis_url}\",\"ttl\":\"24h\"}}")
    settings="{\"sessionProvider\":\"${redis_id}\"}"
  fi

  if $WANT_POSTGRES; then
    local pg_url="postgres://magec:magec@localhost:5432/magec?sslmode=disable"
    [[ "$INSTALL_METHOD" == "2" ]] && pg_url="postgres://magec:magec@postgres:5432/magec?sslmode=disable"
    local embedding_ref="{\"backend\":\"${ollama_backend_id}\",\"model\":\"nomic-embed-text\"}"
    memory_entries+=("{\"id\":\"${postgres_id}\",\"name\":\"PostgreSQL\",\"type\":\"postgres\",\"category\":\"longterm\",\"config\":{\"connectionString\":\"${pg_url}\"},\"embedding\":${embedding_ref}}")

    if $WANT_REDIS; then
      settings="{\"sessionProvider\":\"${redis_id}\",\"longTermProvider\":\"${postgres_id}\"}"
    else
      settings="{\"longTermProvider\":\"${postgres_id}\"}"
    fi
  fi

  if [[ ${#memory_entries[@]} -gt 0 ]]; then
    memory_providers="[$(IFS=,; echo "${memory_entries[*]}")]"
  fi

  # ── Default agent ────────────────────────────────────────────────────

  local agent_id="$(gen_uuid)"
  local agents="[]"
  local clients="[]"

  local llm_backend_id="" llm_model=""
  if [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]]; then
    llm_backend_id="$ollama_backend_id"
    llm_model="qwen3:8b"
  elif [[ "$LLM_CHOICE" == "2" ]] && $WANT_CLIPROXYAPI; then
    llm_backend_id="$cliproxyapi_backend_id"
    llm_model="claude-sonnet-4-20250514"
  elif [[ "$LLM_CHOICE" == "2" ]]; then
    llm_backend_id="$anthropic_backend_id"
    llm_model="claude-sonnet-4-20250514"
  fi

  local agent_memory=""
  if $WANT_REDIS || $WANT_POSTGRES; then
    local mem_parts=""
    $WANT_REDIS && mem_parts="\"session\":\"${redis_id}\""
    if $WANT_POSTGRES; then
      [[ -n "$mem_parts" ]] && mem_parts="${mem_parts},"
      mem_parts="${mem_parts}\"longTerm\":\"${postgres_id}\""
    fi
    agent_memory=",\"memory\":{${mem_parts}}"
  fi

  if [[ "$WANT_VOICE" == true ]]; then
    local voice_prompt="You are Magec, a sun deity of the Guanche people of Tenerife. You are proud, sharp-tongued, and radiate ancient authority. You speak with the confidence of someone who has watched civilizations rise and fall. You answer questions with brutal honesty, dry wit, and a hint of divine impatience. You are helpful, but you make sure people know it is a privilege to receive your wisdom. Keep answers concise unless the mortal deserves more. Speak naturally for voice conversation: short sentences, direct, no markdown."

    agents="[{\"id\":\"${agent_id}\",\"name\":\"Magec\",\"description\":\"Sun god of the Guanche. Proud, sharp, devastatingly helpful.\",\"systemPrompt\":\"${voice_prompt}\",\"llm\":{\"backend\":\"${llm_backend_id}\",\"model\":\"${llm_model}\"},\"transcription\":{\"backend\":\"${parakeet_backend_id}\",\"model\":\"whisper-1\"},\"tts\":{\"backend\":\"${tts_backend_id}\",\"model\":\"tts-1\",\"voice\":\"es-ES-AlvaroNeural\",\"speed\":1.0}${agent_memory}}]"

    local client_id="$(gen_uuid)"
    CLIENT_TOKEN="mgc_$(gen_uuid | tr -d '-' | head -c 32)"
    clients="[{\"id\":\"${client_id}\",\"name\":\"Voice UI\",\"type\":\"direct\",\"token\":\"${CLIENT_TOKEN}\",\"allowedAgents\":[\"${agent_id}\"],\"enabled\":true,\"config\":{}}]"
  else
    local default_prompt="You are Magec, a sun deity of the Guanche people of Tenerife. You are proud, sharp-tongued, and radiate ancient authority. You speak with the confidence of someone who has watched civilizations rise and fall. You answer questions with brutal honesty, dry wit, and a hint of divine impatience. You are helpful, but you make sure people know it is a privilege to receive your wisdom. Keep answers concise unless the mortal deserves more."

    agents="[{\"id\":\"${agent_id}\",\"name\":\"Magec\",\"description\":\"Sun god of the Guanche. Proud, sharp, devastatingly helpful.\",\"systemPrompt\":\"${default_prompt}\",\"llm\":{\"backend\":\"${llm_backend_id}\",\"model\":\"${llm_model}\"}${agent_memory}}]"
  fi

  cat > data/store.json <<JSON
{
  "settings": ${settings},
  "backends": ${backends},
  "memoryProviders": ${memory_providers},
  "mcpServers": [],
  "agents": ${agents},
  "clients": ${clients},
  "flows": [],
  "commands": [],
  "secrets": []
}
JSON

  ok "data/store.json"
}

generate_docker_compose() {
  local services=""
  local volumes=""

  local depends_entries=()

  local need_ollama=false
  [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]] && need_ollama=true
  $WANT_POSTGRES && need_ollama=true

  if $need_ollama; then
    depends_entries+=("      ollama-setup:\n        condition: service_completed_successfully")
  fi
  $WANT_REDIS && depends_entries+=("      redis:\n        condition: service_started")
  $WANT_POSTGRES && depends_entries+=("      postgres:\n        condition: service_started")

  if [[ "$WANT_VOICE" == true ]]; then
    depends_entries+=("      parakeet:\n        condition: service_started")
    depends_entries+=("      tts:\n        condition: service_started")
  fi

  local magec_depends_block=""
  if [[ ${#depends_entries[@]} -gt 0 ]]; then
    magec_depends_block="    depends_on:\n$(printf '%b\n' "${depends_entries[@]}")"
  fi

  services+="  magec:\n"
  services+="    image: ghcr.io/achetronic/magec:latest\n"
  services+="    ports:\n"
  services+="      - \"8080:8080\"\n"
  services+="      - \"8081:8081\"\n"
  services+="    environment:\n"
  services+="      MAGEC_ADMIN_PASSWORD: \${MAGEC_ADMIN_PASSWORD:-}\n"
  services+="      MAGEC_ENCRYPTION_KEY: \${MAGEC_ENCRYPTION_KEY:-}\n"
  services+="    volumes:\n"
  services+="      - ./config.yaml:/app/config.yaml\n"
  services+="      - ./data:/app/data\n"
  [[ -n "$magec_depends_block" ]] && services+="${magec_depends_block}\n"
  services+="    restart: unless-stopped\n"

  volumes+="  magec_data:\n"

  if $WANT_REDIS; then
    services+="\n  redis:\n"
    services+="    image: redis:alpine\n"
    services+="    volumes:\n"
    services+="      - redis_data:/data\n"
    services+="    restart: unless-stopped\n"
    volumes+="  redis_data:\n"
  fi

  if $WANT_POSTGRES; then
    services+="\n  postgres:\n"
    services+="    image: pgvector/pgvector:pg17\n"
    services+="    environment:\n"
    services+="      POSTGRES_USER: magec\n"
    services+="      POSTGRES_PASSWORD: magec\n"
    services+="      POSTGRES_DB: magec\n"
    services+="    volumes:\n"
    services+="      - postgres_data:/var/lib/postgresql/data\n"
    services+="    restart: unless-stopped\n"
    volumes+="  postgres_data:\n"
  fi

  if $need_ollama; then
    services+="\n  ollama:\n"
    services+="    image: ollama/ollama:latest\n"
    services+="    volumes:\n"
    services+="      - ollama_data:/root/.ollama\n"
    services+="    restart: unless-stopped\n"

    if $GPU; then
      services+="    deploy:\n"
      services+="      resources:\n"
      services+="        reservations:\n"
      services+="          devices:\n"
      services+="            - driver: nvidia\n"
      services+="              count: all\n"
      services+="              capabilities: [gpu]\n"
    fi

    local models_to_pull=""
    if [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]]; then
      models_to_pull+="        echo \"Pulling qwen3:8b (LLM)...\"\n        ollama pull qwen3:8b\n"
    fi
    if $WANT_POSTGRES; then
      models_to_pull+="        echo \"Pulling nomic-embed-text (embeddings)...\"\n        ollama pull nomic-embed-text\n"
    fi

    services+="\n  ollama-setup:\n"
    services+="    image: ollama/ollama:latest\n"
    services+="    depends_on:\n"
    services+="      - ollama\n"
    services+="    restart: \"no\"\n"
    services+="    environment:\n"
    services+="      OLLAMA_HOST: http://ollama:11434\n"
    services+="    entrypoint: [\"\"]\n"
    services+="    command:\n"
    services+="      - /bin/sh\n"
    services+="      - -c\n"
    services+="      - |\n"
    services+="        echo \"Waiting for Ollama to be ready...\"\n"
    services+="        until ollama list > /dev/null 2>&1; do\n"
    services+="          sleep 2\n"
    services+="        done\n"
    services+="${models_to_pull}"
    services+="        echo \"Models ready.\"\n"

    volumes+="  ollama_data:\n"
  fi

  if [[ "$WANT_VOICE" == true ]]; then
    services+="\n  parakeet:\n"
    services+="    image: ghcr.io/achetronic/parakeet:latest\n"
    services+="    restart: unless-stopped\n"

    services+="\n  tts:\n"
    services+="    image: travisvn/openai-edge-tts:latest\n"
    services+="    environment:\n"
    services+="      - REQUIRE_API_KEY=False\n"
    services+="    restart: unless-stopped\n"
  fi

  if $WANT_CLIPROXYAPI; then
    services+="\n  cliproxyapi:\n"
    services+="    image: eceasy/cli-proxy-api:latest\n"
    services+="    ports:\n"
    services+="      - \"54545:54545\"\n"
    services+="    volumes:\n"
    services+="      - ./cliproxyapi/config.yaml:/CLIProxyAPI/config/config.yaml\n"
    services+="      - cliproxyapi_auth:/CLIProxyAPI/auth\n"
    services+="    restart: unless-stopped\n"
    volumes+="  cliproxyapi_auth:\n"
  fi

  printf "services:\n" > docker-compose.yaml
  printf '%b' "$services" >> docker-compose.yaml
  printf "\nvolumes:\n" >> docker-compose.yaml
  printf '%b' "$volumes" >> docker-compose.yaml

  ok "docker-compose.yaml"
}

# ═══════════════════════════════════════════════════════════════════════════
#  SUCCESS
# ═══════════════════════════════════════════════════════════════════════════

print_success() {
  cls
  local binary_cmd=""
  if [[ "$INSTALL_METHOD" == "1" ]]; then
    local bin="./magec"
    [[ "$OS" == "windows" ]] && bin="magec.exe"
    binary_cmd="cd ${INSTALL_DIR} && ${bin} --config config.yaml"
  fi

  local C=$'\033[38;5;203m'
  local M=$'\033[38;5;209m'
  local A=$'\033[38;5;214m'
  local W=$'\033[38;5;255m'

  echo
  echo
  printf "                    ${M}▄▄████${C}████▄▄${NC}\n"
  printf "               ${M}▄████████${C}████████████▄${NC}\n"
  printf "            ${A}▄██████${M}██████████${C}██████████▄${NC}\n"
  printf "          ${A}▄████████████${M}██████${C}██${W}██${C}████████▄${NC}\n"
  printf "         ${A}█████████${W}██${A}██${M}████████${C}█████████████${NC}\n"
  printf "         ${A}██████████████${M}████${W}███${M}████${C}█████████${NC}\n"
  printf "         ${A}██████████████████${M}████${W}█${C}███████████${NC}\n"
  printf "          ${A}▀██████${W}█${A}██████${M}████████${C}█████████▀${NC}\n"
  printf "            ${A}▀████████████${W}██${M}████████████▀${NC}\n"
  printf "               ${A}▀████████████${M}████████▀${NC}\n"
  printf "                    ${A}▀▀████${M}████▀▀${NC}\n"
  printf "${NC}\n"
  printf "${GREEN}${BOLD}"
  cat <<'EOF'
          __  __
         |  \/  | __ _  __ _  ___  ___
         | |\/| |/ _` |/ _` |/ _ \/ __|
         | |  | | (_| | (_| |  __/ (__
         |_|  |_|\__,_|\__, |\___|\___|
                       |___/
EOF
  printf "${NC}\n"

  printf "  $(badge " INSTALLED " "$BG_GREEN" "$FG_BLACK")  ${GREEN}${BOLD}Magec is ready!${NC}\n"
  printf "  ${DIM}$(hline '─' "$BOX_W")${NC}\n"
  echo

  box_top
  box_empty
  box_line "  ${BOLD}Admin Panel${NC}  →  ${CYAN}http://localhost:8081${NC}"
  box_empty

  if [[ "$WANT_VOICE" == true ]]; then
    box_line "  ${BOLD}Voice Chat${NC}   →  ${CYAN}http://localhost:8080${NC}"
    box_empty
  fi

  if $SKIP_CONFIG; then
    box_sep
    box_empty
    box_line "  ${GREEN}Config preserved${NC} — your admin password and"
    box_line "  encryption key are unchanged in config.yaml."
    box_empty
  elif [[ -n "$ADMIN_PASSWORD" ]]; then
    box_sep
    box_empty
    box_line "  ${BOLD}Admin password:${NC}  ${CYAN}${ADMIN_PASSWORD}${NC}"
    box_empty
    box_line "  ${DIM}You'll need this to log into the Admin Panel.${NC}"
    box_line "  ${DIM}Saved in config.yaml → server.adminPassword${NC}"
    box_empty
  else
    box_sep
    box_empty
    box_line "  ${YELLOW}No admin password set.${NC} The panel is open."
    box_line "  ${DIM}Set one later in config.yaml or via env var:${NC}"
    box_line "  ${CYAN}MAGEC_ADMIN_PASSWORD=yourpassword${NC}"
    box_empty
  fi

  if ! $SKIP_CONFIG; then
    if [[ -n "$ENCRYPTION_KEY" ]]; then
      box_sep
      box_empty
      box_line "  ${BOLD}Encryption key:${NC}  ${CYAN}${ENCRYPTION_KEY}${NC}"
      box_empty
      box_line "  ${DIM}Used to encrypt API keys and secrets on disk.${NC}"
      box_line "  ${DIM}Saved in config.yaml → server.encryptionKey${NC}"
      box_empty
    else
      box_sep
      box_empty
      box_line "  ${YELLOW}No encryption key set.${NC} Secrets stored in plain text."
      box_line "  ${DIM}Set one later in config.yaml or via env var:${NC}"
      box_line "  ${CYAN}MAGEC_ENCRYPTION_KEY=yourkey${NC}"
      box_empty
    fi
  fi

  if [[ "$INSTALL_METHOD" == "1" ]]; then
    box_sep
    box_empty
    box_line "  ${BOLD}Start Magec:${NC}"
    box_empty
    box_line "  ${CYAN}${binary_cmd}${NC}"
    box_empty
  fi

  box_sep
  box_empty
  box_line "  ${BOLD}Getting started:${NC}"
  box_empty
  box_line "  A default agent (${BOLD}Magec${NC}) has been created for you."
  if [[ "$WANT_VOICE" == true ]]; then
    box_empty
    box_line "  1. Open the Voice Chat in your browser"
    box_line "  2. Enter this pairing token when prompted:"
    box_empty
    box_line "     ${CYAN}${CLIENT_TOKEN:-see data/store.json}${NC}"
    box_empty
    box_line "  3. Say \"Oye Magec\" and start talking!"
  else
    box_empty
    box_line "  1. Open the Admin Panel in your browser"
    box_line "  2. Customize or create more agents"
    box_line "  3. Start chatting!"
  fi
  box_empty

  if $WANT_CLIPROXYAPI; then
    box_sep
    box_empty
    box_line "  ${YELLOW}Action needed:${NC} Log in to your Claude account."
    if [[ "$INSTALL_METHOD" == "2" ]]; then
      box_line "  ${CYAN}${COMPOSE} exec cliproxyapi \\${NC}"
      box_line "  ${CYAN}  /CLIProxyAPI/CLIProxyAPI --no-browser --claude-login${NC}"
    else
      box_line "  ${CYAN}cliproxyapi --claude-login${NC}"
      box_line "  ${DIM}(or via Docker: docker exec magec-cliproxyapi \\${NC}"
      box_line "  ${DIM}  /CLIProxyAPI/CLIProxyAPI --no-browser --claude-login)${NC}"
    fi
    box_empty
    box_line "  A \"${BOLD}Claude (Subscription)${NC}\" backend has been"
    box_line "  pre-configured. After login, it's ready to use."
    box_empty
  fi

  if [[ "$LLM_CHOICE" == "2" || "$LLM_CHOICE" == "3" ]] && ! $WANT_CLIPROXYAPI; then
    box_sep
    box_empty
    box_line "  ${YELLOW}Remember:${NC} Cloud AI providers need API keys."
    box_line "  Add them in Admin Panel → Backends."
    box_empty
  elif [[ "$LLM_CHOICE" == "2" || "$LLM_CHOICE" == "3" ]] && $WANT_CLIPROXYAPI; then
    box_sep
    box_empty
    box_line "  ${DIM}Other cloud providers (OpenAI, Gemini) still${NC}"
    box_line "  ${DIM}need API keys — add them in Admin Panel → Backends.${NC}"
    box_empty
  fi

  if [[ "$INSTALL_METHOD" == "2" ]] && { [[ "$LLM_CHOICE" == "1" || "$LLM_CHOICE" == "3" ]] || $WANT_POSTGRES; }; then
    box_sep
    box_empty
    box_line "  ${DIM}The first start downloads AI models.${NC}"
    box_line "  ${DIM}This may take a few minutes on slower connections.${NC}"
    box_line "  ${DIM}Track progress: ${COMPOSE} logs -f ollama-setup${NC}"
    box_empty
  fi

  box_bottom
  echo
  printf "  ${DIM}Thank you for installing Magec${NC}\n"
  printf "  ${DIM}https://github.com/achetronic/magec${NC}\n"
  echo
}

# ═══════════════════════════════════════════════════════════════════════════
#  RUN
# ═══════════════════════════════════════════════════════════════════════════

if [[ "$INSTALL_METHOD" == "1" ]]; then
  install_binary
else
  install_containers
fi
