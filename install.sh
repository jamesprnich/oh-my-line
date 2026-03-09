#!/bin/bash
# ─────────────────────────────────────────────────────────────────────────────
# oh-my-line installer
# https://github.com/jamesprnich/oh-my-line
# ─────────────────────────────────────────────────────────────────────────────
set -e

REPO="jamesprnich/oh-my-line"
BRANCH="main"
BASE_URL="https://raw.githubusercontent.com/${REPO}/${BRANCH}"

INSTALL_DIR="${HOME}/.oh-my-line"
SETTINGS_FILE="${HOME}/.claude/settings.json"

# ── Colours & symbols ────────────────────────────────────────────────────────
RST='\033[0m'
BOLD='\033[1m'
DIM='\033[2m'
GREEN='\033[32m'
YELLOW='\033[33m'
PURPLE='\033[38;2;124;58;237m'
WHITE='\033[37m'

OK="${GREEN}✓${RST}"
FAIL="${YELLOW}✗${RST}"
DOT="${DIM}·${RST}"

# ── Banner ───────────────────────────────────────────────────────────────────
printf '\n'
printf '  %b%b━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%b\n' "$BOLD" "$PURPLE" "$RST"
printf '  %b%b  oh-my-line %b%bInstall Script%b\n' "$BOLD" "$PURPLE" "$RST" "$DIM" "$RST"
printf '  %b%b━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%b\n' "$BOLD" "$PURPLE" "$RST"
printf '\n'

# ── Prerequisites ────────────────────────────────────────────────────────────
printf '  %bChecking pre-requisites%b\n\n' "$WHITE" "$RST"

prereqs_ok=true

if command -v curl >/dev/null 2>&1; then
    printf '    %b  curl\n' "$OK"
else
    printf '    %b  curl %b— not found%b\n' "$FAIL" "$DIM" "$RST"
    prereqs_ok=false
fi

if command -v git >/dev/null 2>&1; then
    printf '    %b  git\n' "$OK"
else
    printf '    %b  git %b(optional — needed for GitHub segments)%b\n' "$DOT" "$DIM" "$RST"
fi

if [ "$prereqs_ok" = false ]; then
    printf '\n  %bInstall missing dependencies and try again.%b\n\n' "$YELLOW" "$RST"
    exit 1
fi

# ── Detect source ────────────────────────────────────────────────────────────
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]:-$0}")" 2>/dev/null && pwd)"

if [ -f "${SCRIPT_DIR}/VERSION" ] && [ -d "${SCRIPT_DIR}/skills" ]; then
    SOURCE="local"
else
    SOURCE="remote"
fi

fetch() {
    local dest="$1" src="$2"
    if [ "$SOURCE" = "local" ]; then
        cp "${SCRIPT_DIR}/${src}" "$dest"
    else
        curl -fsSL "${BASE_URL}/${src}" -o "$dest"
    fi
}

# ── Version detection ────────────────────────────────────────────────────────
if [ "$SOURCE" = "local" ]; then
    NEW_VERSION=$(cat "${SCRIPT_DIR}/VERSION" 2>/dev/null || echo "unknown")
else
    NEW_VERSION=$(curl -fsSL "${BASE_URL}/VERSION" 2>/dev/null || echo "unknown")
fi
NEW_VERSION=$(echo "$NEW_VERSION" | tr -d '[:space:]')

OLD_VERSION=""
if [ -f "${INSTALL_DIR}/VERSION" ]; then
    OLD_VERSION=$(cat "${INSTALL_DIR}/VERSION" 2>/dev/null | tr -d '[:space:]')
fi

# ── Install files ────────────────────────────────────────────────────────────
printf '\n'
if [ -n "$OLD_VERSION" ] && [ "$OLD_VERSION" != "$NEW_VERSION" ]; then
    printf '  %bUpgrading%b  %bv%s%b → %b%bv%s%b\n\n' "$WHITE" "$RST" "$DIM" "$OLD_VERSION" "$RST" "$BOLD" "$GREEN" "$NEW_VERSION" "$RST"
