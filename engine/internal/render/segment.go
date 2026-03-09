package render

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

// DefaultColors maps segment types to their default hex colors.
var DefaultColors = map[string]string{
	"model":       "#0099ff",
	"dir":         "#2e9599",
	"version":     "#888888",
	"branch":      "#00a000",
	"dir-branch":  "#2e9599",
	"diff-stats":  "#00a000",
	"tokens":      "#ffb055",
	"pct-used":    "#00a000",
	"pct-remain":  "#2e9599",
	"burn-min":    "#e6c800",
	"burn-hr":     "#e6c800",
	"burn-spark":  "#e6c800",
	"cost":        "#e6c800",
	"cost-min":    "#e6c800",
	"cost-hr":     "#e6c800",
	"cost-7d":     "#e6c800",
	"cost-spark":  "#e6c800",
	"effort":      "#00a0a0",
	"text":        "#aaaaaa",
	"sep":         "#666666",
	"custom-icon": "#ffffff",
	"command":     "#aaaaaa",

	// Session
	"session-cost":     "#e6c800",
	"session-duration": "#888888",
	"lines-changed":    "#00a000",
	"cache-hit":        "#2e9599",
	"total-tokens":     "#ffb055",
	"api-wait":         "#888888",

	// Mode & identity
	"vim-mode":     "#e6c800",
	"worktree":     "#c678dd",
	"agent":        "#e6c800",
	"200k-warn":    "#ff5555",
	"cc-version":   "#888888",
	"model-id":     "#888888",

	// Product
	"icon":    "",
	"label":   "#ffffff",
	"tagline": "#888888",
	"message": "#888888",

	// Compact
	"compact-warn": "#ff5555",

	// Rate limits
	"rate-session": "#00a000",
	"rate-weekly":  "#00a000",
	"rate-extra":   "#00a000",
	"rate-opus":    "#7c3aed",
	"rate-spark":   "#00a000",

	// Sparklines
	"ctx-spark":   "#ffb055",
	"ctx-target":  "#ffb055",
	"rate-target": "#00a000",

	// ETAs
	"eta-session":     "#888888",
	"eta-session-min": "#888888",
	"eta-session-hr":  "#888888",
	"eta-weekly":      "#888888",
	"eta-weekly-min":  "#888888",
	"eta-weekly-hr":   "#888888",

	// GitHub
	"gh-pr":          "#2e9599",
	"gh-checks":      "#00a000",
	"gh-reviews":     "#00a000",
	"gh-actions":     "#00a000",
	"gh-notifs":      "#e6c800",
	"gh-issues":      "#ff5555",
	"gh-pr-count":    "#2e9599",
	"gh-pr-comments": "#888888",
	"gh-stars":       "#e6c800",

	// Docker
	"docker":    "#2e9599",
	"docker-db": "#2e9599",

	// Env
	"env": "#c678dd",
}

// RenderSegment renders a single segment and returns the ANSI string.
// Returns "" if the segment produces no output.
func RenderSegment(seg internal.SegmentConf, input *internal.Input, conf *internal.Config) string {
	color := resolveColor(seg)
	pre, post := buildStyle(seg, color)

	// Nerd font icon
	icon := ""
	showIcon := conf.NerdFont
	if seg.Icon != nil {
		showIcon = *seg.Icon
	}
	if showIcon {
		if nf, ok := NerdFontIcons[seg.Type]; ok {
			icon = pre + nf + post + " "
		}
	}

	out := renderSegmentContent(seg, input, conf, pre, post)
	if out == "" {
		return ""
	}

	// Prefix/suffix wrapping (excluded for text, sep, custom-icon)
	var result strings.Builder
	if seg.PadLeft > 0 {
		result.WriteString(strings.Repeat(" ", seg.PadLeft))
	}

	result.WriteString(icon)

	switch seg.Type {
	case "text", "sep", "custom-icon":
		// No prefix/suffix wrapping
	default:
		if seg.Prefix != "" && icon == "" {
			result.WriteString(DIM + seg.Prefix + RST + " ")
		}
	}

	result.WriteString(out)

	switch seg.Type {
	case "text", "sep", "custom-icon":
		// No suffix
	default:
		if seg.Suffix != "" {
			result.WriteString(" " + DIM + seg.Suffix + RST)
		}
	}

	if seg.PadRight > 0 {
		result.WriteString(strings.Repeat(" ", seg.PadRight))
	}

	return result.String()
}

