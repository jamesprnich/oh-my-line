package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

func makeInput() *internal.Input {
	in := &internal.Input{}
	in.Model.DisplayName = "Claude Sonnet 4"
	in.Model.ID = "claude-sonnet-4-20250514"
	in.CWD = "/home/user/my-project"
	in.Version = "1.0.32"
	in.ContextWindow.Size = 200000
	in.ContextWindow.Usage.InputTokens = 50000
	in.ContextWindow.Usage.CacheCreate = 10000
	in.ContextWindow.Usage.CacheRead = 20000
	in.Cost.TotalCostUSD = 2.50
	in.Cost.TotalDurationMs = 300000 // 5 min
	in.Cost.TotalAPIDurationMs = 150000
	in.Cost.TotalLinesAdded = 42
	in.Cost.TotalLinesRemoved = 10
	return in
}

func makeConf() *internal.Config {
	return &internal.Config{
		NerdFont: false,
		Presets:  make(map[string]internal.PresetConf),
	}
}

func TestRenderSegment_Model(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "model"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "Claude Sonnet 4") {
		t.Errorf("model segment should contain model name, got %q", got)
	}
}

func TestRenderSegment_Dir(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "dir"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "my-project") {
		t.Errorf("dir segment should contain dir name, got %q", got)
	}
}

func TestRenderSegment_DirEmpty(t *testing.T) {
	input := makeInput()
	input.CWD = ""
	conf := makeConf()
	seg := internal.SegmentConf{Type: "dir"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("dir segment with empty CWD should be empty, got %q", got)
	}
}

func TestRenderSegment_Version(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "version"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "1.0.32") {
		t.Errorf("version segment should contain version, got %q", got)
	}
}

func TestRenderSegment_Tokens(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "tokens"}
	got := RenderSegment(seg, input, conf)
	// 50000+10000+20000 = 80000 = "80k"
	if !strings.Contains(got, "80k") || !strings.Contains(got, "200k") {
		t.Errorf("tokens segment should contain 80k/200k, got %q", got)
	}
}

func TestRenderSegment_PctUsed(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "pct-used"}
	got := RenderSegment(seg, input, conf)
	// 80000/200000 = 40%
	if !strings.Contains(got, "40%") || !strings.Contains(got, "used") {
		t.Errorf("pct-used should contain 40%% used, got %q", got)
	}
}

func TestRenderSegment_PctRemain(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "pct-remain"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "60%") || !strings.Contains(got, "remain") {
		t.Errorf("pct-remain should contain 60%% remain, got %q", got)
	}
}

func TestRenderSegment_Text(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "text", Content: "hello"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "hello") {
		t.Errorf("text segment should contain content, got %q", got)
	}
}

func TestRenderSegment_TextEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "text"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("text segment with no content should be empty, got %q", got)
	}
}

func TestRenderSegment_Sep(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "sep", Content: "|"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "|") {
		t.Errorf("sep segment should contain separator, got %q", got)
	}
}

func TestRenderSegment_SepDefault(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "sep"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "|") {
		t.Errorf("sep segment with no content should use |, got %q", got)
	}
}

func TestRenderSegment_SessionCost(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "session-cost"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "$2.50") {
		t.Errorf("session-cost should show $2.50, got %q", got)
	}
}

func TestRenderSegment_SessionCostThresholds(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	// < $1 = green
	input.Cost.TotalCostUSD = 0.50
	seg := internal.SegmentConf{Type: "session-cost"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "\033[38;2;0;160;0m") {
		t.Errorf("session-cost <$1 should be green, got %q", got)
	}

	// $1-$5 = yellow
	input.Cost.TotalCostUSD = 3.00
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "\033[38;2;230;200;0m") {
		t.Errorf("session-cost $3 should be yellow, got %q", got)
	}

	// > $5 = red
	input.Cost.TotalCostUSD = 10.00
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "\033[38;2;255;85;85m") {
		t.Errorf("session-cost $10 should be red, got %q", got)
	}
}

func TestRenderSegment_SessionDuration(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "session-duration"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "5m") {
		t.Errorf("session-duration should show 5m, got %q", got)
	}
}

