# Segment Types

## Basic

| Type | Default Colour | Description |
|------|---------------|-------------|
| `model` | `#0099ff` | Current model name (e.g. "Claude Opus 4.6") |
| `dir` | `#2e9599` | Current working directory name |
| `version` | `#888888` | oh-my-line version number |

## Session

| Type | Default Colour | Description |
|------|---------------|-------------|
| `session-cost` | `#e6c800` | Session cost (e.g. "$1.50") |
| `session-duration` | `#888888` | Session elapsed time (e.g. "45m") |
| `lines-changed` | `#00a000` | Lines changed in session (e.g. "+120 -30") |
| `api-wait` | `#888888` | API wait/response time |
| `total-tokens` | `#ffb055` | Total tokens consumed in session |
| `cache-hit` | `#2e9599` | Cache hit percentage |

These segments read from Claude Code's session metadata when available.

## Git

| Type | Default Colour | Description |
|------|---------------|-------------|
| `branch` | `#00a000` | Current git branch |
| `dir-branch` | `#2e9599` | Directory and git branch combined (e.g. "myapp@main") with diff stats |
| `diff-stats` | `#00a000` | Standalone git diff stats (e.g. "+12 -3") |

`dir-branch` automatically shows uncommitted changes as `+additions -deletions` when there are unstaged changes.

## Mode

| Type | Default Colour | Description |
|------|---------------|-------------|
| `vim-mode` | `#e6c800` | Vim mode indicator (normal/insert/visual) |
| `worktree` | `#c678dd` | Git worktree indicator |
| `agent` | `#e6c800` | Agent mode indicator |
| `200k-warn` | `#ff5555` | Warning when context window is 200k (large context) |
| `cc-version` | `#888888` | Claude Code version string |
| `model-id` | `#888888` | Full model identifier |

## Effort

| Type | Default Colour | Description |
|------|---------------|-------------|
| `effort` | `#00a0a0` | Current Claude Code effort level (low/med/high) |

Reads from `CLAUDE_CODE_EFFORT_LEVEL` env var first, then falls back to `effortLevel` in Claude Code's user global settings (`~/.claude/settings.json`). Colour-coded: grey (low), orange (med), green (high).

## GitHub

| Type | Default Colour | Description |
|------|---------------|-------------|
| `gh-pr` | `#2e9599` | Current branch PR number and state (open/draft/merged/closed) |
| `gh-checks` | `#00a000` | CI check summary — pass/fail/pending counts |
| `gh-reviews` | `#00a000` | Review status — approved, changes requested, or review count |
| `gh-pr-comments` | `#888888` | PR comment count |
| `gh-actions` | `#00a000` | Latest workflow run name, status, and age |
| `gh-notifs` | `#e6c800` | Unread GitHub notification count |
| `gh-issues` | `#ff5555` | Open issue count for the repo |
| `gh-pr-count` | `#2e9599` | Open PR count for the repo |
| `gh-stars` | `#e6c800` | Repository star count (formatted as 1.2k, 3.4m) |

!!! info "Requirements"

    GitHub segments require the `gh` CLI to be installed and authenticated. They also need a git repository with a branch — segments return empty outside of git repos.

!!! note "Caching"

    - PR data (`gh-pr`, `gh-checks`, `gh-reviews`, `gh-pr-comments`) — single `gh pr view` call, cached 60s
    - Actions — single `gh run list` call, cached 60s
    - Notifications — `gh api` call, cached 120s
    - Repo stats (`gh-stars`, `gh-issues`, `gh-pr-count`) — cached 5min

## Tokens

| Type | Default Colour | Description |
|------|---------------|-------------|
| `tokens` | `#ffb055` | Context window usage as current/total (e.g. "65k/200k") |
| `pct-used` | `#00a000` | Percentage of context window consumed (e.g. "32% used") |
| `pct-remain` | `#2e9599` | Percentage of context window remaining (e.g. "68% remain") |

## Compact Warning

| Type | Default Colour | Description |
|------|---------------|-------------|
| `compact-warn` | `#ff5555` | Compaction warning — alert-only, hidden when context remaining > threshold |

Configurable via `threshold` option (default: 10%). Only appears when context usage exceeds the threshold:

```json
{ "type": "compact-warn", "threshold": 15 }
```

## Product

