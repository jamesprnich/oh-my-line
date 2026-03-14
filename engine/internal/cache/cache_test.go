package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestReadFile_NonExistent(t *testing.T) {
	content, fresh := ReadFile("/tmp/nonexistent-test-file-12345", 60)
	if content != "" || fresh {
		t.Errorf("non-existent file should return empty/false, got %q/%v", content, fresh)
	}
}

func TestWriteAndReadFile(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test-cache.dat")

	err := WriteFile(tmp, "hello world")
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	content, fresh := ReadFile(tmp, 60)
	if content != "hello world" {
		t.Errorf("content = %q, want 'hello world'", content)
	}
	if !fresh {
		t.Error("just-written file should be fresh")
	}
}

func TestReadFile_Stale(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test-stale.dat")
	os.WriteFile(tmp, []byte("old data"), 0600)

	// Set mtime to 2 minutes ago
	past := time.Now().Add(-2 * time.Minute)
	os.Chtimes(tmp, past, past)

	content, fresh := ReadFile(tmp, 60) // 60s TTL
	if content != "old data" {
		t.Errorf("content = %q, want 'old data'", content)
	}
	if fresh {
		t.Error("2-min-old file should not be fresh with 60s TTL")
	}
}

func TestWriteFile_Atomic(t *testing.T) {
	tmp := filepath.Join(t.TempDir(), "test-atomic.dat")

	// Write initial
	WriteFile(tmp, "first")

	// Overwrite
	err := WriteFile(tmp, "second")
	if err != nil {
		t.Fatalf("WriteFile overwrite failed: %v", err)
	}

	content, _ := ReadFile(tmp, 60)
	if content != "second" {
		t.Errorf("content = %q, want 'second'", content)
	}

	// No .tmp file should remain
	if _, err := os.Stat(tmp + ".tmp"); err == nil {
		t.Error("temp file should not remain after atomic write")
	}
}

func TestBurnFile(t *testing.T) {
	dir := t.TempDir()

	err := WriteBurnFile(dir, 1700000000, 50000)
	if err != nil {
		t.Fatalf("WriteBurnFile failed: %v", err)
	}

	state, err := ReadBurnFile(dir)
	if err != nil {
		t.Fatalf("ReadBurnFile failed: %v", err)
	}

	if state.StartEpoch != 1700000000 {
		t.Errorf("epoch = %d, want 1700000000", state.StartEpoch)
	}
	if state.StartTokens != 50000 {
		t.Errorf("tokens = %d, want 50000", state.StartTokens)
	}
}

func TestReadBurnFile_NonExistent(t *testing.T) {
	_, err := ReadBurnFile("/tmp/nonexistent-test-dir-12345")
	if err == nil {
		t.Error("non-existent burn file should return error")
	}
}

func TestReadBurnFile_InvalidFormat(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "statusline-burn.dat")
	os.WriteFile(path, []byte("invalid"), 0600)

	_, err := ReadBurnFile(dir)
	if err == nil {
		t.Error("invalid burn file should return error")
	}
}

func TestWindowFile(t *testing.T) {
	dir := t.TempDir()

	err := WriteWindowFile(dir, 18000, 604800)
	if err != nil {
		t.Fatalf("WriteWindowFile failed: %v", err)
	}

	state, err := ReadWindowFile(dir)
	if err != nil {
		t.Fatalf("ReadWindowFile failed: %v", err)
	}

	if state.ShortSecs != 18000 {
		t.Errorf("short = %d, want 18000", state.ShortSecs)
	}
	if state.LongSecs != 604800 {
		t.Errorf("long = %d, want 604800", state.LongSecs)
	}
}

func TestReadWindowFile_NonExistent(t *testing.T) {
	_, err := ReadWindowFile("/tmp/nonexistent-test-dir-12345")
	if err == nil {
		t.Error("non-existent window file should return error")
	}
}

// ── AccountKey ──

func TestAccountKey_EmptyIsDefault(t *testing.T) {
	if got := AccountKey(""); got != "default" {
		t.Errorf("AccountKey(\"\") = %q, want \"default\"", got)
	}
}

func TestAccountKey_HomeDotClaudeIsDefault(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	if got := AccountKey(filepath.Join(home, ".claude")); got != "default" {
		t.Errorf("AccountKey(~/.claude) = %q, want \"default\"", got)
	}
}

func TestAccountKey_HomeDotClaudeTrailingSlashIsDefault(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home dir")
	}
	if got := AccountKey(filepath.Join(home, ".claude") + "/"); got != "default" {
		t.Errorf("AccountKey(~/.claude/) = %q, want \"default\"", got)
	}
}

func TestAccountKey_CustomPathProducesHash(t *testing.T) {
	got := AccountKey("/home/user/.claude-work")
	if got == "default" {
		t.Error("custom path should not be \"default\"")
	}
	if len(got) != 8 {
		t.Errorf("hash should be 8 hex chars, got %q (len %d)", got, len(got))
	}
}

func TestAccountKey_Deterministic(t *testing.T) {
	a := AccountKey("/custom/path")
	b := AccountKey("/custom/path")
	if a != b {
		t.Errorf("same input should produce same key: %q vs %q", a, b)
	}
}

func TestAccountKey_DifferentPathsDifferentKeys(t *testing.T) {
	a := AccountKey("/path/a")
	b := AccountKey("/path/b")
	if a == b {
		t.Errorf("different paths should produce different keys: both %q", a)
	}
}

// ── AccountDir ──

func TestAccountDir_DefaultReturnsBase(t *testing.T) {
	base, err := Dir()
	if err != nil {
		t.Fatalf("Dir() failed: %v", err)
	}
	got, err := AccountDir("default")
	if err != nil {
		t.Fatalf("AccountDir(\"default\") failed: %v", err)
	}
	if got != base {
		t.Errorf("AccountDir(\"default\") = %q, want %q", got, base)
	}
}

func TestAccountDir_EmptyReturnsBase(t *testing.T) {
	base, err := Dir()
	if err != nil {
		t.Fatalf("Dir() failed: %v", err)
	}
	got, err := AccountDir("")
	if err != nil {
		t.Fatalf("AccountDir(\"\") failed: %v", err)
	}
	if got != base {
		t.Errorf("AccountDir(\"\") = %q, want %q", got, base)
	}
}

func TestAccountDir_CustomCreatesSubdir(t *testing.T) {
	got, err := AccountDir("a1b2c3d4")
	if err != nil {
		t.Fatalf("AccountDir(\"a1b2c3d4\") failed: %v", err)
	}
	base, _ := Dir()
	want := filepath.Join(base, "acct-a1b2c3d4")
	if got != want {
		t.Errorf("AccountDir(\"a1b2c3d4\") = %q, want %q", got, want)
	}
	fi, err := os.Stat(got)
	if err != nil {
		t.Fatalf("account dir should exist: %v", err)
	}
	if !fi.IsDir() {
		t.Error("account dir should be a directory")
	}
	if fi.Mode().Perm() != 0700 {
		t.Errorf("account dir perms = %o, want 0700", fi.Mode().Perm())
	}
}
