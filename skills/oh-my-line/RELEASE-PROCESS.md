# oh-my-line Release Process

Step-by-step instructions for releasing a new version of oh-my-line. Follow this exactly when the user asks to "do a release", "tag a version", "bump the version", or similar.

## Process

### Phase 1: Analyze

1. Read the current `VERSION` file
2. Run `git log v$(cat VERSION)..HEAD --oneline` to see what's changed since the last release
3. Determine the new version number:
   - **Patch** (0.10.x) — bug fixes, docs, CI changes
   - **Minor** (0.x.0) — new segments, new features, config changes
   - **Major** (x.0.0) — breaking changes
4. Draft the commit message: `release: vX.Y.Z` followed by a bulleted summary of changes
5. Present to the user:
   - Current version
   - Proposed new version with rationale
   - Summary of changes that will be in this release
   - The proposed commit message
6. **Wait for user confirmation before proceeding**

### Phase 2: Execute

1. Update `VERSION` file with the new version number
2. Update `CHANGELOG.md` — add a new entry at the top with the version, date, and changes
3. Check if `site/index.html` references the version (e.g. in install mockups) and update if needed
4. Commit with message: `release: vX.Y.Z`
5. **Do NOT push yet**

### Phase 3: Confirm

Present to the user:
- The version number: `vX.Y.Z`
- Summary of what was changed in the commit
- Explain: "When you confirm, I'll push the commit and tag. CI will automatically build binaries for all platforms and create the GitHub Release. The site will also redeploy."

**Wait for user confirmation before pushing.**

### Phase 4: Ship

1. Push the commit: `git push origin main`
2. Create and push the tag: `git tag vX.Y.Z && git push origin vX.Y.Z`
3. Watch the Release workflow to confirm it passes
4. Report the release URL: `https://github.com/jamesprnich/oh-my-line/releases/tag/vX.Y.Z`

## CI Workflows Triggered

- **Test** — runs on the push to main (Go tests, vet, build)
- **Deploy Site** — runs if site/docs/builder files changed
- **Release** — runs on the tag push (builds 4 platform binaries, creates GitHub Release)

## Files That May Need Updating

| File | What to update |
|------|---------------|
| `VERSION` | Always — the version number |
| `CHANGELOG.md` | Always — new release entry |
| `site/index.html` | If version appears in install mockups |
| `docs/docs/index.md` | Rarely — only if install instructions change |
