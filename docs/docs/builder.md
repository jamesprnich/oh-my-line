# Builder

The builder is a visual config editor that runs entirely in your browser.

## Opening the builder

Visit [jamesprnich.github.io/oh-my-line/builder.html](https://jamesprnich.github.io/oh-my-line/builder.html).

## Interface

### Palette (left panel)

Segments are organised into 6 tabs:

- **Workspace** — text, sep, custom-icon, command, worktree, lines-changed, version
- **Product** — icon, label, tagline, message
- **Claude Code** — model, tokens, burn rates, rate limits, ETAs, costs, sparklines, effort, compact-warn
- **GitHub** — gh-pr, gh-checks, gh-reviews, gh-actions, gh-notifs, gh-issues, gh-pr-count, gh-stars
- **Docker** — docker, docker-db
- **Git** — branch, dir-branch, diff-stats

Each segment shows its name, a mock preview, and a description of what it displays.

Use the **search box** at the top to filter across all tabs by name, type, or description.

**Drag** any segment from the palette onto a line to add it.

### Live Preview (center)

Shows a real-time preview of your statusline using mock data. Updates instantly as you make changes.

The preview uses simulated values:

- Model: "Claude Opus 4.6"
- Tokens: 65k/200k
- Rate limits: sample progress bars
- Product fields: from your segment config

### Presets (above preview)

Click a preset button to load a pre-made collection of segments and lines. Presets replace your current layout entirely — a quick way to start from a known-good configuration and customise from there.

### Line Configuration (below preview)

Each line has its own tab. Click a line tab to configure:

- **Separator** — character between segments (e.g. `|`)
- **Separator style** — dim, colour
- **Background style** — none, solid, gradient, fade, neon
- **Background colour** — hex colour picker

Use **+ Add Line** to add more lines (up to 20). Click **Remove Line** to remove one.

### Segment Editor (right panel)

Click any segment chip in a line to open its editor:

- **Colour** — foreground hex colour (overrides default)
- **Background** — background hex colour
- **Bold** — toggle bold text
- **Dim** — toggle dimmed text

For product segments (`icon`, `label`, `tagline`):

- **Content** — inline value (e.g. an emoji, app name, or tagline)
- **Source** — set to `.product.json` to read from a shared product identity file

For `message` segments:

- **Messages** — one message per line, rotated over time
- **Interval** — seconds between rotation

For `command` segments:

- **Command** — the shell command to run
- **Cache** — seconds to cache output

For `text`, `sep`, `custom-icon` segments:

- **Content** — the text, character, or emoji to display

### JSON Output (bottom)

The generated `oh-my-line.json` config updates live as you make changes. Click **Copy** to copy it to your clipboard, then paste into your config file.

Use **Import JSON** to load an existing `oh-my-line.json` config.

## Workflow

1. **Open the builder** in your browser
2. **Drag segments** from the palette onto lines
3. **Reorder** segments by dragging within a line
4. **Style** individual segments by clicking them
5. **Configure** line-level options (background, separator)
6. **Try presets** to start from a pre-made layout
7. **Copy** the JSON output
8. **Paste** into your `oh-my-line.json`
9. The statusline updates automatically on the next refresh

## Tips

- Start with a preset, then customise from there
- Use `sep` segments or line-level separators (not both) to avoid double separators
- The `command` segment is powerful — show anything from weather to Docker stats
- GitHub segments require the `gh` CLI — the builder shows mock data
- Rate limit and ETA segments need an OAuth session to show real data; the builder shows mock data
- Keep statuslines to 1-2 lines for best readability
