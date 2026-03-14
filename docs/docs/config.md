# Config Reference

Configuration lives in `oh-my-line.json`. The engine checks your working directory first, then per-account config (if `CLAUDE_CONFIG_DIR` is set), then falls back to `~/.oh-my-line/config.json`.

Product identity (icon, label, tagline) is set inline on each segment via the `content` field. Optionally, you can reference a shared `.product.json` file via `source: ".product.json"` — see [Product identity segments](#product-identity-segments) below.

## Config lookup order

```
1. {cwd}/oh-my-line.json                    ← project config (untrusted)
2. {CLAUDE_CONFIG_DIR}/oh-my-line.json       ← per-account config (trusted)
3. ~/.oh-my-line/config.json                 ← global config (trusted)
```

Step 2 is skipped when `CLAUDE_CONFIG_DIR` is unset or is the default (`~/.claude`). See [Multi-account setup](#multi-account-setup) for details.

If no config file is found anywhere, the statusline shows just the model name.

## oh-my-line.json format

```json
{
  "nerdFont": false,
  "oh-my-lines": [
    {
      "separator": "|",
      "separatorStyle": { "dim": true },
      "segments": [
        { "type": "model" },
        { "type": "tokens" },
        { "type": "icon", "content": "🚀" },
        { "type": "label", "content": "oh-my-line" },
        { "type": "message", "messages": ["Pull. Run. Done.", "No cloud required."], "interval": 300 }
      ]
    }
  ]
}
```

| Field | Type | Description |
|-------|------|-------------|
| `nerdFont` | bool | Enable Nerd Font icons on all segments (default: `false`). Requires a [Nerd Font](https://www.nerdfonts.com/) in your terminal. |
| `oh-my-lines` | array | Array of line objects (max 8). Each renders as one row. |

## Product identity segments

Product identity (icon, label, tagline) is configured per-segment, not at the top level.

### Inline content

Set the value directly on the segment:

```json
{ "type": "icon", "content": "🚀" }
{ "type": "label", "content": "oh-my-line" }
{ "type": "tagline", "content": "Built with Claude Code" }
```

### From .product.json (shared product identity)

Reference the shared `.product.json` file:

```json
{ "type": "icon", "source": ".product.json" }
{ "type": "label", "source": ".product.json" }
```

When `source` is set, the engine reads the corresponding field from `.product.json` at the repo root. This file follows the [product-json-spec](https://github.com/jamesprnich/product-json-spec) and is shared with other tools — oh-my-line reads but never writes it.

`content` takes priority over `source` — if both are set, `content` wins.

### Message segment

Rotating messages are configured directly on the segment:

```json
{ "type": "message", "messages": ["Pull. Run. Done.", "No cloud required.", "Models on tap."], "interval": 300 }
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `messages` | array | `[]` | Strings rotated by the message segment |
| `interval` | number | `300` | Seconds between message rotation |

## Line objects

Each entry in `oh-my-lines[]` is either a **normal line** with segments, or a **special line type**.

### Normal line

```json
{
  "separator": "|",
  "separatorStyle": { "dim": true },
  "segments": [
    { "type": "model" },
    { "type": "tokens" }
  ]
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `segments` | array | `[]` | Array of segment objects |
| `separator` | string | — | Character(s) placed between segments |
| `separatorStyle` | object | — | `{ "dim": bool, "color": "#hex" }` |
| `backgroundStyle` | string | `"none"` | `none`, `solid`, `gradient`, `fade`, `neon` |
| `background` | string | — | Hex colour for the background |
| `padding` | object | — | `{ "left": N, "right": N }` |

#### Background styles

| Style | Effect |
|-------|--------|
| `none` | No background (transparent) |
| `solid` | Flat colour fill |
| `gradient` | Colour fades to black over 60 characters |
| `fade` | Colour with block fade trail: `████▓▓▒▒░░` |
| `neon` | Dark blue-purple base (`#1a1a2e`) |

### Special line types

#### Art

Multi-line ASCII art:

```json
{
  "type": "art",
  "style": { "color": "#00ff00", "dim": true },
  "lines": [
    "╔══════════╗",
    "║  MyApp   ║",
    "╚══════════╝"
  ]
}
```

#### Rule

Horizontal divider:

```json
{
  "type": "rule",
  "char": "─",
  "width": 120,
  "style": { "color": "#555555", "dim": true }
}
```

#### Spacer

Blank line:

```json
{ "type": "spacer" }
```

## Segment objects

```json
{
  "type": "model",
  "style": {
    "color": "#ff6600",
    "background": "#1a1a2e",
    "bold": true,
    "dim": false
  }
}
```

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | `"text"` | Segment type (see [Segment Types](segments.md)) |
| `style.color` | string | per-type | Foreground hex colour |
| `style.background` | string | — | Background hex colour |
| `style.bold` | bool | `false` | Bold text |
| `style.dim` | bool | `false` | Dimmed text |
| `prefix` | string | — | Dim text prepended before segment output |
| `suffix` | string | — | Dim text appended after segment output |
| `icon` | bool | — | Per-segment Nerd Font override (`true`/`false`). Overrides global `nerdFont` setting. |
| `content` | string | — | For `text`, `sep`, `custom-icon`, `command`, `icon`, `label`, `tagline` types |
| `source` | string | — | For `icon`, `label`, `tagline`: `".product.json"` to read from shared file |
| `messages` | array | — | For `message` type: array of strings to rotate |
| `interval` | number | `300` | For `message` type: seconds between rotation |
| `cache` | number | `60` | For `command` type: cache TTL in seconds |
| `timeout` | number | `3` | For `command` type: execution timeout in seconds |

If no `style.color` is set, the segment's registered default colour is used.

### Configurable segment options

Some segments expose tunable options. Set these directly on the segment object:

```json
{ "type": "compact-warn", "threshold": 20 }
{ "type": "burn-min", "warmup": 60 }
{ "type": "ctx-target", "warn": 40, "critical": 70, "width": 12 }
{ "type": "command", "content": "date +%H:%M", "cache": 10, "timeout": 5 }
```

| Option | Applies to | Default | Description |
|--------|-----------|---------|-------------|
| `threshold` | compact-warn | `10` | % remaining to trigger warning |
| `warmup` | burn-min, burn-hr | `30` | Seconds before showing data |
| `width` | burn-spark, ctx-spark, rate-spark, ctx-target, rate-target | `8` | Sparkline bar count |
| `warn` | ctx-target, rate-target | `50` | % for amber zone |
| `critical` | ctx-target, rate-target | `80` | % for red zone |
| `barStyle` | rate-session, rate-weekly, rate-extra, rate-opus | `"dot"` | Bar style: `"dot"` (●○) or `"block"` (▓░) |
| `showReset` | rate-session, rate-weekly, rate-extra, rate-opus | varies | Show reset time/date. Default: `true` for session, `false` for weekly/extra/opus |

## Multi-project setup

Use per-project `oh-my-line.json` files for different configurations with a shared global fallback:

```
~/.oh-my-line/config.json              ← global default layout
~/projects/webapp/oh-my-line.json      ← webapp config
~/projects/api-server/oh-my-line.json  ← api config
```

The engine finds the nearest config by checking the current working directory first, then the global path. Each project can have its own segments and layout while sharing the same engine installation.

## Multi-account setup

If you run multiple Claude Code accounts simultaneously using `CLAUDE_CONFIG_DIR`, oh-my-line automatically isolates per-account state so sessions never cross-contaminate.

**No configuration needed** — the engine detects `CLAUDE_CONFIG_DIR` and creates separate cache subdirectories automatically. Optionally, each account can have its own statusline config.

### How it works

The engine derives an account key from `CLAUDE_CONFIG_DIR` via SHA-256 hash. The default account (`~/.claude` or unset) uses the base cache path unchanged — zero behavior change for single-account users.

**Runtime state** (cache, cost) is isolated per account automatically:

```
/tmp/claude-{uid}/                     ← default account (base path)
/tmp/claude-{uid}/acct-{hash}/         ← additional accounts
~/.oh-my-line/cost/                    ← default account cost data
~/.oh-my-line/cost/acct-{hash}/        ← additional account cost data
```

**Config** can optionally be customized per account by placing `oh-my-line.json` in the account's config directory:

```
{CLAUDE_CONFIG_DIR}/oh-my-line.json    ← per-account config (trusted)
```

This is checked after the project config but before the global fallback. If no per-account config exists, the global `~/.oh-my-line/config.json` is used. The per-account config is trusted (can run `command` segments) because the user controls their own config directory.

### What's isolated per account

| Data | Why |
|------|-----|
| Burn rate tracking | Token consumption rate is per-session |
| Rate limit cache | Each account has its own usage limits |
| ETA projections | Derived from per-account consumption |
| Sparkline history | Per-account time series data |
| Cost tracking | Daily cost logs are per-account |
| OAuth tokens | Credentials read from account's config directory |
| Settings (effort level) | Read from account's `settings.json` |

### What's shared across accounts

| Data | Why |
|------|-----|
| Claude Code version | Same binary regardless of account |
| GitHub PR/issue data | Repo-scoped, not account-scoped |
| Docker container status | Machine-scoped, not account-scoped |
| Command segment output | Command-scoped, not account-scoped |

### Per-account config example

```
~/.oh-my-line/config.json                 ← global default (all accounts)
~/.claude-work/oh-my-line.json            ← work account only
```

Terminal 1 (default account — uses global config):

```bash
claude
```

Terminal 2 (work account — uses per-account config + isolated state):

```bash
CLAUDE_CONFIG_DIR=~/.claude-work claude
```

Both sessions render their own rate limits, burn rates, and ETAs independently. The work account can have a completely different statusline layout.

!!! warning "Command segments require a trusted config"

    `command` segments only execute from trusted configs — `~/.oh-my-line/config.json` (global) and `{CLAUDE_CONFIG_DIR}/oh-my-line.json` (per-account). Project-level configs cannot run shell commands — this prevents cloned repos from executing arbitrary code on your machine.

## Usage proxy

When multiple Claude Code sessions poll the Anthropic API simultaneously, they may trigger rate limiting. A usage proxy fetches once and serves all clients. oh-my-line consumes the [ai-usage-proxy-spec](https://github.com/jamesprnich/ai-usage-proxy-spec).

### Configuration

Set the **base URL** of your proxy. oh-my-line appends `/api/proxy/anthropic/subscription/` automatically.

**Environment variable** (takes precedence):

```bash
export OML_USAGE_PROXY_CLAUDE_CODE="http://localhost:8787"
```

**Config file:**

```json
{
  "usageProxy": {
    "claudeCode": "http://localhost:8787"
  },
  "oh-my-lines": [...]
}
```

When set, all rate-limit segments (`rate-session`, `rate-weekly`, `rate-extra`, `rate-opus`) and ETA segments fetch from the proxy instead of the Anthropic API directly. No authentication headers are sent to the proxy.

See [Usage Proxy](usage-proxy.md) for details.

## Debug logging

Enable debug logging to troubleshoot rate limits, proxy connections, and cache behavior.

**Environment variable** (takes precedence):

```bash
export OML_DEBUG=1
```

**Config file:**

```json
{
  "debug": true,
  "oh-my-lines": [...]
}
```

Logs are written to `~/.oh-my-line/debug.log` and auto-truncated at 100KB.

## Security: trusted configs

Trusted configs — `~/.oh-my-line/config.json` (global) and `{CLAUDE_CONFIG_DIR}/oh-my-line.json` (per-account) — can run shell commands via `command` segments. Project-level configs cannot execute arbitrary code.
