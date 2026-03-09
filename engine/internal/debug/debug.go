package debug

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	enabled bool
	logPath string
)

const maxLogSize = 100 * 1024 // 100KB

// Init enables debug logging. Call once at startup after resolving config.
// Env var OML_DEBUG=1 takes precedence, then config "debug": true.
func Init(envSet, confSet bool) {
	enabled = envSet || confSet
	if !enabled {
		return
	}

	// Log next to the binary in ~/.oh-my-line/
	home, _ := os.UserHomeDir()
	if home == "" {
		enabled = false
		return
	}
	omlDir := filepath.Join(home, ".oh-my-line")
	os.MkdirAll(omlDir, 0700)
	logPath = filepath.Join(omlDir, "debug.log")

	// Auto-truncate if over max size
	if fi, err := os.Stat(logPath); err == nil && fi.Size() > maxLogSize {
		// Keep the last half
		data, err := os.ReadFile(logPath)
		if err == nil && len(data) > maxLogSize/2 {
			os.WriteFile(logPath, data[len(data)-maxLogSize/2:], 0600)
		}
	}
}

// Enabled returns whether debug logging is active.
func Enabled() bool {
	return enabled
}

// Log writes a single debug line. No-op if debug is disabled.
// Usage: debug.Log("rate", "cache=miss proxy=%s status=%d", url, code)
func Log(component, format string, args ...any) {
	if !enabled {
		return
	}
	ts := time.Now().Format("2006-01-02T15:04:05")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("%s [%s] %s\n", ts, component, msg)

	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(line)
}
