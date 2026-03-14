package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/jamesprnich/oh-my-line/engine/internal"
	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/config"
	"github.com/jamesprnich/oh-my-line/engine/internal/datasource"
	"github.com/jamesprnich/oh-my-line/engine/internal/debug"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

func main() {
	// Read stdin (bounded to 1MB to prevent memory exhaustion)
	data, err := io.ReadAll(io.LimitReader(os.Stdin, 1<<20))
	if err != nil || len(data) == 0 {
		fmt.Print("Claude")
		return
	}

	var input internal.Input
	if err := json.Unmarshal(data, &input); err != nil {
		fmt.Print("Claude")
		return
	}

	if input.Model.DisplayName == "" {
		input.Model.DisplayName = "Claude"
	}
	if input.ContextWindow.Size == 0 {
		input.ContextWindow.Size = 200000
	}

	// Resolve account config directory from CLAUDE_CONFIG_DIR
	configDir := os.Getenv("CLAUDE_CONFIG_DIR")
	if configDir == "" {
		if home, err := os.UserHomeDir(); err == nil {
			configDir = filepath.Join(home, ".claude")
		}
	}

	// Load config (configDir enables per-account config lookup)
	conf, err := config.LoadWithProduct(input.CWD, configDir)
	if err != nil {
		fmt.Print(input.Model.DisplayName)
		return
	}

	conf.ConfigDir = configDir
	conf.AccountKey = cache.AccountKey(configDir)

	// Init debug logging: env var OML_DEBUG=1 takes precedence, then config
	debug.Init(os.Getenv("OML_DEBUG") == "1", conf.Debug)

	// Detect terminal width from $COLUMNS (set by the shell)
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if w, err := strconv.Atoi(cols); err == nil && w > 0 {
			conf.TermWidth = w
		}
	}

	// Populate runtime data from datasources
	conf.Runtime = computeRuntimeData(conf, &input)

	// Render
	output := render.RenderStatusline(conf, &input)
	fmt.Print(output)
}

