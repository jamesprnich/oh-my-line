package render

import (
	"fmt"
	"strings"
)

// FormatTokens formats a token count as "42", "65k", or "1.5M".
func FormatTokens(n int) string {
	if n >= 1000000 {
		return fmt.Sprintf("%.1fM", float64(n)/1000000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprintf("%d", n)
}

// FormatDuration formats minutes as "~45m", "~2h 30m", "~3d 5h".
func FormatDuration(mins int) string {
	if mins <= 0 {
		return ""
	}
	if mins >= 1440 {
		return fmt.Sprintf("~%dd %dh", mins/1440, (mins%1440)/60)
	}
	if mins >= 60 {
		return fmt.Sprintf("~%dh %dm", mins/60, mins%60)
	}
	return fmt.Sprintf("~%dm", mins)
}

// BuildBar builds a progress bar with colored fill.
// barStyle "block" uses ▓░, anything else (including "") uses ●○.
func BuildBar(pct, width int, barStyle string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := pct * width / 100
	empty := width - filled

	var bc string
	switch {
	case pct >= 90:
		bc = HexFG("#ff5555")
	case pct >= 70:
		bc = HexFG("#e6c800")
	case pct >= 50:
		bc = HexFG("#ffb055")
	default:
		bc = HexFG("#00a000")
	}

	filledChar, emptyChar := "●", "○"
	if barStyle == "block" {
		filledChar, emptyChar = "▓", "░"
	}

	return bc + strings.Repeat(filledChar, filled) + DIM + strings.Repeat(emptyChar, empty) + RST
}

// FormatCost formats a USD cost as "~$2.50".
func FormatCost(usd float64) string {
	if usd >= 100 {
		return fmt.Sprintf("~$%.0f", usd)
	}
	if usd >= 10 {
		return fmt.Sprintf("~$%.1f", usd)
	}
	return fmt.Sprintf("~$%.2f", usd)
}
