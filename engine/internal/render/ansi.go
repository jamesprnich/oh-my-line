package render

import (
	"fmt"
	"unicode/utf8"
)

const (
	RST   = "\033[0m"
	BOLD  = "\033[1m"
	DIM   = "\033[2m"
	NODIM = "\033[22m"
)

// HexFG returns a 24-bit foreground ANSI escape for a "#RRGGBB" hex color.
func HexFG(hex string) string {
	r, g, b := ParseHex(hex)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

// HexBG returns a 24-bit background ANSI escape for a "#RRGGBB" hex color.
func HexBG(hex string) string {
	r, g, b := ParseHex(hex)
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// VisibleLen returns the number of visible columns in s after stripping
// ANSI CSI sequences (\033[...m) and OSC sequences (\033]...\007).
// Counts Unicode runes (not bytes) for correct multi-byte character handling.
func VisibleLen(s string) int {
	n := 0
	i := 0
	for i < len(s) {
		if s[i] == '\033' {
			if i+1 < len(s) && s[i+1] == '[' {
				// CSI sequence: skip until 'm'
				i += 2
				for i < len(s) && s[i] != 'm' {
					i++
				}
				if i < len(s) {
					i++ // skip 'm'
				}
				continue
			}
			if i+1 < len(s) && s[i+1] == ']' {
				// OSC sequence: skip until BEL (\007)
				i += 2
				for i < len(s) && s[i] != '\007' {
					i++
				}
				if i < len(s) {
					i++ // skip BEL
				}
				continue
			}
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		n++
		i += size
	}
	return n
}

// ParseHex parses "#RRGGBB" into r, g, b components.
func ParseHex(hex string) (uint8, uint8, uint8) {
	if len(hex) > 0 && hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) != 6 {
		return 255, 255, 255
	}
	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return r, g, b
}