func resolveColor(seg internal.SegmentConf) string {
	if seg.Style != nil && seg.Style.Color != "" {
		return seg.Style.Color
	}
	if c, ok := DefaultColors[seg.Type]; ok {
		return c
	}
	return ""
}

func buildStyle(seg internal.SegmentConf, color string) (pre, post string) {
	if color != "" {
		pre += HexFG(color)
	}
	if seg.Style != nil && seg.Style.Background != "" {
		pre += HexBG(seg.Style.Background)
	}
	if seg.Style != nil && seg.Style.Bold {
		pre += BOLD
	}
	if seg.Style != nil && seg.Style.Dim {
		pre += DIM
	}
	post = RST
	return
}

// renderSegmentContent dispatches to the appropriate segment renderer.
func renderSegmentContent(seg internal.SegmentConf, input *internal.Input, conf *internal.Config, pre, post string) string {
	switch seg.Type {
	case "model":
		return pre + input.Model.DisplayName + post

	case "dir":
		if input.CWD == "" {
			return ""
		}
		return pre + filepath.Base(input.CWD) + post

	case "version":
		if input.Version == "" {
			return ""
		}
		return pre + input.Version + post

	case "tokens":
		used := input.CurrentTokens()
		return pre + FormatTokens(used) + "/" + FormatTokens(input.ContextWindow.Size) + post

	case "pct-used":
		if input.ContextWindow.Size <= 0 {
			return ""
		}
		pct := input.CurrentTokens() * 100 / input.ContextWindow.Size
		return pre + fmt.Sprintf("%d%%", pct) + post + " " + DIM + "used" + RST

	case "pct-remain":
		if input.ContextWindow.Size <= 0 {
			return ""
		}
		pct := 100 - (input.CurrentTokens() * 100 / input.ContextWindow.Size)
		return pre + fmt.Sprintf("%d%%", pct) + post + " " + DIM + "remain" + RST

	case "effort":
		return renderEffort(pre, post)

	case "cost":
		// If runtime data has computed cost, use it (pricing model based)
		if conf.Runtime != nil && conf.Runtime.CostCtx != "" {
			return conf.Runtime.CostCtx
		}
		// Fallback to stdin cost
		if input.Cost.TotalCostUSD <= 0 {
			return ""
		}
		return pre + FormatCost(input.Cost.TotalCostUSD) + post

	case "text":
		if seg.Content == "" {
			return ""
		}
		return pre + seg.Content + post

	case "sep":
		content := seg.Content
		if content == "" {
			content = "|"
		}
		return DIM + content + RST

	case "custom-icon":
		if seg.Content == "" {
			return ""
		}
		return pre + seg.Content + post + " "

	// ── Phase 2 segments (stubs for dispatch, full impl later) ──

	case "dir-branch":
		return renderDirBranch(input, pre, post)

	case "branch":
		branch := DetectBranch(input.CWD)
		if branch == "" {
			return ""
		}
		return pre + branch + post

	case "diff-stats":
		return renderDiffStats(input)

	case "session-cost":
		return renderSessionCost(input)

	case "session-duration":
		return renderSessionDuration(input, pre, post)

	case "lines-changed":
		return renderLinesChanged(input)

	case "api-wait":
		return renderAPIWait(input)

	case "total-tokens":
		ti := input.ContextWindow.TotalInputTokens
		to := input.ContextWindow.TotalOutputTokens
		if ti == 0 && to == 0 {
			return ""
		}
		return pre + FormatTokens(ti) + "↑ " + FormatTokens(to) + "↓" + post

	case "cache-hit":
		total := input.ContextWindow.Usage.InputTokens +
			input.ContextWindow.Usage.CacheCreate +
			input.ContextWindow.Usage.CacheRead
		if total == 0 {
			return ""
		}
		pct := input.ContextWindow.Usage.CacheRead * 100 / total
		return pre + "cache " + fmt.Sprintf("%d%%", pct) + post

	case "vim-mode":
		return renderVimMode(input)

	case "worktree":
		if input.Worktree.Name == "" {
			return ""
		}
		wc := HexFG("#39d2c0")
		out := wc + input.Worktree.Name + RST
		if input.Worktree.Branch != "" {
			out += DIM + "@" + RST + wc + input.Worktree.Branch + RST
		}
		return out

	case "agent":
		if input.Agent.Name == "" {
			return ""
		}
		return DIM + "agent:" + RST + HexFG("#39d2c0") + input.Agent.Name + RST

	case "200k-warn":
		if !input.Exceeds200k {
			return ""
		}
		return HexFG("#ff5555") + "⚠ >200k" + RST

	case "cc-version":
		if input.Version == "" {
			return ""
		}
		return DIM + "v" + RST + input.Version

	case "model-id":
		if input.Model.ID == "" {
			return ""
		}
		return pre + input.Model.ID + post

	case "icon":
		v := seg.Content
		if v == "" {
			v = conf.MetaIcon
		}
		if v == "" {
			return ""
		}
		return v + " "

	case "label":
		v := seg.Content
		if v == "" {
			v = conf.MetaLabel
		}
		if v == "" {
			return ""
		}
		return pre + BOLD + v + post

	case "tagline":
		v := seg.Content
		if v == "" {
			v = conf.MetaTagline
		}
		if v == "" {
			return ""
		}
		return DIM + v + RST

	case "message":
		if conf.CurrentMessage == "" {
			return ""
		}
		return "         " + DIM + "— " + conf.CurrentMessage + " —" + RST

	case "compact-warn":
		if input.ContextWindow.Size <= 0 {
			return ""
		}
		threshold := seg.Threshold
		if threshold <= 0 {
			threshold = 10 // default: warn when <=10% remaining
		}
		remainPct := 100 - (input.CurrentTokens() * 100 / input.ContextWindow.Size)
		if remainPct > threshold {
			return ""
		}
		return pre + fmt.Sprintf("⚠ COMPACTING SOON %d%%", remainPct) + post

	// ── Runtime data segments (populated by datasources before rendering) ──

	case "burn-min", "burn-hr":
		if conf.Runtime == nil {
			return DIM + "—/" + strings.TrimPrefix(seg.Type, "burn-") + RST
		}
		warmup := seg.Warmup
		if warmup <= 0 {
			warmup = 30
		}
		switch seg.Type {
		case "burn-min":
			if !conf.Runtime.BurnHasData || conf.Runtime.BurnElapsed < warmup {
				return DIM + "—/min" + RST
			}
			return pre + FormatTokens(conf.Runtime.BurnRateMin) + "/min" + post
		case "burn-hr":
			if !conf.Runtime.BurnHasData || conf.Runtime.BurnElapsed < warmup {
				return DIM + "—/hr" + RST
			}
			return pre + FormatTokens(conf.Runtime.BurnRateHr) + "/hr" + post
		}
		return ""

	case "rate-session":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.RateSession

	case "rate-weekly":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.RateWeekly

	case "rate-extra":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.RateExtra

	case "rate-opus":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.RateOpus

	case "burn-spark":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.BurnSpark

	case "ctx-spark":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.CtxSpark

	case "rate-spark":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.RateSpark

	case "ctx-target":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.CtxTarget

	case "rate-target":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.RateTarget

	case "eta-session":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.ETASession

	case "eta-session-min":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.ETASessionMin

	case "eta-session-hr":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.ETASessionHr

	case "eta-weekly":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.ETAWeekly

	case "eta-weekly-min":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.ETAWeeklyMin

	case "eta-weekly-hr":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.ETAWeeklyHr

	case "cost-min":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.CostMin

	case "cost-hr":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.CostHr

	case "cost-7d":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.Cost7d

	case "cost-spark":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.CostSpark

	case "gh-pr", "gh-checks", "gh-reviews", "gh-actions", "gh-notifs",
		"gh-issues", "gh-pr-count", "gh-pr-comments", "gh-stars":
		if conf.Runtime == nil {
			return ""
		}
		switch seg.Type {
		case "gh-pr":
			return conf.Runtime.GhPR
		case "gh-checks":
			return conf.Runtime.GhChecks
		case "gh-reviews":
			return conf.Runtime.GhReviews
		case "gh-actions":
			return conf.Runtime.GhActions
		case "gh-notifs":
			return conf.Runtime.GhNotifs
		case "gh-issues":
			return conf.Runtime.GhIssues
		case "gh-pr-count":
			return conf.Runtime.GhPRCount
		case "gh-pr-comments":
			return conf.Runtime.GhPRComments
		case "gh-stars":
			return conf.Runtime.GhStars
		}
		return ""

	case "docker":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.Docker

	case "docker-db":
		if conf.Runtime == nil {
			return ""
		}
		return conf.Runtime.DockerDB

	case "env":
		if seg.Content == "" {
			return ""
		}
		val := os.Getenv(seg.Content)
		if val == "" {
			return ""
		}
		return pre + val + post

	case "command":
		if conf.Runtime == nil || !conf.Trusted {
			return ""
		}
		if seg.Content == "" {
			return ""
		}
		if conf.Runtime.CommandCache != nil {
			if v, ok := conf.Runtime.CommandCache[seg.Content]; ok {
				if v == "" {
					return ""
				}
				return pre + v + post
			}
		}
		return ""

	default:
		return ""
	}
}

