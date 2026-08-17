package cmd

import (
	"bytes"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsMissingFileAndSuggestsClosestSibling(t *testing.T) {
	tmp := t.TempDir()
	existing := filepath.Join(tmp, "companion_observability_contract.md")
	missing := filepath.Join(tmp, "companion_observability_contract_v3.md")
	mustWriteFile(t, existing)

	var out bytes.Buffer
	err := run(
		[]string{missing},
		&out,
		func(string) error {
			t.Fatal("clipboard should not be written")
			return nil
		},
		func() bool { return true },
		func(string) (string, error) {
			t.Fatal("picker should not run")
			return "", nil
		},
	)

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("run() error = %v, want not-exist error", err)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("run() error = %v, want missing path", err)
	}
	if !strings.Contains(err.Error(), "did you mean") || !strings.Contains(err.Error(), existing) {
		t.Fatalf("run() error = %v, want suggestion %s", err, existing)
	}
	if out.Len() != 0 {
		t.Fatalf("output = %q, want empty", out.String())
	}
}

func TestRunRejectsMissingFileWithoutWeakSuggestion(t *testing.T) {
	tmp := t.TempDir()
	missing := filepath.Join(tmp, "missing_v3.md")
	mustWriteFile(t, filepath.Join(tmp, "unrelated.md"))
	mustMkdir(t, filepath.Join(tmp, "missing_v2.md"))

	err := run(
		[]string{missing},
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

	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("run() error = %v, want not-exist error", err)
	}
	if !strings.Contains(err.Error(), missing) {
		t.Fatalf("run() error = %v, want missing path", err)
	}
	if strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("run() error = %v, want no weak suggestion", err)
	}
}

func TestDirectoryArgsRejectsMissingInput(t *testing.T) {
	tmp := t.TempDir()
	existing := filepath.Join(tmp, "existing")
	missing := filepath.Join(tmp, "missing")
	mustMkdir(t, existing)

	_, err := directoryArgs([]string{existing, missing})
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("directoryArgs() error = %v, want not-exist error", err)
	}
}

func TestValidatePathDoesNotSuggestFileForMissingDirectory(t *testing.T) {
	tmp := t.TempDir()
	mustWriteFile(t, filepath.Join(tmp, "missing"))
	missingDirectory := filepath.Join(tmp, "mssing") + string(filepath.Separator)

	err := validatePath(missingDirectory)
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("validatePath() error = %v, want not-exist error", err)
	}
	if strings.Contains(err.Error(), "did you mean") {
		t.Fatalf("validatePath() error = %v, want no file suggestion", err)
	}
}

func TestEditDistanceHonorsMaximum(t *testing.T) {
	tests := []struct {
		name     string
		left     string
		right    string
		maximum  int
		expected int
	}{
		{name: "within limit", left: "contract_v3.md", right: "contract.md", maximum: 3, expected: 3},
		{name: "outside limit", left: "kitten", right: "sitting", maximum: 2, expected: 3},
		{name: "unicode", left: "café.md", right: "cafe.md", maximum: 1, expected: 1},
		{name: "empty", left: "", right: "abc", maximum: 3, expected: 3},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := editDistance([]rune(test.left), []rune(test.right), test.maximum)
			if got != test.expected {
				t.Fatalf("editDistance() = %d, want %d", got, test.expected)
			}
		})
	}
}
