# Custom Segments

oh-my-line ships with 65 built-in segment types. For anything else, use the `command` segment — it runs any shell command and displays the output. No forking or recompiling needed.

## Command segments

```json
{ "type": "command", "content": "date +%H:%M", "cache": 10 }
```

| Field | Default | Description |
|-------|---------|-------------|
| `content` | — | Shell command to execute |
| `cache` | `60` | Seconds to cache the output |
| `timeout` | `3` | Max execution time in seconds |

The command runs via `bash -c` (or `sh -c` if bash isn't available). If it errors or returns empty, the segment is silently hidden.

!!! warning "Trusted configs only"

    Command segments only run from trusted configs — `~/.oh-my-line/config.json` (global) and `{CLAUDE_CONFIG_DIR}/oh-my-line.json` (per-account). Project-level configs cannot execute shell commands — this prevents untrusted repos from running arbitrary code.

## Examples

**Current time:**
```json
{ "type": "command", "content": "date +%H:%M", "cache": 10 }
```

**Weather:**
```json
{ "type": "command", "content": "curl -s 'wttr.in/?format=%c%t'", "cache": 300 }
```

**Commit count:**
```json
{ "type": "command", "content": "git rev-list --count HEAD", "cache": 30 }
```

**Running containers:**
```json
{ "type": "command", "content": "docker ps -q | wc -l | tr -d ' '", "cache": 30 }
```

**Node version:**
```json
{ "type": "command", "content": "node -v", "cache": 3600 }
```

**Disk usage:**
```json
{ "type": "command", "content": "df -h / | awk 'NR==2{print $5}'", "cache": 60 }
```

## Other custom segment types

Beyond `command`, these built-in types cover most remaining use cases:

- **`text`** — static text: `{ "type": "text", "content": "oh-my-line" }`
- **`sep`** — visual separator: `{ "type": "sep", "content": "//" }`
- **`custom-icon`** — any emoji: `{ "type": "custom-icon", "content": "🔥" }`
- **`message`** — rotating messages: `{ "type": "message", "messages": ["A", "B", "C"], "interval": 300 }`
- **`env`** — environment variable value: `{ "type": "env", "content": "USER" }`

See [Segment Types](segments.md) for the full catalog.