// renderSessionCost renders cost with color thresholds (green <$1, yellow $1-$5, red >$5).
func renderSessionCost(input *internal.Input) string {
	c := input.Cost.TotalCostUSD
	if c <= 0 {
		return ""
	}
	var cc string
	switch {
	case c >= 5:
		cc = HexFG("#ff5555")
	case c >= 1:
		cc = HexFG("#e6c800")
	default:
		cc = HexFG("#00a000")
	}
	return cc + fmt.Sprintf("$%.2f", c) + RST
}

// renderSessionDuration formats ms as "5m", "1h 30m".
func renderSessionDuration(input *internal.Input, pre, post string) string {
	ms := input.Cost.TotalDurationMs
	if ms <= 0 {
		return ""
	}
	secs := ms / 1000
	if secs < 60 {
		return ""
	}
	mins := secs / 60
	hrs := mins / 60
	rem := mins % 60
	var dur string
	if hrs > 0 {
		dur = fmt.Sprintf("%dh %dm", hrs, rem)
	} else {
		dur = fmt.Sprintf("%dm", mins)
	}
	return pre + dur + post
}

// renderLinesChanged shows green adds, red removes (GitHub-style colors).
func renderLinesChanged(input *internal.Input) string {
	added := input.Cost.TotalLinesAdded
	removed := input.Cost.TotalLinesRemoved
	if added == 0 && removed == 0 {
		return ""
	}
	var out string
	if added > 0 {
		out += HexFG("#3fb950") + fmt.Sprintf("+%d", added) + RST
	}
	if added > 0 && removed > 0 {
		out += " "
	}
	if removed > 0 {
		out += HexFG("#f85149") + fmt.Sprintf("-%d", removed) + RST
	}
	return out
}

