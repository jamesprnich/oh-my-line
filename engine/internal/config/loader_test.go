package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParse_ModernFormat(t *testing.T) {
	json := `{
		"nerdFont": true,
		"oh-my-lines": [
			{
				"separator": "|",
				"segments": [
					{"type": "model"},
					{"type": "tokens"}
				]
			}
		]
	}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !conf.NerdFont {
		t.Error("nerdFont should be true")
	}
	if len(conf.Lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(conf.Lines))
	}
	if len(conf.Lines[0].Segments) != 2 {
		t.Errorf("expected 2 segments, got %d", len(conf.Lines[0].Segments))
	}
	if conf.Lines[0].Separator != "|" {
		t.Errorf("separator should be |, got %q", conf.Lines[0].Separator)
	}
}

func TestParse_LegacyFormat(t *testing.T) {
	json := `{
		"icon": "🤖",
		"label": "My App",
		"tagline": "Cool stuff",
		"statusline": {
			"lines": [
				{
					"segments": [{"type": "model"}]
				}
			],
			"presets": {
				"main": {"backgroundStyle": "solid", "backgroundColor": "#1a1a2e"}
			},
			"messages": ["hello", "world"],
			"messageInterval": 60
		}
	}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(conf.Lines) != 1 {
		t.Fatalf("legacy lines should be normalized, got %d", len(conf.Lines))
	}
	if conf.MetaIcon != "🤖" {
		t.Errorf("MetaIcon = %q, want 🤖", conf.MetaIcon)
	}
	if conf.MetaLabel != "My App" {
		t.Errorf("MetaLabel = %q, want My App", conf.MetaLabel)
	}
	if conf.MetaTagline != "Cool stuff" {
		t.Errorf("MetaTagline = %q, want Cool stuff", conf.MetaTagline)
	}
	if len(conf.ResolvedMessages) != 2 {
		t.Errorf("messages should have 2, got %d", len(conf.ResolvedMessages))
	}
	if conf.MessageInterval != 60 {
		t.Errorf("interval = %d, want 60", conf.MessageInterval)
	}
	if _, ok := conf.Presets["main"]; !ok {
		t.Error("presets should include 'main'")
	}
}

func TestParse_SegmentIdentityOverrides(t *testing.T) {
	json := `{
		"icon": "🤖",
		"label": "Default",
		"oh-my-lines": [
			{
				"segments": [
					{"type": "icon", "content": "🔥"},
					{"type": "label", "content": "Override"},
					{"type": "tagline", "content": "New tagline"}
				]
			}
		]
	}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if conf.MetaIcon != "🔥" {
		t.Errorf("MetaIcon = %q, want 🔥 (segment override)", conf.MetaIcon)
	}
	if conf.MetaLabel != "Override" {
		t.Errorf("MetaLabel = %q, want Override", conf.MetaLabel)
	}
	if conf.MetaTagline != "New tagline" {
		t.Errorf("MetaTagline = %q, want New tagline", conf.MetaTagline)
	}
}

func TestParse_MessageSegment(t *testing.T) {
	json := `{
		"oh-my-lines": [
			{
				"segments": [
					{"type": "message", "messages": ["a", "b", "c"], "interval": 120}
				]
			}
		]
	}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(conf.ResolvedMessages) != 3 {
		t.Errorf("messages = %d, want 3", len(conf.ResolvedMessages))
	}
	if conf.MessageInterval != 120 {
		t.Errorf("interval = %d, want 120", conf.MessageInterval)
	}
	if conf.CurrentMessage == "" {
		t.Error("CurrentMessage should be resolved")
	}
}

