package datasource

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/debug"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// UsageData represents the Anthropic subscription usage response.
// Aligned to https://github.com/jamesprnich/ai-usage-proxy-spec
type UsageData struct {
	FiveHour     *WindowData `json:"five_hour"`
	SevenDay     *WindowData `json:"seven_day"`
	SevenDayOpus *WindowData `json:"seven_day_opus"`
	ExtraUsage   *ExtraData  `json:"extra_usage"`
	Meta         *UsageMeta  `json:"meta,omitempty"`
}

// WindowData holds utilization and reset time for a single rate-limit window.
type WindowData struct {
	Utilization float64 `json:"utilization"`
	ResetsAt    string  `json:"resets_at"`
}

// ExtraData holds extra usage billing data (pay-as-you-go overage).
type ExtraData struct {
	IsEnabled    bool    `json:"is_enabled"`
	Utilization  float64 `json:"utilization"`
	UsedCredits  float64 `json:"used_credits"`  // cents consumed this month
	MonthlyLimit float64 `json:"monthly_limit"` // cents cap for extra usage this month
}

// UsageMeta carries proxy-level metadata alongside cached usage data.
type UsageMeta struct {
	Source      string `json:"source"`
	RateLimited bool   `json:"rate_limited"`
	LastUpdated string `json:"last_updated"`
}

// RateLimitResult holds the computed rate limit data.
type RateLimitResult struct {
	SessionPct     int
	SessionPctRaw  float64
	SessionReset   string
	WeeklyPct      int
	WeeklyPctRaw   float64
	WeeklyReset    string
	ExtraEnabled   bool
	ExtraPct       int
	ExtraUsed      float64
	ExtraLimit     float64
	OpusPct        int
	OpusPctRaw     float64
	OpusReset      string
	Stale          bool   // yellow ⚠ — data exists but is old or proxy rate-limited
	Unreachable    bool   // red ⚠ — cannot connect to data source
	HasData        bool

	// Auto-detected window durations
	ShortLabel string
	LongLabel  string
}

// Default rate-limit window durations (in seconds).
const (
	DefaultShortWindowSecs = 18000  // 5 hours
	DefaultLongWindowSecs  = 604800 // 7 days
)

