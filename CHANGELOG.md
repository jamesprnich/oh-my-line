# Changelog

## [0.11.0] — 2026-03-13

### Multi-account support

When multiple Claude Code accounts run simultaneously via `CLAUDE_CONFIG_DIR`, each account now gets fully isolated state — no cross-contamination of cache, OAuth tokens, cost tracking, or settings.

**Runtime isolation** — Automatic, zero-config. The engine derives an account key from `CLAUDE_CONFIG_DIR` via SHA-256 hash and creates per-account cache subdirectories:

- Burn rate tracking — per-account token consumption
- Rate limit cache — per-account usage limits and background fetches
- ETA projections — per-account time-to-limit forecasts
- Sparkline history — per-account time series data
- Cost tracking — per-account daily cost logs (`~/.oh-my-line/cost/acct-{key}/`)
- OAuth tokens — credentials resolved from account's own config directory
- Settings — effort level read from account's `settings.json`

Shared data (version cache, GitHub, Docker, command segments) remains in the base cache directory.

**Per-account config** — Optional. Place `oh-my-line.json` in the account's `CLAUDE_CONFIG_DIR` for a custom statusline layout per account. New config lookup order:

1. `{cwd}/oh-my-line.json` — project config (untrusted)
2. `{CLAUDE_CONFIG_DIR}/oh-my-line.json` — per-account config (trusted)
3. `~/.oh-my-line/config.json` — global config (trusted)

Step 2 is skipped when `CLAUDE_CONFIG_DIR` is unset or is the default (`~/.claude`). Default account behavior is completely unchanged — zero-migration upgrade.

### Tests

- 42 new tests: account key derivation, directory isolation, OAuth scoping, multi-account isolation for burn/cost/sparkline/rate-limits, per-account config loading with trust/priority/edge cases, security properties, integration tests
- Total test count: 269 (up from 227)

### Documentation

- Config reference: multi-account setup section with isolation tables, per-account config example
- Troubleshooting: multi-account issues table, per-account cache diagnostics, updated temp files tables (shared vs account-scoped), OAuth chain scoped to CLAUDE_CONFIG_DIR
- Skill file: updated config lookup table, install locations, cache diagnostics
- Release audit checklist: rewritten for Go engine (replaced stale bash-era references)

## [0.10.1] — 2026-03-09

### Fixes
- Add `max` effort level support for Opus 4.6 (red-orange indicator)
- Cap ETA at window remaining — shows `>5h`/`>7d` instead of absurd values
- Disable Nerd Font icons by default in builder presets
- Pin Nerd Font CDN to v3.3.0, add local font name fallback
- Fix builder Site nav link (`/site/` → `/`)
- Fix palette tab where "Git" also highlighted "GitHub"

### Infrastructure
- Add release workflow for automated binary builds on tag push
- Add RELEASE-PROCESS.md and RELEASE-AUDIT.md to skill sub-specs

## [0.10.0] — 2026-03-07

First public release. Compiled Go binary — no shell dependencies beyond `curl` and `bash` for the installer.

### Engine
- Compiled Go binary with 4-phase architecture: Config, Data, Render, Output
- 65 segment types across 14 categories
- Config format: `oh-my-line.json` with `oh-my-lines` array of line definitions
- Config lookup: project-level (`oh-my-line.json`) overrides global (`~/.oh-my-line/config.json`)
- Security: project-level configs cannot execute shell commands (trusted global config only)
- Background rate-limit fetch via detached `curl` subprocess (survives parent process exit)
- TTL-based cache layer for expensive lookups (API calls, file I/O)
- Debug logging — `OML_DEBUG=1` or `"debug": true` in config, writes to `~/.oh-my-line/debug.log`

### Segments

**Basic** — model name, directory, version

**Session** — session cost, duration, lines changed, API wait, total tokens, cache hit

**Git** — branch, dir@branch with diff stats

**Mode** — vim mode, worktree, agent, 200k warning, CC version, model ID

**Effort** — effort level (low/med/high) from env var or settings

**GitHub** — 9 types: PR, checks, reviews, notifications, comments, actions, issues, PR count, stars

**Tokens** — absolute and percentage context window usage, compact warning

**Product** — icon, label, tagline, rotating messages from [`.product.json`](https://github.com/jamesprnich/product-json-spec)

**Docker** — container health summary and DB container status

**Custom** — static text, separators, custom icons, cached shell commands, environment variables

**Burn rate** — per-minute/hour token burn, sparkline trends

**Rate limits** — session/weekly/extra/opus usage bars with dot (●○) or block (▓░) style, red/yellow status indicators, zoom bar at >95%

**Cost** — input cost estimates with model-aware pricing, burn rates, 7-day totals, daily sparkline

**ETA** — time-to-limit forecasts using 3 methods (window avg, 5min, 60min) across session and weekly windows

**Sparklines** — 8-bar Unicode sparklines for burn rate, context fill, and rate limits, plus 3-color target zone variants

### Usage proxy
- Aligned to [ai-usage-proxy-spec](https://github.com/jamesprnich/ai-usage-proxy-spec)
- Configure `usageProxy.claudeCode` (base URL) in config or `OML_USAGE_PROXY_CLAUDE_CODE` env var
- Fetches from `{baseURL}/api/proxy/anthropic/subscription/` — no auth headers needed
- Reduces API calls when running multiple concurrent Claude Code sessions

### Line types
- Normal lines with segments, configurable separators
- Background styles: solid, fade, gradient, neon
- Rule lines (repeating character)
- Spacer lines
- ASCII art lines

### Tooling
- Visual config builder (`builder.html`) with drag-and-drop, live preview, 13 presets
- One-command installer with version detection and upgrade support
- Go test suite: 192 tests
- CI: GitHub Actions with Go test/vet/build on Ubuntu and macOS
- Claude Code skill for configuration assistance (`/oh-my-line`)
- Nerd Font icon support with global toggle and per-segment override
- Prefix/suffix support on segments