| Type | Default Colour | Description |
|------|---------------|-------------|
| `icon` | — | Product icon — set `content` inline or `source: ".product.json"` |
| `label` | `#ffffff` | Product name — set `content` inline or `source: ".product.json"` (rendered bold) |
| `tagline` | `#888888` | Product subtitle — set `content` inline or `source: ".product.json"` (rendered dim) |
| `message` | `#888888` | Rotating message — set `messages` array and `interval` on the segment |

### Inline content vs source

Product segments (`icon`, `label`, `tagline`) support two ways to set their value:

- **`content`** — inline value directly on the segment: `{ "type": "icon", "content": "🦙" }`
- **`source`** — read from `.product.json` ([product-json-spec](https://github.com/jamesprnich/product-json-spec) file at repo root): `{ "type": "icon", "source": ".product.json" }`

If both are set, `content` wins.

### Message segment

```json
{ "type": "message", "messages": ["Pull. Run. Done.", "No cloud required.", "Models on tap."], "interval": 300 }
```

- `messages` — array of strings to cycle through
- `interval` — seconds between rotation (default: 300)

## Docker

| Type | Default Colour | Description |
|------|---------------|-------------|
| `docker` | `#2e9599` | Container summary (e.g. "3/3 up" or "1 unhealthy") |
| `docker-db` | `#2e9599` | Database container status (e.g. "postgres: up") |

Only active when a Docker Compose file (`docker-compose.yml`, `compose.yml`, etc.) is found in the current working directory. Detects database containers by image/service name (postgres, mysql, redis, mongo).

Cached for 30 seconds. Requires the `docker` CLI.

## Custom

| Type | Default Colour | Description |
|------|---------------|-------------|
| `text` | `#aaaaaa` | Static text from `content` field |
| `sep` | `#666666` | Visual separator from `content` field (default: `|`) |
| `custom-icon` | `#ffffff` | User-chosen emoji from `content` field |
| `command` | `#aaaaaa` | Output of a shell command (see below) |

### Command segments

Run any shell command and display the output. Results are cached to avoid re-running on every refresh.

```json
{ "type": "command", "content": "date +%H:%M", "cache": 10 }
{ "type": "command", "content": "curl -s 'wttr.in/?format=%c%t'", "cache": 300 }
```

- `content` — shell command to execute
- `cache` — seconds to cache output (default: 60)
- Timeout: 3 seconds max execution
- Fails silently (segment skipped if command errors or returns empty)

!!! tip "Command ideas"

    - `date +%H:%M` — current time (cache: 10s)
    - `curl -s 'wttr.in/?format=%c%t'` — weather (cache: 300s)
    - `git rev-list --count HEAD` — commit count (cache: 30s)
    - `docker ps -q | wc -l | tr -d ' '` — running containers (cache: 30s)

## Burn Rate

| Type | Default Colour | Description |
|------|---------------|-------------|
| `burn-min` | `#e6c800` | Token burn rate per minute (e.g. "2.1k/min") |
| `burn-hr` | `#e6c800` | Token burn rate per hour (e.g. "126k/hr") |

Tracks token consumption over time. Resets automatically when tokens drop (new context window).

Configurable via `warmup` option (default: 30 seconds) — suppresses noisy initial values:

```json
{ "type": "burn-min", "warmup": 60 }
```

## Sparklines

| Type | Default Colour | Description |
|------|---------------|-------------|
| `burn-spark` | `#e6c800` | Token burn rate trend (8-bar Unicode sparkline) |
| `ctx-spark` | `#ffb055` | Context fill trend (8-bar Unicode sparkline) |
| `rate-spark` | `#00a000` | Rate limit usage trend (8-bar Unicode sparkline) |

Sparklines show historical trends as Unicode block characters (▁▂▃▄▅▆▇█). Configurable width via `width` option (default: 8 bars).

## Target Sparklines

| Type | Default Colour | Description |
|------|---------------|-------------|
| `ctx-target` | `#ffb055` | Context fill with 3-color target zones (green/amber/red) |
| `rate-target` | `#00a000` | Weekly rate limit with 3-color target zones |

Zone thresholds are configurable:

```json
{ "type": "ctx-target", "warn": 50, "critical": 80, "width": 10 }
```

## Cost

| Type | Default Colour | Description |
|------|---------------|-------------|
| `cost` | `#e6c800` | Estimated input cost for this context window (e.g. "~$0.42") |
| `cost-min` | `#e6c800` | Dollar burn rate per minute (e.g. "~$0.03/min") |
| `cost-hr` | `#e6c800` | Dollar burn rate per hour (e.g. "~$1.80/hr") |
| `cost-7d` | `#e6c800` | Accumulated cost over past 7 days (e.g. "~$12.50 (7d)") |
| `cost-spark` | `#e6c800` | Daily cost sparkline — 7 bars, one per day |

Computes input-side cost estimates using public API pricing (per 1M tokens): Opus $15/$18.75/$1.50, Sonnet $3/$3.75/$0.30, Haiku $0.25/$0.30/$0.025 (input/cache write/cache read). Model is matched case-insensitively; unknown models default to Sonnet pricing.

Costs are prefixed with `~` since output tokens aren't available in the status JSON, making this an input-only estimate.

Daily costs are logged atomically to `~/.oh-my-line/cost/YYYY-MM-DD.dat` files, safe for concurrent Claude instances. Files older than 8 days are pruned automatically.

## Rate Limits

| Type | Default Colour | Description |
|------|---------------|-------------|
| `rate-session` | `#00a000` | Session rate limit: progress bar + percentage + reset time |
| `rate-weekly` | `#00a000` | Weekly rate limit: progress bar + percentage + reset time |
| `rate-extra` | `#00a000` | Extra usage: progress bar + dollars spent/limit |
| `rate-opus` | `#7c3aed` | Opus-specific rate limit (when available) |

Fetches usage data from the Anthropic API via OAuth, or from a [usage proxy](usage-proxy.md) if configured (aligned to [ai-usage-proxy-spec](https://github.com/jamesprnich/ai-usage-proxy-spec)). Proxy responses cached for 60s; direct API responses for 5 minutes.

!!! info "Adaptive window labels"

    Display labels (e.g. "5h", "7d") auto-adapt to your subscription plan's actual window sizes. The type names (`rate-session`, `rate-weekly`) remain fixed for config compatibility.

### Bar style

Choose between dot (default) and block characters:

```json
{ "type": "rate-session", "barStyle": "dot" }
{ "type": "rate-session", "barStyle": "block" }
```

| Style | Filled | Empty |
|-------|--------|-------|
| `dot` (default) | ● | ○ |
| `block` | ▓ | ░ |

### Show reset time

Control whether the reset countdown is shown:

```json
{ "type": "rate-session", "showReset": false }
```

Default is `true` for `rate-session` and `false` for `rate-weekly`, `rate-extra`, and `rate-opus`.

### Status indicators

| Icon | Colour | Meaning |
|------|--------|---------|
| ⚠ | Red | Cannot connect to data source (API or proxy unreachable) |
| ⚠ | Yellow | Data is stale (cache expired, or proxy reports rate limiting) |

Progress bar colours:

- :green_circle: Green — under 50%
- :orange_circle: Orange — 50-69%
- :yellow_circle: Yellow — 70-89%
- :red_circle: Red — 90%+

## ETA

| Type | Default Colour | Description |
|------|---------------|-------------|
| `eta-session` | `#888888` | Time until session limit full (window average rate) |
| `eta-session-min` | `#888888` | Time until session limit full (short-term ~5min rate) |
| `eta-session-hr` | `#888888` | Time until session limit full (long-term ~60min rate) |
| `eta-weekly` | `#888888` | Time until weekly limit full (window average rate) |
| `eta-weekly-min` | `#888888` | Time until weekly limit full (short-term ~5min rate) |
| `eta-weekly-hr` | `#888888` | Time until weekly limit full (long-term ~60min rate) |

Labels and calculations auto-adapt to your subscription plan's actual window sizes. Type names (`eta-session`, `eta-weekly`) remain fixed for config compatibility.

Three forecasting methods:

- **Window average** — rate = utilization / time elapsed since window start
- **Short-term (~5min)** — rate from a rolling ~5 minute reference point
- **Long-term (~60min)** — rate from a rolling ~60 minute reference point

!!! info "ETA warm-up"

    ETA segments show nothing until enough data is collected — 2 minutes for short-term, 10 minutes for long-term.

## Environment

| Type | Default Colour | Description |
|------|---------------|-------------|
| `env` | `#c678dd` | Display the value of an environment variable |

Set the variable name in the `content` field:

```json
{ "type": "env", "content": "USER" }
{ "type": "env", "content": "VIRTUAL_ENV" }
```

Returns empty (hidden) if the variable is unset or empty.
