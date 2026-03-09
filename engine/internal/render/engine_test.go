package render

import (
	"strings"
	"testing"

	"github.com/jamesprnich/oh-my-line/engine/internal"
)

func TestRenderStatusline_Basic(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Lines = []internal.LineConf{
		{
			Segments: []internal.SegmentConf{
				{Type: "model"},
				{Type: "tokens"},
			},
		},
	}
	got := RenderStatusline(conf, input)
	if !strings.Contains(got, "Claude Sonnet 4") {
		t.Errorf("should contain model name, got %q", got)
	}
}

func TestRenderStatusline_MultiLine(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Lines = []internal.LineConf{
		{
			Segments: []internal.SegmentConf{{Type: "model"}},
		},
		{
			Segments: []internal.SegmentConf{{Type: "version"}},
		},
	}
	got := RenderStatusline(conf, input)
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Errorf("should have 2 lines, got %d", len(lines))
	}
}

func TestRenderStatusline_SpacerLine(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Lines = []internal.LineConf{
		{
			Segments: []internal.SegmentConf{{Type: "model"}},
		},
		{
			Type: "spacer",
		},
		{
			Segments: []internal.SegmentConf{{Type: "version"}},
		},
	}
	got := RenderStatusline(conf, input)
	lines := strings.Split(got, "\n")
	if len(lines) != 3 {
		t.Errorf("should have 3 lines (including spacer), got %d", len(lines))
	}
	if lines[1] != "" {
		t.Errorf("spacer line should be empty, got %q", lines[1])
	}
}

func TestRenderStatusline_NoConfig(t *testing.T) {
	input := makeInput()
	conf := makeConf()
	conf.Lines = nil
	got := RenderStatusline(conf, input)
	if got != "Claude Sonnet 4" {
		t.Errorf("with no lines, should return model name, got %q", got)
	}
}

func TestRenderStatusline_AllEmptySegments(t *testing.T) {
	input := makeInput()
	input.CWD = ""
	input.Version = ""
	conf := makeConf()
	conf.Lines = []internal.LineConf{
		{
			Segments: []internal.SegmentConf{
				{Type: "dir"},
				{Type: "version"},
			},
		},
	}
	got := RenderStatusline(conf, input)
	// Both segments produce empty output, so line is empty, fallback to model name
	if got != "Claude Sonnet 4" {
		t.Errorf("with all empty segments should fallback, got %q", got)
	}
}
