package clip

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteUsesClipboardCommandsInOrder(t *testing.T) {
	tmp := t.TempDir()
	capture := filepath.Join(tmp, "capture")
	args := filepath.Join(tmp, "args")
	t.Setenv("PATH", tmp)
	t.Setenv("YP_CAPTURE", capture)
	t.Setenv("YP_ARGS", args)

	writeClipboardScript(t, filepath.Join(tmp, "pbcopy"))
	writeClipboardScript(t, filepath.Join(tmp, "xclip"))
	writeClipboardScript(t, filepath.Join(tmp, "wl-copy"))

	if err := Write("hello"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := mustReadFile(t, capture)
	if got != "hello" {
		t.Fatalf("capture = %q, want hello", got)
	}
	gotArgs := mustReadFile(t, args)
	if gotArgs != "" {
		t.Fatalf("args = %q, want empty for pbcopy", gotArgs)
	}
}

func TestWriteFallsBackToXclip(t *testing.T) {
	tmp := t.TempDir()
	capture := filepath.Join(tmp, "capture")
	args := filepath.Join(tmp, "args")
	t.Setenv("PATH", tmp)
	t.Setenv("YP_CAPTURE", capture)
	t.Setenv("YP_ARGS", args)

	writeClipboardScript(t, filepath.Join(tmp, "xclip"))

	if err := Write("hello"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := mustReadFile(t, capture)
	if got != "hello" {
		t.Fatalf("capture = %q, want hello", got)
	}
	gotArgs := mustReadFile(t, args)
	if gotArgs != "-selection\nclipboard\n" {
		t.Fatalf("args = %q, want xclip clipboard args", gotArgs)
	}
}

func TestWriteFallsBackToWlCopy(t *testing.T) {
	tmp := t.TempDir()
	capture := filepath.Join(tmp, "capture")
	args := filepath.Join(tmp, "args")
	t.Setenv("PATH", tmp)
	t.Setenv("YP_CAPTURE", capture)
	t.Setenv("YP_ARGS", args)

	writeClipboardScript(t, filepath.Join(tmp, "wl-copy"))

	if err := Write("hello"); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	got := mustReadFile(t, capture)
	if got != "hello" {
		t.Fatalf("capture = %q, want hello", got)
	}
	gotArgs := mustReadFile(t, args)
	if gotArgs != "" {
		t.Fatalf("args = %q, want empty for wl-copy", gotArgs)
	}
}

func TestWriteReturnsNoClipboardToolError(t *testing.T) {
	t.Setenv("PATH", t.TempDir())

	if err := Write("hello"); !errors.Is(err, errNoClipboardTool) {
		t.Fatalf("Write() error = %v, want %v", err, errNoClipboardTool)
	}
}

func writeClipboardScript(t *testing.T, path string) {
	t.Helper()
	script := `#!/bin/sh
: > "$YP_CAPTURE"
: > "$YP_ARGS"
if [ "$#" -gt 0 ]; then
  printf '%s\n' "$@" > "$YP_ARGS"
fi
while IFS= read -r line || [ -n "$line" ]; do
  printf '%s' "$line" >> "$YP_CAPTURE"
done
`
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write script %s: %v", path, err)
	}
}

func mustReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
