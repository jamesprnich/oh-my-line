package datasource

import (
	"strings"
	"testing"

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

// Suppress unused import warnings
var _ = render.RST
