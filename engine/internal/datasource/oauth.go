package datasource

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/jamesprnich/oh-my-line/engine/internal/debug"
	"github.com/jamesprnich/oh-my-line/engine/internal/platform"
)

type credentialFile struct {
	ClaudeAiOauth struct {
		AccessToken string `json:"accessToken"`
	} `json:"claudeAiOauth"`
}

// GetOAuthToken resolves the OAuth token from the environment, keychain,
// credentials file, or GNOME keyring.
func GetOAuthToken() string {
	// 1. Environment variable
	if t := os.Getenv("CLAUDE_CODE_OAUTH_TOKEN"); t != "" {
		debug.Log("oauth", "source=env")
		return t
	}

	// 2. macOS Keychain
	if platform.ExecAvailable() {
		if out, err := platform.RunCommand(`security find-generic-password -s "Claude Code-credentials" -w 2>/dev/null`, 3); err == nil {
			out = strings.TrimSpace(out)
			if out != "" {
				if t := extractAccessToken(out); t != "" {
					debug.Log("oauth", "source=keychain")
					return t
				}
			}
		}
	}

	// 3. Credentials file (must be owner-only permissions)
	home, _ := os.UserHomeDir()
	if home != "" {
		credPath := filepath.Join(home, ".claude", ".credentials.json")
		if fi, err := os.Stat(credPath); err == nil {
			if fi.Mode().Perm()&0077 != 0 {
				debug.Log("oauth", "credentials file has insecure permissions %o, skipping", fi.Mode().Perm())
			} else if data, err := os.ReadFile(credPath); err == nil {
				if t := extractAccessToken(string(data)); t != "" {
					debug.Log("oauth", "source=credentials-file")
					return t
				}
			}
		}
	}

	// 4. GNOME Keyring (secret-tool)
	if platform.ExecAvailable() {
		if out, err := platform.RunCommand(`secret-tool lookup service "Claude Code-credentials" 2>/dev/null`, 2); err == nil {
			out = strings.TrimSpace(out)
			if out != "" {
				if t := extractAccessToken(out); t != "" {
					debug.Log("oauth", "source=gnome-keyring")
					return t
				}
			}
		}
	}

	debug.Log("oauth", "no token found")
	return ""
}

func extractAccessToken(data string) string {
	var cred credentialFile
	if err := json.Unmarshal([]byte(data), &cred); err != nil {
		return ""
	}
	t := cred.ClaudeAiOauth.AccessToken
	if t == "" || t == "null" {
		return ""
	}
	return t
}
