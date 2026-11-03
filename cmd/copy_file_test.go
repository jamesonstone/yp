package cmd

import (
	"bytes"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCopyFlagCopiesExactFileBytes(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "notes.txt")
	payload := []byte("alpha\nbeta\n\xff")
	if err := os.WriteFile(filePath, payload, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	for _, args := range [][]string{
		{"-c", filePath},
		{"--copy", filePath},
		{filePath, "-c"},
	} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var copied string
			var out bytes.Buffer
			err := run(args, &out, func(value string) error {
				copied = value
				return nil
			}, pickerUnavailable, pickerMustNotRun(t))
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}
			if copied != string(payload) {
				t.Fatalf("copied = %q, want exact file bytes", copied)
			}
			if strings.Contains(out.String(), string(payload)) {
				t.Fatalf("output = %q, want size confirmation without contents", out.String())
			}
			want := "📋 " + filePath + " (12 bytes, 3 lines)\n"
			if out.String() != want {
				t.Fatalf("output = %q, want %q", out.String(), want)
			}
		})
	}
}

func TestRunCopyFlagReportsEmptyFile(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "empty.txt")
	if err := os.WriteFile(filePath, nil, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var copied string
	var out bytes.Buffer
	err := run([]string{"-c", filePath}, &out, func(value string) error {
		copied = value
		return nil
	}, pickerUnavailable, pickerMustNotRun(t))
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != "" {
		t.Fatalf("copied = %q, want empty", copied)
	}
	if out.String() != "📋 "+filePath+" (0 bytes, 0 lines)\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunCopyFlagCopiesPickedFile(t *testing.T) {
	tmp := t.TempDir()
	dir := filepath.Join(tmp, "src")
	mustMkdir(t, dir)
	filePath := filepath.Join(dir, "picked.txt")
	payload := []byte("one line")
	if err := os.WriteFile(filePath, payload, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	var copied string
	var out bytes.Buffer
	err := run(
		[]string{"-c", dir + string(os.PathSeparator)},
		&out,
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(path string) (string, error) {
			if path != dir {
				t.Fatalf("picker path = %s, want %s", path, dir)
			}
			return filePath, nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != string(payload) {
		t.Fatalf("copied = %q, want file contents", copied)
	}
	if out.String() != "📋 "+filePath+" (8 bytes, 1 line)\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunCopyFlagRejectsDirectory(t *testing.T) {
	tmp := t.TempDir()
	assertClipboardUntouched(t, []string{"-c", tmp}, tmp+" is a directory")
	assertClipboardUntouched(t, []string{"--copy", tmp + string(os.PathSeparator)}, tmp+" is a directory")
}

func TestRunCopyFlagRejectsMissingPathAndSuggestsSibling(t *testing.T) {
	tmp := t.TempDir()
	existing := filepath.Join(tmp, "companion_observability_contract.md")
	missing := filepath.Join(tmp, "companion_observability_contract_v3.md")
	mustWriteFile(t, existing)

	var out bytes.Buffer
	err := run([]string{"-c", missing}, &out, clipboardMustStay(t), pickerUnavailable, pickerMustNotRun(t))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("run() error = %v, want not-exist error", err)
	}
	if !strings.Contains(err.Error(), missing) || !strings.Contains(err.Error(), existing) || !strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("run() error = %v, want suggestion %s", err, existing)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func TestRunCopyFlagRejectsMultiplePaths(t *testing.T) {
	tmp := t.TempDir()
	first := filepath.Join(tmp, "first.txt")
	second := filepath.Join(tmp, "second.txt")
	mustWriteFile(t, first)
	mustWriteFile(t, second)
	assertClipboardUntouched(t, []string{"-c", first, second}, "-c/--copy expects exactly one file")
	assertClipboardUntouched(t, []string{"-c", filepath.Join(tmp, "*")}, "-c/--copy expects exactly one file")
}

func TestRunCopyFlagRejectsPickedDirectoryAndDirectoryList(t *testing.T) {
	tmp := t.TempDir()
	assertPickedCopyRejected(t, tmp, tmp, tmp+" is a directory")
	list := filepath.Join(tmp, "a") + "\n" + filepath.Join(tmp, "b")
	assertPickedCopyRejected(t, tmp, list, "-c/--copy expects exactly one file")
}

func TestRunCopyFlagRejectsBinaryAndNonRegularFiles(t *testing.T) {
	tmp := t.TempDir()
	binaryPath := filepath.Join(tmp, "blob.bin")
	if err := os.WriteFile(binaryPath, []byte{'a', 0, 'b'}, 0o644); err != nil {
		t.Fatalf("write binary: %v", err)
	}
	assertClipboardUntouched(t, []string{"--copy", binaryPath}, binaryPath+" is a binary file")

	info, err := os.Stat("/dev/null")
	if err != nil || info.Mode().IsRegular() {
		t.Skip("/dev/null is not a non-regular file here")
	}
	assertClipboardUntouched(t, []string{"-c", "/dev/null"}, "/dev/null is not a regular file")
}

func TestRunCopyFlagCopiesSymlinkTarget(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "target.txt")
	linkPath := filepath.Join(tmp, "link.txt")
	payload := []byte("through link\n")
	if err := os.WriteFile(filePath, payload, 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Symlink(filePath, linkPath); err != nil {
		t.Fatalf("symlink: %v", err)
	}

	var copied string
	err := run([]string{"-c", linkPath}, &bytes.Buffer{}, func(value string) error {
		copied = value
		return nil
	}, pickerUnavailable, pickerMustNotRun(t))
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != string(payload) {
		t.Fatalf("copied = %q, want symlink target bytes", copied)
	}
}

func TestRunHelpDoesNotCopy(t *testing.T) {
	for _, args := range [][]string{{"--help"}, {"-h"}, {"-c", "--help"}} {
		t.Run(strings.Join(args, " "), func(t *testing.T) {
			var out bytes.Buffer
			err := run(args, &out, clipboardMustStay(t), pickerUnavailable, pickerMustNotRun(t))
			if err != nil {
				t.Fatalf("run() error = %v", err)
			}
			if !strings.Contains(out.String(), "--copy") || !strings.Contains(out.String(), "-c") {
				t.Fatalf("help = %q, want copy flags", out.String())
			}
		})
	}
}

func TestRunUnknownFlagDoesNotCopy(t *testing.T) {
	assertClipboardUntouched(t, []string{"--contents"}, "unknown flag: --contents")
}

func TestRunCopiesLeadingDashFilenameAsPath(t *testing.T) {
	tmp := t.TempDir()
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir %s: %v", tmp, err)
	}
	t.Setenv("PWD", tmp)
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})
	mustWriteFile(t, filepath.Join(tmp, "--help"))

	var copied string
	err = run([]string{"./--help"}, &bytes.Buffer{}, func(value string) error {
		copied = value
		return nil
	}, pickerUnavailable, pickerMustNotRun(t))
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	want := filepath.Join(tmp, "--help")
	if copied != want {
		t.Fatalf("copied = %s, want %s", copied, want)
	}
}

func assertClipboardUntouched(t *testing.T, args []string, want string) {
	t.Helper()
	var out bytes.Buffer
	err := run(args, &out, clipboardMustStay(t), func() bool { return false }, pickerMustNotRun(t))
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("run(%q) error = %v, want %q", args, err, want)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func assertPickedCopyRejected(t *testing.T, dir, picked, want string) {
	t.Helper()
	var out bytes.Buffer
	err := run(
		[]string{"-c", dir + string(os.PathSeparator)},
		&out,
		clipboardMustStay(t),
		func() bool { return true },
		func(path string) (string, error) {
			if path != dir {
				t.Fatalf("picker path = %s, want %s", path, dir)
			}
			return picked, nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("run() error = %v, want %q", err, want)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func clipboardMustStay(t *testing.T) func(string) error {
	t.Helper()
	return func(string) error {
		t.Fatal("clipboard should not be written")
		return nil
	}
}

func pickerUnavailable() bool { return false }

func pickerMustNotRun(t *testing.T) func(string) (string, error) {
	t.Helper()
	return func(string) (string, error) {
		t.Fatal("picker should not run")
		return "", nil
	}
}
