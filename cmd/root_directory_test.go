package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestRunCopiesDirectoryWhenPickerIsUnavailable(t *testing.T) {
	tmp := t.TempDir()
	var copied string
	var out bytes.Buffer

	err := run(
		[]string{tmp + string(os.PathSeparator)},
		&out,
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return false },
		func(string) (string, error) {
			t.Fatal("picker should not run when unavailable")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != tmp {
		t.Fatalf("copied = %s, want %s", copied, tmp)
	}
	if out.String() != "📋 "+tmp+"\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunCopiesExistingDirectoryWithoutTrailingSeparator(t *testing.T) {
	root := t.TempDir()
	currentDir := filepath.Join(root, "yp")
	targetDir := filepath.Join(root, "wu")
	mustMkdir(t, currentDir)
	mustMkdir(t, targetDir)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(currentDir); err != nil {
		t.Fatalf("chdir %s: %v", currentDir, err)
	}
	t.Setenv("PWD", currentDir)
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	var copied string
	var out bytes.Buffer

	err = run(
		[]string{filepath.Join("..", "wu")},
		&out,
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(string) (string, error) {
			t.Fatal("picker should not run without a trailing path separator")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != targetDir {
		t.Fatalf("copied = %s, want %s", copied, targetDir)
	}
	if out.String() != "📋 "+targetDir+"\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunUsesPickerForDirectoryWithTrailingSeparator(t *testing.T) {
	root := t.TempDir()
	currentDir := filepath.Join(root, "yp")
	targetDir := filepath.Join(root, "wu")
	mustMkdir(t, currentDir)
	mustMkdir(t, targetDir)
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(currentDir); err != nil {
		t.Fatalf("chdir %s: %v", currentDir, err)
	}
	t.Setenv("PWD", currentDir)
	t.Cleanup(func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	})

	picked := targetDir + string(os.PathSeparator) + "picked.txt"
	var copied string
	var out bytes.Buffer

	err = run(
		[]string{filepath.Join("..", "wu") + string(os.PathSeparator)},
		&out,
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(path string) (string, error) {
			if path != targetDir {
				t.Fatalf("picker path = %s, want %s", path, targetDir)
			}
			return picked, nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != picked {
		t.Fatalf("copied = %s, want %s", copied, picked)
	}
	if out.String() != "📋 "+picked+"\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunUsesPickerForExplicitDotWithTrailingSeparator(t *testing.T) {
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

	picked := filepath.Join(tmp, "picked.txt")
	var copied string
	err = run(
		[]string{"." + string(os.PathSeparator)},
		&bytes.Buffer{},
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(path string) (string, error) {
			if path != tmp {
				t.Fatalf("picker path = %s, want %s", path, tmp)
			}
			return picked, nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != picked {
		t.Fatalf("copied = %s, want %s", copied, picked)
	}
}

func TestRunCopiesCurrentDirectoryForExplicitDot(t *testing.T) {
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

	var copied string
	var out bytes.Buffer
	err = run(
		[]string{"."},
		&out,
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(string) (string, error) {
			t.Fatal("picker should not run for explicit dot")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != tmp {
		t.Fatalf("copied = %s, want %s", copied, tmp)
	}
	if out.String() != "📋 "+tmp+"\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunWithoutArgsStillUsesPicker(t *testing.T) {
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

	picked := filepath.Join(tmp, "picked.txt")
	var copied string
	err = run(
		nil,
		&bytes.Buffer{},
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(path string) (string, error) {
			if path != tmp {
				t.Fatalf("picker path = %s, want %s", path, tmp)
			}
			return picked, nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != picked {
		t.Fatalf("copied = %s, want %s", copied, picked)
	}
}

func TestRunSkipsPickerForNonDirectory(t *testing.T) {
	tmp := t.TempDir()
	filePath := filepath.Join(tmp, "file.txt")
	mustWriteFile(t, filePath)
	var copied string

	err := run(
		[]string{filePath},
		&bytes.Buffer{},
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(string) (string, error) {
			t.Fatal("picker should not run for files")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}
	if copied != filePath {
		t.Fatalf("copied = %s, want %s", copied, filePath)
	}
}
