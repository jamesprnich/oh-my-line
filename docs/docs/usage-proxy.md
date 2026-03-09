# Usage Proxy

When multiple Claude Code terminal sessions run simultaneously, each one polls the Anthropic usage API independently. With 10+ sessions this can trigger rate limiting (shown as a ⚠ icon). A usage proxy fetches once and serves all clients.

oh-my-line consumes the [ai-usage-proxy-spec](https://github.com/jamesprnich/ai-usage-proxy-spec) — a standardised API for cached AI provider usage data.

## Configuration

Set the **base URL** of your proxy. oh-my-line appends `/api/proxy/anthropic/subscription/` automatically per the spec.

### Environment variable (takes precedence)

```bash
export OML_USAGE_PROXY_CLAUDE_CODE="http://localhost:8787"
```

### Config file (`oh-my-line.json`)

```json
{
  "usageProxy": {
    "claudeCode": "http://localhost:8787"
  },
  "oh-my-lines": [...]
}
```

The environment variable always takes precedence over the config file. If neither is set, oh-my-line falls back to the direct Anthropic API with OAuth authentication.

## How It Works

1. oh-my-line checks `OML_USAGE_PROXY_CLAUDE_CODE` env var first, then `usageProxy.claudeCode` in config
2. If a proxy URL is set: `GET {baseURL}/api/proxy/anthropic/subscription/` with no auth headers
3. If no proxy: direct Anthropic API call with OAuth token and beta headers (existing behavior)
4. Proxy responses are cached for 60 seconds; direct API responses for 5 minutes
5. Background refresh keeps the cache warm without blocking renders

## Spec Alignment

oh-my-line expects the response shape defined in the [ai-usage-proxy-spec](https://github.com/jamesprnich/ai-usage-proxy-spec). Key details:

- **Endpoint**: `GET /api/proxy/anthropic/subscription/`
- **Auth**: None required
- **Success**: 200 with `five_hour`, `seven_day`, `seven_day_opus`, `extra_usage`, and `meta` fields
- **Stale data**: 200 with `meta.rate_limited: true` — client shows ⚠ indicator
- **Errors**: 502/503 with RFC 9457 Problem Details — client shows unreachable ⚠

See the [spec README](https://github.com/jamesprnich/ai-usage-proxy-spec) for the full response shape and field reference.
