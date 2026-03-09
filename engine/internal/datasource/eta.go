package datasource

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// ETAResult holds computed ETA values.
type ETAResult struct {
	Session    string
	SessionMin string
	SessionHr  string
	Weekly     string
	WeeklyMin  string
	WeeklyHr   string
}

// ComputeETAs calculates ETA using 3 methods for both windows.
func ComputeETAs(sessionPctRaw, weeklyPctRaw float64, sessionReset, weeklyReset string,
	shortSecs, longSecs int, shortLabel, longLabel string) *ETAResult {

	result := &ETAResult{}

	cacheDir, err := cache.Dir()
	if err != nil {
		return result
	}

	now := time.Now().Unix()

	// Method 1: Window average
	if sessionReset != "" && sessionReset != "null" {
		if ep := isoToEpoch(sessionReset); ep > 0 {
			winElapsed := int(now - (ep - int64(shortSecs)))
			if winElapsed > 60 {
				if mins := etaMins(sessionPctRaw, 0, winElapsed); mins > 0 {
					result.Session = render.FormatDuration(mins)
				}
			}
		}
	}
	if weeklyReset != "" && weeklyReset != "null" {
		if ep := isoToEpoch(weeklyReset); ep > 0 {
			winElapsed := int(now - (ep - int64(longSecs)))
			if winElapsed > 60 {
				if mins := etaMins(weeklyPctRaw, 0, winElapsed); mins > 0 {
					result.Weekly = render.FormatDuration(mins)
				}
			}
		}
	}

	// Method 2: Per-minute (short-term ~5min reference)
	refShortFile := filepath.Join(cacheDir, "statusline-eta-short.dat")
	if state := readETARef(refShortFile); state != nil {
		age := int(now - state.epoch)
		if age >= 120 {
			if mins := etaMins(sessionPctRaw, state.sessionPct, age); mins > 0 {
				result.SessionMin = render.FormatDuration(mins)
			}
			if mins := etaMins(weeklyPctRaw, state.weeklyPct, age); mins > 0 {
				result.WeeklyMin = render.FormatDuration(mins)
			}
		}
		if age >= 300 || sessionPctRaw+5 < state.sessionPct {
			writeETARef(refShortFile, now, sessionPctRaw, weeklyPctRaw)
		}
	} else {
		writeETARef(refShortFile, now, sessionPctRaw, weeklyPctRaw)
	}

	// Method 3: Per-hour (long-term ~60min reference)
	refLongFile := filepath.Join(cacheDir, "statusline-eta-long.dat")
	if state := readETARef(refLongFile); state != nil {
		age := int(now - state.epoch)
		if age >= 600 {
			if mins := etaMins(sessionPctRaw, state.sessionPct, age); mins > 0 {
				result.SessionHr = render.FormatDuration(mins)
			}
			if mins := etaMins(weeklyPctRaw, state.weeklyPct, age); mins > 0 {
				result.WeeklyHr = render.FormatDuration(mins)
			}
		}
		if age >= 3600 || sessionPctRaw+5 < state.sessionPct {
			writeETARef(refLongFile, now, sessionPctRaw, weeklyPctRaw)
		}
	} else {
		writeETARef(refLongFile, now, sessionPctRaw, weeklyPctRaw)
	}

	return result
}

// RenderETASegment renders an ETA segment with window label.
func RenderETASegment(segType string, eta *ETAResult, shortLabel, longLabel string) string {
	pre := render.HexFG("#39d2c0")
	post := render.RST

	switch segType {
	case "eta-session":
		if eta.Session == "" {
			return render.DIM + shortLabel + " —" + render.RST
		}
		return render.DIM + shortLabel + " " + pre + eta.Session + post

	case "eta-session-min":
		if eta.SessionMin == "" {
			return render.DIM + shortLabel + "/min —" + render.RST
		}
		return render.DIM + shortLabel + "/min " + pre + eta.SessionMin + post

	case "eta-session-hr":
		if eta.SessionHr == "" {
			return render.DIM + shortLabel + "/hr —" + render.RST
		}
		return render.DIM + shortLabel + "/hr " + pre + eta.SessionHr + post

	case "eta-weekly":
		if eta.Weekly == "" {
			return render.DIM + longLabel + " —" + render.RST
		}
		return render.DIM + longLabel + " " + pre + eta.Weekly + post

	case "eta-weekly-min":
		if eta.WeeklyMin == "" {
			return render.DIM + longLabel + "/min —" + render.RST
		}
		return render.DIM + longLabel + "/min " + pre + eta.WeeklyMin + post

	case "eta-weekly-hr":
		if eta.WeeklyHr == "" {
			return render.DIM + longLabel + "/hr —" + render.RST
		}
		return render.DIM + longLabel + "/hr " + pre + eta.WeeklyHr + post
	}
	return ""
}

func etaMins(curPct, refPct float64, elapsedSecs int) int {
	delta := curPct - refPct
	if delta <= 0 || elapsedSecs <= 0 {
		return -1
	}
	rate := delta / float64(elapsedSecs)
	remain := 100.0 - curPct
	if remain <= 0 {
		return 0
	}
	return int((remain / rate) / 60)
}

type etaRef struct {
	epoch      int64
	sessionPct float64
	weeklyPct  float64
}

func readETARef(path string) *etaRef {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	parts := strings.SplitN(strings.TrimSpace(string(data)), "|", 3)
	if len(parts) != 3 {
		return nil
	}
	ep, _ := strconv.ParseInt(parts[0], 10, 64)
	sess, _ := strconv.ParseFloat(parts[1], 64)
	wkly, _ := strconv.ParseFloat(parts[2], 64)
	return &etaRef{epoch: ep, sessionPct: sess, weeklyPct: wkly}
}

func writeETARef(path string, epoch int64, sessionPct, weeklyPct float64) {
	os.WriteFile(path, []byte(fmt.Sprintf("%d|%f|%f", epoch, sessionPct, weeklyPct)), 0600)
}
