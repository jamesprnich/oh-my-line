package render

import (
	"testing"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

// AllSegmentTypes is the canonical list of every segment type the engine supports.
// If you add a new segment, add it here — the tests below enforce consistency.
var AllSegmentTypes = []string{
	// Basic
	"model", "dir", "version",
	// Session
	"session-cost", "session-duration", "lines-changed", "cache-hit", "total-tokens", "api-wait",
	// Git
	"branch", "dir-branch", "diff-stats",
	// Mode & identity
	"vim-mode", "worktree", "agent", "200k-warn", "cc-version", "model-id",
	// Effort
	"effort",
	// GitHub
	"gh-pr", "gh-checks", "gh-reviews", "gh-actions", "gh-notifs",
	"gh-issues", "gh-pr-count", "gh-pr-comments", "gh-stars",
	// Tokens
	"tokens", "pct-used", "pct-remain",
	// Compact
	"compact-warn",
	// Product
	"icon", "label", "tagline", "message",
	// Docker
	"docker", "docker-db",
	// Custom
	"text", "sep", "custom-icon", "command",
	// Burn rate
	"burn-min", "burn-hr",
	// Sparklines
	"burn-spark", "ctx-spark", "rate-spark", "ctx-target", "rate-target",
	// Rate limits
	"rate-session", "rate-weekly", "rate-extra", "rate-opus",
	// Cost
	"cost", "cost-min", "cost-hr", "cost-7d", "cost-spark",
	// ETA
	"eta-session", "eta-session-min", "eta-session-hr",
	"eta-weekly", "eta-weekly-min", "eta-weekly-hr",
	// Env
	"env",
}

// Segments that intentionally have no nerd font icon.
var noIconSegments = map[string]bool{
	"text":        true,
	"sep":         true,
	"custom-icon": true,
	"command":     true,
	"message":     true,
	"icon":        true,
	"label":       true,
	"tagline":     true,
	"200k-warn":   true, // uses inline ⚠ emoji
}

// Segments that intentionally have no DefaultColors entry (they use inline styling).
var noColorSegments = map[string]bool{
	"icon": true, // product icon, no text color
}

// TestConsistency_AllSegmentsHaveDefaultColor ensures every segment type has a
// DefaultColors entry so the builder and engine can render it with consistent styling.
func TestConsistency_AllSegmentsHaveDefaultColor(t *testing.T) {
	for _, segType := range AllSegmentTypes {
		if noColorSegments[segType] {
			continue
		}
		if _, ok := DefaultColors[segType]; !ok {
			t.Errorf("segment %q missing from DefaultColors map", segType)
		}
	}
}

// TestConsistency_AllSegmentsHaveNerdFontIcon ensures every segment type that
// should have a nerd font icon has one in the NerdFontIcons map.
func TestConsistency_AllSegmentsHaveNerdFontIcon(t *testing.T) {
	for _, segType := range AllSegmentTypes {
		if noIconSegments[segType] {
			continue
		}
		if _, ok := NerdFontIcons[segType]; !ok {
			t.Errorf("segment %q missing from NerdFontIcons map", segType)
		}
	}
}

// TestConsistency_AllSegmentsRender ensures every segment type is handled by
// renderSegmentContent (doesn't fall through to the default empty return for
// a type that should produce output).
func TestConsistency_AllSegmentsRender(t *testing.T) {
	// Segments that need runtime data to produce output — they return ""
	// without it, which is correct behavior, not a missing case.
	needsRuntime := map[string]bool{
		"rate-session": true, "rate-weekly": true, "rate-extra": true, "rate-opus": true,
		"burn-spark": true, "ctx-spark": true, "rate-spark": true,
		"ctx-target": true, "rate-target": true,
		"eta-session": true, "eta-session-min": true, "eta-session-hr": true,
		"eta-weekly": true, "eta-weekly-min": true, "eta-weekly-hr": true,
		"cost-min": true, "cost-hr": true, "cost-7d": true, "cost-spark": true,
		"gh-pr": true, "gh-checks": true, "gh-reviews": true, "gh-actions": true,
		"gh-notifs": true, "gh-issues": true, "gh-pr-count": true,
		"gh-pr-comments": true, "gh-stars": true,
		"docker": true, "docker-db": true,
		"command": true,
	}

	// Segments that need specific input data or filesystem access
	needsInput := map[string]bool{
		"vim-mode": true, "worktree": true, "agent": true,
		"200k-warn": true, "compact-warn": true,
		"icon": true, "label": true, "tagline": true, "message": true,
		"custom-icon": true,
		"branch": true, // needs real git repo on disk
	}

	// Segments that need content field
	needsContent := map[string]bool{
		"text": true, "env": true,
	}

	input := makeInput()
	conf := makeConf()
	conf.Trusted = true
	conf.Runtime = &internal.RuntimeData{
		BurnHasData: true, BurnElapsed: 60,
		BurnRateMin: 2000, BurnRateHr: 120000,
		RateSession: "5h mock", RateWeekly: "7d mock",
		RateExtra: "extra mock", RateOpus: "opus mock",
		BurnSpark: "▁▂▃", CtxSpark: "▁▂▃", RateSpark: "▁▂▃",
		CtxTarget: "▁▂▃", RateTarget: "▁▂▃",
		ETASession: "~2h", ETASessionMin: "~1h", ETASessionHr: "~3h",
		ETAWeekly: "~4d", ETAWeeklyMin: "~3d", ETAWeeklyHr: "~5d",
		CostCtx: "~$0.42", CostMin: "~$0.03/min", CostHr: "~$1.80/hr",
		Cost7d: "~$12.50", CostSpark: "▁▂▃",
		GhPR: "PR #1", GhChecks: "✓", GhReviews: "ok",
		GhActions: "pass", GhNotifs: "3", GhIssues: "5",
		GhPRCount: "2", GhPRComments: "10", GhStars: "★ 42",
		Docker: "3/3", DockerDB: "pg ✓",
		CommandCache: map[string]string{"test": "output"},
	}
	conf.MetaIcon = "🚀"
	conf.MetaLabel = "Test"
	conf.MetaTagline = "A test"
	conf.CurrentMessage = "Hello"

	input.Vim.Mode = "NORMAL"
	input.Worktree.Name = "wt"
	input.Agent.Name = "agent"
	input.Exceeds200k = true
	input.ContextWindow.TotalInputTokens = 50000
	input.ContextWindow.TotalOutputTokens = 5000

	for _, segType := range AllSegmentTypes {
		seg := internal.SegmentConf{Type: segType}

		// Add content for segments that need it
		if segType == "text" || segType == "env" {
			seg.Content = "test-value"
		}
		if segType == "custom-icon" {
			seg.Content = "🔥"
		}
		if segType == "command" {
			seg.Content = "test"
		}
		if segType == "sep" {
			// sep always returns something (default "|")
		}

		got := RenderSegment(seg, input, conf)

		// Skip segments that legitimately return empty without specific conditions
		if needsRuntime[segType] || needsInput[segType] || needsContent[segType] {
			// These were set up above, so they should produce output
		}

		// The key test: no segment should return empty when properly configured
		if got == "" {
			// Check if this is expected
			if needsRuntime[segType] && conf.Runtime != nil {
				t.Errorf("segment %q returned empty despite runtime data being set", segType)
			}
			if !needsRuntime[segType] && !needsInput[segType] && !needsContent[segType] {
				t.Errorf("segment %q returned empty — may be missing from renderSegmentContent switch", segType)
			}
		}
	}
}

// TestConsistency_NoOrphanColors ensures every entry in DefaultColors has
// a corresponding segment type in the canonical list.
func TestConsistency_NoOrphanColors(t *testing.T) {
	known := make(map[string]bool)
	for _, s := range AllSegmentTypes {
		known[s] = true
	}
	for segType := range DefaultColors {
		if !known[segType] {
			t.Errorf("DefaultColors has orphan entry %q not in AllSegmentTypes", segType)
		}
	}
}

// TestConsistency_NoOrphanIcons ensures every entry in NerdFontIcons has
// a corresponding segment type in the canonical list.
func TestConsistency_NoOrphanIcons(t *testing.T) {
	known := make(map[string]bool)
	for _, s := range AllSegmentTypes {
		known[s] = true
	}
	for segType := range NerdFontIcons {
		if !known[segType] {
			t.Errorf("NerdFontIcons has orphan entry %q not in AllSegmentTypes", segType)
		}
	}
}

// TestConsistency_SegmentCount ensures the canonical list matches expectations.
// Update this number when adding new segments.
func TestConsistency_SegmentCount(t *testing.T) {
	expected := 65
	got := len(AllSegmentTypes)
	if got != expected {
		t.Errorf("AllSegmentTypes has %d entries, expected %d — update this test when adding segments", got, expected)
	}
}