// FetchRateLimits fetches and parses rate limit data with caching.
// If proxyURL is non-empty, it fetches from the proxy with a plain GET (no auth headers).
func FetchRateLimits(proxyURL string) *RateLimitResult {
	result := &RateLimitResult{}

	cacheDir, err := cache.Dir()
	if err != nil {
		return result
	}

	// Load window durations
	shortSecs := DefaultShortWindowSecs
	longSecs := DefaultLongWindowSecs
	if ws, err := cache.ReadWindowFile(cacheDir); err == nil {
		shortSecs = ws.ShortSecs
		longSecs = ws.LongSecs
	}
	result.ShortLabel = secsToLabel(shortSecs)
	result.LongLabel = secsToLabel(longSecs)

	// Cache TTL: 60s when using a proxy (cheap local call), 300s for direct Anthropic API
	cacheTTL := 300
	src := "direct"
	if proxyURL != "" {
		cacheTTL = 60
		src = "proxy"
	}
	debug.Log("rate", "source=%s ttl=%ds proxy=%q", src, cacheTTL, proxyURL)

	// Try cache
	cacheFile := filepath.Join(cacheDir, "statusline-usage-cache.json")
	tmpFile := cacheFile + ".tmp"
	pidFile := cacheFile + ".pid"
	errFile := cacheFile + ".err"

	// Check if background fetch completed
	if _, err := os.Stat(tmpFile); err == nil {
		if _, err := os.Stat(pidFile); err != nil {
			// Background done — promote if valid
			if data, err := os.ReadFile(tmpFile); err == nil {
				var test UsageData
				if json.Unmarshal(data, &test) == nil && (test.FiveHour != nil || test.SevenDay != nil) {
					os.Rename(tmpFile, cacheFile)
					os.Remove(errFile) // clear any previous error
					debug.Log("rate", "promoted tmp→cache size=%d", len(data))
				} else {
					os.Remove(tmpFile)
					debug.Log("rate", "discarded invalid tmp size=%d content=%q", len(data), truncate(string(data), 200))
				}
			}
		} else {
			debug.Log("rate", "tmp exists but pid file still present, bg-fetch in progress")
		}
	}

	// Check for connection error from last background fetch
	if _, err := os.Stat(errFile); err == nil {
		errContent, errFresh := cache.ReadFile(errFile, cacheTTL)
		if errFresh && errContent != "" {
			result.Unreachable = true
			debug.Log("rate", "unreachable: err file present content=%q", truncate(errContent, 100))
		} else if errContent != "" {
			debug.Log("rate", "stale err file (age > %ds), ignoring", cacheTTL)
		}
	}

	var usageData *UsageData
	needsRefresh := true

	if content, fresh := cache.ReadFile(cacheFile, cacheTTL); fresh && content != "" {
		var ud UsageData
		if json.Unmarshal([]byte(content), &ud) == nil && (ud.FiveHour != nil || ud.SevenDay != nil) {
			usageData = &ud
			needsRefresh = false
			debug.Log("rate", "cache=hit session=%.1f%% weekly=%.1f%%",
				safeUtil(ud.FiveHour), safeUtil(ud.SevenDay))
		}
	}

	if needsRefresh {
		// Serve stale cache
		if data, err := os.ReadFile(cacheFile); err == nil {
			var ud UsageData
			if json.Unmarshal(data, &ud) == nil && (ud.FiveHour != nil || ud.SevenDay != nil) {
				usageData = &ud
				result.Stale = true
				fi, _ := os.Stat(cacheFile)
				age := 0
				if fi != nil {
					age = int(time.Since(fi.ModTime()).Seconds())
				}
				debug.Log("rate", "cache=stale age=%ds serving old data", age)
			}
		} else {
			debug.Log("rate", "cache=miss no existing data")
		}

		debug.Log("rate", "launching bg-fetch source=%s", src)
		launchBGRefresh(cacheDir, pidFile, tmpFile, errFile, proxyURL)
	}

	if usageData == nil {
		debug.Log("rate", "result: no data stale=%v unreachable=%v", result.Stale, result.Unreachable)
		return result
	}

	result.HasData = true

	// Proxy signals upstream rate limiting — data is valid but stale
	if usageData.Meta != nil && usageData.Meta.RateLimited {
		result.Stale = true
	}

	now := time.Now().Unix()

	// Parse session window
	if usageData.FiveHour != nil {
		result.SessionPctRaw = usageData.FiveHour.Utilization
		result.SessionPct = int(math.Round(usageData.FiveHour.Utilization))
		result.SessionReset = usageData.FiveHour.ResetsAt

		// Window duration auto-detection
		if ep := isoToEpoch(usageData.FiveHour.ResetsAt); ep > 0 {
			ttl := int(ep - now)
			if ttl > 0 && ttl <= 36000 && ttl > shortSecs {
				shortSecs = ttl
				result.ShortLabel = secsToLabel(shortSecs)
				cache.WriteWindowFile(cacheDir, shortSecs, longSecs)
			}
		}
	}

	// Parse weekly window
	if usageData.SevenDay != nil {
		result.WeeklyPctRaw = usageData.SevenDay.Utilization
		result.WeeklyPct = int(math.Round(usageData.SevenDay.Utilization))
		result.WeeklyReset = usageData.SevenDay.ResetsAt

		if ep := isoToEpoch(usageData.SevenDay.ResetsAt); ep > 0 {
			ttl := int(ep - now)
			if ttl > 0 && ttl <= 1209600 && ttl > longSecs {
				longSecs = ttl
				result.LongLabel = secsToLabel(longSecs)
				cache.WriteWindowFile(cacheDir, shortSecs, longSecs)
			}
		}
	}

	// Parse extra usage
	if usageData.ExtraUsage != nil && usageData.ExtraUsage.IsEnabled {
		result.ExtraEnabled = true
		result.ExtraPct = int(math.Round(usageData.ExtraUsage.Utilization))
		result.ExtraUsed = usageData.ExtraUsage.UsedCredits / 100
		result.ExtraLimit = usageData.ExtraUsage.MonthlyLimit / 100
	}

	// Parse opus tier
	if usageData.SevenDayOpus != nil && usageData.SevenDayOpus.Utilization > 0 {
		result.OpusPctRaw = usageData.SevenDayOpus.Utilization
		result.OpusPct = int(math.Round(usageData.SevenDayOpus.Utilization))
		result.OpusReset = usageData.SevenDayOpus.ResetsAt
	}

	return result
}