elif [ -n "$OLD_VERSION" ]; then
    printf '  %bReinstalling%b  %bv%s%b\n\n' "$WHITE" "$RST" "$DIM" "$NEW_VERSION" "$RST"
else
    printf '  %bInstalling%b  %b%bv%s%b\n\n' "$WHITE" "$RST" "$BOLD" "$GREEN" "$NEW_VERSION" "$RST"
fi

if [ "$SOURCE" = "remote" ]; then
    printf '    %b  Downloading from %bgithub.com/%s%b\n' "$DOT" "$DIM" "$REPO" "$RST"
fi

mkdir -p "${INSTALL_DIR}"
fetch "${INSTALL_DIR}/VERSION" "VERSION"

# ── Install Go binary ────────────────────────────────────────────────────────
COMMAND_PATH="~/.oh-my-line/oh-my-line"

install_go_binary() {
    local os_name arch binary_name
    os_name="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"
    case "$arch" in
        x86_64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) return 1 ;;
    esac
    binary_name="oh-my-line-${os_name}-${arch}"

    if [ "$SOURCE" = "local" ] && [ -f "${SCRIPT_DIR}/engine/dist/${binary_name}" ]; then
        cp "${SCRIPT_DIR}/engine/dist/${binary_name}" "${INSTALL_DIR}/oh-my-line"
        chmod +x "${INSTALL_DIR}/oh-my-line"
        return 0
    fi

    local url="https://github.com/${REPO}/releases/latest/download/${binary_name}"
    if curl -fsSL "$url" -o "${INSTALL_DIR}/oh-my-line" 2>/dev/null; then
        chmod +x "${INSTALL_DIR}/oh-my-line"
        return 0
    fi
    return 1
}

if install_go_binary; then
    printf '    %b  oh-my-line engine installed\n' "$OK"
else
    printf '    %b  Engine binary not available for this platform\n' "$FAIL"
    printf '       Build from source: cd engine && go build -o ~/.oh-my-line/oh-my-line ./cmd/oh-my-line/\n'
    exit 1
fi

# ── Install Claude Code skill (global) ───────────────────────────────────────
# Skills go in ~/.claude/skills/ so /oh-my-line commands work in any project.
# oh-my-line is a global tool, not per-project — the skill must be global too.
SKILL_DIR="${HOME}/.claude/skills/oh-my-line"
mkdir -p "${SKILL_DIR}"
fetch "${SKILL_DIR}/SKILL.md" "skills/oh-my-line/SKILL.md"
fetch "${SKILL_DIR}/CONFIG-REFERENCE.md" "skills/oh-my-line/CONFIG-REFERENCE.md"

printf '    %b  oh-my-line %bv%s%b installed\n' "$OK" "$DIM" "$NEW_VERSION" "$RST"
printf '    %b  skill installed to %b%s/%b\n' "$OK" "$DIM" "${SKILL_DIR}" "$RST"

# ── Configure settings.json ──────────────────────────────────────────────────
mkdir -p "$(dirname "$SETTINGS_FILE")"

if [ -f "$SETTINGS_FILE" ]; then
    if grep -q "$COMMAND_PATH" "$SETTINGS_FILE" 2>/dev/null; then
        printf '    %b  Claude Code user global settings %b(already configured)%b\n' "$OK" "$DIM" "$RST"
    elif command -v jq >/dev/null 2>&1; then
        tmp=$(mktemp)
        jq --arg cmd "$COMMAND_PATH" '.statusLine = {"type": "command", "command": $cmd}' "$SETTINGS_FILE" > "$tmp"
        mv "$tmp" "$SETTINGS_FILE"
        printf '    %b  Claude Code user global settings updated\n' "$OK"
    else
        printf '    %b  settings.json exists — add statusLine manually:\n' "$DOT"
        printf '       "statusLine": {"type": "command", "command": "%s"}\n' "$COMMAND_PATH"
    fi
