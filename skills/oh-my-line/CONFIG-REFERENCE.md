# oh-my-line Config Reference

Segment types, config options, line styles, and Nerd Font icons. Read this when customizing a user's statusline config.

## Config Format

Config lives in `oh-my-line.json` (project-level) or `~/.oh-my-line/config.json` (global). Project-level overrides global when present.

```json
{
  "nerdFont": false,
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

Each line in `oh-my-lines` is rendered as a separate statusline row. Segments are rendered left-to-right within a line.

### Top-Level Options

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `nerdFont` | bool | `false` | Enable Nerd Font icons on all segments |
| `debug` | bool | `false` | Enable debug logging to `~/.oh-my-line/debug.log` |
| `usageProxy` | object | — | Proxy URLs for usage data (see below) |
| `oh-my-lines` | array | `[]` | Array of line objects (max 8) |

#### Usage Proxy

Aligned to [ai-usage-proxy-spec](https://github.com/jamesprnich/ai-usage-proxy-spec). Set the base URL; oh-my-line appends `/api/proxy/anthropic/subscription/`.

```json
{
  "usageProxy": {
    "claudeCode": "http://localhost:8787"
  }
}
```

Env var `OML_USAGE_PROXY_CLAUDE_CODE` takes precedence over config.

### Segment Object

```json
{
  "type": "pct-used",
  "style": { "color": "#ff0000", "bold": true, "dim": false, "background": "#222222" },
  "prefix": "ctx:",
  "suffix": "filled",
  "icon": true
}
```

- `type` — required, must match a registered segment type
- `style` — optional, overrides default colors
- All other fields are passed as the **opts bag** (read via `_opt` in segments)

## Available Segment Types

| Type | Category | Description |
|------|----------|-------------|
| `model` | 10-basic | Current model name |
| `dir` | 10-basic | Working directory |
| `version` | 10-basic | oh-my-line version number |
| `session-cost` | 11-session | Session cost ($X.XX) |
| `session-duration` | 11-session | Session duration |
| `lines-changed` | 11-session | Lines changed in session |
| `api-wait` | 11-session | API wait time |
| `total-tokens` | 11-session | Total tokens in session |
| `cache-hit` | 11-session | Cache hit percentage |
| `branch` | 12-git | Git branch |
| `dir-branch` | 12-git | Directory@branch |
| `diff-stats` | 12-git | Git diff stats (+N -N) |
| `vim-mode` | 13-mode | Vim mode indicator |
| `worktree` | 13-mode | Worktree indicator |
| `agent` | 13-mode | Agent mode indicator |
| `200k-warn` | 13-mode | 200k context warning |
| `cc-version` | 13-mode | Claude Code version |
| `model-id` | 13-mode | Model identifier |
| `effort` | 14-effort | Effort level (low/med/high) |
| `gh-pr` | 16-github | Current PR info |
| `gh-checks` | 16-github | CI check status |
| `gh-reviews` | 16-github | PR review status |
| `gh-actions` | 16-github | Actions run status |
| `gh-notifs` | 16-github | Unread notification count |
| `gh-pr-comments` | 16-github | PR comment count |
| `gh-issues` | 16-github | Open issue count |
| `gh-pr-count` | 16-github | Open PR count |
| `gh-stars` | 16-github | Repo star count |
| `tokens` | 20-tokens | Token usage (current/total) |
| `pct-used` | 20-tokens | Context % consumed |
| `pct-remain` | 20-tokens | Context % remaining |
| `compact-warn` | 22-compact-warn | Compaction warning alert |
| `icon` | 30-product | Product emoji |
| `label` | 30-product | Product name |
| `tagline` | 30-product | Product subtitle |
| `message` | 30-product | Rotating message |
| `docker` | 32-docker | Container summary (3/3 up) |
| `docker-db` | 32-docker | DB container status (pg: up) |
| `text` | 40-custom | Static text |
| `sep` | 40-custom | Visual separator |
| `custom-icon` | 40-custom | Custom emoji |
| `command` | 40-custom | Shell command output |
| `env` | 40-custom | Environment variable value |
| `burn-min` | 50-burn-rate | Tokens/minute |
| `burn-hr` | 50-burn-rate | Tokens/hour |
| `burn-spark` | 55-sparklines | Burn rate sparkline trend |
| `ctx-spark` | 55-sparklines | Context fill sparkline trend |
| `rate-spark` | 55-sparklines | Rate limit sparkline trend |
| `ctx-target` | 56-target-sparks | Context fill with target zones |
| `rate-target` | 56-target-sparks | Weekly rate limit with target zones |
| `rate-session` | 60-rate-limits | Session rate limit bar |
| `rate-weekly` | 60-rate-limits | Weekly rate limit bar |
| `rate-extra` | 60-rate-limits | Extra usage bar |
| `rate-opus` | 60-rate-limits | Opus-specific rate limit |
| `cost` | 65-cost | Input cost estimate (~$X.XX) |
| `cost-min` | 65-cost | Dollar burn rate/min |
| `cost-hr` | 65-cost | Dollar burn rate/hr |
| `cost-7d` | 65-cost | 7-day accumulated cost |
| `cost-spark` | 65-cost | Daily cost sparkline (7d) |
| `eta-session` | 70-eta | Session ETA (avg) |
| `eta-session-min` | 70-eta | Session ETA (~5min rate) |
| `eta-session-hr` | 70-eta | Session ETA (~60min rate) |
| `eta-weekly` | 70-eta | Weekly ETA (avg) |
| `eta-weekly-min` | 70-eta | Weekly ETA (~5min rate) |
| `eta-weekly-hr` | 70-eta | Weekly ETA (~60min rate) |

## Configurable Segment Options

Set these directly on the segment object in config:

```json
{ "type": "compact-warn", "threshold": 20 }
{ "type": "burn-min", "warmup": 60, "prefix": "burn:" }
{ "type": "ctx-target", "warn": 40, "critical": 70, "width": 12 }
{ "type": "command", "content": "date +%H:%M", "cache": 30, "timeout": 5 }
```

| Option | Applies to | Default | Description |
|--------|-----------|---------|-------------|
| `prefix` | Most types (not text/sep/custom-icon) | — | Dim text before segment |
| `suffix` | Most types (not text/sep/custom-icon) | — | Dim text after segment |
| `content` | text, sep, custom-icon, command, env, icon, label, tagline | — | User content |
| `threshold` | compact-warn | `10` | % remaining to trigger warning |
| `warmup` | burn-min, burn-hr | `30` | Seconds before showing data |
| `width` | burn-spark, ctx-spark, rate-spark, ctx-target, rate-target | `8` | Sparkline bar count |
| `warn` | ctx-target, rate-target | `50` | % for amber zone |
| `critical` | ctx-target, rate-target | `80` | % for red zone |
| `barStyle` | rate-session, rate-weekly, rate-extra, rate-opus | `"dot"` | Bar style: `"dot"` (●○) or `"block"` (▓░) |
| `cache` | command | `60` | Cache TTL (seconds) |
| `timeout` | command | `3` | Execution timeout (seconds) |

## Nerd Font Icons

Opt-in feature that prepends Nerd Font glyphs to segments. Requires a [Nerd Font](https://www.nerdfonts.com/) installed in the user's terminal.

**Enable globally:**
```json
{ "nerdFont": true, "oh-my-lines": [ ... ] }
```

**Per-segment override:**
```json
{ "type": "model", "icon": false }
```

When a nerd icon is shown, the segment's `prefix` is automatically suppressed (the icon replaces it). Suffixes are kept.

The icon map is built into the engine. The builder has a "Nerd Font icons" toggle.

## Line Types

### Normal Lines

Standard lines with segments. Support these properties:

| Property | Values | Description |
|----------|--------|-------------|
| `separator` | any string | Character(s) between segments |
| `separatorStyle` | `{ "dim": true }` | Style for separator |
| `background` | `{ "style": "solid", "color": "#hex" }` | Background style |

Background styles: `solid`, `fade`, `gradient`, `neon`.

### Rule Lines

Repeating character across the full width:

```json
{ "type": "rule", "char": "─", "style": { "color": "#444444", "dim": true } }
```

### Spacer Lines

Empty line for vertical spacing:

```json
{ "type": "spacer" }
```

### ASCII Art Lines

Multi-line ASCII art:

```json
{ "type": "art", "content": ["line1", "line2"], "style": { "color": "#hex" } }
```

## Product Identity

Product segments (`icon`, `label`, `tagline`, `message`) can pull from a shared `.product.json` file at the repo root. The file follows the [product-json-spec](https://github.com/jamesprnich/product-json-spec):

```json
{
  "$schema": "https://raw.githubusercontent.com/jamesprnich/product-json-spec/main/schema.json",
  "icon": "🚀",
  "name": "My Project",
  "tagline": "A cool project",
  "extensions": {
    "oh-my-line": {
      "messages": ["Tip one", "Tip two"],
      "messageInterval": 300
    }
  }
}
```

oh-my-line reads standard spec fields (`icon`, `name`, `tagline`) directly. Tool-specific fields (`messages`, `messageInterval`) go under `extensions.oh-my-line` per the spec.

Reference it via `source` on product segments:
```json
{ "type": "icon", "source": ".product.json" }
{ "type": "message", "source": ".product.json" }
```

Or override inline:
```json
{ "type": "icon", "content": "🔥" }
```
