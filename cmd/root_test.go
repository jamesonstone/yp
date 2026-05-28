package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolvePathMatchesShellFunction(t *testing.T) {
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

	mustMkdir(t, filepath.Join(tmp, "dir"))
	mustWriteFile(t, filepath.Join(tmp, "file.txt"))

	tests := []struct {
		name string
		path string
		want string
	}{
		{
			name: "empty path defaults to working directory",
			path: "",
			want: tmp,
		},
		{
			name: "relative directory resolves through pwd",
			path: "dir",
			want: filepath.Join(tmp, "dir"),
		},
		{
			name: "relative file resolves through parent pwd",
			path: "file.txt",
			want: filepath.Join(tmp, "file.txt"),
		},
		{
			name: "relative missing file under existing parent",
			path: "missing.txt",
			want: filepath.Join(tmp, "missing.txt"),
		},
		{
			name: "relative missing nested path with missing parent falls back to root basename",
			path: filepath.Join("missing", "leaf.txt"),
			want: string(os.PathSeparator) + "leaf.txt",
		},
		{
			name: "dot segments are resolved by parent directory lookup",
			path: filepath.Join(tmp, "dir", "..", "file.txt"),
			want: filepath.Join(tmp, "file.txt"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolvePath(tc.path)
			if err != nil {
				t.Fatalf("resolvePath() error = %v", err)
			}
			if got != tc.want {
				t.Fatalf("resolvePath() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestRunCopiesDirectoryWhenPickerIsUnavailable(t *testing.T) {
	tmp := t.TempDir()
	var copied string
	var out bytes.Buffer

	err := run(
		[]string{tmp},
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

func TestRunUsesPickerForExistingDirectory(t *testing.T) {
	tmp := t.TempDir()
	picked := tmp + string(os.PathSeparator) + "picked.txt"
	var copied string
	var out bytes.Buffer

	err := run(
		[]string{tmp},
		&out,
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
	if out.String() != "📋 "+picked+"\n" {
		t.Fatalf("output = %q", out.String())
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

func TestRunMatchesShellFunctionForLeadingDashInput(t *testing.T) {
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
	err = run(
		[]string{"--help"},
		&bytes.Buffer{},
		func(value string) error {
			copied = value
			return nil
		},
		func() bool { return false },
		func(string) (string, error) {
			t.Fatal("picker should not run")
			return "", nil
		},
	)
	if err != nil {
		t.Fatalf("run() error = %v", err)
	}

	want := tmp + string(os.PathSeparator)
	if copied != want {
		t.Fatalf("copied = %s, want %s", copied, want)
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
