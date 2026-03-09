---
name: oh-my-line
description: "[Skill] Manage oh-my-line statusline — setup, customize segments, troubleshoot."
argument-hint: "<install|customize|troubleshoot>"
user-invocable: true
---

# oh-my-line Skill

Manage the oh-my-line statusline engine.

## Bootstrap

If the user doesn't have oh-my-line installed:

```bash
curl -fsSL https://raw.githubusercontent.com/jamesprnich/oh-my-line/main/install.sh | bash
```

This installs the compiled Go binary, builder, and skill files globally.

## Install Locations

oh-my-line is a **global tool** — installed once, available everywhere. The skill files are installed to Claude Code's global skills directory so `/oh-my-line` commands work in any project, any terminal.

| Path | Purpose |
|------|---------|
| `~/.oh-my-line/oh-my-line` | Compiled Go binary (the engine) |
| `~/.oh-my-line/config.json` | Global statusline config |
| [Builder](https://jamesprnich.github.io/oh-my-line/builder.html) | Visual config builder (hosted) |
| `~/.oh-my-line/VERSION` | Installed version identifier |
| `~/.claude/skills/oh-my-line/` | Global skill files (this file + sub-specs) |
| `~/.claude/settings.json` | Claude Code settings — `statusLine` key |
| `oh-my-line.json` | Project-level config (overrides global) |
| `.product.json` | Shared product identity at repo root |

## Update Check

Before running any command, check for updates:

```bash
LOCAL_VERSION=$(cat ~/.oh-my-line/VERSION 2>/dev/null | tr -d '[:space:]')
REMOTE_VERSION=$(curl -fsSL --max-time 5 https://raw.githubusercontent.com/jamesprnich/oh-my-line/main/VERSION 2>/dev/null | tr -d '[:space:]')
```

If newer → tell user and show the install command. Also re-fetch this skill:

```bash
curl -fsSL --max-time 5 https://raw.githubusercontent.com/jamesprnich/oh-my-line/main/skills/oh-my-line/SKILL.md -o ~/.claude/skills/oh-my-line/SKILL.md 2>/dev/null
```

## Usage

```
/oh-my-line install         Set up the statusline engine from scratch
/oh-my-line customize     Add/remove/restyle segments in oh-my-line.json
/oh-my-line troubleshoot  Diagnose statusline issues
```

No argument → ask the user what they want to do.

## Sub-Specs

These companion files contain detailed reference material. Read them when needed — not upfront.

| File | When to read |
|------|-------------|
| `~/.claude/skills/oh-my-line/CONFIG-REFERENCE.md` | When customizing configs — has all segment types, options, line styles |

---

## Command: `install`

### Step 1: Check if already installed

```bash
[ -f ~/.oh-my-line/oh-my-line ] && echo "INSTALLED" || echo "NOT INSTALLED"
```

If installed → offer to customize instead. If not → Step 2.

### Step 2: Run the installer

```bash
curl -fsSL https://raw.githubusercontent.com/jamesprnich/oh-my-line/main/install.sh | bash
```

### Step 3: Nerd Fonts (recommended)

Ask the user if they want Nerd Font icons in their statusline. If yes:

1. Direct them to install a Nerd Font from [nerdfonts.com](https://www.nerdfonts.com/) — popular choices are "JetBrainsMono Nerd Font" or "FiraCode Nerd Font"
2. Tell them to select the installed Nerd Font in their terminal app's font settings
3. Set `"nerdFont": true` in their config file (`~/.oh-my-line/config.json`)

If no, leave `"nerdFont": false` (the default) — the statusline works fine without icons.

### Step 4: Customize config (optional)

Ask the user in plain text:

1. **Where?** — Project-specific or global (`~/.oh-my-line/config.json`)? Default: global.
2. **Product identity** — Icon, label, tagline. All optional.
3. **Complexity** — Minimal (model + dir + tokens) or full (rate limits, ETAs, burn rate)?

Read `CONFIG-REFERENCE.md` for available segment types when building the config.

**Minimal config:**
```json
{
  "oh-my-lines": [
    {
      "separator": "|",
      "separatorStyle": { "dim": true },
      "segments": [
        { "type": "model" },
        { "type": "dir-branch" },
        { "type": "tokens" }
      ]
    }
  ]
}
```

### Step 5: Summary

Tell the user:
- Statusline appears at the bottom of Claude Code terminal, refreshes automatically
- Customize via `oh-my-line.json` or the [builder](https://jamesprnich.github.io/oh-my-line/builder.html)
- Rate limit segments require an OAuth session

---

## Command: `customize`

### Step 1: Find and read config

Check `cwd` for `oh-my-line.json`, then `~/.oh-my-line/config.json`. If none → suggest setup.

### Step 2: Read CONFIG-REFERENCE.md

Read `~/.claude/skills/oh-my-line/CONFIG-REFERENCE.md` for the full segment type catalog and config options.

### Step 3: Understand intent and apply

Common requests:

- **Add a segment** → add to the `segments` array. Check CONFIG-REFERENCE.md for available types.
- **Remove a segment** → remove from config.
- **Change style** → add/modify `style` object: `{ "color": "#hex", "background": "#hex", "bold": true, "dim": true }`
- **Add a line** → add to `oh-my-lines` array.
- **Add a command segment** → `{ "type": "command", "content": "the-command", "cache": 60 }`
- **Add rotating messages** → `{ "type": "message", "messages": ["A", "B"], "interval": 300 }`
- **Change product identity** → update `content` on icon/label/tagline, or use `"source": ".product.json"`

For visual configuration, point the user to the [Builder](https://jamesprnich.github.io/oh-my-line/builder.html).

### Step 4: Verify

```bash
echo '{"model":{"display_name":"Test"},"context_window":{"context_window_size":200000,"current_usage":{"input_tokens":50000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"cwd":"'$(pwd)'"}' | ~/.oh-my-line/oh-my-line 2>&1 | cat -v
```

---

## Command: `troubleshoot`

### Step 1: Gather symptoms

| Symptom | Investigation |
|---------|--------------|
| Statusline missing | Check `~/.claude/settings.json` has `statusLine` key |
| Shows "Claude" or model name only | No config found — check file exists and location |
| Rate limits empty / "?" | OAuth session issue |
| Segments showing warning icon | API error — usually transient |
| ETA segments not showing | Need 2-10 minutes of data collection |
| Command segment blank | Command failing or returning empty |
| Stale data | Clear cache files |

### Step 2: Diagnostics

**Config check:**
```bash
[ -f "$(pwd)/oh-my-line.json" ] && echo "Project: $(pwd)/oh-my-line.json" || \
[ -f ~/.oh-my-line/config.json ] && echo "Global: ~/.oh-my-line/config.json" || echo "No config"
```

**Engine test:**
```bash
echo '{"model":{"display_name":"Test"},"context_window":{"context_window_size":200000,"current_usage":{"input_tokens":50000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"cwd":"'$(pwd)'"}' | ~/.oh-my-line/oh-my-line 2>&1 | cat -v
```

**Cache state:**
```bash
ls -la /tmp/claude-$(id -u)/statusline-* 2>/dev/null || echo "No cache files"
```

### Step 3: Common fixes

- **Missing config** → run `/oh-my-line install`
- **Invalid JSON** → fix syntax, show user what was wrong
- **Cache stale** → `rm /tmp/claude-$(id -u)/statusline-*.dat /tmp/claude-$(id -u)/statusline-*.json`
- **Binary not running** → check `~/.oh-my-line/oh-my-line` exists and is executable (`chmod +x`)
- **OAuth missing** → restart Claude Code, log in fresh