func TestRenderSegment_SessionDurationShort(t *testing.T) {
	input := makeInput()
	input.Cost.TotalDurationMs = 30000 // 30s
	conf := makeConf()
	seg := internal.SegmentConf{Type: "session-duration"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("session-duration < 60s should be empty, got %q", got)
	}
}

func TestRenderSegment_LinesChanged(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "lines-changed"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "+42") || !strings.Contains(got, "-10") {
		t.Errorf("lines-changed should show +42 -10, got %q", got)
	}
}

func TestRenderSegment_LinesChangedZero(t *testing.T) {
	input := makeInput()
	input.Cost.TotalLinesAdded = 0
	input.Cost.TotalLinesRemoved = 0
	conf := makeConf()
	seg := internal.SegmentConf{Type: "lines-changed"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("lines-changed with 0/0 should be empty, got %q", got)
	}
}

func TestRenderSegment_CacheHit(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "cache-hit"}
	got := RenderSegment(seg, input, conf)
	// cache_read=20000, total=80000 => 25%
	if !strings.Contains(got, "cache 25%") {
		t.Errorf("cache-hit should show cache 25%%, got %q", got)
	}
}

func TestRenderSegment_VimMode(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "vim-mode"}

	// No mode set
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("vim-mode with no mode should be empty, got %q", got)
	}

	// NORMAL = green
	input.Vim.Mode = "NORMAL"
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "NORMAL") || !strings.Contains(got, "\033[38;2;0;160;0m") {
		t.Errorf("vim-mode NORMAL should be green, got %q", got)
	}

	// INSERT = yellow
	input.Vim.Mode = "INSERT"
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "INSERT") || !strings.Contains(got, "\033[38;2;230;200;0m") {
		t.Errorf("vim-mode INSERT should be yellow, got %q", got)
	}
}

func TestRenderSegment_Worktree(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "worktree"}

	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("worktree with no name should be empty, got %q", got)
	}

	input.Worktree.Name = "feat-123"
	input.Worktree.Branch = "feature/123"
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "feat-123") || !strings.Contains(got, "feature/123") {
		t.Errorf("worktree should show name@branch, got %q", got)
	}
}

func TestRenderSegment_Agent(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "agent"}

	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("agent with no name should be empty, got %q", got)
	}

	input.Agent.Name = "test-runner"
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "agent:") || !strings.Contains(got, "test-runner") {
		t.Errorf("agent should show agent:name, got %q", got)
	}
}

func TestRenderSegment_200kWarn(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "200k-warn"}

	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("200k-warn without exceeds flag should be empty, got %q", got)
	}

	input.Exceeds200k = true
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "⚠ >200k") {
		t.Errorf("200k-warn should show warning, got %q", got)
	}
}

func TestRenderSegment_CCVersion(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "cc-version"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "v") || !strings.Contains(got, "1.0.32") {
		t.Errorf("cc-version should show v1.0.32, got %q", got)
	}
}

func TestRenderSegment_ModelID(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "model-id"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "claude-sonnet-4-20250514") {
		t.Errorf("model-id should show model ID, got %q", got)
	}
}

func TestRenderSegment_CompactWarn(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "compact-warn"}

	// 40% used = 60% remaining > 10% threshold
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("compact-warn should be empty with 60%% remaining, got %q", got)
	}

	// 95% used = 5% remaining
	input.ContextWindow.Usage.InputTokens = 180000
	input.ContextWindow.Usage.CacheCreate = 10000
	input.ContextWindow.Usage.CacheRead = 0
	got = RenderSegment(seg, input, conf)
	if !strings.Contains(got, "COMPACTING SOON") {
		t.Errorf("compact-warn should show warning at 5%% remaining, got %q", got)
	}
}

func TestRenderSegment_CustomIcon(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "custom-icon", Content: "🚀"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "🚀") {
		t.Errorf("custom-icon should show content, got %q", got)
	}
}

