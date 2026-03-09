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

const sparkBars = 8

var sparkChars = []rune("▁▂▃▄▅▆▇█")

// SparkResult holds computed sparkline data for all types.
type SparkResult struct {
	BurnSpark  string
	CtxSpark   string
	RateSpark  string
	CtxTarget  string
	RateTarget string
}

// ComputeSparklines computes sparkline segments.
func ComputeSparklines(burnRateMin, ctxPct, sessionPct int, shortSecs, longSecs int) *SparkResult {
	result := &SparkResult{}

	cacheDir, err := cache.Dir()
	if err != nil {
		return result
	}

	now := time.Now().Unix()

	// Bucket sizes
	burnBucket := 120  // 2 min
	ctxBucket := 300   // 5 min
	rateBucket := shortSecs / 8
	if rateBucket < 60 {
		rateBucket = 60
	}

	// Read spark cache
	sparkFile := filepath.Join(cacheDir, "statusline-spark.dat")
	burnState := readSparkState(sparkFile, "burn")
	ctxState := readSparkState(sparkFile, "ctx")
	rateState := readSparkState(sparkFile, "rate")

	// Update buckets
	burnState = updateBucket(burnState, burnBucket, burnRateMin, now)
	ctxState = updateBucket(ctxState, ctxBucket, ctxPct, now)
	rateState = updateBucket(rateState, rateBucket, sessionPct, now)

	// Write spark cache
	writeSparkCache(sparkFile, map[string]*sparkState{
		"burn": burnState,
		"ctx":  ctxState,
		"rate": rateState,
	})

	// Render sparklines
	burnFilled := filledCount(burnState.vals)
	ctxFilled := filledCount(ctxState.vals)
	rateFilled := filledCount(rateState.vals)

	result.BurnSpark = renderSparkSegment(burnState.vals, burnFilled, burnBucket,
		render.HexFG("#e6c800"), render.RST)
	result.CtxSpark = renderSparkSegment(ctxState.vals, ctxFilled, ctxBucket,
		render.HexFG("#ffb055"), render.RST)
	result.RateSpark = renderSparkSegment(rateState.vals, rateFilled, rateBucket,
		render.HexFG("#39d2c0"), render.RST)

	// Target sparklines
	targetFile := filepath.Join(cacheDir, "statusline-target-spark.dat")
	rate7dBucket := longSecs / 14
	if rate7dBucket < 3600 {
		rate7dBucket = 3600
	}

	ctxTState := readSparkState(targetFile, "ctx")
	rate7dState := readSparkState(targetFile, "rate7d")

	ctxTState = updateBucket(ctxTState, ctxBucket, ctxPct, now)
	rate7dState = updateBucket(rate7dState, rate7dBucket, sessionPct, now)

	writeSparkCache(targetFile, map[string]*sparkState{
		"ctx":    ctxTState,
		"rate7d": rate7dState,
	})

	ctxTFilled := filledCount(ctxTState.vals)
	rate7dFilled := filledCount(rate7dState.vals)

	result.CtxTarget = renderTargetSpark(ctxTState.vals, ctxTFilled, ctxBucket, 50, 80)
	result.RateTarget = renderTargetSpark(rate7dState.vals, rate7dFilled, rate7dBucket, 50, 80)

	return result
}

type sparkState struct {
	epoch int64
	vals  []int
}

func readSparkState(path, key string) *sparkState {
	data, err := os.ReadFile(path)
	if err != nil {
		return &sparkState{vals: make([]int, sparkBars)}
	}
	for _, line := range strings.Split(string(data), "\n") {
		parts := strings.SplitN(line, "|", 3)
		if len(parts) == 3 && parts[0] == key {
			ep, _ := strconv.ParseInt(parts[1], 10, 64)
			vals := parseCSV(parts[2])
			return &sparkState{epoch: ep, vals: vals}
		}
	}
	return &sparkState{vals: make([]int, sparkBars)}
}

func parseCSV(s string) []int {
	parts := strings.Split(s, ",")
	vals := make([]int, 0, sparkBars)
	for _, p := range parts {
		v, _ := strconv.Atoi(strings.TrimSpace(p))
		vals = append(vals, v)
	}
	// Pad/trim to sparkBars
	for len(vals) < sparkBars {
		vals = append([]int{0}, vals...)
	}
	if len(vals) > sparkBars {
		vals = vals[len(vals)-sparkBars:]
	}
	return vals
}