else
    cat > "$SETTINGS_FILE" << EOF
{
  "statusLine": {
    "type": "command",
    "command": "${COMMAND_PATH}"
  }
}
EOF
    printf '    %b  Claude Code user global settings created\n' "$OK"
fi

# ── Create starter config.json ────────────────────────────────────────────────
if [ -f "${INSTALL_DIR}/config.json" ]; then
    printf '    %b  config.json %b(already exists, kept)%b\n' "$OK" "$DIM" "$RST"
else
    cat > "${INSTALL_DIR}/config.json" << 'EOF'
{
  "nerdFont": false,
  "oh-my-lines": [
    {
      "separatorStyle": { "dim": true },
      "segments": [
        { "type": "rate-session", "showReset": false },
        { "type": "eta-session" },
        { "type": "sep" },
        { "type": "rate-weekly" },
        { "type": "sep" },
        { "type": "effort" },
        { "type": "sep" },
        { "type": "text", "content": "CC" },
        { "type": "cc-version" }
      ]
    },
    {
      "separatorStyle": { "dim": true },
      "segments": [
        { "type": "dir-branch" },
        { "type": "sep" },
        { "type": "diff-stats" },
        { "type": "sep" },
        { "type": "tokens" },
        { "type": "sep" },
        { "type": "text", "content": "Context" },
        { "type": "pct-used" }
      ]
    },
    {
      "segments": [
        { "type": "burn-min" },
        { "type": "burn-hr" },
        { "type": "compact-warn" },
        { "type": "docker" }
      ]
    },
    {
      "backgroundStyle": "gradient",
      "background": "#1a1a4a",
      "segments": [
        { "type": "icon", "content": "🚀" },
        { "type": "label", "content": "oh-my-line" },
        { "type": "tagline", "content": "Built with Claude Code" },
        { "type": "message", "messages": ["Hallucination? I call it jazz", "Tokens in, vibes out", "Pair programming, minus the small talk"], "interval": 300 }
      ]
    }
  ]
}
EOF
    printf '    %b  Starter config.json created\n' "$OK"
fi

# ── Verify ───────────────────────────────────────────────────────────────────
printf '\n'
printf '  %bVerifying oh-my-line%b\n\n' "$WHITE" "$RST"

test_output=$(echo '{"model":{"display_name":"Test Model"},"context_window":{"context_window_size":200000,"current_usage":{"input_tokens":50000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"cwd":"'"$(pwd)"'"}' | "${INSTALL_DIR}/oh-my-line" 2>&1)

if [ -n "$test_output" ]; then
    printf '    %b  oh-my-line runs OK\n' "$OK"
else
    printf '    %b  oh-my-line returned empty output — check the install\n' "$FAIL"
    exit 1
fi

# ── Done ─────────────────────────────────────────────────────────────────────
printf '\n'
printf '  %b%b━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%b\n' "$BOLD" "$GREEN" "$RST"
printf '  %b%b  Done!%b  oh-my-line is live. Enjoy!\n' "$BOLD" "$GREEN" "$RST"
printf '  %b%b━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━%b\n' "$BOLD" "$GREEN" "$RST"
printf '\n'
printf '    %bCustomise%b    ~/.oh-my-line/config.json\n' "$WHITE" "$RST"
printf '    %bBuilder%b      https://jamesprnich.github.io/oh-my-line/builder.html\n' "$WHITE" "$RST"
printf '    %bDocs%b         https://jamesprnich.github.io/oh-my-line/docs/\n' "$WHITE" "$RST"
printf '\n'
printf '    %bSome segments need external data (OAuth for rate limits,%b\n' "$DIM" "$RST"
printf '    %bgh CLI for GitHub). They populate automatically once available.%b\n' "$DIM" "$RST"
printf '\n'
