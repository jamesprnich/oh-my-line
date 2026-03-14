# Troubleshooting

## Common issues

| Symptom | Cause | Fix |
|---------|-------|-----|
| Just shows model name | No config file found | Create `oh-my-line.json` in your project root or `~/.oh-my-line/config.json` |
| Rate limit segments empty | Not logged in via OAuth | Run `claude` and log in — the engine needs an OAuth session |
| Rate limit shows red ⚠ | Cannot connect to API or proxy | Check network, verify proxy URL, see debug log |
| Rate limit shows yellow ⚠ | Data is stale or proxy rate-limited | Data is still shown but may be outdated; will auto-recover |
| Rate limit shows "?" | API call failed or token expired | Check `curl` is installed; restart Claude Code to refresh OAuth |
| ETA segments not appearing | Not enough data yet | Wait 2min (short-term) or 10min (long-term) for warm-up |
| Command segment missing | Config is untrusted (project-level) | Command segments only run from trusted configs — see [security](#security-trusted-vs-untrusted-configs) |
| Command segment missing | Command errored or returned empty | Test manually: `bash -c "your command"` |
| Command segment stale | Cached output | Delete cache files (see below) or wait for cache TTL |
| Statusline blank/error | Invalid config file | Check JSON syntax with `jq . < oh-my-line.json` |
| Burn rate shows 0 | New context window just started | Wait for token consumption to begin |
| Statusline not updating | Claude Code not piping data | Restart Claude Code |
| Wrong project identity | Found a different config file | Check config lookup path (see below) |
| GitHub segments empty | Not in a git repo, no PR, or `gh` not installed | Run `gh auth status` to check |
| GitHub segments show stale data | Cached response | Delete cache files (see below) |

## Debugging config lookup

The engine checks for config files in this order:

```bash
# Check which config the engine would use
[ -f "$(pwd)/oh-my-line.json" ] && echo "Project: $(pwd)/oh-my-line.json"
[ -n "$CLAUDE_CONFIG_DIR" ] && [ -f "$CLAUDE_CONFIG_DIR/oh-my-line.json" ] && echo "Account: $CLAUDE_CONFIG_DIR/oh-my-line.json"
[ -f ~/.oh-my-line/config.json ] && echo "Global: ~/.oh-my-line/config.json"
```

## Validating your config

```bash
# Check JSON syntax
jq . < oh-my-line.json

# Check structure (new format)
jq '."oh-my-lines" | length' < oh-my-line.json

# List configured segment types
jq '[ ."oh-my-lines"[].segments[]?.type ] | unique' < oh-my-line.json
```

## Testing the engine directly

Pipe test data to the engine to verify it works:

```bash
echo '{"model":{"display_name":"Test Model"},"context_window":{"context_window_size":200000,"current_usage":{"input_tokens":50000,"cache_creation_input_tokens":0,"cache_read_input_tokens":0}},"cwd":"'$(pwd)'"}' | ~/.oh-my-line/oh-my-line
```

You should see your statusline rendered with ANSI colours.

## Testing without colours

Strip ANSI codes to see raw text output:

```bash
echo '...' | ~/.oh-my-line/oh-my-line | sed 's/\x1b\[[0-9;]*m//g'
```

## Rate limit / OAuth issues

Rate limit segments (`rate-session`, `rate-weekly`, `rate-extra`, `rate-opus`) and ETA segments require an active OAuth session from Claude Code.

The engine resolves the OAuth token using this chain:

1. `CLAUDE_CODE_OAUTH_TOKEN` environment variable
2. macOS Keychain (`security find-generic-password`)
3. Credentials file (`{CLAUDE_CONFIG_DIR}/.credentials.json` — defaults to `~/.claude/.credentials.json`)
4. GNOME Keyring (Linux, via `secret-tool`)

In multi-account setups, the credentials file path is scoped to the account's `CLAUDE_CONFIG_DIR`, so each account resolves its own OAuth token.

If rate limits show nothing:

1. Verify you're logged into Claude Code with an account (not just an API key)
2. Check that `curl` is installed
3. Try restarting Claude Code to refresh the OAuth session

## Usage proxy issues

If you've configured a usage proxy and rate-limit segments show ⚠ indicators:

1. **Red ⚠ (unreachable)** — the proxy URL cannot be reached
    - Verify the proxy is running: `curl -s YOUR_PROXY_URL | jq .`
    - Check the URL is correct in your config or env var
    - Enable debug logging to see the exact error: `export OML_DEBUG=1`

2. **Yellow ⚠ (stale)** — the proxy is responding but data is old
    - The proxy may be rate-limited upstream (check `meta.rate_limited` in response)
    - Data is still displayed — it's just not the latest

3. **Clear stale state** — delete cached files to force a fresh fetch:
    ```bash
    rm -f /tmp/claude-$(id -u)/statusline-usage-cache.json*
    ```

## Debug logging

Enable debug logging to troubleshoot rate limits, proxy connections, OAuth, and cache behavior:

```bash
export OML_DEBUG=1
```

Or set in config:

```json
{ "debug": true }
```

Logs are written to `~/.oh-my-line/debug.log`. Check the log after a statusline render:

```bash
tail -20 ~/.oh-my-line/debug.log
```

The log includes timestamps, component tags, and key=value data for each event. Auto-truncates at 100KB.

## Temp files

The engine writes transient state to `/tmp/claude-{uid}/` (created with ownership and symlink checks).

For multi-account setups (using `CLAUDE_CONFIG_DIR`), account-scoped files are stored in subdirectories:

```
/tmp/claude-{uid}/                         ← default account (shared + account-scoped)
/tmp/claude-{uid}/acct-{hash}/             ← additional account (account-scoped only)
```

**Shared files** (always in base directory):

| File | Purpose | TTL |
|------|---------|-----|
| `statusline-version.txt` | Cached Claude Code version string | 1hr |
| `statusline-cmd-*.dat` | Cached command segment outputs | Per config |
| `statusline-gh-*.json` | Cached GitHub CLI responses | 60s-5min |
| `statusline-docker.json` | Cached Docker container status | 30s |

**Account-scoped files** (in base dir for default account, `acct-{hash}/` for others):

| File | Purpose | TTL |
|------|---------|-----|
| `statusline-usage-cache.json` | Cached API/proxy response for rate limits | 5min |
| `statusline-usage-cache.json.tmp` | In-progress background fetch result | Transient |
| `statusline-usage-cache.json.pid` | PID of background fetch process | Transient |
| `statusline-usage-cache.json.err` | Connection error marker | 5min |
| `statusline-burn.dat` | Burn rate tracking (start epoch + tokens) | Session |
| `statusline-eta-short.dat` | Short-term ETA reference point | ~5min |
| `statusline-eta-long.dat` | Long-term ETA reference point | ~60min |
| `statusline-spark-*.dat` | Sparkline bucket data | Session |

Debug log (separate location):

| File | Purpose |
|------|---------|
| `~/.oh-my-line/debug.log` | Debug log (when `OML_DEBUG=1` or `"debug": true`) — auto-truncated at 100KB |

To find your cache directory:

```bash
# Default account
ls /tmp/claude-$(id -u)/statusline-* 2>/dev/null

# All accounts (including multi-account subdirs)
ls /tmp/claude-$(id -u)/acct-*/statusline-* 2>/dev/null
```

To reset all cached state:

```bash
# Default account only
rm -f /tmp/claude-$(id -u)/statusline-*

# All accounts
rm -rf /tmp/claude-$(id -u)/acct-*
```

## How the statusline refreshes

The statusline is **not** on a fixed timer. Claude Code calls the engine on every prompt submission and response — the `statusCommandHook` pipes fresh context data (model, tokens, cwd) to the engine each time. If you stop interacting with Claude Code, the statusline stays on its last rendered value.

**Rate limit data** works differently — it's too expensive to fetch from the API on every render. Instead:

1. The first render that needs rate limits launches a background `curl` subprocess
2. The subprocess writes its result to a cache file (TTL: 5 minutes)
3. Subsequent renders read from cache instantly
4. When the cache expires, the next render launches a fresh background fetch
5. While the new fetch is in-flight (~1-2 seconds), the old cached data is shown with a yellow ⚠ indicator

This means rate limit data may lag by a few seconds after a cache expiry, but it self-corrects without any user action.

## Why do I see a yellow ⚠ on rate limits?

The yellow ⚠ means the engine is showing **stale cached data** while a background refresh is in progress. This is normal and resolves itself within a few seconds.

You'll typically see it:

- When you first open a session (no cache yet, first fetch in progress)
- After 5 minutes of inactivity followed by a new prompt (cache expired)
- When switching between projects (different cache contexts)

**No action needed** — the indicator disappears once the background fetch completes. If it persists for more than 30 seconds, check your network connection or see [Debug logging](#debug-logging).

A **red ⚠** is different — it means the API or proxy is unreachable. See [Usage proxy issues](#usage-proxy-issues) for details.

## Security: trusted vs untrusted configs

The engine has three config paths with different trust levels:

| Path | Trusted | `command` segments |
|------|---------|-------------------|
| `{cwd}/oh-my-line.json` | No | Silently blocked |
| `{CLAUDE_CONFIG_DIR}/oh-my-line.json` | Yes | Executed |
| `~/.oh-my-line/config.json` | Yes | Executed |

**Why?** `command` segments run arbitrary shell commands (`bash -c "..."`). A cloned repo could include a malicious `oh-my-line.json` that executes code every time the statusline renders. Trusted configs — the global config and per-account configs — are safe because the user explicitly controls them.

Project-level configs work for everything else — layout, segments, styling, product identity. Only shell command execution is restricted.

If your command segments aren't appearing, check which config the engine is loading — if it's a project-level `oh-my-line.json`, move the command segments to a trusted config (`~/.oh-my-line/config.json` or your per-account config).

## Multi-account issues

If you run multiple Claude Code accounts via `CLAUDE_CONFIG_DIR`, each account gets isolated cache, OAuth, and cost tracking automatically. See [Multi-account setup](config.md#multi-account-setup) for details.

| Symptom | Cause | Fix |
|---------|-------|-----|
| Rate limits from wrong account | Cache cross-contamination (pre-v0.11) | Clear all cache: `rm -rf /tmp/claude-$(id -u)/statusline-* /tmp/claude-$(id -u)/acct-*` |
| Effort level wrong in second account | Reading default `~/.claude/settings.json` | Upgrade to v0.11+ — effort is now read from account's config dir |
| Cost data mixed between accounts | Shared cost directory (pre-v0.11) | Upgrade to v0.11+ — cost dirs are now per-account |

To check which account key the engine derived:

```bash
# The engine logs the account key in debug mode
OML_DEBUG=1 echo '{"model":{"display_name":"Test"},...}' | ~/.oh-my-line/oh-my-line 2>&1
tail -5 ~/.oh-my-line/debug.log
```

To list per-account cache directories:

```bash
ls -d /tmp/claude-$(id -u)/acct-* 2>/dev/null || echo "No multi-account dirs (single account)"
```

## Running the test suite

```bash
cd engine && go test ./...
```

## Getting help

- [GitHub Issues](https://github.com/jamesprnich/oh-my-line/issues) — report bugs or request features
- [Builder](https://jamesprnich.github.io/oh-my-line/builder.html) — visual config builder
