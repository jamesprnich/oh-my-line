package render

import (
	"fmt"
	"strings"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

// RenderLine renders a single line (normal, art, rule, or spacer).
func RenderLine(line internal.LineConf, input *internal.Input, conf *internal.Config, lineIdx int) string {
	switch line.Type {
	case "spacer":
		return ""
	case "rule":
		return renderRule(line)
	case "art":
		return renderArt(line)
	default:
		return renderNormalLine(line, input, conf, lineIdx)
	}
}

func renderRule(line internal.LineConf) string {
	char := line.Char
	if char == "" {
		char = "─"
	}
	width := line.Width
	if width <= 0 {
		width = 120
	}
	color := "#555555"
	dim := true
	if line.Style != nil {
		if line.Style.Color != "" {
			color = line.Style.Color
		}
		dim = line.Style.Dim
	}
	fg := HexFG(color)
	dm := ""
	if dim {
		dm = DIM
	}
	return fg + dm + strings.Repeat(char, width) + RST
}

func renderArt(line internal.LineConf) string {
	if len(line.Lines) == 0 {
		return ""
	}
	fg := ""
	dm := ""
	if line.Style != nil {
		if line.Style.Color != "" {
			fg = HexFG(line.Style.Color)
		}
		if line.Style.Dim {
			dm = DIM
		}
	}
	var rows []string
	for _, row := range line.Lines {
		rows = append(rows, fg+dm+row+RST)
	}
	return strings.Join(rows, "\n")
}

func renderNormalLine(line internal.LineConf, input *internal.Input, conf *internal.Config, lineIdx int) string {
	// Resolve preset
	bgStyle := line.BackgroundStyle
	bgHex := line.Background
	labelColor := ""
	if line.Preset != "" && conf.Presets != nil {
		if p, ok := conf.Presets[line.Preset]; ok {
			if bgStyle == "" {
				bgStyle = p.BackgroundStyle
			}
			if bgHex == "" {
				bgHex = p.BackgroundColor
			}
			labelColor = p.LabelColor
		}
	}

	// Render segments
	var parts []string
	segIdx := 0
	for _, seg := range line.Segments {
		rendered := RenderSegment(seg, input, conf)
		if rendered != "" {
			if conf.EmitMarkers {
				rendered = fmt.Sprintf("\x1b]9;%d;%d\x07", lineIdx, segIdx) + rendered
			}
			parts = append(parts, rendered)
		}
		segIdx++
	}
	if len(parts) == 0 {
		return ""
	}

	// Build separator
	sep := " "
	if line.Separator != "" {
		sepColor := DIM
		if line.SeparatorStyle != nil {
			if line.SeparatorStyle.Color != "" {
				sepColor = HexFG(line.SeparatorStyle.Color)
			}
			if line.SeparatorStyle.Dim {
				sepColor += DIM
			}
		}
		sep = " " + sepColor + line.Separator + RST + " "
	}

	// Active background escape
	activeBG := ""
	lfg := ""
	switch bgStyle {
	case "solid", "fade", "gradient":
		if bgHex != "" {
			activeBG = HexBG(bgHex)
		}
	case "neon":
		activeBG = "\033[48;2;26;26;46m"
	}
	if labelColor != "" {
		lfg = HexFG(labelColor)
	}

	// Build line output
	var lo strings.Builder

	// Background prefix
	if activeBG != "" {
		lo.WriteString(activeBG + lfg + " ")
	}

	// Pad left
	if line.Padding != nil && line.Padding.Left > 0 {
		lo.WriteString(strings.Repeat(" ", line.Padding.Left))
	}

	// Separator with background restore
	bgSep := sep
	if activeBG != "" {
		bgSep = sep + activeBG + lfg
	}

	// Join segments
	for i, part := range parts {
		if i > 0 {
			lo.WriteString(bgSep)
		}
		lo.WriteString(part)
		// Re-apply line background after segment resets
		if activeBG != "" {
			lo.WriteString(activeBG + lfg)
		}
	}

	// Pad right
	if line.Padding != nil && line.Padding.Right > 0 {
		lo.WriteString(strings.Repeat(" ", line.Padding.Right))
	}

	// Background fill to terminal width
	termW := conf.TermWidth
	if termW <= 0 {
		termW = 120
	}
	contentLen := VisibleLen(lo.String())
	remaining := termW - contentLen
	if remaining < 0 {
		remaining = 0
	}

	switch bgStyle {
	case "fade":
		if bgHex != "" {
			fadeChars := "████▓▓▒▒░░  "
			fadeLen := len([]rune(fadeChars))
			bgAsFG := HexFG(bgHex)
			lo.WriteString("  " + RST + bgAsFG + fadeChars + RST)
			pad := remaining - 2 - fadeLen
			if pad > 0 {
				lo.WriteString(strings.Repeat(" ", pad))
			}
		}
	case "gradient":
		if bgHex != "" {
			r, g, b := ParseHex(bgHex)
			// Calculate gradient steps to fill remaining space
			steps := (remaining - 2) / 2 // each step is 2 chars
			if steps < 4 {
				steps = 4
			}
			if steps > 40 {
				steps = 40
			}
			var grd strings.Builder
			for s := 0; s < steps; s++ {
				gr := int(r) * (steps - s) / steps
				gg := int(g) * (steps - s) / steps
				gb := int(b) * (steps - s) / steps
				grd.WriteString(fmt.Sprintf("\033[48;2;%d;%d;%dm  ", gr, gg, gb))
			}
			lo.WriteString("  " + grd.String() + RST)
		}
	case "solid", "neon":
		if remaining > 0 {
			lo.WriteString(strings.Repeat(" ", remaining) + RST)
		}
	}

	return lo.String()
}
