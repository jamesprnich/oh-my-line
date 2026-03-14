package internal

// Input represents the JSON stdin from Claude Code.
type Input struct {
	Model struct {
		DisplayName string `json:"display_name"`
		ID          string `json:"id"`
	} `json:"model"`
	ContextWindow struct {
		Size  int `json:"context_window_size"`
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			CacheCreate  int `json:"cache_creation_input_tokens"`
			CacheRead    int `json:"cache_read_input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"current_usage"`
		TotalInputTokens  int `json:"total_input_tokens"`
		TotalOutputTokens int `json:"total_output_tokens"`
	} `json:"context_window"`
	CWD     string `json:"cwd"`
	Version string `json:"version"`
	Cost    struct {
		TotalCostUSD      float64 `json:"total_cost_usd"`
		TotalDurationMs   int     `json:"total_duration_ms"`
		TotalAPIDurationMs int    `json:"total_api_duration_ms"`
		TotalLinesAdded   int     `json:"total_lines_added"`
		TotalLinesRemoved int     `json:"total_lines_removed"`
	} `json:"cost"`
	Exceeds200k bool `json:"exceeds_200k_tokens"`
	Vim         struct {
		Mode string `json:"mode"`
	} `json:"vim"`
	Agent struct {
		Name string `json:"name"`
	} `json:"agent"`
	Worktree struct {
		Name   string `json:"name"`
		Branch string `json:"branch"`
	} `json:"worktree"`

	// Runtime data passed via input JSON (used by WASM builder for mock data)
	Runtime *RuntimeData `json:"runtime,omitempty"`
}

// CurrentTokens returns input_tokens + cache_create + cache_read.
func (in *Input) CurrentTokens() int {
	return in.ContextWindow.Usage.InputTokens +
		in.ContextWindow.Usage.CacheCreate +
		in.ContextWindow.Usage.CacheRead
}

// Style holds segment styling options.
type Style struct {
	Color      string `json:"color"`
	Background string `json:"background"`
	Bold       bool   `json:"bold"`
	Dim        bool   `json:"dim"`
}

// SegmentConf represents a single segment in a line.
type SegmentConf struct {
	Type    string  `json:"type"`
	Style   *Style  `json:"style"`
	Content string  `json:"content"`
	PadLeft int     `json:"padLeft"`
	PadRight int    `json:"padRight"`
	Prefix  string  `json:"prefix"`
	Suffix  string  `json:"suffix"`
	Icon      *bool   `json:"icon"`      // per-segment nerd font override
	BarStyle  string  `json:"barStyle"`  // "dot" (default ●○) or "block" (▓░)
	ShowReset *bool   `json:"showReset"` // show reset time/date (default: true for session, false for weekly)

	// Segment-specific options
	Source   string   `json:"source"`
	Messages []string `json:"messages"`
	Interval int      `json:"interval"`
	Cache    int      `json:"cache"`
	Timeout  int      `json:"timeout"`
	Threshold int     `json:"threshold"`
	Warmup   int      `json:"warmup"`
	Width    int      `json:"width"`
	Warn     int      `json:"warn"`
	Critical int      `json:"critical"`
}

// SepStyle holds separator styling.
type SepStyle struct {
	Dim   bool   `json:"dim"`
	Color string `json:"color"`
}

// Padding holds line-level padding.
type Padding struct {
	Left  int `json:"left"`
	Right int `json:"right"`
}

// LineConf represents a single line in the config.
type LineConf struct {
	// Normal line fields
	Separator       string       `json:"separator"`
	SeparatorStyle  *SepStyle    `json:"separatorStyle"`
	BackgroundStyle string       `json:"backgroundStyle"`
	Background      string       `json:"background"`
	Padding         *Padding     `json:"padding"`
	Preset          string       `json:"preset"`
	Segments        []SegmentConf `json:"segments"`

	// Special line type fields
	Type  string   `json:"type"`  // "art", "rule", "spacer", or "" for normal
	Lines []string `json:"lines"` // art lines
	Char  string   `json:"char"`  // rule character
	Width int      `json:"width"` // rule width
	Style *Style   `json:"style"` // art/rule style
}