func TestRenderSegment_NerdFontIcon(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.NerdFont = true
	seg := internal.SegmentConf{Type: "model"}
	got := RenderSegment(seg, input, conf)
	// Should contain the nerd font icon for model
	if !strings.Contains(got, NerdFontIcons["model"]) {
		t.Errorf("model with nerdFont should include icon, got %q", got)
	}
}

func TestRenderSegment_IconOverride(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.NerdFont = true
	f := false
	seg := internal.SegmentConf{Type: "model", Icon: &f}
	got := RenderSegment(seg, input, conf)
	if strings.Contains(got, NerdFontIcons["model"]) {
		t.Errorf("model with icon=false should not include icon, got %q", got)
	}
}

func TestRenderSegment_PrefixSuffix(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "model", Prefix: "M:", Suffix: "!"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "M:") || !strings.Contains(got, "!") {
		t.Errorf("model with prefix/suffix should include them, got %q", got)
	}
}

func TestRenderSegment_Padding(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "text", Content: "x", PadLeft: 2, PadRight: 3}
	got := RenderSegment(seg, input, conf)
	if !strings.HasPrefix(got, "  ") {
		t.Errorf("text with padLeft=2 should start with 2 spaces, got %q", got)
	}
	if !strings.HasSuffix(got, "   ") {
		t.Errorf("text with padRight=3 should end with 3 spaces, got %q", got)
	}
}

func TestRenderSegment_CustomStyle(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{
		Type: "text",
		Content: "styled",
		Style: &internal.Style{Color: "#ff0000", Bold: true},
	}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "\033[38;2;255;0;0m") {
		t.Errorf("custom color should be applied, got %q", got)
	}
	if !strings.Contains(got, BOLD) {
		t.Errorf("bold should be applied, got %q", got)
	}
}

func TestRenderSegment_RuntimeData(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		RateSession: "5h ●●●○○○ 45%",
		GhPR:        "PR #123 open",
		Docker:      "3/3 up",
	}

	tests := []struct {
		segType string
		want    string
	}{
		{"rate-session", "5h ●●●○○○ 45%"},
		{"gh-pr", "PR #123 open"},
		{"docker", "3/3 up"},
	}
	for _, tt := range tests {
		seg := internal.SegmentConf{Type: tt.segType}
		got := RenderSegment(seg, input, conf)
		if got != tt.want {
			t.Errorf("segment %s = %q, want %q", tt.segType, got, tt.want)
		}
	}
}

func TestRenderSegment_RuntimeNil(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = nil

	runtimeSegments := []string{
		"rate-session", "rate-weekly", "rate-extra", "rate-opus",
		"burn-spark", "ctx-spark", "rate-spark", "ctx-target", "rate-target",
		"eta-session", "eta-weekly",
		"cost-min", "cost-hr", "cost-7d", "cost-spark",
		"gh-pr", "gh-checks", "gh-reviews", "gh-actions",
		"docker", "docker-db",
	}
	for _, segType := range runtimeSegments {
		seg := internal.SegmentConf{Type: segType}
		got := RenderSegment(seg, input, conf)
		if got != "" && !strings.Contains(got, "—") {
			t.Errorf("segment %s with nil runtime should be empty or dash, got %q", segType, got)
		}
	}
}

func TestRenderSegment_Unknown(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "nonexistent"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("unknown segment type should return empty, got %q", got)
	}
}

func TestRenderSegment_Label(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.MetaLabel = "My App"
	seg := internal.SegmentConf{Type: "label"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "My App") || !strings.Contains(got, BOLD) {
		t.Errorf("label should show bold name, got %q", got)
	}
}

func TestRenderSegment_Tagline(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.MetaTagline = "Build cool stuff"
	seg := internal.SegmentConf{Type: "tagline"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "Build cool stuff") || !strings.Contains(got, DIM) {
		t.Errorf("tagline should show dim text, got %q", got)
	}
}

func TestRenderSegment_Message(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.CurrentMessage = "Hello World"
	seg := internal.SegmentConf{Type: "message"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "Hello World") {
		t.Errorf("message should show current message, got %q", got)
	}
}

func TestRenderSegment_MessageEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	seg := internal.SegmentConf{Type: "message"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("message with no current message should be empty, got %q", got)
	}
}

