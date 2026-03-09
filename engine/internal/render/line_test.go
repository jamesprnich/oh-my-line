package render

import (
	"strings"
	"testing"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

func TestRenderLine_Rule(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		Type: "rule",
		Char: "═",
		Width: 10,
	}
	got := RenderLine(line, input, conf, 0)
	if count := strings.Count(got, "═"); count != 10 {
		t.Errorf("rule should have 10 chars, got %d in %q", count, got)
	}
}

func TestRenderLine_RuleDefaults(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{Type: "rule"}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "─") {
		t.Errorf("rule with no char should use ─, got %q", got)
	}
}

func TestRenderLine_Art(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		Type:  "art",
		Lines: []string{"  ╭───╮", "  │ ! │", "  ╰───╯"},
		Style: &internal.Style{Color: "#ff0000"},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "╭───╮") {
		t.Errorf("art should contain art lines, got %q", got)
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Errorf("art should have 3 lines, got %d", len(lines))
	}
}

func TestRenderLine_ArtEmpty(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{Type: "art"}
	got := RenderLine(line, input, conf, 0)
	if got != "" {
		t.Errorf("art with no lines should be empty, got %q", got)
	}
}

func TestRenderLine_Spacer(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{Type: "spacer"}
	got := RenderLine(line, input, conf, 0)
	if got != "" {
		t.Errorf("spacer should return empty, got %q", got)
	}
}

func TestRenderLine_Normal(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		Segments: []internal.SegmentConf{
			{Type: "model"},
			{Type: "version"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "Claude Sonnet 4") || !strings.Contains(got, "1.0.32") {
		t.Errorf("normal line should contain segments, got %q", got)
	}
}

func TestRenderLine_NormalWithSeparator(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		Separator: "|",
		Segments: []internal.SegmentConf{
			{Type: "model"},
			{Type: "version"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "|") {
		t.Errorf("line with separator should contain |, got %q", got)
	}
}

func TestRenderLine_SolidBackground(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		BackgroundStyle: "solid",
		Background:      "#1a1a2e",
		Segments: []internal.SegmentConf{
			{Type: "model"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "\033[48;2;26;26;46m") {
		t.Errorf("solid bg should contain bg escape, got %q", got)
	}
}

func TestRenderLine_NeonBackground(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		BackgroundStyle: "neon",
		Segments: []internal.SegmentConf{
			{Type: "model"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	// neon uses hardcoded dark bg
	if !strings.Contains(got, "\033[48;2;26;26;46m") {
		t.Errorf("neon bg should contain dark bg escape, got %q", got)
	}
}

func TestRenderLine_FadeBackground(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		BackgroundStyle: "fade",
		Background:      "#2e9599",
		Segments: []internal.SegmentConf{
			{Type: "model"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "████▓▓▒▒░░") {
		t.Errorf("fade bg should contain fade chars, got %q", got)
	}
}

func TestRenderLine_GradientBackground(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		BackgroundStyle: "gradient",
		Background:      "#2e9599",
		Segments: []internal.SegmentConf{
			{Type: "model"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	// Gradient produces 20 bg color steps
	if !strings.Contains(got, "\033[48;2;") {
		t.Errorf("gradient bg should contain bg escapes, got %q", got)
	}
}

func TestRenderLine_Preset(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Presets["myPreset"] = internal.PresetConf{
		BackgroundStyle: "solid",
		BackgroundColor: "#1a1a2e",
	}
	line := internal.LineConf{
		Preset: "myPreset",
		Segments: []internal.SegmentConf{
			{Type: "model"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "\033[48;2;26;26;46m") {
		t.Errorf("preset should apply bg, got %q", got)
	}
}

func TestRenderLine_EmptySegments(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		Segments: []internal.SegmentConf{},
	}
	got := RenderLine(line, input, conf, 0)
	if got != "" {
		t.Errorf("line with empty segments should return empty, got %q", got)
	}
}

func TestRenderLine_Padding(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	line := internal.LineConf{
		Padding: &internal.Padding{Left: 3, Right: 5},
		Segments: []internal.SegmentConf{
			{Type: "text", Content: "X"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "   ") {
		t.Errorf("line with padding should have spaces, got %q", got)
	}
}

// ── TermWidth tests ──

func TestRenderLine_SolidBackgroundTermWidth(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.TermWidth = 60

	line := internal.LineConf{
		BackgroundStyle: "solid",
		Background:      "#1a1a2e",
		Segments: []internal.SegmentConf{
			{Type: "text", Content: "Hello"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	visLen := VisibleLen(got)
	// Should pad to exactly TermWidth
	if visLen < 58 || visLen > 62 {
		t.Errorf("solid bg with TermWidth=60 should produce ~60 visible chars, got %d", visLen)
	}
}

func TestRenderLine_GradientTermWidth(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.TermWidth = 80

	line := internal.LineConf{
		BackgroundStyle: "gradient",
		Background:      "#1a1a4a",
		Segments: []internal.SegmentConf{
			{Type: "text", Content: "Test"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	visLen := VisibleLen(got)
	// Gradient should not massively exceed TermWidth
	if visLen > 90 {
		t.Errorf("gradient with TermWidth=80 should not exceed ~80 visible chars, got %d", visLen)
	}
}

func TestRenderLine_GradientDefaultTermWidth(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	// TermWidth = 0, should default to 120

	line := internal.LineConf{
		BackgroundStyle: "gradient",
		Background:      "#1a1a4a",
		Segments: []internal.SegmentConf{
			{Type: "text", Content: "Test"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	visLen := VisibleLen(got)
	if visLen > 130 {
		t.Errorf("gradient with default TermWidth should not exceed ~120 visible chars, got %d", visLen)
	}
}

func TestRenderLine_NeonTermWidth(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.TermWidth = 50

	line := internal.LineConf{
		BackgroundStyle: "neon",
		Segments: []internal.SegmentConf{
			{Type: "text", Content: "Test"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	visLen := VisibleLen(got)
	if visLen < 48 || visLen > 52 {
		t.Errorf("neon with TermWidth=50 should produce ~50 visible chars, got %d", visLen)
	}
}

func TestRenderLine_FadeTermWidth(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.TermWidth = 80

	line := internal.LineConf{
		BackgroundStyle: "fade",
		Background:      "#2e9599",
		Segments: []internal.SegmentConf{
			{Type: "text", Content: "Test"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	visLen := VisibleLen(got)
	if visLen > 90 {
		t.Errorf("fade with TermWidth=80 should not exceed ~80 visible chars, got %d", visLen)
	}
}

// ── OSC markers ──

func TestRenderLine_EmitMarkers(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.EmitMarkers = true

	line := internal.LineConf{
		Segments: []internal.SegmentConf{
			{Type: "model"},
			{Type: "version"},
		},
	}
	got := RenderLine(line, input, conf, 0)
	if !strings.Contains(got, "\x1b]9;0;0\x07") {
		t.Errorf("should contain OSC marker for seg 0, got %q", got)
	}
	if !strings.Contains(got, "\x1b]9;0;1\x07") {
		t.Errorf("should contain OSC marker for seg 1, got %q", got)
	}
}

func TestRenderLine_EmitMarkersLineIdx(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.EmitMarkers = true

	line := internal.LineConf{
		Segments: []internal.SegmentConf{
			{Type: "model"},
		},
	}
	got := RenderLine(line, input, conf, 3)
	if !strings.Contains(got, "\x1b]9;3;0\x07") {
		t.Errorf("marker should use lineIdx=3, got %q", got)
	}
}
