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
3. Credentials file (`~/.claude/.credentials.json`)
4. GNOME Keyring (Linux, via `secret-tool`)

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

The engine writes transient state to `/tmp/claude-{uid}/` (created with ownership and symlink checks):

| File | Purpose | TTL |
|------|---------|-----|
| `statusline-usage-cache.json` | Cached API/proxy response for rate limits | 5min |
| `statusline-usage-cache.json.tmp` | In-progress background fetch result | Transient |
| `statusline-usage-cache.json.pid` | PID of background fetch process | Transient |
| `statusline-usage-cache.json.err` | Connection error marker | 5min |
| `statusline-burn.dat` | Burn rate tracking (start epoch + tokens) | Session |
| `statusline-eta-short.dat` | Short-term ETA reference point | ~5min |
| `statusline-eta-long.dat` | Long-term ETA reference point | ~60min |
| `statusline-version.txt` | Cached Claude Code version string | 1hr |
| `statusline-cmd-*.dat` | Cached command segment outputs | Per config |
| `statusline-gh-*.json` | Cached GitHub CLI responses | 60s-5min |
| `statusline-docker.json` | Cached Docker container status | 30s |
| `statusline-spark-*.dat` | Sparkline bucket data | Session |

Debug log (separate location):

| File | Purpose |
|------|---------|
| `~/.oh-my-line/debug.log` | Debug log (when `OML_DEBUG=1` or `"debug": true`) — auto-truncated at 100KB |

To find your cache directory:

```bash
ls /tmp/claude-$(id -u)/statusline-* 2>/dev/null
```

To reset all cached state:

```bash
rm -f /tmp/claude-$(id -u)/statusline-*
```

## Running the test suite

```bash
cd engine && go test ./...
```

## Getting help

- [GitHub Issues](https://github.com/jamesprnich/oh-my-line/issues) — report bugs or request features
- [Builder](https://jamesprnich.github.io/oh-my-line/builder.html) — visual config builder