func TestRenderSegment_Env(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	// Set a test env var
	t.Setenv("OML_TEST_VAR", "hello-world")

	seg := internal.SegmentConf{Type: "env", Content: "OML_TEST_VAR"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "hello-world") {
		t.Errorf("env segment should contain value, got %q", got)
	}
}

func TestRenderSegment_EnvEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	seg := internal.SegmentConf{Type: "env", Content: "OML_NONEXISTENT_VAR_12345"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("env segment with unset var should be empty, got %q", got)
	}
}

func TestRenderSegment_EnvNoContent(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	seg := internal.SegmentConf{Type: "env"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("env segment with no content should be empty, got %q", got)
	}
}

// ── Command segment ──

func TestRenderSegment_Command(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Trusted = true
	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"date +%H:%M": "14:30",
		},
	}

	seg := internal.SegmentConf{Type: "command", Content: "date +%H:%M"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "14:30") {
		t.Errorf("command segment should contain cached output, got %q", got)
	}
}

func TestRenderSegment_CommandEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Trusted = true
	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"failing-cmd": "",
		},
	}

	seg := internal.SegmentConf{Type: "command", Content: "failing-cmd"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("command with empty output should be empty, got %q", got)
	}
}

func TestRenderSegment_CommandUntrusted(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Trusted = false
	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{"date": "14:30"},
	}

	seg := internal.SegmentConf{Type: "command", Content: "date"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("command segment in untrusted config should be empty, got %q", got)
	}
}

func TestRenderSegment_CommandNoContent(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Trusted = true
	conf.Runtime = &internal.RuntimeData{CommandCache: map[string]string{}}

	seg := internal.SegmentConf{Type: "command"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("command with no content should be empty, got %q", got)
	}
}

func TestRenderSegment_CommandNotInCache(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Trusted = true
	conf.Runtime = &internal.RuntimeData{CommandCache: map[string]string{}}

	seg := internal.SegmentConf{Type: "command", Content: "uncached-cmd"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("command not in cache should be empty, got %q", got)
	}
}

func TestRenderSegment_CommandNilRuntime(t *testing.T) {
	// SECURITY: When Runtime is nil (as it would be when computeRuntimeData
	// skips command execution for untrusted configs), commands must be blocked
	// even if Trusted is somehow true.
	input := makeInput()
	conf := makeConf()
	conf.Trusted = true
	conf.Runtime = nil

	seg := internal.SegmentConf{Type: "command", Content: "echo exploit"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("SECURITY: nil Runtime must block command segments, got %q", got)
	}
}

// ── All GitHub segment dispatch ──

func TestRenderSegment_AllGitHubSegments(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		GhPR:         "PR #42 open",
		GhChecks:     "✓ 8/8",
		GhReviews:    "2 approved",
		GhActions:    "passing",
		GhNotifs:     "3 notifs",
		GhIssues:     "5 open",
		GhPRCount:    "3 PRs",
		GhPRComments: "12 comments",
		GhStars:      "★ 142",
	}

	tests := []struct {
		segType string
		want    string
	}{
		{"gh-pr", "PR #42 open"},
		{"gh-checks", "✓ 8/8"},
		{"gh-reviews", "2 approved"},
		{"gh-actions", "passing"},
		{"gh-notifs", "3 notifs"},
		{"gh-issues", "5 open"},
		{"gh-pr-count", "3 PRs"},
		{"gh-pr-comments", "12 comments"},
		{"gh-stars", "★ 142"},
	}
	for _, tt := range tests {
		seg := internal.SegmentConf{Type: tt.segType}
		got := RenderSegment(seg, input, conf)
		if got != tt.want {
			t.Errorf("segment %s = %q, want %q", tt.segType, got, tt.want)
		}
	}
}

func TestRenderSegment_AllGitHubSegmentsEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{}

	ghTypes := []string{"gh-pr", "gh-checks", "gh-reviews", "gh-actions",
		"gh-notifs", "gh-issues", "gh-pr-count", "gh-pr-comments", "gh-stars"}
	for _, segType := range ghTypes {
		seg := internal.SegmentConf{Type: segType}
		got := RenderSegment(seg, input, conf)
		if got != "" {
			t.Errorf("segment %s with empty runtime should be empty, got %q", segType, got)
		}
	}
}

