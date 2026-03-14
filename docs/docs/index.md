# Getting Started

**oh-my-line** is a real-time statusline engine for [Claude Code](https://docs.anthropic.com/en/docs/claude-code). It displays model info, token usage, rate limits, burn rates, ETAs, product identity, and custom content — all configurable via JSON.

The engine (`oh-my-line`) is a compiled Go binary that reads JSON on stdin and outputs ANSI-formatted text. Any tool that pipes JSON on stdin can use it.

## Install

### One-liner

```bash
curl -fsSL https://raw.githubusercontent.com/jamesprnich/oh-my-line/main/install.sh | bash
```

### Clone and run

```bash
git clone https://github.com/jamesprnich/oh-my-line.git
cd oh-my-line && ./install.sh
```

The installer:

1. Downloads files to `~/.oh-my-line/`
2. Installs Claude Code skill files to `~/.claude/skills/oh-my-line/` — enables `/oh-my-line` commands for setup, customization, and troubleshooting
3. Adds `statusLine` to Claude Code's user global settings (`~/.claude/settings.json`)
4. Creates a starter `~/.oh-my-line/config.json`
5. Verifies the engine works

The statusline appears at the bottom of Claude Code on next launch.

### Upgrade

Re-run the same install command — the installer detects your existing version and upgrades in place. Your config is preserved.

```bash
curl -fsSL https://raw.githubusercontent.com/jamesprnich/oh-my-line/main/install.sh | bash
```

??? note "Manual install"

    Run the installer or build from source (`go build -o oh-my-line`), then place the binary and supporting files in `~/.oh-my-line/`. Add to Claude Code's user global settings (`~/.claude/settings.json`):

    ```json
    {
      "statusLine": {
        "type": "command",
        "command": "~/.oh-my-line/oh-my-line"
      }
    }
    ```

## Prerequisites

- **macOS or Linux** (x86_64 or arm64)
- **Claude Code** with OAuth session — required for rate limit and ETA segments

## Customize

Use the [Builder](https://jamesprnich.github.io/oh-my-line/builder.html) to visually add, remove, and style segments — then copy the JSON into your config file.

Or edit `~/.oh-my-line/config.json` (or `oh-my-line.json` in your project root) directly:

```json
{
  "oh-my-lines": [
    {
      "separator": "|",
      "separatorStyle": { "dim": true },
      "segments": [
        { "type": "model" },
        { "type": "dir-branch" },
        { "type": "tokens" },
        { "type": "rate-session" },
        { "type": "rate-weekly" }
      ]
    }
  ]
}
```

If no config file is found anywhere, the statusline falls back to showing the model name.

## Architecture

```
~/.oh-my-line/
├── oh-my-line          # Compiled engine binary
├── config.json         # User config
└── VERSION
```

The engine is a compiled Go binary. It reads JSON config and renders all segments internally — there are no shell scripts to source. All segment types (model, tokens, rate limits, burn rates, ETAs, etc.) are built into the binary.

## How It Works

### Stdin format

Claude Code pipes a JSON object on stdin every few seconds:

```json
{
  "model": {
    "display_name": "Claude Opus 4.6"
  },
  "context_window": {
    "context_window_size": 200000,
    "current_usage": {
      "input_tokens": 50000,
      "cache_creation_input_tokens": 10000,
      "cache_read_input_tokens": 5000
    }
  },
  "cwd": "/Users/you/project"
}
```

The engine computes `current_tokens` as the sum of `input_tokens + cache_creation_input_tokens + cache_read_input_tokens`.

### Engine flow

```
Phase 1: Read stdin      Parse incoming JSON, early exit if empty
Phase 2: Parse           Extract model, tokens, cwd, and other fields
Phase 3: Config          oh-my-line.json lookup (project → global)
Phase 4: Compute         Evaluate all active segments internally
Phase 5: Render          Compose lines with separators and styles
Phase 6: Output          Print final ANSI-formatted string to stdout
```

### Config lookup

The engine checks for config files in this order:

```
{cwd}/oh-my-line.json                    ← project config (untrusted)
{CLAUDE_CONFIG_DIR}/oh-my-line.json       ← per-account config (trusted)
~/.oh-my-line/config.json                 ← global config (trusted)
```

The per-account step is skipped when `CLAUDE_CONFIG_DIR` is unset or is the default (`~/.claude`).

## Uninstall

Remove the `statusLine` key from Claude Code's user global settings (`~/.claude/settings.json`), then delete the install directory:

```bash
rm -rf ~/.oh-my-line
```
