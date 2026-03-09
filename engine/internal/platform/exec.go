//go:build !js

package platform

import (
	"context"
	"os/exec"
	"time"
)

// RunCommand executes a shell command with timeout and returns stdout.
func RunCommand(cmd string, timeoutSecs int) (string, error) {
	if timeoutSecs <= 0 {
		timeoutSecs = 3
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSecs)*time.Second)
	defer cancel()

	out, err := exec.CommandContext(ctx, "bash", "-c", cmd).Output()
	return string(out), err
}

// ExecAvailable returns true — real exec is available on non-WASM.
func ExecAvailable() bool {
	return true
}
