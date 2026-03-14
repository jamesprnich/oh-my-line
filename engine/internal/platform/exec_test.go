//go:build !js

package platform

import (
	"strings"
	"testing"
	"time"
)

func TestRunCommand_BasicOutput(t *testing.T) {
	out, err := RunCommand("echo hello", 3)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if strings.TrimSpace(out) != "hello" {
		t.Errorf("output = %q, want hello", out)
	}
}

func TestRunCommand_MultilineOutput(t *testing.T) {
	out, err := RunCommand("printf 'line1\nline2\nline3'", 3)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d: %q", len(lines), out)
	}
}

func TestRunCommand_StderrNotCaptured(t *testing.T) {
	// SECURITY: stderr must NOT appear in output — it could contain
	// error messages with sensitive info (paths, env vars, etc.)
	out, _ := RunCommand("echo visible; echo secret >&2", 3)
	if strings.Contains(out, "secret") {
		t.Fatal("SECURITY: stderr should not appear in stdout capture")
	}
	if !strings.Contains(out, "visible") {
		t.Errorf("stdout should be captured, got %q", out)
	}
}

func TestRunCommand_FailedCommandReturnsError(t *testing.T) {
	_, err := RunCommand("false", 3)
	if err == nil {
		t.Error("failed command should return error")
	}
}

func TestRunCommand_NonexistentCommand(t *testing.T) {
	_, err := RunCommand("command_that_does_not_exist_xyz123", 3)
	if err == nil {
		t.Error("nonexistent command should return error")
	}
}

func TestRunCommand_TimeoutKillsProcess(t *testing.T) {
	// SECURITY: A command that hangs must be killed by the timeout.
	// Without this, a malicious command segment could block the statusline forever.
	start := time.Now()
	_, err := RunCommand("sleep 30", 1)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("timed-out command should return error")
	}
	// Should complete in ~1s, not 30s. Allow 3s for CI slowness.
	if elapsed > 3*time.Second {
		t.Errorf("timeout not enforced: took %v (expected ~1s)", elapsed)
	}
}

func TestRunCommand_DefaultTimeout(t *testing.T) {
	// timeoutSecs <= 0 should default to 3, not hang forever.
	start := time.Now()
	_, err := RunCommand("sleep 30", 0)
	elapsed := time.Since(start)

	if err == nil {
		t.Error("should timeout and return error")
	}
	if elapsed > 5*time.Second {
		t.Errorf("default timeout not applied: took %v", elapsed)
	}
}

func TestRunCommand_EmptyOutput(t *testing.T) {
	out, err := RunCommand("true", 3)
	if err != nil {
		t.Fatalf("true should succeed: %v", err)
	}
	if out != "" {
		t.Errorf("true should produce empty output, got %q", out)
	}
}

func TestRunCommand_SpecialCharacters(t *testing.T) {
	// Verify output with special chars isn't mangled.
	out, err := RunCommand(`printf '{"key": "value"}'`, 3)
	if err != nil {
		t.Fatalf("RunCommand failed: %v", err)
	}
	if strings.TrimSpace(out) != `{"key": "value"}` {
		t.Errorf("special chars mangled: %q", out)
	}
}
