# Contributing to oh-my-line

## Getting Started

```bash
git clone https://github.com/jamesprnich/oh-my-line.git
cd oh-my-line
```

**Requirements:** Go 1.23+, git.

## Project Structure

```
engine/                # Go engine
  cmd/oh-my-line/      # CLI entry point
  cmd/wasm/            # WASM build for browser use
  internal/            # Core packages
    config/            #   Config loader
    datasource/        #   Data source implementations
    render/            #   Segment rendering, ANSI, line building
    cache/             #   Cache layer
    debug/             #   Debug logging
    platform/          #   Platform-specific exec
skills/oh-my-line/     # Claude Code skill files (canonical source)
docs/                  # MkDocs Material documentation site
site/                  # Homepage (WASM-powered)
builder.html           # Visual config builder (static, no build step)
install.sh             # One-command installer
```

### Skill Files

The `skills/oh-my-line/` directory contains two files that teach Claude Code how to work with oh-my-line:

- **`SKILL.md`** — the main skill: install, customize, troubleshoot
- **`CONFIG-REFERENCE.md`** — segment type catalog, config options, line styles, Nerd Font icons

`RELEASE-AUDIT.md` is also in this directory but is a development-only checklist — not installed to user machines.

These are **source files** in the repo — this is where they're authored and versioned. But oh-my-line is a **global tool**, not a per-project dependency. Users install it once and it works across every project and terminal session. The skill files need to follow the same principle.

When the installer runs, it copies these files to `~/.claude/skills/oh-my-line/` — Claude Code's **global** skills directory. This means `/oh-my-line` commands are available in any Claude Code session, regardless of which project the user is working in. If the skills were installed at the project level (`.claude/skills/` inside a repo), they'd only work in that one repo, which defeats the purpose.

The update check in SKILL.md also re-fetches the skill file from GitHub on each invocation, so users pick up new instructions even between full upgrades.

If you're developing oh-my-line itself, you don't need to manually copy skill files to `~/.claude/skills/` — just run the installer locally and it handles everything.

## Architecture

The Go engine (`engine/`) follows this flow:

1. **Read stdin** — parse the JSON blob piped by Claude Code
2. **Config** — load and parse `oh-my-line.json` (`internal/config/`)
3. **Compute** — collect data from sources: git, docker, GitHub API, cost/rate files, etc. (`internal/datasource/`)
4. **Render** — resolve each configured segment to its output string, apply ANSI colors, icons, prefix/suffix (`internal/render/`)
5. **Output** — assemble the final line(s) and print the ANSI string

Key packages:
- `internal/render/segment.go` — segment type registry and render dispatch
- `internal/render/engine.go` — top-level render orchestration
- `internal/render/line.go` — multi-line layout and separator logic
- `internal/datasource/` — one file per data source (git, docker, cost, ratelimit, etc.)
- `internal/cache/` — TTL-based caching layer for expensive lookups
- `internal/config/` — config file discovery and parsing

## Adding a Segment

New segments are added in Go code. Two files are involved:

1. **Register the segment type** in `engine/internal/render/segment.go` — add a case to the segment registry with the type name and default color.

2. **Add a data source** (if needed) in `engine/internal/datasource/` — create a new file that implements the data-fetching logic. Existing files like `github.go`, `docker.go`, and `cost.go` are good references.

Key rules:
- Segment render functions receive config options and return a string (empty string = nothing to show)
- Use the cache package (`internal/cache/`) for expensive lookups (API calls, file I/O)
- Data sources should be testable in isolation — add a `_test.go` alongside your source file
- Register a default color for each new segment type so the builder and default configs work out of the box

## Running Tests

```bash
# All tests
cd engine && go test ./...

# Specific package
cd engine && go test ./internal/render/
cd engine && go test ./internal/datasource/

# Verbose output
cd engine && go test -v ./...

# Integration tests
cd engine && go test -run TestIntegration
```

Test files live alongside their source files (standard Go convention). See `engine/internal/render/segment_test.go` and `engine/internal/datasource/datasource_test.go` for examples.

## Linting

```bash
cd engine && go vet ./...
```

## Commit Messages

Follow conventional commits:

```
feat: add new segment type
fix: handle edge case in token formatting
chore: update CI config
```

## Pull Requests

1. Branch from `main`
2. Add tests for new functionality
3. Ensure `cd engine && go test ./...` passes
4. Ensure `cd engine && go vet ./...` passes
5. Keep PRs focused — one feature or fix per PR