// renderAPIWait shows API wait percentage with color thresholds.
func renderAPIWait(input *internal.Input) string {
	if input.Cost.TotalAPIDurationMs <= 0 || input.Cost.TotalDurationMs <= 0 {
		return ""
	}
	secs := input.Cost.TotalDurationMs / 1000
	if secs < 60 {
		return ""
	}
	pct := input.Cost.TotalAPIDurationMs * 100 / input.Cost.TotalDurationMs
	var cc string
	switch {
	case pct > 70:
		cc = HexFG("#e6c800")
	case pct < 30:
		cc = DIM
	default:
		cc = ""
	}
	return cc + fmt.Sprintf("API %d%%", pct) + RST
}

// renderVimMode shows vim mode with color coding.
func renderVimMode(input *internal.Input) string {
	if input.Vim.Mode == "" {
		return ""
	}
	var cc string
	switch input.Vim.Mode {
	case "NORMAL":
		cc = HexFG("#00a000")
	case "INSERT":
		cc = HexFG("#e6c800")
	default:
		cc = HexFG("#dcdcdc")
	}
	return cc + input.Vim.Mode + RST
}

// renderEffort renders the effort segment.
func renderEffort(pre, post string) string {
	level := os.Getenv("CLAUDE_CODE_EFFORT_LEVEL")
	if level == "" {
		// Try reading from settings
		home, _ := os.UserHomeDir()
		if home != "" {
			data, err := os.ReadFile(filepath.Join(home, ".claude", "settings.json"))
			if err == nil {
				// Simple extraction — avoid json dependency for one field
				level = extractJSONString(string(data), "effortLevel")
			}
		}
	}
	switch level {
	case "low", "min":
		level = "low"
	case "med", "medium":
		level = "med"
	case "max":
		level = "max"
	default:
		level = "high"
	}
	var ec string
	switch level {
	case "low":
		ec = HexFG("#888888")
	case "med":
		ec = HexFG("#ffb055")
	case "max":
		ec = HexFG("#ff4500")
	default:
		ec = HexFG("#00a000")
	}
	return "effort: " + ec + level + RST
}

