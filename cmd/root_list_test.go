package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCopiesSubdirectoriesForStarPath(t *testing.T) {
	tmp := t.TempDir()
	parent := filepath.Join(tmp, "parent")
	mustMkdir(t, parent)
	dirA := filepath.Join(parent, "alpha")
	dirB := filepath.Join(parent, "bravo")
	hiddenDir := filepath.Join(parent, ".hidden")
	mustMkdir(t, dirA)
	mustMkdir(t, dirB)
	mustMkdir(t, hiddenDir)
	mustWriteFile(t, filepath.Join(parent, "file.txt"))

	var copied string
	var out bytes.Buffer
	err := run(
		[]string{filepath.Join(parent, "*")},
		&out,
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(string) (string, error) {
			t.Fatal("picker should not run for star paths")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	want := strings.Join([]string{dirA, dirB}, "\n")
	if copied != want {
		t.Fatalf("copied = %q, want %q", copied, want)
	}
	if out.String() != "📋 "+want+"\n" {
		t.Fatalf("output = %q", out.String())
	}
}

func TestRunCopiesExpandedDirectoryArgs(t *testing.T) {
	tmp := t.TempDir()
	dirA := filepath.Join(tmp, "alpha")
	dirB := filepath.Join(tmp, "bravo")
	filePath := filepath.Join(tmp, "file.txt")
	mustMkdir(t, dirA)
	mustMkdir(t, dirB)
	mustWriteFile(t, filePath)

	var copied string
	err := run(
		[]string{dirB, filePath, dirA},
		&bytes.Buffer{},
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return true },
		func(string) (string, error) {
			t.Fatal("picker should not run for expanded directory args")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	want := strings.Join([]string{dirB, dirA}, "\n")
	if copied != want {
		t.Fatalf("copied = %q, want %q", copied, want)
	}
}

func TestRunFallsBackToFirstArgWhenMultipleArgsHaveNoDirectories(t *testing.T) {
	tmp := t.TempDir()
	first := filepath.Join(tmp, "first.txt")
	second := filepath.Join(tmp, "second.txt")
	mustWriteFile(t, first)
	mustWriteFile(t, second)

	var copied string
	err := run(
		[]string{first, second},
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
	if copied != first {
		t.Fatalf("copied = %s, want %s", copied, first)
	}
}

func TestRunErrorsWhenStarPathHasNoSubdirectories(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "file.txt"))

	err := run(
		[]string{filepath.Join(tmp, "*")},
		&bytes.Buffer{},
		func(string) error {
			t.Fatal("clipboard should not be written")
			return nil
		},
		func() bool { return false },
		func(string) (string, error) {
			t.Fatal("picker should not run")
			return "", nil
		},
	)
	if err == nil {
		t.Fatal("run() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "no subdirectories found") {
		t.Fatalf("run() error = %v", err)
	}
}