func TestParse_PresetsInitialized(t *testing.T) {
	json := `{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if conf.Presets == nil {
		t.Error("Presets should be initialized even if empty")
	}
}

func TestParse_InvalidJSON(t *testing.T) {
	_, err := Parse([]byte("not json"))
	if err == nil {
		t.Error("should fail on invalid JSON")
	}
}

func TestParse_SegmentOptions(t *testing.T) {
	json := `{
		"oh-my-lines": [{
			"segments": [{
				"type": "model",
				"style": {"color": "#ff0000", "bold": true},
				"prefix": "M:",
				"suffix": "!",
				"padLeft": 2,
				"padRight": 3
			}]
		}]
	}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	seg := conf.Lines[0].Segments[0]
	if seg.Style == nil || seg.Style.Color != "#ff0000" || !seg.Style.Bold {
		t.Errorf("style not parsed correctly: %+v", seg.Style)
	}
	if seg.Prefix != "M:" {
		t.Errorf("prefix = %q, want M:", seg.Prefix)
	}
	if seg.Suffix != "!" {
		t.Errorf("suffix = %q, want !", seg.Suffix)
	}
	if seg.PadLeft != 2 || seg.PadRight != 3 {
		t.Errorf("padding = %d/%d, want 2/3", seg.PadLeft, seg.PadRight)
	}
}

func TestParse_SpecialLineTypes(t *testing.T) {
	json := `{
		"oh-my-lines": [
			{"type": "rule", "char": "═", "width": 50},
			{"type": "spacer"},
			{"type": "art", "lines": ["line1", "line2"]}
		]
	}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(conf.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(conf.Lines))
	}
	if conf.Lines[0].Type != "rule" || conf.Lines[0].Char != "═" || conf.Lines[0].Width != 50 {
		t.Errorf("rule not parsed: %+v", conf.Lines[0])
	}
	if conf.Lines[1].Type != "spacer" {
		t.Errorf("spacer type = %q", conf.Lines[1].Type)
	}
	if conf.Lines[2].Type != "art" || len(conf.Lines[2].Lines) != 2 {
		t.Errorf("art not parsed: %+v", conf.Lines[2])
	}
}

func TestParse_EmptyConfig(t *testing.T) {
	json := `{}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(conf.Lines) != 0 {
		t.Errorf("empty config should have 0 lines, got %d", len(conf.Lines))
	}
}

// ── Security: Trusted config flag ────────────────────────────────────────────
// Only ~/.oh-my-line/config.json is trusted (can run command segments).
// Project-level oh-my-line.json must NEVER be trusted — a cloned repo could
// contain malicious command segments that execute arbitrary code.

func TestLoad_ProjectConfigIsUntrusted(t *testing.T) {
	dir := t.TempDir()
	configJSON := `{"oh-my-lines": [{"segments": [{"type": "command", "content": "echo pwned"}]}]}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	conf, err := Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Trusted {
		t.Fatal("SECURITY: project-level config must NOT be trusted — command segments could execute arbitrary code")
	}
}

func TestLoad_ProjectConfigCommandSegmentBlocked(t *testing.T) {
	// End-to-end: a project config with a command segment should parse but
	// the Trusted flag must be false, ensuring the render layer blocks execution.
	dir := t.TempDir()
	configJSON := `{"oh-my-lines": [{"segments": [
		{"type": "command", "content": "rm -rf /"},
		{"type": "model"}
	]}]}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	conf, err := Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Trusted {
		t.Fatal("SECURITY: project config with command segments must not be trusted")
	}
	// Verify the command segment is still in the config (it's not stripped,
	// just blocked at render time via the Trusted flag)
	found := false
	for _, line := range conf.Lines {
		for _, seg := range line.Segments {
			if seg.Type == "command" {
				found = true
			}
		}
	}
	if !found {
		t.Error("command segment should still be in parsed config (blocked at render, not parse)")
	}
}

func TestLoad_GlobalConfigIsTrusted(t *testing.T) {
	// We can't easily test the real ~/.oh-my-line/config.json path in a unit test,
	// but we verify the loader's candidate list logic: the global path gets trusted=true.
	// This test creates a config at the global path in a temp HOME.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	configJSON := `{"oh-my-lines": [{"segments": [{"type": "command", "content": "date"}]}]}`
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(configJSON), 0644)

	// Load with a cwd that has no config — should fall back to global
	conf, err := Load(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !conf.Trusted {
		t.Fatal("SECURITY: global config (~/.oh-my-line/config.json) must be trusted")
	}
}

