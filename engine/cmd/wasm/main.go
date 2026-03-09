//go:build js && wasm

package main

import (
	"encoding/json"
	"syscall/js"

	"github.com/jamesprnich/oh-my-line/engine/internal"
	"github.com/jamesprnich/oh-my-line/engine/internal/render"
)

func renderStatusline(this js.Value, args []js.Value) any {
	if len(args) < 2 {
		return "Claude"
	}

	configJSON := args[0].String()
	inputJSON := args[1].String()

	// Parse config
	var conf internal.Config
	if err := json.Unmarshal([]byte(configJSON), &conf); err != nil {
		return "Claude"
	}

	// Normalize config (same as config/loader.go Parse, but inline for WASM)
	if len(conf.Lines) == 0 && conf.Statusline != nil && len(conf.Statusline.Lines) > 0 {
		conf.Lines = conf.Statusline.Lines
		conf.Presets = conf.Statusline.Presets
		conf.ResolvedMessages = conf.Statusline.Messages
		conf.MessageInterval = conf.Statusline.MessageInterval
	}
	if conf.Presets == nil {
		conf.Presets = make(map[string]internal.PresetConf)
	}
	conf.MetaIcon = conf.Icon
	conf.MetaLabel = conf.Label
	conf.MetaTagline = conf.Tagline

	// Scan segments for identity overrides
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
	conf.CurrentMessage = render.ResolveCurrentMessage(conf.ResolvedMessages, conf.MessageInterval)
	conf.EmitMarkers = true

	// Optional 3rd arg: terminal width (from builder preview container)
	if len(args) >= 3 && args[2].Type() == js.TypeNumber {
		conf.TermWidth = args[2].Int()
	}

	// Parse input
	var input internal.Input
	if err := json.Unmarshal([]byte(inputJSON), &input); err != nil {
		return "Claude"
	}
	if input.Model.DisplayName == "" {
		input.Model.DisplayName = "Claude"
	}
	if input.ContextWindow.Size == 0 {
		input.ContextWindow.Size = 200000
	}

	// If input includes mock runtime data (from builder preview), wire it up
	if input.Runtime != nil {
		conf.Runtime = input.Runtime
	}

	output := render.RenderStatusline(&conf, &input)
	return output
}

func main() {
	c := make(chan struct{})
	js.Global().Set("omlRender", js.FuncOf(renderStatusline))
	<-c
}