func launchBGRefresh(cacheDir, pidFile, tmpFile, errFile, proxyURL string) {
	// Check if already running (reuses shouldLaunchBG from github.go)
	if !shouldLaunchBG(pidFile) {
		return
	}

	// Curl writes to a staging file, then atomically renames to .tmp
	// to prevent readers from seeing partially-written data.
	stageFile := tmpFile + ".stage"

	var url string
	var curlArgs []string
	var postCurl string

	if proxyURL != "" {
		// Only allow http/https proxy URLs
		if !strings.HasPrefix(proxyURL, "http://") && !strings.HasPrefix(proxyURL, "https://") {
			debug.Log("rate", "invalid proxy URL scheme, ignoring: %s", proxyURL)
			return
		}
		// Append spec endpoint path: ai-usage-proxy-spec
		url = strings.TrimRight(proxyURL, "/") + "/api/proxy/anthropic/subscription/"
		curlArgs = []string{"-s", "-m", "8", "-o", stageFile, url}
	} else {
		token := GetOAuthToken()
		if token == "" {
			return
		}
		ccVersion := getClaudeVersion(cacheDir)
		url = "https://api.anthropic.com/api/oauth/usage"
		// Pass auth header via curl config file to avoid token exposure
		// in /proc/*/cmdline.
		cfgFile := filepath.Join(cacheDir, "statusline-curl-cfg.tmp")
		curlCfg := fmt.Sprintf("header = \"Authorization: Bearer %s\"\n", token)
		if err := os.WriteFile(cfgFile, []byte(curlCfg), 0600); err != nil {
			debug.Log("rate", "failed to write curl config: %v", err)
			return
		}
		curlArgs = []string{
			"-s", "-m", "8", "-o", stageFile,
			"-H", "Accept: application/json",
			"-H", "Content-Type: application/json",
			"-H", "anthropic-beta: oauth-2025-04-20",
			"-H", "User-Agent: claude-code/" + ccVersion,
			"-K", cfgFile,
			url,
		}
		postCurl = fmt.Sprintf("rm -f %s; ", shellQuote(cfgFile))
	}

	debug.Log("rate", "launching bg-fetch subprocess url=%s", url)

	// Launch a detached curl subprocess that outlives this process.
	// A goroutine would be killed when main() returns.
	// On success, atomically rename stage→tmp. On failure, record error.
	script := fmt.Sprintf(
		`curl %s; ec=$?; %sif [ $ec -ne 0 ]; then rm -f %s; echo unreachable > %s; else mv -f %s %s; rm -f %s; fi; rm -f %s`,
		shellJoinArgs(curlArgs), postCurl,
		shellQuote(stageFile), shellQuote(errFile),
		shellQuote(stageFile), shellQuote(tmpFile), shellQuote(errFile), shellQuote(pidFile),
	)
	cmd := exec.Command("sh", "-c", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		debug.Log("rate", "bg-fetch start err=%v", err)
		return
	}
	os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0600)
}

func getClaudeVersion(cacheDir string) string {
	vcPath := filepath.Join(cacheDir, "statusline-version.txt")
	if content, fresh := cache.ReadFile(vcPath, 3600); fresh && content != "" {
		return strings.TrimSpace(content)
	}

	out, err := exec.Command("claude", "--version").Output()
	if err == nil {
		v := strings.Fields(strings.TrimSpace(string(out)))
		if len(v) > 0 {
			cache.WriteFile(vcPath, v[0])
			return v[0]
		}
	}
	return "unknown"
}

func isoToEpoch(iso string) int64 {
	if iso == "" || iso == "null" {
		return 0
	}
	// Try standard formats (RFC3339Nano handles fractional seconds from proxy spec)
	for _, layout := range []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02T15:04:05+00:00",
	} {
		if t, err := time.Parse(layout, iso); err == nil {
			return t.Unix()
		}
	}
	return 0
}

func secsToLabel(secs int) string {
	if secs >= 86400 && secs%86400 == 0 {
		return fmt.Sprintf("%dd", secs/86400)
	}
	if secs >= 3600 && secs%3600 == 0 {
		return fmt.Sprintf("%dh", secs/3600)
	}
	if secs >= 3600 {
		return fmt.Sprintf("%dh", (secs+1800)/3600)
	}
	return fmt.Sprintf("%dm", secs/60)
}

// shellQuote quotes a string for safe use in a shell command.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// shellJoinArgs joins args with shell quoting.
func shellJoinArgs(args []string) string {
	quoted := make([]string, len(args))
	for i, a := range args {
		quoted[i] = shellQuote(a)
	}
	return strings.Join(quoted, " ")
}

// FormatResetTime formats an ISO timestamp for display.
func FormatResetTime(iso, style string) string {
	if iso == "" || iso == "null" {
		return ""
	}
	ep := isoToEpoch(iso)
	if ep == 0 {
		return ""
	}
	t := time.Unix(ep, 0).Local()
	switch style {
	case "time":
		return strings.ToLower(strings.TrimLeft(t.Format("3:04PM"), " "))
	case "datetime":
		return strings.ToLower(t.Format("Jan 2, 3:04PM"))
	default:
		return strings.ToLower(t.Format("Jan 2"))
	}
}

