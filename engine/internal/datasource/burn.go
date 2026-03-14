package datasource

import (
	"time"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// BurnResult holds computed burn rate data.
type BurnResult struct {
	RateMin int  // tokens per minute
	RateHr  int  // tokens per hour
	Elapsed int  // seconds since tracking started
	HasData bool
}

// ComputeBurnRate calculates token burn rate from the cache file.
func ComputeBurnRate(currentTokens int, accountKey string) *BurnResult {
	result := &BurnResult{}

	if currentTokens <= 0 {
		return result
	}

	cacheDir, err := cache.AccountDir(accountKey)
	if err != nil {
		return result
	}

	now := time.Now().Unix()
	needsReset := false

	state, err := cache.ReadBurnFile(cacheDir)
	if err != nil {
		needsReset = true
	} else if currentTokens < state.StartTokens {
		needsReset = true // context reset
	}

	if needsReset {
		cache.WriteBurnFile(cacheDir, now, currentTokens)
		return result
	}

	elapsed := int(now - state.StartEpoch)
	result.Elapsed = elapsed

	if elapsed >= 30 {
		delta := currentTokens - state.StartTokens
		result.RateMin = delta * 60 / elapsed
		result.RateHr = delta * 3600 / elapsed
		result.HasData = true
	}

	return result
}

// RenderBurnSegment renders a burn rate segment.
func RenderBurnSegment(segType string, burn *BurnResult, warmup int) string {
	if warmup <= 0 {
		warmup = 30
	}

	switch segType {
	case "burn-min":
		if !burn.HasData || burn.Elapsed < warmup {
			return render.DIM + "—/min" + render.RST
		}
		return render.HexFG("#e6c800") + render.FormatTokens(burn.RateMin) + "/min" + render.RST

	case "burn-hr":
		if !burn.HasData || burn.Elapsed < warmup {
			return render.DIM + "—/hr" + render.RST
		}
		return render.HexFG("#e6c800") + render.FormatTokens(burn.RateHr) + "/hr" + render.RST
	}

	return ""
}