// ── All Docker segment dispatch ──

func TestRenderSegment_DockerDB(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		DockerDB: "pg ✓",
	}

	seg := internal.SegmentConf{Type: "docker-db"}
	got := RenderSegment(seg, input, conf)
	if got != "pg ✓" {
		t.Errorf("docker-db = %q, want %q", got, "pg ✓")
	}
}

func TestRenderSegment_DockerEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{}

	for _, segType := range []string{"docker", "docker-db"} {
		seg := internal.SegmentConf{Type: segType}
		got := RenderSegment(seg, input, conf)
		if got != "" {
			t.Errorf("segment %s with empty runtime should be empty, got %q", segType, got)
		}
	}
}

// ── All rate limit segments dispatch ──

func TestRenderSegment_AllRateLimitSegments(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		RateSession: "5h ●●●○○○ 45%",
		RateWeekly:  "7d ●●○○○○ 28%",
		RateExtra:   "extra ○○○○○○ $0/$50",
		RateOpus:    "opus ●○○○○○ 10%",
	}

	tests := []struct {
		segType string
		want    string
	}{
		{"rate-session", "5h ●●●○○○ 45%"},
		{"rate-weekly", "7d ●●○○○○ 28%"},
		{"rate-extra", "extra ○○○○○○ $0/$50"},
		{"rate-opus", "opus ●○○○○○ 10%"},
	}
	for _, tt := range tests {
		seg := internal.SegmentConf{Type: tt.segType}
		got := RenderSegment(seg, input, conf)
		if got != tt.want {
			t.Errorf("segment %s = %q, want %q", tt.segType, got, tt.want)
		}
	}
}

// ── All sparkline segments dispatch ──

func TestRenderSegment_AllSparklineSegments(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		BurnSpark:  "▁▂▃▄▅▆▇█",
		CtxSpark:   "▁▁▂▃▃▄▅▆",
		RateSpark:  "▁▁▁▂▂▃▃▄",
		CtxTarget:  "▁▂▃▄",
		RateTarget: "▂▃▄▅",
	}

	tests := []struct {
		segType string
		want    string
	}{
		{"burn-spark", "▁▂▃▄▅▆▇█"},
		{"ctx-spark", "▁▁▂▃▃▄▅▆"},
		{"rate-spark", "▁▁▁▂▂▃▃▄"},
		{"ctx-target", "▁▂▃▄"},
		{"rate-target", "▂▃▄▅"},
	}
	for _, tt := range tests {
		seg := internal.SegmentConf{Type: tt.segType}
		got := RenderSegment(seg, input, conf)
		if got != tt.want {
			t.Errorf("segment %s = %q, want %q", tt.segType, got, tt.want)
		}
	}
}

// ── All ETA segments dispatch ──

func TestRenderSegment_AllETASegments(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		ETASession:    "~2h 15m",
		ETASessionMin: "~1h 30m",
		ETASessionHr:  "~3h 0m",
		ETAWeekly:     "~4d 8h",
		ETAWeeklyMin:  "~3d 12h",
		ETAWeeklyHr:   "~5d 2h",
	}

	tests := []struct {
		segType string
		want    string
	}{
		{"eta-session", "~2h 15m"},
		{"eta-session-min", "~1h 30m"},
		{"eta-session-hr", "~3h 0m"},
		{"eta-weekly", "~4d 8h"},
		{"eta-weekly-min", "~3d 12h"},
		{"eta-weekly-hr", "~5d 2h"},
	}
	for _, tt := range tests {
		seg := internal.SegmentConf{Type: tt.segType}
		got := RenderSegment(seg, input, conf)
		if got != tt.want {
			t.Errorf("segment %s = %q, want %q", tt.segType, got, tt.want)
		}
	}
}

// ── All cost segments dispatch ──

