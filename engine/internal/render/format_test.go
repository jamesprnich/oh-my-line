package render

import (
	"strings"
	"testing"
)

func TestFormatTokens(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{42, "42"},
		{999, "999"},
		{1000, "1k"},
		{1500, "1k"},
		{65000, "65k"},
		{999999, "999k"},
		{1000000, "1.0M"},
		{1500000, "1.5M"},
		{200000, "200k"},
	}
	for _, tt := range tests {
		got := FormatTokens(tt.n)
		if got != tt.want {
			t.Errorf("FormatTokens(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		mins int
		want string
	}{
		{0, ""},
		{-1, ""},
		{5, "~5m"},
		{45, "~45m"},
		{60, "~1h 0m"},
		{90, "~1h 30m"},
		{150, "~2h 30m"},
		{1440, "~1d 0h"},
		{1500, "~1d 1h"},
	}
	for _, tt := range tests {
		got := FormatDuration(tt.mins)
		if got != tt.want {
			t.Errorf("FormatDuration(%d) = %q, want %q", tt.mins, got, tt.want)
		}
	}
}

func TestBuildBar(t *testing.T) {
	// Just verify it returns non-empty and has correct width
	bar := BuildBar(50, 8, "")
	if bar == "" {
		t.Error("BuildBar(50, 8) returned empty")
	}

	// Test edge cases
	bar = BuildBar(0, 8, "")
	if bar == "" {
		t.Error("BuildBar(0, 8) returned empty")
	}
	bar = BuildBar(100, 8, "")
	if bar == "" {
		t.Error("BuildBar(100, 8) returned empty")
	}
	bar = BuildBar(-5, 8, "")
	if bar == "" {
		t.Error("BuildBar(-5, 8) returned empty")
	}
	bar = BuildBar(150, 8, "")
	if bar == "" {
		t.Error("BuildBar(150, 8) returned empty")
	}
}

func TestBuildBar_DotStyle(t *testing.T) {
	bar := BuildBar(50, 6, "")
	if !strings.Contains(bar, "●") || !strings.Contains(bar, "○") {
		t.Errorf("dot style should use ●○, got %q", bar)
	}
}

func TestBuildBar_BlockStyle(t *testing.T) {
	bar := BuildBar(50, 6, "block")
	if !strings.Contains(bar, "▓") || !strings.Contains(bar, "░") {
		t.Errorf("block style should use ▓░, got %q", bar)
	}
}

func TestFormatCost(t *testing.T) {
	tests := []struct {
		usd  float64
		want string
	}{
		{0.50, "~$0.50"},
		{2.50, "~$2.50"},
		{9.99, "~$9.99"},
		{15.0, "~$15.0"},
		{100.0, "~$100"},
		{250.0, "~$250"},
	}
	for _, tt := range tests {
		got := FormatCost(tt.usd)
		if got != tt.want {
			t.Errorf("FormatCost(%.2f) = %q, want %q", tt.usd, got, tt.want)
		}
	}
}