func TestLoad_ProjectConfigTakesPriorityButStaysUntrusted(t *testing.T) {
	// When both project and global configs exist, project wins but must be untrusted.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create global config (trusted)
	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	// Create project config (untrusted)
	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "command", "content": "echo attack"}]}]}`), 0644)

	conf, err := Load(projectDir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Trusted {
		t.Fatal("SECURITY: project config must NOT be trusted even when global config also exists")
	}
	// Verify it loaded the project config (has command segment), not the global one
	hasCommand := false
	for _, line := range conf.Lines {
		for _, seg := range line.Segments {
			if seg.Type == "command" {
				hasCommand = true
			}
		}
	}
	if !hasCommand {
		t.Error("should have loaded project config (with command segment), not global")
	}
}

func TestLoad_TrustedFlagCannotBeSetViaJSON(t *testing.T) {
	// SECURITY: Even if a project config includes "trusted": true in JSON,
	// it must NOT override the loader's trust decision. The Trusted field
	// has json:"-" to prevent this, but we test the behaviour explicitly.
	dir := t.TempDir()
	configJSON := `{
		"trusted": true,
		"oh-my-lines": [{"segments": [{"type": "command", "content": "echo hacked"}]}]
	}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	conf, err := Load(dir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Trusted {
		t.Fatal("SECURITY: Trusted flag must NEVER be settable via JSON — json:\"-\" tag may have been removed")
	}
}

func TestParse_TrustedDefaultsFalse(t *testing.T) {
	// The Trusted field must default to false after parsing.
	// If it ever defaults to true, all configs become trusted.
	conf, err := Parse([]byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if conf.Trusted {
		t.Fatal("SECURITY: Trusted must default to false after Parse — only Load should set it to true")
	}
}

// ── Config lookup fallback chain ─────────────────────────────────────────────

func TestLoad_FallbackToGlobalWhenNoProjectConfig(t *testing.T) {
	// No oh-my-line.json in cwd — should fall back to global config and load its content.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	configJSON := `{"oh-my-lines": [{"segments": [{"type": "model"}, {"type": "tokens"}]}]}`
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(configJSON), 0644)

	// cwd has no oh-my-line.json
	emptyCwd := t.TempDir()
	conf, err := Load(emptyCwd, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(conf.Lines) != 1 {
		t.Fatalf("expected 1 line from global config, got %d", len(conf.Lines))
	}
	if len(conf.Lines[0].Segments) != 2 {
		t.Errorf("expected 2 segments from global config, got %d", len(conf.Lines[0].Segments))
	}
	if conf.Lines[0].Segments[0].Type != "model" {
		t.Errorf("expected model segment from global config, got %q", conf.Lines[0].Segments[0].Type)
	}
}

func TestLoad_NoConfigAnywhere(t *testing.T) {
	// No config in cwd or global — should return error.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	emptyCwd := t.TempDir()
	_, err := Load(emptyCwd, "")
	if err == nil {
		t.Fatal("expected error when no config exists anywhere")
	}
}

func TestLoad_ProjectConfigContentTakesPriority(t *testing.T) {
	// Both project and global exist — project content should be loaded, not global.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Global config has "model" segment
	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	// Project config has "tokens" segment
	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "tokens"}]}]}`), 0644)

	conf, err := Load(projectDir, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if len(conf.Lines[0].Segments) != 1 || conf.Lines[0].Segments[0].Type != "tokens" {
		t.Errorf("expected tokens segment from project config, got %q", conf.Lines[0].Segments[0].Type)
	}
}

func TestLoad_EmptyCwdFallsBackToGlobal(t *testing.T) {
	// Empty string cwd — should skip project lookup entirely, load global.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	conf, err := Load("", "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !conf.Trusted {
		t.Error("global config should be trusted")
	}
	if conf.Lines[0].Segments[0].Type != "model" {
		t.Errorf("expected model segment from global, got %q", conf.Lines[0].Segments[0].Type)
	}
}

func TestLoad_InvalidProjectConfigSkipsToGlobal(t *testing.T) {
	// Project config has invalid JSON — should skip it and fall back to global.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "oh-my-line.json"), []byte(`{not valid json`), 0644)

	conf, err := Load(projectDir, "")
	if err != nil {
		t.Fatalf("Load failed — should have fallen back to global: %v", err)
	}
	if !conf.Trusted {
		t.Error("should have loaded global config (trusted) after invalid project config")
	}
	if conf.Lines[0].Segments[0].Type != "model" {
		t.Errorf("expected model segment from global fallback, got %q", conf.Lines[0].Segments[0].Type)
	}
}

// ── Config option parsing ────────────────────────────────────────────────────

func TestParse_NerdFontFlag(t *testing.T) {
	cases := []struct {
		json string
		want bool
	}{
		{`{"nerdFont": true, "oh-my-lines": []}`, true},
		{`{"nerdFont": false, "oh-my-lines": []}`, false},
		{`{"oh-my-lines": []}`, false},
	}
	for _, tc := range cases {
		conf, err := Parse([]byte(tc.json))
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if conf.NerdFont != tc.want {
			t.Errorf("json=%q: NerdFont=%v, want %v", tc.json, conf.NerdFont, tc.want)
		}
	}
}

func TestParse_DebugFlag(t *testing.T) {
	conf, err := Parse([]byte(`{"debug": true, "oh-my-lines": []}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if !conf.Debug {
		t.Error("debug should be true")
	}
}

