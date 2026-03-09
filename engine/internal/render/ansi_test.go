package render

import "testing"

func TestParseHex(t *testing.T) {
	tests := []struct {
		input    string
		r, g, b  uint8
	}{
		{"#ff0000", 255, 0, 0},
		{"ff0000", 255, 0, 0},
		{"#00ff00", 0, 255, 0},
		{"#0000ff", 0, 0, 255},
		{"#ffffff", 255, 255, 255},
		{"#000000", 0, 0, 0},
		{"#2e9599", 46, 149, 153},
		// Invalid
		{"#fff", 255, 255, 255},
		{"", 255, 255, 255},
	}
	for _, tt := range tests {
		r, g, b := ParseHex(tt.input)
		if r != tt.r || g != tt.g || b != tt.b {
			t.Errorf("ParseHex(%q) = (%d,%d,%d), want (%d,%d,%d)", tt.input, r, g, b, tt.r, tt.g, tt.b)
		}
	}
}

func TestHexFG(t *testing.T) {
	got := HexFG("#ff0000")
	want := "\033[38;2;255;0;0m"
	if got != want {
		t.Errorf("HexFG(#ff0000) = %q, want %q", got, want)
	}
}

func TestHexBG(t *testing.T) {
	got := HexBG("#00ff00")
	want := "\033[48;2;0;255;0m"
	if got != want {
		t.Errorf("HexBG(#00ff00) = %q, want %q", got, want)
	}
}

func TestVisibleLen(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"", 0},
		{"\033[38;2;255;0;0mred\033[0m", 3},
		{"\033[1m\033[38;2;0;255;0mbold green\033[0m", 10},
		{"\033[48;2;26;26;70m  \033[0m", 2},
		// OSC marker sequences
		{"\033]9;0;1\007hello", 5},
		// Mixed
		{"\033[38;2;255;0;0mA\033[0m B \033[1mC\033[0m", 5},
	}
	for _, tt := range tests {
		got := VisibleLen(tt.input)
		if got != tt.want {
			t.Errorf("VisibleLen(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestConstants(t *testing.T) {
	if RST != "\033[0m" {
		t.Errorf("RST = %q", RST)
	}
	if BOLD != "\033[1m" {
		t.Errorf("BOLD = %q", BOLD)
	}
	if DIM != "\033[2m" {
		t.Errorf("DIM = %q", DIM)
	}
}