// computeRuntimeData runs datasource computations for configured segments.
func computeRuntimeData(conf *internal.Config, input *internal.Input) *internal.RuntimeData {
	rt := &internal.RuntimeData{}

	// Collect configured segment types and per-type options
	types := make(map[string]bool)
	barStyles := make(map[string]string)
	showResets := make(map[string]*bool)
	for _, line := range conf.Lines {
		for _, seg := range line.Segments {
			types[seg.Type] = true
			if seg.BarStyle != "" {
				barStyles[seg.Type] = seg.BarStyle
			}
			if seg.ShowReset != nil {
				showResets[seg.Type] = seg.ShowReset
			}
		}
	}

	// Burn rate
	if types["burn-min"] || types["burn-hr"] || types["burn-spark"] ||
		types["cost-min"] || types["cost-hr"] {
		burn := datasource.ComputeBurnRate(input.CurrentTokens(), conf.AccountKey)
		rt.BurnRateMin = burn.RateMin
		rt.BurnRateHr = burn.RateHr
		rt.BurnElapsed = burn.Elapsed
		rt.BurnHasData = burn.HasData
	}

	// Rate limits
	if types["rate-session"] || types["rate-weekly"] || types["rate-extra"] || types["rate-opus"] ||
		types["eta-session"] || types["eta-session-min"] || types["eta-session-hr"] ||
		types["eta-weekly"] || types["eta-weekly-min"] || types["eta-weekly-hr"] ||
		types["rate-spark"] || types["rate-target"] {
		// Resolve usage proxy: env var takes precedence over config
		proxyURL := os.Getenv("OML_USAGE_PROXY_CLAUDE_CODE")
		if proxyURL == "" {
			proxyURL = conf.UsageProxy["claudeCode"]
		}
		rl := datasource.FetchRateLimits(proxyURL, conf.AccountKey, conf.ConfigDir)
		rt.RateSession = datasource.RenderRateLimitSegment("rate-session", rl, barStyles["rate-session"], resolveShowReset(showResets, "rate-session", true))
		rt.RateWeekly = datasource.RenderRateLimitSegment("rate-weekly", rl, barStyles["rate-weekly"], resolveShowReset(showResets, "rate-weekly", false))
		rt.RateExtra = datasource.RenderRateLimitSegment("rate-extra", rl, barStyles["rate-extra"], false)
		rt.RateOpus = datasource.RenderRateLimitSegment("rate-opus", rl, barStyles["rate-opus"], resolveShowReset(showResets, "rate-opus", false))
		rt.SessionPct = rl.SessionPctRaw
		rt.WeeklyPct = rl.WeeklyPctRaw
		rt.SessionReset = rl.SessionReset
		rt.WeeklyReset = rl.WeeklyReset
		rt.ShortLabel = rl.ShortLabel
		rt.LongLabel = rl.LongLabel
		rt.RateHasData = rl.HasData

		// ETAs (depend on rate limits)
		if types["eta-session"] || types["eta-session-min"] || types["eta-session-hr"] ||
			types["eta-weekly"] || types["eta-weekly-min"] || types["eta-weekly-hr"] {
			eta := datasource.ComputeETAs(rl.SessionPctRaw, rl.WeeklyPctRaw,
				rl.SessionReset, rl.WeeklyReset,
				datasource.DefaultShortWindowSecs, datasource.DefaultLongWindowSecs,
				rl.ShortLabel, rl.LongLabel, conf.AccountKey)
			rt.ETASession = datasource.RenderETASegment("eta-session", eta, rl.ShortLabel, rl.LongLabel)
			rt.ETASessionMin = datasource.RenderETASegment("eta-session-min", eta, rl.ShortLabel, rl.LongLabel)
			rt.ETASessionHr = datasource.RenderETASegment("eta-session-hr", eta, rl.ShortLabel, rl.LongLabel)
			rt.ETAWeekly = datasource.RenderETASegment("eta-weekly", eta, rl.ShortLabel, rl.LongLabel)
			rt.ETAWeeklyMin = datasource.RenderETASegment("eta-weekly-min", eta, rl.ShortLabel, rl.LongLabel)
			rt.ETAWeeklyHr = datasource.RenderETASegment("eta-weekly-hr", eta, rl.ShortLabel, rl.LongLabel)
		}
	}

	// Sparklines
	if types["burn-spark"] || types["ctx-spark"] || types["rate-spark"] ||
		types["ctx-target"] || types["rate-target"] {
		ctxPct := 0
		if input.ContextWindow.Size > 0 {
			ctxPct = input.CurrentTokens() * 100 / input.ContextWindow.Size
		}
		sessionPct := int(rt.SessionPct)
		spark := datasource.ComputeSparklines(rt.BurnRateMin, ctxPct, sessionPct,
			datasource.DefaultShortWindowSecs, datasource.DefaultLongWindowSecs, conf.AccountKey)
		rt.BurnSpark = spark.BurnSpark
		rt.CtxSpark = spark.CtxSpark
		rt.RateSpark = spark.RateSpark
		rt.CtxTarget = spark.CtxTarget
		rt.RateTarget = spark.RateTarget
	}

	// Cost tracking
	if types["cost"] || types["cost-min"] || types["cost-hr"] || types["cost-7d"] || types["cost-spark"] {
		cost := datasource.ComputeCost(
			input.Model.DisplayName,
			input.ContextWindow.Usage.InputTokens,
			input.ContextWindow.Usage.CacheCreate,
			input.ContextWindow.Usage.CacheRead,
			rt.BurnRateMin, rt.BurnRateHr,
			conf.AccountKey,
		)
		rt.CostCtx = datasource.RenderCostSegment("cost", cost)
		rt.CostMin = datasource.RenderCostSegment("cost-min", cost)
		rt.CostHr = datasource.RenderCostSegment("cost-hr", cost)
		rt.Cost7d = datasource.RenderCostSegment("cost-7d", cost)
		rt.CostSpark = datasource.RenderCostSegment("cost-spark", cost)
	}

	// GitHub segments
	ghTypes := map[string]bool{}
	for _, t := range []string{"gh-pr", "gh-checks", "gh-reviews", "gh-actions", "gh-notifs",
		"gh-issues", "gh-pr-count", "gh-pr-comments", "gh-stars"} {
		if types[t] {
			ghTypes[t] = true
		}
	}
	if len(ghTypes) > 0 {
		branch := ""
		if input.CWD != "" {
			branch = render.DetectBranch(input.CWD)
		}
		gh := datasource.FetchGitHub(input.CWD, branch, ghTypes)
		rt.GhPR = gh.PR
		rt.GhChecks = gh.Checks
		rt.GhReviews = gh.Reviews
		rt.GhActions = gh.Actions
		rt.GhNotifs = gh.Notifs
		rt.GhIssues = gh.Issues
		rt.GhPRCount = gh.PRCount
		rt.GhPRComments = gh.PRComments
		rt.GhStars = gh.Stars
	}

	// Docker segments
	dockerTypes := map[string]bool{}
	if types["docker"] {
		dockerTypes["docker"] = true
	}
	if types["docker-db"] {
		dockerTypes["docker-db"] = true
	}
	if len(dockerTypes) > 0 {
		dk := datasource.FetchDocker(input.CWD, dockerTypes)
		rt.Docker = dk.Summary
		rt.DockerDB = dk.DB
	}

	// Command segments
	if types["command"] {
		var specs []datasource.CommandSpec
		for _, line := range conf.Lines {
			for _, seg := range line.Segments {
				if seg.Type == "command" {
					specs = append(specs, datasource.CommandSpec{
						Content: seg.Content,
						Cache:   seg.Cache,
						Timeout: seg.Timeout,
					})
				}
			}
		}
		rt.CommandCache = datasource.BuildCommandCache(conf.Trusted, specs, datasource.ExecCommand)
	}

	return rt
}

// resolveShowReset returns the user's override if set, otherwise the default.
func resolveShowReset(overrides map[string]*bool, segType string, defaultVal bool) bool {
	if v, ok := overrides[segType]; ok && v != nil {
		return *v
	}
	return defaultVal
}