func TestRenderSegment_AllCostSegments(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		CostCtx:   "~$0.42",
		CostMin:   "~$0.03/min",
		CostHr:    "~$1.80/hr",
		Cost7d:    "~$12.50 (7d)",
		CostSpark: "▁▂▃▄▅▆▇",
	}

	tests := []struct {
		segType string
		want    string
	}{
		{"cost", "~$0.42"},
		{"cost-min", "~$0.03/min"},
		{"cost-hr", "~$1.80/hr"},
		{"cost-7d", "~$12.50 (7d)"},
		{"cost-spark", "▁▂▃▄▅▆▇"},
	}
	for _, tt := range tests {
		seg := internal.SegmentConf{Type: tt.segType}
		got := RenderSegment(seg, input, conf)
		if got != tt.want {
			t.Errorf("segment %s = %q, want %q", tt.segType, got, tt.want)
		}
	}
}

// ── Cost fallback to stdin ──

func TestRenderSegment_CostFallbackToStdin(t *testing.T) {
	input := makeInput()
	input.Cost.TotalCostUSD = 3.75
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{CostCtx: ""} // empty runtime cost

	seg := internal.SegmentConf{Type: "cost"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "3.75") {
		t.Errorf("cost should fall back to stdin cost, got %q", got)
	}
}

func TestRenderSegment_CostZero(t *testing.T) {
	input := makeInput()
	input.Cost.TotalCostUSD = 0
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{}

	seg := internal.SegmentConf{Type: "cost"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("cost with zero should be empty, got %q", got)
	}
}

// ── Burn rate segments ──

func TestRenderSegment_BurnMinWithData(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		BurnHasData: true,
		BurnElapsed: 60,
		BurnRateMin: 2100,
	}

	seg := internal.SegmentConf{Type: "burn-min"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "2k/min") {
		t.Errorf("burn-min should show rate, got %q", got)
	}
}

func TestRenderSegment_BurnHrWithData(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		BurnHasData: true,
		BurnElapsed: 60,
		BurnRateHr:  126000,
	}

	seg := internal.SegmentConf{Type: "burn-hr"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "126k/hr") {
		t.Errorf("burn-hr should show rate, got %q", got)
	}
}

func TestRenderSegment_BurnMinWarmup(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		BurnHasData: true,
		BurnElapsed: 10, // under default 30s warmup
		BurnRateMin: 5000,
	}

	seg := internal.SegmentConf{Type: "burn-min"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "—") {
		t.Errorf("burn-min during warmup should show dash, got %q", got)
	}
}

func TestRenderSegment_BurnMinCustomWarmup(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Runtime = &internal.RuntimeData{
		BurnHasData: true,
		BurnElapsed: 45, // over custom 10s warmup
		BurnRateMin: 3000,
	}

	seg := internal.SegmentConf{Type: "burn-min", Warmup: 10}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "3k/min") {
		t.Errorf("burn-min past custom warmup should show rate, got %q", got)
	}
}

// ── Mode segments with varied inputs ──

func TestRenderSegment_VimModeInsert(t *testing.T) {
	input := makeInput()
	input.Vim.Mode = "INSERT"
	conf := makeConf()

	seg := internal.SegmentConf{Type: "vim-mode"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "INSERT") {
		t.Errorf("vim-mode should show INSERT, got %q", got)
	}
}

func TestRenderSegment_VimModeVisual(t *testing.T) {
	input := makeInput()
	input.Vim.Mode = "VISUAL"
	conf := makeConf()

	seg := internal.SegmentConf{Type: "vim-mode"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "VISUAL") {
		t.Errorf("vim-mode should show VISUAL, got %q", got)
	}
}

func TestRenderSegment_WorktreeWithBranch(t *testing.T) {
	input := makeInput()
	input.Worktree.Name = "feature-x"
	input.Worktree.Branch = "main"
	conf := makeConf()

	seg := internal.SegmentConf{Type: "worktree"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "feature-x") || !strings.Contains(got, "main") {
		t.Errorf("worktree should show name and branch, got %q", got)
	}
}