func updateBucket(state *sparkState, bucketSize, curVal int, now int64) *sparkState {
	if state.epoch == 0 || len(state.vals) == 0 {
		vals := make([]int, sparkBars)
		vals[sparkBars-1] = curVal
		return &sparkState{epoch: now, vals: vals}
	}

	elapsed := int(now - state.epoch)
	vals := make([]int, sparkBars)
	copy(vals, state.vals)

	if elapsed >= bucketSize {
		shifts := elapsed / bucketSize
		if shifts > sparkBars {
			shifts = sparkBars
		}
		for s := 0; s < shifts; s++ {
			vals = append(vals[1:], 0)
		}
		vals[sparkBars-1] = curVal
		newEpoch := state.epoch + int64(shifts*bucketSize)
		return &sparkState{epoch: newEpoch, vals: vals}
	}

	vals[sparkBars-1] = curVal
	return &sparkState{epoch: state.epoch, vals: vals}
}

func filledCount(vals []int) int {
	count := 0
	for _, v := range vals {
		if v > 0 {
			count++
		}
	}
	return count
}

func writeSparkCache(path string, states map[string]*sparkState) {
	var lines []string
	for key, state := range states {
		csv := valsToCSV(state.vals)
		lines = append(lines, fmt.Sprintf("%s|%d|%s", key, state.epoch, csv))
	}
	os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0600)
}

func valsToCSV(vals []int) string {
	parts := make([]string, len(vals))
	for i, v := range vals {
		parts[i] = strconv.Itoa(v)
	}
	return strings.Join(parts, ",")
}

func renderSpark(vals []int, filled int) string {
	unfilled := sparkBars - filled
	var out strings.Builder
	min, max := minMax(vals, unfilled)
	rng := max - min

	for i := 0; i < sparkBars; i++ {
		if i < unfilled {
			out.WriteRune('·')
		} else if rng == 0 {
			out.WriteRune('▄')
		} else {
			v := vals[i]
			idx := (v - min) * 7 / rng
			if idx < 0 {
				idx = 0
			}
			if idx > 7 {
				idx = 7
			}
			out.WriteRune(sparkChars[idx])
		}
	}
	return out.String()
}

func renderSparkSegment(vals []int, filled, bucketSize int, pre, post string) string {
	if len(vals) == 0 {
		return ""
	}
	spark := renderSpark(vals, filled)
	if filled < 3 {
		barsNeeded := 3 - filled
		secsLeft := barsNeeded * bucketSize
		minsLeft := (secsLeft + 59) / 60
		return pre + spark + fmt.Sprintf(" %dm", minsLeft) + post
	}
	return pre + spark + post
}

func renderTargetSpark(vals []int, filled, bucketSize, warn, crit int) string {
	if len(vals) == 0 {
		return ""
	}

	unfilled := sparkBars - filled
	min, max := minMax(vals, unfilled)
	rng := max - min

	grn := render.HexFG("#00b400")
	amb := render.HexFG("#e6c800")
	red := render.HexFG("#ff5555")

	var out strings.Builder
	prevColor := ""

	for i := 0; i < sparkBars; i++ {
		if i < unfilled {
			out.WriteRune('·')
			continue
		}

		v := vals[i]
		var c string
		switch {
		case v > crit:
			c = red
		case v > warn:
			c = amb
		default:
			c = grn
		}
		if c != prevColor {
			out.WriteString(c)
			prevColor = c
		}

		if rng == 0 {
			out.WriteRune('▄')
		} else {
			idx := (v - min) * 7 / rng
			if idx < 0 {
				idx = 0
			}
			if idx > 7 {
				idx = 7
			}
			out.WriteRune(sparkChars[idx])
		}
	}

	result := out.String() + render.RST
	if filled < 3 {
		barsNeeded := 3 - filled
		secsLeft := barsNeeded * bucketSize
		if bucketSize >= 3600 {
			hrsLeft := (secsLeft + 3599) / 3600
			result += fmt.Sprintf(" %dh", hrsLeft)
		} else {
			minsLeft := (secsLeft + 59) / 60
			result += fmt.Sprintf(" %dm", minsLeft)
		}
	}
	return result
}

func minMax(vals []int, unfilled int) (int, int) {
	lo := math.MaxInt
	hi := 0
	for i := unfilled; i < len(vals); i++ {
		if vals[i] < lo {
			lo = vals[i]
		}
		if vals[i] > hi {
			hi = vals[i]
		}
	}
	if lo == math.MaxInt {
		lo = 0
	}
	return lo, hi
}
