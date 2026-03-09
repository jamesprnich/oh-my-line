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

	conf, err := LoadWithProduct(dir)
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

	conf, err := LoadWithProduct(dir)
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

