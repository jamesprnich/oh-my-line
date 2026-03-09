package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/jamesprnich/oh-my-line/engine/internal"
	"github.com/jamesprnich/oh-my-line/engine/internal/config"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// TestIntegration_BasicConfig tests end-to-end config parsing and rendering.
func TestIntegration_BasicConfig(t *testing.T) {
	configJSON := `{
		"nerdFont": false,
		"oh-my-lines": [
			{
				"separator": "|",
				"segments": [
					{"type": "model"},
					{"type": "tokens"},
					{"type": "pct-used"}
				]
			}
		]
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude Sonnet 4"
	input.ContextWindow.Size = 200000
	input.ContextWindow.Usage.InputTokens = 80000

	output := render.RenderStatusline(conf, input)
	if !strings.Contains(output, "Claude Sonnet 4") {
		t.Errorf("output should contain model name: %q", output)
	}
	if !strings.Contains(output, "80k/200k") {
		t.Errorf("output should contain token counts: %q", output)
	}
	if !strings.Contains(output, "40%") {
		t.Errorf("output should contain percentage: %q", output)
	}
}

// TestIntegration_MultiLine tests multi-line rendering.
func TestIntegration_MultiLine(t *testing.T) {
	configJSON := `{
		"oh-my-lines": [
			{
				"segments": [{"type": "model"}, {"type": "dir"}]
			},
			{"type": "rule", "char": "─", "width": 40},
			{
				"segments": [{"type": "tokens"}, {"type": "pct-remain"}]
			}
		]
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude Opus 4"
	input.CWD = "/home/user/project"
	input.ContextWindow.Size = 200000
	input.ContextWindow.Usage.InputTokens = 100000

	output := render.RenderStatusline(conf, input)
	lines := strings.Split(output, "\n")
	if len(lines) != 3 {
		t.Errorf("should have 3 lines, got %d", len(lines))
	}
}

// TestIntegration_LegacyConfig tests that legacy config format renders correctly.
func TestIntegration_LegacyConfig(t *testing.T) {
	configJSON := `{
		"icon": "🔧",
		"label": "Dev Tool",
		"tagline": "Build things",
		"statusline": {
			"lines": [
				{
					"segments": [
						{"type": "icon"},
						{"type": "label"},
						{"type": "tagline"}
					]
				}
			],
			"messages": ["Go fast", "Stay safe"],
			"messageInterval": 60
		}
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"

	output := render.RenderStatusline(conf, input)
	if !strings.Contains(output, "🔧") {
		t.Errorf("output should contain icon: %q", output)
	}
	if !strings.Contains(output, "Dev Tool") {
		t.Errorf("output should contain label: %q", output)
	}
	if !strings.Contains(output, "Build things") {
		t.Errorf("output should contain tagline: %q", output)
	}
}

// TestIntegration_StdinJSON tests that the input JSON parses correctly.
func TestIntegration_StdinJSON(t *testing.T) {
	stdinJSON := `{
		"model": {"display_name": "Claude Opus 4", "id": "claude-opus-4-20250514"},
		"context_window": {
			"context_window_size": 200000,
			"current_usage": {
				"input_tokens": 50000,
				"cache_creation_input_tokens": 10000,
				"cache_read_input_tokens": 20000,
				"output_tokens": 5000
			}
		},
		"cwd": "/home/user/my-project",
		"version": "1.0.32",
		"cost": {
			"total_cost_usd": 2.50,
			"total_duration_ms": 300000,
			"total_api_duration_ms": 150000,
			"total_lines_added": 42,
			"total_lines_removed": 10
		}
	}`

	var input internal.Input
	if err := json.Unmarshal([]byte(stdinJSON), &input); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if input.Model.DisplayName != "Claude Opus 4" {
		t.Errorf("model name = %q", input.Model.DisplayName)
	}
	if input.ContextWindow.Size != 200000 {
		t.Errorf("ctx size = %d", input.ContextWindow.Size)
	}
	if input.CurrentTokens() != 80000 {
		t.Errorf("current tokens = %d, want 80000", input.CurrentTokens())
	}
	if input.CWD != "/home/user/my-project" {
		t.Errorf("cwd = %q", input.CWD)
	}
	if input.Cost.TotalCostUSD != 2.50 {
		t.Errorf("cost = %f", input.Cost.TotalCostUSD)
	}
	if input.Cost.TotalLinesAdded != 42 {
		t.Errorf("lines added = %d", input.Cost.TotalLinesAdded)
	}
}

// TestIntegration_NerdFont tests that nerd font icons are included when enabled.
func TestIntegration_NerdFont(t *testing.T) {
	configJSON := `{
		"nerdFont": true,
		"oh-my-lines": [
			{
				"segments": [{"type": "model"}]
			}
		]
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"

	output := render.RenderStatusline(conf, input)
	if !strings.Contains(output, render.NerdFontIcons["model"]) {
		t.Errorf("nerdFont output should contain model icon: %q", output)
	}
}

// TestIntegration_EmptyInput tests fallback with empty/default input.
func TestIntegration_EmptyInput(t *testing.T) {
	configJSON := `{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`
	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	if !strings.Contains(output, "Claude") {
		t.Errorf("output should contain Claude: %q", output)
	}
}

// TestIntegration_BackgroundStyles tests all 4 background styles render.
func TestIntegration_BackgroundStyles(t *testing.T) {
	styles := []string{"solid", "fade", "gradient", "neon"}
	for _, style := range styles {
		configJSON := `{
			"oh-my-lines": [{
				"backgroundStyle": "` + style + `",
				"background": "#2e9599",
				"segments": [{"type": "model"}]
			}]
		}`
		conf, err := config.Parse([]byte(configJSON))
		if err != nil {
			t.Fatalf("Parse %s failed: %v", style, err)
		}

		input := &internal.Input{}
		input.Model.DisplayName = "Claude"

		output := render.RenderStatusline(conf, input)
		if output == "" {
			t.Errorf("%s style produced empty output", style)
		}
		if !strings.Contains(output, "Claude") {
			t.Errorf("%s style should contain Claude: %q", style, output)
		}
	}
}

// TestIntegration_BarStylePropagation tests that barStyle "block" on a rate-session
// segment config is correctly parsed and available.
func TestIntegration_BarStylePropagation(t *testing.T) {
	configJSON := `{
		"oh-my-lines": [{
			"segments": [{"type": "rate-session", "barStyle": "block"}]
		}]
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if conf.Lines[0].Segments[0].BarStyle != "block" {
		t.Errorf("expected barStyle %q, got %q", "block", conf.Lines[0].Segments[0].BarStyle)
	}
}

// TestIntegration_UsageProxyConfig tests that usageProxy config is correctly parsed.
func TestIntegration_UsageProxyConfig(t *testing.T) {
	configJSON := `{
		"usageProxy": {"claudeCode": "http://localhost:8787/usage"},
		"oh-my-lines": [{"segments": [{"type": "model"}]}]
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if conf.UsageProxy["claudeCode"] != "http://localhost:8787/usage" {
		t.Errorf("expected usageProxy claudeCode %q, got %q",
			"http://localhost:8787/usage", conf.UsageProxy["claudeCode"])
	}
}
