package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jamesprnich/oh-my-line/engine/internal"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

// Load finds and parses the oh-my-line config file.
// Priority: cwd/oh-my-line.json > ~/.oh-my-line/config.json
func Load(cwd string) (*internal.Config, error) {
	home, _ := os.UserHomeDir()

	type candidate struct {
		path    string
		trusted bool
	}

	var candidates []candidate
	if cwd != "" {
		candidates = append(candidates, candidate{filepath.Join(cwd, "oh-my-line.json"), false})
	}
	if home != "" {
		candidates = append(candidates, candidate{filepath.Join(home, ".oh-my-line", "config.json"), true})
	}

	for _, c := range candidates {
		data, err := os.ReadFile(c.path)
		if err != nil {
			continue
		}
		conf, err := Parse(data)
		if err != nil {
			continue
		}
		conf.Trusted = c.trusted
		return conf, nil
	}

	return nil, fmt.Errorf("no config found")
}

// Parse parses config JSON and normalizes it to the internal format.
func Parse(data []byte) (*internal.Config, error) {
	var conf internal.Config
	if err := json.Unmarshal(data, &conf); err != nil {
		return nil, err
	}

	// Normalize: if "oh-my-lines" is present, use it directly.
	// If "statusline.lines" is present (legacy format), copy to Lines.
	if len(conf.Lines) == 0 && conf.Statusline != nil && len(conf.Statusline.Lines) > 0 {
		conf.Lines = conf.Statusline.Lines
		conf.Presets = conf.Statusline.Presets
		conf.ResolvedMessages = conf.Statusline.Messages
		conf.MessageInterval = conf.Statusline.MessageInterval
	}

	// Resolve presets if not already set
	if conf.Presets == nil {
		conf.Presets = make(map[string]internal.PresetConf)
	}

	// Resolve identity from config-level fields
	conf.MetaIcon = conf.Icon
	conf.MetaLabel = conf.Label
	conf.MetaTagline = conf.Tagline

	// Scan segments for identity overrides (content > source)
	for _, line := range conf.Lines {
		for _, seg := range line.Segments {
			switch seg.Type {
			case "icon":
				if seg.Content != "" {
					conf.MetaIcon = seg.Content
				}
			case "label":
				if seg.Content != "" {
					conf.MetaLabel = seg.Content
				}
			case "tagline":
				if seg.Content != "" {
					conf.MetaTagline = seg.Content
				}
			case "message":
				if len(seg.Messages) > 0 {
					conf.ResolvedMessages = seg.Messages
					if seg.Interval > 0 {
						conf.MessageInterval = seg.Interval
					}
				}
			}
		}
	}

	// Resolve product.json sources
	// (deferred — needs cwd context, handled in LoadWithProduct)

	// Resolve current message
	conf.CurrentMessage = render.ResolveCurrentMessage(conf.ResolvedMessages, conf.MessageInterval)

	return &conf, nil
}

// LoadWithProduct loads config and resolves .product.json references.
func LoadWithProduct(cwd string) (*internal.Config, error) {
	conf, err := Load(cwd)
	if err != nil {
		return nil, err
	}

	// Try to load .product.json for source resolution
	// Follows https://github.com/jamesprnich/product-json-spec
	if cwd != "" {
		productPath := filepath.Join(cwd, ".product.json")
		if productData, err := os.ReadFile(productPath); err == nil {
			var product struct {
				Icon       string `json:"icon"`
				Name       string `json:"name"`
				Tagline    string `json:"tagline"`
				Extensions map[string]json.RawMessage `json:"extensions"`
			}
			if json.Unmarshal(productData, &product) == nil {
				// Parse oh-my-line extension data (messages, messageInterval)
				var omlExt struct {
					Messages        []string `json:"messages"`
					MessageInterval int      `json:"messageInterval"`
				}
				if raw, ok := product.Extensions["oh-my-line"]; ok {
					json.Unmarshal(raw, &omlExt)
				}

				// Scan segments for source: ".product.json"
				for _, line := range conf.Lines {
					for _, seg := range line.Segments {
						if seg.Source != ".product.json" {
							continue
						}
						switch seg.Type {
						case "icon":
							if seg.Content == "" && product.Icon != "" {
								conf.MetaIcon = product.Icon
							}
						case "label":
							if seg.Content == "" && product.Name != "" {
								conf.MetaLabel = product.Name
							}
						case "tagline":
							if seg.Content == "" && product.Tagline != "" {
								conf.MetaTagline = product.Tagline
							}
						case "message":
							if len(seg.Messages) == 0 && len(omlExt.Messages) > 0 {
								conf.ResolvedMessages = omlExt.Messages
								if omlExt.MessageInterval > 0 {
									conf.MessageInterval = omlExt.MessageInterval
								}
								conf.CurrentMessage = render.ResolveCurrentMessage(conf.ResolvedMessages, conf.MessageInterval)
							}
						}
					}
				}
			}
		}
	}

	return conf, nil
}