func TestRenderSegment_WorktreeNoBranch(t *testing.T) {
	input := makeInput()
	input.Worktree.Name = "hotfix"
	input.Worktree.Branch = ""
	conf := makeConf()

	seg := internal.SegmentConf{Type: "worktree"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "hotfix") {
		t.Errorf("worktree should show name, got %q", got)
	}
	if strings.Contains(got, "@") {
		t.Errorf("worktree without branch should not contain @, got %q", got)
	}
}

// ── Product segments with meta fallback ──

func TestRenderSegment_IconFromMeta(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.MetaIcon = "🚀"

	seg := internal.SegmentConf{Type: "icon"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "🚀") {
		t.Errorf("icon should use MetaIcon, got %q", got)
	}
}

func TestRenderSegment_IconContentOverridesMeta(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.MetaIcon = "🚀"

	seg := internal.SegmentConf{Type: "icon", Content: "🔥"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "🔥") {
		t.Errorf("icon content should override meta, got %q", got)
	}
}

func TestRenderSegment_IconEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	seg := internal.SegmentConf{Type: "icon"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("icon with no content or meta should be empty, got %q", got)
	}
}

func TestRenderSegment_LabelFromMeta(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.MetaLabel = "MyApp"

	seg := internal.SegmentConf{Type: "label"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "MyApp") {
		t.Errorf("label should use MetaLabel, got %q", got)
	}
}

func TestRenderSegment_LabelEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	seg := internal.SegmentConf{Type: "label"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("label with no content or meta should be empty, got %q", got)
	}
}

func TestRenderSegment_TaglineFromMeta(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.MetaTagline = "Fast and simple"

	seg := internal.SegmentConf{Type: "tagline"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "Fast and simple") {
		t.Errorf("tagline should use MetaTagline, got %q", got)
	}
}

func TestRenderSegment_TaglineEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()

	seg := internal.SegmentConf{Type: "tagline"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("tagline with no content or meta should be empty, got %q", got)
	}
}

// ── Compact warn thresholds ──

func TestRenderSegment_CompactWarnCustomThreshold(t *testing.T) {
	input := makeInput()
	input.ContextWindow.Size = 200000
	input.ContextWindow.Usage.InputTokens = 170000
	input.ContextWindow.Usage.CacheCreate = 0
	input.ContextWindow.Usage.CacheRead = 0
	conf := makeConf()

	// 15% remaining, default threshold 10% — should NOT trigger
	seg := internal.SegmentConf{Type: "compact-warn"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("compact-warn at 15%% remaining with 10%% threshold should be empty, got %q", got)
	}

	// Custom threshold 20% — should trigger
	seg2 := internal.SegmentConf{Type: "compact-warn", Threshold: 20}
	got2 := RenderSegment(seg2, input, conf)
	if !strings.Contains(got2, "COMPACTING") {
		t.Errorf("compact-warn at 15%% remaining with 20%% threshold should trigger, got %q", got2)
	}
}

// ── API wait thresholds ──

func TestRenderSegment_APIWaitHigh(t *testing.T) {
	input := makeInput()
	input.Cost.TotalDurationMs = 600000
	input.Cost.TotalAPIDurationMs = 480000 // 80%
	conf := makeConf()

	seg := internal.SegmentConf{Type: "api-wait"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "80%") {
		t.Errorf("api-wait should show 80%%, got %q", got)
	}
}

func TestRenderSegment_APIWaitTooShort(t *testing.T) {
	input := makeInput()
	input.Cost.TotalDurationMs = 30000 // 30s, under 60s threshold
	input.Cost.TotalAPIDurationMs = 15000
	conf := makeConf()

	seg := internal.SegmentConf{Type: "api-wait"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("api-wait under 60s should be empty, got %q", got)
	}
}

// ── Total tokens ──

func TestRenderSegment_TotalTokensZero(t *testing.T) {
	input := makeInput()
	input.ContextWindow.TotalInputTokens = 0
	input.ContextWindow.TotalOutputTokens = 0
	conf := makeConf()

	seg := internal.SegmentConf{Type: "total-tokens"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("total-tokens with zero should be empty, got %q", got)
	}
}

