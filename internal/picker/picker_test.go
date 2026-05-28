package picker

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunReturnsCurrentDirectoryWhenFzfCancels(t *testing.T) {
	tmp := t.TempDir()
	fzf := writeExecutable(t, filepath.Join(tmp, "fzf"), `#!/bin/sh
cat >/dev/null
exit 130
`)
	t.Setenv("PATH", filepath.Dir(fzf)+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := Run(tmp)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if got != tmp {
		t.Fatalf("Run() = %s, want %s", got, tmp)
	}
}

func TestRunReturnsPickedPath(t *testing.T) {
	tmp := t.TempDir()
	fzf := writeExecutable(t, filepath.Join(tmp, "fzf"), `#!/bin/sh
cat >/dev/null
printf '📄	file.txt\n'
`)
	t.Setenv("PATH", filepath.Dir(fzf)+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := Run(tmp)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := tmp + string(os.PathSeparator) + "file.txt"
	if got != want {
		t.Fatalf("Run() = %s, want %s", got, want)
	}
}

func TestRunReturnsSubdirectoriesWhenStarActionPicked(t *testing.T) {
	tmp := t.TempDir()
	mustMkdir(t, filepath.Join(tmp, "alpha"))
	mustMkdir(t, filepath.Join(tmp, "bravo"))
	mustMkdir(t, filepath.Join(tmp, ".hidden"))
	mustWriteFile(t, filepath.Join(tmp, "file.txt"), 0o644)

	fzf := writeExecutable(t, filepath.Join(tmp, "fzf"), `#!/bin/sh
cat >/dev/null
printf '✳️	*\n'
`)
	t.Setenv("PATH", filepath.Dir(fzf)+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := Run(tmp)
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}

	want := strings.Join([]string{
		filepath.Join(tmp, "alpha"),
		filepath.Join(tmp, "bravo"),
	}, "\n")
	if got != want {
		t.Fatalf("Run() = %q, want %q", got, want)
	}
}

func TestHelperListMatchesShellFunctionRows(t *testing.T) {
	tmp := t.TempDir()
	mustMkdir(t, filepath.Join(tmp, "dir"))
	mustWriteFile(t, filepath.Join(tmp, ".hidden"), 0o644)
	mustWriteFile(t, filepath.Join(tmp, "exec.sh"), 0o755)
	mustWriteFile(t, filepath.Join(tmp, "file.txt"), 0o644)
	mustSymlink(t, filepath.Join(tmp, "file.txt"), filepath.Join(tmp, "link"))

	state := filepath.Join(tmp, "state")
	if err := os.WriteFile(state, []byte(tmp), 0o600); err != nil {
		t.Fatalf("write state: %v", err)
	}
	helper, err := writeHelper(state)
	if err != nil {
		t.Fatalf("writeHelper() error = %v", err)
	}
	defer os.Remove(helper)

	output, err := runHelper(helper, "list")
	if err != nil {
		t.Fatalf("helper list error = %v", err)
	}

	for _, want := range []string{
		"📁\t..",
		"📁\tdir",
		"📄\t.hidden",
		"⚙️\texec.sh",
		"📄\tfile.txt",
		"🔗\tlink",
		"✳️\t*",
	} {
		if !strings.Contains(output, want+"\n") {
			t.Fatalf("helper output missing %q:\n%s", want, output)
		}
	}
}

func TestTargetFromPickedPreservesShellConcatenation(t *testing.T) {
	got, err := targetFromPicked(string(os.PathSeparator), "📄\tfile.txt")
	if err != nil {
		t.Fatalf("targetFromPicked() error = %v", err)
	}
	want := string(os.PathSeparator) + string(os.PathSeparator) + "file.txt"
	if got != want {
		t.Fatalf("targetFromPicked() = %s, want %s", got, want)
	}
}

func TestTargetFromPickedStarReturnsSubdirectoryList(t *testing.T) {
	tmp := t.TempDir()
	mustMkdir(t, filepath.Join(tmp, "alpha"))
	mustMkdir(t, filepath.Join(tmp, "bravo"))
	mustMkdir(t, filepath.Join(tmp, ".hidden"))
	mustWriteFile(t, filepath.Join(tmp, "file.txt"), 0o644)

	got, err := targetFromPicked(tmp, "✳️\t*")
	if err != nil {
		t.Fatalf("targetFromPicked() error = %v", err)
	}

	want := strings.Join([]string{
		filepath.Join(tmp, "alpha"),
		filepath.Join(tmp, "bravo"),
	}, "\n")
	if got != want {
		t.Fatalf("targetFromPicked() = %q, want %q", got, want)
	}
}

func TestPickedNameKeepsTabsAfterFirstDelimiter(t *testing.T) {
	got := pickedName("📄\tname\twith-tab")
	if got != "name\twith-tab" {
		t.Fatalf("pickedName() = %q", got)
	}
}

func runHelper(path string, args ...string) (string, error) {
	data, err := exec.Command(path, args...).Output()
	return string(data), err
}

func writeExecutable(t *testing.T, path, script string) string {
	t.Helper()
	if err := os.WriteFile(path, []byte(script), 0o755); err != nil {
		t.Fatalf("write script %s: %v", path, err)
	}
	return path
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path string, mode os.FileMode) {
	t.Helper()
	if err := os.WriteFile(path, []byte("x"), mode); err != nil {
		t.Fatalf("write file %s: %v", path, err)
	}
}

func mustSymlink(t *testing.T, target, path string) {
	t.Helper()
	if err := os.Symlink(target, path); err != nil {
		t.Fatalf("symlink %s -> %s: %v", path, target, err)
	}
}