func TestParse_UsageProxy(t *testing.T) {
	conf, err := Parse([]byte(`{"usageProxy": {"claudeCode": "http://localhost:8787"}, "oh-my-lines": []}`))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if conf.UsageProxy == nil || conf.UsageProxy["claudeCode"] != "http://localhost:8787" {
		t.Errorf("usageProxy not parsed: %+v", conf.UsageProxy)
	}
}

func TestParse_LineBackgroundStyles(t *testing.T) {
	json := `{"oh-my-lines": [
		{"backgroundStyle": "gradient", "background": "#1a1a4a", "segments": [{"type": "model"}]},
		{"backgroundStyle": "solid", "background": "#ff0000", "segments": [{"type": "model"}]},
		{"backgroundStyle": "neon", "segments": [{"type": "model"}]}
	]}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	if len(conf.Lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(conf.Lines))
	}
	if conf.Lines[0].BackgroundStyle != "gradient" || conf.Lines[0].Background != "#1a1a4a" {
		t.Errorf("line 0: style=%q bg=%q", conf.Lines[0].BackgroundStyle, conf.Lines[0].Background)
	}
	if conf.Lines[1].BackgroundStyle != "solid" || conf.Lines[1].Background != "#ff0000" {
		t.Errorf("line 1: style=%q bg=%q", conf.Lines[1].BackgroundStyle, conf.Lines[1].Background)
	}
	if conf.Lines[2].BackgroundStyle != "neon" {
		t.Errorf("line 2: style=%q", conf.Lines[2].BackgroundStyle)
	}
}

func TestParse_SeparatorStyles(t *testing.T) {
	json := `{"oh-my-lines": [{"separator": "|", "separatorStyle": {"dim": true, "color": "#888888"}, "segments": [{"type": "model"}]}]}`
	conf, err := Parse([]byte(json))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}
	line := conf.Lines[0]
	if line.Separator != "|" {
		t.Errorf("separator = %q, want |", line.Separator)
	}
	if line.SeparatorStyle == nil || !line.SeparatorStyle.Dim || line.SeparatorStyle.Color != "#888888" {
		t.Errorf("separatorStyle not parsed: %+v", line.SeparatorStyle)
	}
}

// ── .product.json edge cases ─────────────────────────────────────────────────

func TestLoadWithProduct_ContentOverridesSource(t *testing.T) {
	// When both content and source are set, content wins.
	dir := t.TempDir()

	configJSON := `{"oh-my-lines": [{"segments": [
		{"type": "label", "content": "Hardcoded", "source": ".product.json"}
	]}]}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	productJSON := `{"name": "From Product"}`
	os.WriteFile(filepath.Join(dir, ".product.json"), []byte(productJSON), 0644)

	conf, err := LoadWithProduct(dir, "")
	if err != nil {
		t.Fatalf("LoadWithProduct failed: %v", err)
	}
	// content is set, so source should be ignored (loader.go line 149: seg.Content == "")
	if conf.MetaLabel != "Hardcoded" {
		t.Errorf("MetaLabel = %q, want Hardcoded (content should override source)", conf.MetaLabel)
	}
}