// ── Session duration hours ──

func TestRenderSegment_SessionDurationHours(t *testing.T) {
	input := makeInput()
	input.Cost.TotalDurationMs = 5400000 // 90 min
	conf := makeConf()

	seg := internal.SegmentConf{Type: "session-duration"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "1h 30m") {
		t.Errorf("session-duration should show hours, got %q", got)
	}
}

// ── 200k warn off ──

func TestRenderSegment_200kWarnOff(t *testing.T) {
	input := makeInput()
	input.Exceeds200k = false
	conf := makeConf()

	seg := internal.SegmentConf{Type: "200k-warn"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("200k-warn should be empty when not exceeding, got %q", got)
	}
}

// ── DirBranch ──

func TestRenderSegment_DirBranchNoCWD(t *testing.T) {
	input := makeInput()
	input.CWD = ""
	conf := makeConf()

	seg := internal.SegmentConf{Type: "dir-branch"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("dir-branch with no CWD should be empty, got %q", got)
	}
}

// ── DiffStats zero ──

func TestRenderSegment_DiffStatsZero(t *testing.T) {
	input := makeInput()
	input.Cost.TotalLinesAdded = 0
	input.Cost.TotalLinesRemoved = 0
	conf := makeConf()

	seg := internal.SegmentConf{Type: "diff-stats"}
	got := RenderSegment(seg, input, conf)
	if got != "" {
		t.Errorf("diff-stats with zero should be empty, got %q", got)
	}
}

// ── Nerd font icon suppresses prefix ──

func TestRenderSegment_NerdFontSuppressesPrefix(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.NerdFont = true

	seg := internal.SegmentConf{Type: "model", Prefix: "model:"}
	got := RenderSegment(seg, input, conf)
	if strings.Contains(got, "model:") {
		t.Errorf("nerd font icon should suppress prefix, got %q", got)
	}
	if !strings.Contains(got, NerdFontIcons["model"]) {
		t.Errorf("should contain nerd font icon, got %q", got)
	}
}

// ── Per-segment icon override false ──

func TestRenderSegment_NerdFontDisabledPerSegment(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.NerdFont = true

	f := false
	seg := internal.SegmentConf{Type: "model", Icon: &f}
	got := RenderSegment(seg, input, conf)
	if strings.Contains(got, NerdFontIcons["model"]) {
		t.Errorf("per-segment icon=false should suppress nerd font, got %q", got)
	}
}

// ── Effort with configDir ──

func TestRenderEffort_ReadsFromConfigDir(t *testing.T) {
	dir := t.TempDir()

	// Write a settings.json with effortLevel
	settingsJSON := `{"effortLevel": "low"}`
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte(settingsJSON), 0644)

	input := makeInput()
	conf := makeConf()
	conf.ConfigDir = dir

	// Clear env var so it falls through to settings file
	t.Setenv("CLAUDE_CODE_EFFORT_LEVEL", "")

	seg := internal.SegmentConf{Type: "effort"}
	got := RenderSegment(seg, input, conf)
	if !strings.Contains(got, "low") {
		t.Errorf("effort should read 'low' from configDir settings, got %q", got)
	}
}

func TestRenderEffort_DifferentConfigDirs(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "settings.json"), []byte(`{"effortLevel": "low"}`), 0644)
	os.WriteFile(filepath.Join(dir2, "settings.json"), []byte(`{"effortLevel": "max"}`), 0644)

	input := makeInput()
	t.Setenv("CLAUDE_CODE_EFFORT_LEVEL", "")

	conf1 := makeConf()
	conf1.ConfigDir = dir1
	seg := internal.SegmentConf{Type: "effort"}
	got1 := RenderSegment(seg, input, conf1)

	conf2 := makeConf()
	conf2.ConfigDir = dir2
	got2 := RenderSegment(seg, input, conf2)

	if !strings.Contains(got1, "low") {
		t.Errorf("dir1 effort should be 'low', got %q", got1)
	}
	if !strings.Contains(got2, "max") {
		t.Errorf("dir2 effort should be 'max', got %q", got2)
	}
}