// RuntimeData holds computed data from datasources, populated before rendering.
// This keeps render/ as pure functions that don't do I/O.
// JSON tags allow WASM builder to pass mock data through input JSON.
type RuntimeData struct {
	// Burn rate
	BurnRateMin int  `json:"burn_rate_min"`
	BurnRateHr  int  `json:"burn_rate_hr"`
	BurnElapsed int  `json:"burn_elapsed"`
	BurnHasData bool `json:"burn_has_data"`

	// Rate limits (pre-rendered ANSI strings)
	RateSession string `json:"rate_session"`
	RateWeekly  string `json:"rate_weekly"`
	RateExtra   string `json:"rate_extra"`
	RateOpus    string `json:"rate_opus"`

	// Rate limit raw data for downstream (sparklines, ETAs)
	SessionPct    float64 `json:"session_pct"`
	WeeklyPct     float64 `json:"weekly_pct"`
	SessionReset  string  `json:"session_reset"`
	WeeklyReset   string  `json:"weekly_reset"`
	ShortLabel    string  `json:"short_label"`
	LongLabel     string  `json:"long_label"`
	RateHasData   bool    `json:"rate_has_data"`

	// Sparklines (pre-rendered)
	BurnSpark  string `json:"burn_spark"`
	CtxSpark   string `json:"ctx_spark"`
	RateSpark  string `json:"rate_spark"`
	CtxTarget  string `json:"ctx_target"`
	RateTarget string `json:"rate_target"`

	// ETAs (pre-rendered)
	ETASession    string `json:"eta_session"`
	ETASessionMin string `json:"eta_session_min"`
	ETASessionHr  string `json:"eta_session_hr"`
	ETAWeekly     string `json:"eta_weekly"`
	ETAWeeklyMin  string `json:"eta_weekly_min"`
	ETAWeeklyHr   string `json:"eta_weekly_hr"`

	// Cost (pre-rendered)
	CostCtx   string `json:"cost_ctx"`
	CostMin   string `json:"cost_min"`
	CostHr    string `json:"cost_hr"`
	Cost7d    string `json:"cost_7d"`
	CostSpark string `json:"cost_spark"`

	// GitHub (pre-rendered)
	GhPR         string `json:"gh_pr"`
	GhChecks     string `json:"gh_checks"`
	GhReviews    string `json:"gh_reviews"`
	GhActions    string `json:"gh_actions"`
	GhNotifs     string `json:"gh_notifs"`
	GhIssues     string `json:"gh_issues"`
	GhPRCount    string `json:"gh_pr_count"`
	GhPRComments string `json:"gh_pr_comments"`
	GhStars      string `json:"gh_stars"`

	// Docker (pre-rendered)
	Docker   string `json:"docker"`
	DockerDB string `json:"docker_db"`

	// Command cache
	CommandCache map[string]string `json:"command_cache,omitempty"`
}

// PresetConf defines preset line styling.
type PresetConf struct {
	BackgroundStyle string `json:"backgroundStyle"`
	BackgroundColor string `json:"backgroundColor"`
	LabelColor      string `json:"labelColor"`
	AccentColor     string `json:"accentColor"`
}

// Config represents the full oh-my-line.json configuration.
type Config struct {
	NerdFont   bool              `json:"nerdFont"`
	Debug      bool              `json:"debug,omitempty"`
	UsageProxy map[string]string `json:"usageProxy,omitempty"`
	Lines      []LineConf        `json:"oh-my-lines"`

	// Legacy format fields (examples/full.json)
	Icon    string `json:"icon"`
	Label   string `json:"label"`
	Tagline string `json:"tagline"`
	Statusline *struct {
		Lines           []LineConf            `json:"lines"`
		Presets         map[string]PresetConf `json:"presets"`
		Messages        []string              `json:"messages"`
		MessageInterval int                   `json:"messageInterval"`
	} `json:"statusline"`

	// Resolved after loading
	Presets         map[string]PresetConf `json:"-"`
	ResolvedMessages []string             `json:"-"`
	MessageInterval  int                  `json:"-"`
	MetaIcon         string               `json:"-"`
	MetaLabel        string               `json:"-"`
	MetaTagline      string               `json:"-"`
	CurrentMessage   string               `json:"-"`
	Trusted          bool                 `json:"-"`
	EmitMarkers      bool                 `json:"-"`
	TermWidth        int                  `json:"-"`
	AccountKey       string               `json:"-"` // derived from CLAUDE_CONFIG_DIR
	ConfigDir        string               `json:"-"` // resolved CLAUDE_CONFIG_DIR path
	Runtime          *RuntimeData         `json:"-"`
}
