package cache

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// Dir returns the secure cache directory path.
// Creates it if needed with mode 0700.
func Dir() (string, error) {
	d := fmt.Sprintf("/tmp/claude-%d", os.Getuid())

	// Reject symlinks and verify ownership
	fi, err := os.Lstat(d)
	if err == nil {
		if fi.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("cache dir is symlink")
		}
		if !fi.IsDir() {
			return "", fmt.Errorf("cache dir is not a directory")
		}
		if stat, ok := fi.Sys().(*syscall.Stat_t); ok {
			if stat.Uid != uint32(os.Getuid()) {
				return "", fmt.Errorf("cache dir owned by uid %d, expected %d", stat.Uid, os.Getuid())
			}
		}
		return d, nil
	}

	if err := os.Mkdir(d, 0700); err != nil {
		return "", err
	}
	return d, nil
}

// AccountKey derives a short account identifier from CLAUDE_CONFIG_DIR.
// Returns "default" for ~/.claude or empty (backward compat).
func AccountKey(configDir string) string {
	if configDir == "" {
		return "default"
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		defaultDir := filepath.Join(home, ".claude")
		// Normalize both paths for comparison
		cleanConfig := filepath.Clean(configDir)
		cleanDefault := filepath.Clean(defaultDir)
		if cleanConfig == cleanDefault {
			return "default"
		}
	}
	h := sha256.Sum256([]byte(configDir))
	return fmt.Sprintf("%x", h[:4]) // 8 hex chars
}

// AccountDir returns a per-account cache subdirectory.
// "default" or "" returns the base Dir() unchanged. Others get acct-{key}/ subdirectory.
func AccountDir(accountKey string) (string, error) {
	base, err := Dir()
	if err != nil {
		return "", err
	}
	if accountKey == "" || accountKey == "default" {
		return base, nil
	}
	d := filepath.Join(base, "acct-"+accountKey)
	if err := os.MkdirAll(d, 0700); err != nil {
		return "", err
	}
	return d, nil
}

// ReadFile reads a cache file if it exists and is within TTL.
// Returns content and true if fresh, or content and false if stale.
func ReadFile(path string, ttlSecs int) (string, bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	age := time.Since(fi.ModTime()).Seconds()
	return string(data), int(age) < ttlSecs
}

// WriteFile writes data to a cache file atomically.
func WriteFile(path, data string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(data), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

// BurnState holds the burn rate tracking state.
type BurnState struct {
	StartEpoch  int64
	StartTokens int
}

// ReadBurnFile reads the burn tracking state file.
func ReadBurnFile(cacheDir string) (*BurnState, error) {
	path := filepath.Join(cacheDir, "statusline-burn.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(strings.TrimSpace(string(data)), "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid burn file format")
	}
	epoch, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, err
	}
	tokens, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	return &BurnState{StartEpoch: epoch, StartTokens: tokens}, nil
}

// WriteBurnFile writes the burn tracking state file.
func WriteBurnFile(cacheDir string, epoch int64, tokens int) error {
	path := filepath.Join(cacheDir, "statusline-burn.dat")
	return os.WriteFile(path, []byte(fmt.Sprintf("%d|%d", epoch, tokens)), 0600)
}

// WindowState holds auto-detected window durations.
type WindowState struct {
	ShortSecs int // session window (e.g. 18000 = 5h)
	LongSecs  int // weekly window (e.g. 604800 = 7d)
}

// ReadWindowFile reads cached window durations.
func ReadWindowFile(cacheDir string) (*WindowState, error) {
	path := filepath.Join(cacheDir, "statusline-window-dur.dat")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	parts := strings.SplitN(strings.TrimSpace(string(data)), "|", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid window file format")
	}
	short, _ := strconv.Atoi(parts[0])
	long, _ := strconv.Atoi(parts[1])
	if short <= 0 || long <= 0 {
		return nil, fmt.Errorf("invalid window values")
	}
	return &WindowState{ShortSecs: short, LongSecs: long}, nil
}

// WriteWindowFile writes cached window durations.
func WriteWindowFile(cacheDir string, short, long int) error {
	path := filepath.Join(cacheDir, "statusline-window-dur.dat")
	return os.WriteFile(path, []byte(fmt.Sprintf("%d|%d", short, long)), 0600)
}
