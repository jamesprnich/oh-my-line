package datasource

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jamesprnich/oh-my-line/engine/internal/cache"
	"github.com/jamesprnich/oh-my-line/engine/internal/debug"
	"github.com/jamesprnich/oh-my-line/engine/internal/platform"
)

// CommandSpec describes a command segment to execute.
type CommandSpec struct {
	Content string
	Cache   int
	Timeout int
}

// CommandExecutor is a function that runs a shell command and returns output.
type CommandExecutor func(cmd string, cacheTTL, timeout int) string

// BuildCommandCache executes command segments and returns a cache map.
// Returns nil if trusted is false — commands are NEVER executed for untrusted configs.
// The executor is injected to allow testing without actual shell execution.
func BuildCommandCache(trusted bool, specs []CommandSpec, exec CommandExecutor) map[string]string {
	if !trusted {
		return nil
	}
	if len(specs) == 0 {
		return nil
	}
	cache := make(map[string]string)
	for _, s := range specs {
		if s.Content == "" {
			continue
		}
		if _, done := cache[s.Content]; !done {
			cache[s.Content] = exec(s.Content, s.Cache, s.Timeout)
		}
	}
	return cache
}

// ExecCommand executes a shell command with caching.
// Only callable for trusted configs.
func ExecCommand(cmd string, cacheTTL, timeout int) string {
	if cmd == "" {
		return ""
	}
	if cacheTTL <= 0 {
		cacheTTL = 60
	}
	if timeout <= 0 {
		timeout = 3
	} else if timeout > 30 {
		timeout = 30
	}

	cacheDir, err := cache.Dir()
	if err != nil {
		return ""
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(cmd)))[:16]
	cacheFile := filepath.Join(cacheDir, "statusline-cmd-"+hash+".dat")

	// Check cache
	if content, fresh := cache.ReadFile(cacheFile, cacheTTL); fresh && content != "" {
		debug.Log("cmd", "cache=hit cmd=%q", cmd)
		return strings.TrimSpace(content)
	}

	// Execute
	result, err := platform.RunCommand(cmd, timeout)
	if err != nil || result == "" {
		debug.Log("cmd", "exec err=%v cmd=%q", err, cmd)
		return ""
	}

	result = strings.TrimSpace(result)
	cache.WriteFile(cacheFile, result)
	debug.Log("cmd", "exec ok size=%d cmd=%q", len(result), cmd)
	return result
}