// RenderRateLimitSegment renders a rate limit segment.
// barStyle "block" uses ▓░ characters, anything else uses ●○ (default).
// showReset controls whether the reset time/date is appended.
func RenderRateLimitSegment(segType string, rl *RateLimitResult, barStyle string, showReset bool) string {
	if rl == nil {
		return ""
	}

	white := render.HexFG("#dcdcdc")
	cyan := render.HexFG("#2e9599")
	bw := 6

	// Status suffix: red ⚠ for unreachable, yellow ⚠ for stale
	status := ""
	if rl.Unreachable {
		status = " " + render.HexFG("#ff5555") + "⚠" + render.RST
	} else if rl.Stale {
		status = " " + render.HexFG("#e6c800") + "⚠" + render.RST
	}

	switch segType {
	case "rate-session":
		if !rl.HasData {
			return white + rl.ShortLabel + render.RST + " " + render.DIM + "—" + render.RST + status
		}
		resetStr := ""
		if showReset {
			resetStr = FormatResetTime(rl.SessionReset, "time")
		}
		if rl.SessionPct >= 95 {
			return buildZoomBar(rl.SessionPctRaw, rl.ShortLabel, resetStr, barStyle) + status
		}
		bar := render.BuildBar(rl.SessionPct, bw, barStyle)
		out := white + rl.ShortLabel + render.RST + " " + bar + " " + cyan + fmt.Sprintf("%d%%", rl.SessionPct) + render.RST
		if resetStr != "" {
			out += " " + render.DIM + "@" + resetStr + render.RST
		}
		return out + status

	case "rate-weekly":
		if !rl.HasData {
			return white + rl.LongLabel + render.RST + " " + render.DIM + "—" + render.RST + status
		}
		resetStr := ""
		if showReset {
			resetStr = FormatResetTime(rl.WeeklyReset, "datetime")
		}
		if rl.WeeklyPct >= 95 {
			return buildZoomBar(rl.WeeklyPctRaw, rl.LongLabel, resetStr, barStyle) + status
		}
		bar := render.BuildBar(rl.WeeklyPct, bw, barStyle)
		out := white + rl.LongLabel + render.RST + " " + bar + " " + cyan + fmt.Sprintf("%d%%", rl.WeeklyPct) + render.RST
		if resetStr != "" {
			out += " " + render.DIM + "@" + resetStr + render.RST
		}
		return out + status

	case "rate-extra":
		if !rl.ExtraEnabled {
			return ""
		}
		bar := render.BuildBar(rl.ExtraPct, bw, barStyle)
		return white + "extra" + render.RST + " " + bar + " " + cyan +
			fmt.Sprintf("$%.2f/$%.2f", rl.ExtraUsed, rl.ExtraLimit) + render.RST + status

	case "rate-opus":
		if rl.OpusPct == 0 {
			return ""
		}
		resetStr := ""
		if showReset {
			resetStr = FormatResetTime(rl.OpusReset, "datetime")
		}
		if rl.OpusPct >= 95 {
			return buildZoomBar(rl.OpusPctRaw, "opus", resetStr, barStyle) + status
		}
		bar := render.BuildBar(rl.OpusPct, bw, barStyle)
		out := white + "opus" + render.RST + " " + bar + " " + cyan + fmt.Sprintf("%d%%", rl.OpusPct) + render.RST
		if resetStr != "" {
			out += " " + render.DIM + "@" + resetStr + render.RST
		}
		return out + status
	}
	return ""
}

// buildZoomBar builds a magnified bar for >=95% usage.
func buildZoomBar(rawPct float64, label, resetStr, barStyle string) string {
	zoomW := 10
	zoomRange := 5.0

	remainPct := 100.0 - rawPct
	if remainPct < 0 {
		remainPct = 0
	}

	filled := int(remainPct * float64(zoomW) / zoomRange)
	if filled < 0 {
		filled = 0
	}
	if filled > zoomW {
		filled = zoomW
	}
	empty := zoomW - filled

	var bc string
	switch {
	case remainPct <= 1:
		bc = render.HexFG("#ff5555")
	case remainPct <= 2:
		bc = render.HexFG("#e6c800")
	default:
		bc = render.HexFG("#00a000")
	}

	white := render.HexFG("#dcdcdc")
	dimC := render.HexFG("#666666")

	filledChar, emptyChar := "●", "○"
	if barStyle == "block" {
		filledChar, emptyChar = "▓", "░"
	}

	out := white + label + render.RST + " " + bc + strings.Repeat(filledChar, filled) + dimC + strings.Repeat(emptyChar, empty) + render.RST +
		" " + bc + fmt.Sprintf("%.1f%%", remainPct) + render.RST + " " + render.DIM + "left" + render.RST
	if resetStr != "" {
		out += " " + render.DIM + "@" + resetStr + render.RST
	}
	return out
}

func safeUtil(w *WindowData) float64 {
	if w == nil {
		return 0
	}
	return w.Utilization
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
