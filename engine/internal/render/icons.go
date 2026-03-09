package render

// NerdFontIcons maps segment types to nerd font icon characters.
var NerdFontIcons = map[string]string{
	// Core
	"model":   "\U000F06A9", // 󰚩
	"dir":     "\uF07B",     //
	"version": "\uF02B",     //

	// Git
	"branch":     "\uE725", //
	"dir-branch": "\uE725", //
	"diff-stats": "\uF440", //

	// Tokens & context
	"tokens":     "\uF080", //
	"pct-used":   "\uF200", //
	"pct-remain": "\uF200", //

	// Burn rate
	"burn-min":   "\uF06D", //
	"burn-hr":    "\uF06D", //
	"burn-spark": "\uF06D", //

	// Rate limits
	"rate-session": "\uF0E4", //
	"rate-weekly":  "\uF073", //
	"rate-extra":   "\uF055", //
	"rate-spark":   "\uF0E4", //
	"rate-opus":    "\uF219", //

	// ETAs
	"eta-session":     "\uF017", //
	"eta-session-min": "\uF017", //
	"eta-session-hr":  "\uF017", //
	"eta-weekly":      "\uF017", //
	"eta-weekly-min":  "\uF017", //
	"eta-weekly-hr":   "\uF017", //

	// Cost
	"cost":       "\uF155", //
	"cost-min":   "\uF155", //
	"cost-hr":    "\uF155", //
	"cost-7d":    "\uF155", //
	"cost-spark": "\uF155", //

	// Sparkline targets
	"ctx-spark":   "\uF1FE", //
	"ctx-target":  "\uF05B", //
	"rate-target": "\uF05B", //

	// Docker
	"docker":    "\uF308", //
	"docker-db": "\uF1C0", //

	// GitHub
	"gh-pr":          "\uE726", //
	"gh-checks":      "\uF013", //
	"gh-reviews":     "\uF00C", //
	"gh-actions":     "\uF04B", //
	"gh-notifs":      "\uF0F3", //
	"gh-issues":      "\uF41B", //
	"gh-pr-count":    "\uE726", //
	"gh-pr-comments": "\uF075", //
	"gh-stars":       "\uF005", //

	// Session
	"session-cost":     "\uF155", //
	"session-duration": "\uF017", //
	"lines-changed":    "\uF440", //
	"cache-hit":        "\U000F06BC", // 󰆼
	"total-tokens":     "\uF0EC", //
	"api-wait":         "\uF254", //

	// Mode & identity
	"vim-mode":     "\uE7C5", //
	"worktree":     "\uE725", //
	"agent":        "\U000F06A9", // 󰚩
	"200k-warn":    "\uF071", //
	"cc-version":   "\uF02B", //
	"model-id":     "\U000F06A9", // 󰚩

	// Effort & warnings
	"effort":       "\U000F02A0", // 󰊠
	"compact-warn": "\uF071",     //

	// Env
	"env": "\uF462", //
}
