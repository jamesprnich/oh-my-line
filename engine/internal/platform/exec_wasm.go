//go:build js

package platform

import "fmt"

// RunCommand is a no-op on WASM — no exec available.
func RunCommand(cmd string, timeoutSecs int) (string, error) {
	return "", fmt.Errorf("exec not available in WASM")
}

// ExecAvailable returns false — no exec in WASM.
func ExecAvailable() bool {
	return false
}