func TestLoadWithProduct_MissingProductJson(t *testing.T) {
	// Config references .product.json but file doesn't exist — should not error.
	dir := t.TempDir()

	configJSON := `{"oh-my-lines": [{"segments": [
		{"type": "icon", "source": ".product.json"},
		{"type": "label", "source": ".product.json"}
	]}]}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)
	// No .product.json created

	conf, err := LoadWithProduct(dir, "")
	if err != nil {
		t.Fatalf("LoadWithProduct failed: %v", err)
	}
	if conf.MetaIcon != "" {
		t.Errorf("MetaIcon should be empty when .product.json missing, got %q", conf.MetaIcon)
	}
	if conf.MetaLabel != "" {
		t.Errorf("MetaLabel should be empty when .product.json missing, got %q", conf.MetaLabel)
	}
}

func TestLoadWithProduct_SpecCompliant(t *testing.T) {
	// Create temp dir with oh-my-line.json and .product.json
	dir := t.TempDir()

	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "icon", "source": ".product.json"},
				{"type": "label", "source": ".product.json"},
				{"type": "tagline", "source": ".product.json"},
				{"type": "message", "source": ".product.json"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	// product-json-spec compliant: messages under extensions.oh-my-line
	productJSON := `{
		"$schema": "https://raw.githubusercontent.com/jamesprnich/product-json-spec/main/schema.json",
		"icon": "🦙",
		"name": "Llama App",
		"tagline": "Run LLMs locally",
		"extensions": {
			"oh-my-line": {
				"messages": ["Pull. Run. Done.", "No cloud required."],
				"messageInterval": 120
			}
		}
	}`
	os.WriteFile(filepath.Join(dir, ".product.json"), []byte(productJSON), 0644)

	conf, err := LoadWithProduct(dir, "")
	if err != nil {
		t.Fatalf("LoadWithProduct failed: %v", err)
	}
	if conf.MetaIcon != "🦙" {
		t.Errorf("MetaIcon = %q, want 🦙", conf.MetaIcon)
	}
	if conf.MetaLabel != "Llama App" {
		t.Errorf("MetaLabel = %q, want Llama App", conf.MetaLabel)
	}
	if conf.MetaTagline != "Run LLMs locally" {
		t.Errorf("MetaTagline = %q, want Run LLMs locally", conf.MetaTagline)
	}
	if len(conf.ResolvedMessages) != 2 {
		t.Errorf("ResolvedMessages = %d, want 2", len(conf.ResolvedMessages))
	}
	if conf.MessageInterval != 120 {
		t.Errorf("MessageInterval = %d, want 120", conf.MessageInterval)
	}
	if conf.CurrentMessage == "" {
		t.Error("CurrentMessage should be resolved from product.json extensions")
	}
}

func TestLoadWithProduct_NoExtensions(t *testing.T) {
	// product.json without extensions — messages should not be set
	dir := t.TempDir()

	configJSON := `{
		"oh-my-lines": [{
			"segments": [
				{"type": "icon", "source": ".product.json"},
				{"type": "message", "source": ".product.json"}
			]
		}]
	}`
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(configJSON), 0644)

	productJSON := `{
		"icon": "🚀",
		"name": "Rocket"
	}`
	os.WriteFile(filepath.Join(dir, ".product.json"), []byte(productJSON), 0644)

	conf, err := LoadWithProduct(dir, "")
	if err != nil {
		t.Fatalf("LoadWithProduct failed: %v", err)
	}
	if conf.MetaIcon != "🚀" {
		t.Errorf("MetaIcon = %q, want 🚀", conf.MetaIcon)
	}
	if len(conf.ResolvedMessages) != 0 {
		t.Errorf("ResolvedMessages should be empty without extensions, got %d", len(conf.ResolvedMessages))
	}
}

// ── Per-account config (CLAUDE_CONFIG_DIR) ───────────────────────────────────

func TestLoad_AccountConfigIsTrusted(t *testing.T) {
	// A config in CLAUDE_CONFIG_DIR should be trusted (user controls it).
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	configDir := t.TempDir() // simulates a non-default CLAUDE_CONFIG_DIR
	configJSON := `{"oh-my-lines": [{"segments": [{"type": "command", "content": "date"}]}]}`
	os.WriteFile(filepath.Join(configDir, "oh-my-line.json"), []byte(configJSON), 0644)

	conf, err := Load(t.TempDir(), configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if !conf.Trusted {
		t.Fatal("SECURITY: config in CLAUDE_CONFIG_DIR should be trusted — user controls it")
	}
}

func TestLoad_AccountConfigPriorityBetweenProjectAndGlobal(t *testing.T) {
	// Priority: project (untrusted) > account (trusted) > global (trusted)
	// When no project config exists, account config should win over global.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Global config has "model" segment
	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	// Account config has "tokens" segment
	configDir := t.TempDir()
	os.WriteFile(filepath.Join(configDir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "tokens"}]}]}`), 0644)

	// cwd has no config
	conf, err := Load(t.TempDir(), configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Lines[0].Segments[0].Type != "tokens" {
		t.Errorf("expected tokens from account config, got %q", conf.Lines[0].Segments[0].Type)
	}
	if !conf.Trusted {
		t.Error("account config should be trusted")
	}
}

