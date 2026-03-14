package datasource

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// CostResult holds computed cost data.
type CostResult struct {
	CtxFmt   string
	MinFmt   string
	HrFmt    string
	Day7Fmt  string
	SparkStr string
}

// pricing holds per-model token costs in cents per 1M tokens.
type pricing struct {
	input  int // standard input tokens
	cacheW int // cache write tokens
	cacheR int // cache read tokens
}

var pricingModels = map[string]pricing{
	"opus":   {1500, 1875, 150},
	"sonnet": {300, 375, 30},
	"haiku":  {25, 30, 3},
}

// ComputeCost calculates context cost, burn cost, 7-day totals, and sparkline.
func ComputeCost(modelName string, inputTokens, cacheCreate, cacheRead, burnRateMin, burnRateHr int, accountKey string) *CostResult {
	result := &CostResult{}

	// Determine pricing
	p := pricingModels["sonnet"] // default
	lower := strings.ToLower(modelName)
	if strings.Contains(lower, "opus") {
		p = pricingModels["opus"]
	} else if strings.Contains(lower, "haiku") {
		p = pricingModels["haiku"]
	}

	// Current context cost (in cents)
	ctxCents := float64(inputTokens)*float64(p.input)/1e6 +
		float64(cacheCreate)*float64(p.cacheW)/1e6 +
		float64(cacheRead)*float64(p.cacheR)/1e6

	if ctxCents > 0 {
		result.CtxFmt = formatDollars(ctxCents)
	}

	// Burn cost rates
	if burnRateMin > 0 {
		costMinCents := float64(burnRateMin) * float64(p.input) / 1e6
		result.MinFmt = formatDollars(costMinCents)
		costHrCents := float64(burnRateHr) * float64(p.input) / 1e6
		result.HrFmt = formatDollars(costHrCents)
	}

	// 7-day cost tracking
	home, _ := os.UserHomeDir()
	if home == "" {
		return result
	}

	costDir := filepath.Join(home, ".oh-my-line", "cost")
	if accountKey != "" && accountKey != "default" {
		costDir = filepath.Join(home, ".oh-my-line", "cost", "acct-"+accountKey)
	}
	os.MkdirAll(costDir, 0700)

	now := time.Now()
	today := now.Format("2006-01-02")
	todayFile := filepath.Join(costDir, today+".dat")

	// Track delta via per-process baseline
	cacheDir, err := cache.AccountDir(accountKey)
	if err != nil {
		return result
	}

	// Baseline file tracks the previous token state to detect context resets.
	// Format: epoch|inputTokens|cacheCreate|cacheRead|costCents
	baseFile := filepath.Join(cacheDir, "statusline-cost-base.dat")
	curHash := fmt.Sprintf("%d|%d|%d", inputTokens, cacheCreate, cacheRead)
	baseLine := fmt.Sprintf("%d|%s|%.4f", now.Unix(), curHash, ctxCents)

	if data, err := os.ReadFile(baseFile); err == nil {
		pp := strings.SplitN(strings.TrimSpace(string(data)), "|", 5)
		if len(pp) == 5 {
			prevHash := pp[1] + "|" + pp[2] + "|" + pp[3]
			prevCents, _ := strconv.ParseFloat(pp[4], 64)
			if prevHash != curHash {
				prevTotal := sumTokensFromHash(prevHash)
				curTotal := inputTokens + cacheCreate + cacheRead
				if curTotal < prevTotal && prevCents > 0 {
					// Tokens decreased — context was reset. Log the previous cost.
					appendToFile(todayFile, fmt.Sprintf("%.2f", prevCents))
				}
			}
		}
	}
	os.WriteFile(baseFile, []byte(baseLine), 0600)

	// Prune old files (>8 days)
	cutoff := now.AddDate(0, 0, -8).Format("2006-01-02")
	if entries, err := os.ReadDir(costDir); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".dat") {
				name := strings.TrimSuffix(e.Name(), ".dat")
				if name < cutoff {
					os.Remove(filepath.Join(costDir, e.Name()))
				}
			}
		}
	}

	// 7-day total
	totalCents := ctxCents
	for i := 0; i < 7; i++ {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		dayFile := filepath.Join(costDir, day+".dat")
		if data, err := os.ReadFile(dayFile); err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				if v, err := strconv.ParseFloat(strings.TrimSpace(line), 64); err == nil {
					totalCents += v
				}
			}
		}
	}
	result.Day7Fmt = formatDollars(totalCents)

	// 7-day sparkline
	var sparkVals []int
	for i := 6; i >= 0; i-- {
		day := now.AddDate(0, 0, -i).Format("2006-01-02")
		dayFile := filepath.Join(costDir, day+".dat")
		total := 0
		if data, err := os.ReadFile(dayFile); err == nil {
			for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
				if v, err := strconv.ParseFloat(strings.TrimSpace(line), 64); err == nil {
					total += int(math.Round(v))
				}
			}
		}
		sparkVals = append(sparkVals, total)
	}
	result.SparkStr = renderCostSpark(sparkVals)

	return result
}

// RenderCostSegment renders a cost segment.
func RenderCostSegment(segType string, cost *CostResult) string {
	pre := render.HexFG("#e6c800")
	post := render.RST

	switch segType {
	case "cost":
		if cost.CtxFmt == "" {
			return ""
		}
		return pre + "~" + cost.CtxFmt + post

	case "cost-min":
		if cost.MinFmt == "" {
			return render.DIM + "~—/min" + render.RST
		}
		return pre + "~" + cost.MinFmt + "/min" + post

	case "cost-hr":
		if cost.HrFmt == "" {
			return render.DIM + "~—/hr" + render.RST
		}
		return pre + "~" + cost.HrFmt + "/hr" + post

	case "cost-7d":
		if cost.Day7Fmt == "" {
			return ""
		}
		return pre + "~" + cost.Day7Fmt + " (7d)" + post

	case "cost-spark":
		if cost.SparkStr == "" {
			return ""
		}
		return pre + cost.SparkStr + post
	}
	return ""
}

func formatDollars(cents float64) string {
	if cents < 1 {
		return "$0.00"
	}
	if cents < 10000 {
		return fmt.Sprintf("$%.2f", cents/100)
	}
	return fmt.Sprintf("$%.0f", cents/100)
}

func sumTokensFromHash(hash string) int {
	parts := strings.Split(hash, "|")
	total := 0
	for _, p := range parts {
		v, _ := strconv.Atoi(p)
		total += v
	}
	return total
}

func appendToFile(path, line string) {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line + "\n")
}

func renderCostSpark(vals []int) string {
	if len(vals) == 0 {
		return ""
	}
	max := 0
	for _, v := range vals {
		if v > max {
			max = v
		}
	}
	var out strings.Builder
	for _, v := range vals {
		if v == 0 {
			out.WriteRune('·')
		} else if max == 0 {
			out.WriteRune('▄')
		} else {
			idx := v * 7 / max
			if idx > 7 {
				idx = 7
			}
			if idx < 1 {
				idx = 1
			}
			out.WriteRune(sparkChars[idx-1])
		}
	}
	return out.String()
}