// extractJSONString does a simple key lookup in a JSON string. No json import needed.
func extractJSONString(data, key string) string {
	needle := `"` + key + `":`
	idx := strings.Index(data, needle)
	if idx < 0 {
		return ""
	}
	rest := strings.TrimSpace(data[idx+len(needle):])
	if len(rest) == 0 || rest[0] != '"' {
		return ""
	}
	rest = rest[1:]
	end := strings.Index(rest, `"`)
	if end < 0 {
		return ""
	}
	return rest[:end]
}

// DetectBranch reads the git HEAD file to detect the current branch.
func DetectBranch(cwd string) string {
	if cwd == "" {
		return ""
	}
	dir := cwd
	for {
		headPath := filepath.Join(dir, ".git", "HEAD")
		data, err := os.ReadFile(headPath)
		if err == nil {
			s := strings.TrimSpace(string(data))
			if strings.HasPrefix(s, "ref: refs/heads/") {
				return strings.TrimPrefix(s, "ref: refs/heads/")
			}
			if len(s) >= 8 {
				return s[:8] // detached HEAD
			}
			return ""
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

func renderDirBranch(input *internal.Input, pre, post string) string {
	if input.CWD == "" {
		return ""
	}
	dir := filepath.Base(input.CWD)
	branch := DetectBranch(input.CWD)
	if branch != "" {
		return pre + dir + post + DIM + "@" + RST + HexFG("#00a000") + branch + RST
	}
	return pre + dir + post
}

func renderDiffStats(input *internal.Input) string {
	added := input.Cost.TotalLinesAdded
	removed := input.Cost.TotalLinesRemoved
	if added == 0 && removed == 0 {
		return ""
	}
	out := DIM + "(" + RST
	if added > 0 {
		out += HexFG("#00a000") + fmt.Sprintf("+%d", added) + RST
	}
	if added > 0 && removed > 0 {
		out += DIM + " " + RST
	}
	if removed > 0 {
		out += HexFG("#ff5555") + fmt.Sprintf("-%d", removed) + RST
	}
	out += DIM + ")" + RST
	return out
}

// ResolveCurrentMessage picks the current rotating message based on epoch time.
func ResolveCurrentMessage(messages []string, interval int) string {
	if len(messages) == 0 {
		return ""
	}
	if interval <= 0 {
		interval = 300
	}
	now := time.Now().Unix()
	idx := (now / int64(interval)) % int64(len(messages))
	return messages[idx]
}