func TestLoad_ProjectConfigOverridesAccountConfig(t *testing.T) {
	// Project config (untrusted) should take priority over account config (trusted).
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Account config
	configDir := t.TempDir()
	os.WriteFile(filepath.Join(configDir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "tokens"}]}]}`), 0644)

	// Project config
	projectDir := t.TempDir()
	os.WriteFile(filepath.Join(projectDir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	conf, err := Load(projectDir, configDir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Lines[0].Segments[0].Type != "model" {
		t.Errorf("expected model from project config, got %q", conf.Lines[0].Segments[0].Type)
	}
	if conf.Trusted {
		t.Fatal("SECURITY: project config must stay untrusted even when account config exists")
	}
}

func TestLoad_DefaultConfigDirSkipsAccountLookup(t *testing.T) {
	// When configDir is ~/.claude (the default), the account config lookup
	// should be skipped — the global fallback already handles this case.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Create oh-my-line.json in ~/.claude — this should NOT be found
	defaultConfigDir := filepath.Join(tmpHome, ".claude")
	os.MkdirAll(defaultConfigDir, 0755)
	os.WriteFile(filepath.Join(defaultConfigDir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "tokens"}]}]}`), 0644)

	// No global config, no project config
	_, err := Load(t.TempDir(), defaultConfigDir)
	if err == nil {
		t.Fatal("should not find config via default configDir — that path is not a valid config location")
	}
}

func TestLoad_EmptyConfigDirSkipsAccountLookup(t *testing.T) {
	// Empty configDir should skip account lookup entirely.
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	// Only global config exists
	globalDir := filepath.Join(tmpHome, ".oh-my-line")
	os.MkdirAll(globalDir, 0755)
	os.WriteFile(filepath.Join(globalDir, "config.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "model"}]}]}`), 0644)

	conf, err := Load(t.TempDir(), "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if conf.Lines[0].Segments[0].Type != "model" {
		t.Errorf("expected model from global config, got %q", conf.Lines[0].Segments[0].Type)
	}
}

func TestLoad_ConfigDirSameAsCwdNotDuplicated(t *testing.T) {
	// If configDir == cwd, the config should only be checked once (as project/untrusted).
	origHome := os.Getenv("HOME")
	tmpHome := t.TempDir()
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", origHome)

	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "oh-my-line.json"), []byte(`{"oh-my-lines": [{"segments": [{"type": "command", "content": "date"}]}]}`), 0644)

	conf, err := Load(dir, dir)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	// Must be untrusted — loaded as project config, not account config
	if conf.Trusted {
		t.Fatal("SECURITY: when configDir == cwd, config must load as untrusted project config")
	}
}

