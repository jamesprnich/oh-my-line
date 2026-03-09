package render

import (
	"strings"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

// RenderStatusline renders the full statusline from config and input.
func RenderStatusline(conf *internal.Config, input *internal.Input) string {
	var lines []string

	for lineIdx, line := range conf.Lines {
		rendered := RenderLine(line, input, conf, lineIdx)
		switch line.Type {
		case "spacer":
			lines = append(lines, "")
		default:
			if rendered != "" {
				lines = append(lines, rendered)
			}
		}
	}

	if len(lines) == 0 {
		return input.Model.DisplayName
	}

	return strings.Join(lines, "\n")
}
