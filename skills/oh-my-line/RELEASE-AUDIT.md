# oh-my-line Release Audit

Pre-release quality and coverage checklist. Run this before tagging a release to catch stale docs, wrong counts, outdated paths, and inconsistencies across all project files.

## When to use

Read this file and follow the checklist when the user asks to:
- Review documentation quality or accuracy
- Audit for release readiness
- Check that docs match the code
- Run a pre-release checklist

## Automated checks first

Before the manual checklist, run the automated tools:

```bash
# 1. Go tests
cd engine && go test ./... -count=1

# 2. Go vet
cd engine && go vet ./...
```

Fix any failures before proceeding with the manual checklist below.

## How it works

Each section lists **what to check**, **where to check it**, and **what to look for**. Work through each section, report all findings, and fix issues only if the user asks.

---

## 1. Segment counts and types

The single source of truth is the `DefaultColors` map and `RenderSegment` switch in `engine/internal/render/segment.go`.

**Collect ground truth:**
```bash
grep -c '"#' engine/internal/render/segment.go  # approximate count from DefaultColors
```

**Cross-check these locations — all must show the correct total count:**

| File | What to check |
|------|--------------|
| `site/index.html` | Meta description, segment heading, feature cards |
| `docs/docs/segments.md` | Every registered type must appear in a table |
| `skills/oh-my-line/CONFIG-REFERENCE.md` | "Available Segment Types" table must list every type |
| `builder.html` | Every type must appear in a palette tab (`SEGMENT_TABS` array) |
| `CHANGELOG.md` | Segment list in latest release entry |

**Also verify:**
- No types exist in docs that don't exist in code (phantom types)
- No types exist in code that aren't documented anywhere (undocumented types)

---

## 2. Paths and install locations

**Ground truth:** `install.sh` defines where files go.

| Path | Check these files mention it correctly |
|------|---------------------------------------|
| `~/.oh-my-line/` | docs/index.md, SKILL.md, site/index.html |
| `~/.oh-my-line/oh-my-line` | SKILL.md install table, site/index.html install tree |
| `~/.claude/skills/oh-my-line/` | SKILL.md, CONTRIBUTING.md, install.sh, site/index.html install tree |
| `~/.oh-my-line/config.json` | docs/config.md, SKILL.md, site/index.html step 3 caption |
| `~/.claude/settings.json` | docs/index.md, SKILL.md |

**Watch for:**
- `.claude/skills/` without `~/` prefix (should be `~/.claude/skills/`)
- `~/.claude/` as a config location (wrong — global config is `~/.oh-my-line/config.json`)
- Project-level `.claude/skills/` references (skill files are global, not project-level)

---

## 3. Version number

**Ground truth:** `VERSION` file.

**Check these match:**
- `site/index.html` — agent install mockup output
- `CHANGELOG.md` — latest entry heading
- Any version references in docs

---

## 4. Architecture and engine phases

**Ground truth:** `engine/cmd/oh-my-line/main.go` and `engine/internal/render/engine.go`.

**Check these describe the same flow:**
- `CONTRIBUTING.md` — Architecture section
- `docs/docs/index.md` — Engine flow section

Both should match the actual flow in `main.go`: read stdin → parse → config → compute → render → output.

---

## 5. Skill files

**Ground truth:** `skills/oh-my-line/` directory.

**Check:**
- `install.sh` fetches all skill files (SKILL.md, CONFIG-REFERENCE.md)
- SKILL.md sub-specs table lists all companion files
- Skill files reference correct paths (`~/.claude/skills/oh-my-line/`)
- SKILL.md update check re-fetches from correct GitHub URL

---

## 6. Test infrastructure

**Ground truth:** `engine/**/*_test.go` files.

**Count tests:**
```bash
grep -r 'func Test' engine/ | wc -l
```

**Check these match:**
- `CHANGELOG.md` — test count
- `docs/docs/troubleshooting.md` — test command

---

## 7. Builder accuracy

**Ground truth:** `builder.html`.

**Check:**
- Preset count matches what docs/site say
- Tab names match `docs/docs/builder.md` palette section
- Every registered segment type appears in at least one palette tab
- `CONFIG_PRESETS` object key count matches any "N presets" claims

---

## 8. Terminology consistency

**Check for mixed terminology across all files:**

```bash
# Rate limit window names — should be "session"/"weekly", not "5h"/"7d" in segment names
grep -rni 'rate-5h\|rate-7d' skills/ docs/ site/ CONTRIBUTING.md

# "builder" should be "Config Builder" in user-facing text
grep -rni '"builder"' docs/ site/ README.md | grep -vi 'config builder\|builder\.html\|builder\.md'

# Product identity — no stale placeholders outside examples/
grep -rni 'ollama\|Run LLMs locally' docs/ site/ builder.html skills/
```

---

## 9. Documentation completeness and navigation

**Check mkdocs.yml nav matches actual files:**

```bash
# Every .md file in docs/docs/ should appear in mkdocs.yml nav
ls docs/docs/*.md | xargs -I{} basename {} | sort
grep '\.md' docs/mkdocs.yml | sed 's/.*: //' | sort
```

Any file present on disk but missing from nav is undiscoverable. Any nav entry without a matching file will break the docs build.

### Documentation quality (manual review)

For each documentation file, assess:

| Criterion | What to look for |
|-----------|-----------------|
| **Completeness** | Does it cover all features? Any gaps? |
| **Accuracy** | Do examples match actual behavior? |
| **Clarity** | Is it easy to follow? Are concepts explained before being used? |
| **Consistency** | Same terminology throughout? Same formatting patterns? |
| **Freshness** | Any references to removed features or old behavior? |
| **Audience** | Is it written for the right reader (user vs contributor vs AI agent)? |

**Files to review:**
- `CONTRIBUTING.md` — developer audience, should be actionable
- `CHANGELOG.md` — release notes, should be comprehensive
- `docs/docs/*.md` — user docs, should be clear and complete
- `skills/oh-my-line/*.md` — AI agent audience, should be precise and unambiguous

---

## 10. Example configs

**Ground truth:** `examples/` directory.

**Check:**
- All examples use `"oh-my-lines": []` format (not legacy `"statusline": { "lines": [] }`)
- JSON is valid and parseable
- Segment types used in examples actually exist in the engine
- Examples demonstrate realistic, useful configurations

---

## 11. Stale files and artifacts

**Check for files that shouldn't exist:**

```bash
# Backup/temp files
find . -name '*.bak' -o -name '*.tmp' -o -name '*.orig' -o -name '*~' | grep -v node_modules

# Files that should be gitignored but aren't
git ls-files -- engine/dist/ '*.wasm.bak'
```

---

## 12. CI/CD workflows

**Ground truth:** `.github/workflows/` directory.

**Check:**
- `test.yml` — runs Go tests, vet, CLI build, WASM build on ubuntu + macOS
- `deploy.yml` — builds site with WASM files included
- Go version in workflows matches `go.mod`
- No warnings in CI output (check latest run annotations)

---

## 13. Website (site/index.html)

**Check:**
- All numbers match ground truth (segment count, preset count)
- Install tree shows current file structure
- Skill path uses `~/` prefix
- Version in mockups matches VERSION file
- Config path references are correct
- Rate limit terminology is current
- No broken links (GitHub URLs, docs links, builder link)
- WASM paths use `engine/` (not `../engine/`)
- Analytics tracking is present

---

## Output format

Report findings as:

```
## Audit Results

### Passed
- [list of checks that passed]

### Issues Found
1. **[file:line]** — description of issue
2. **[file:line]** — description of issue

### Recommendations
- [optional suggestions for improvement]
```
