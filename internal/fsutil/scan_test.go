package fsutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanTableDriven(t *testing.T) {
	tmp := t.TempDir()
	mustMkdir(t, filepath.Join(tmp, "dir"))
	mustWriteFile(t, filepath.Join(tmp, "file.txt"), 0o644)
	mustWriteFile(t, filepath.Join(tmp, ".hidden"), 0o644)
	mustWriteFile(t, filepath.Join(tmp, "exec.sh"), 0o755)
	mustSymlink(t, filepath.Join(tmp, "file.txt"), filepath.Join(tmp, "link"))

	tests := []struct {
		name          string
		includeHidden bool
		wantHidden    bool
		wantKinds     map[string]EntryKind
	}{
		{
			name:          "includes hidden files",
			includeHidden: true,
			wantHidden:    true,
			wantKinds: map[string]EntryKind{
				"dir":      KindDir,
				"file.txt": KindFile,
				"exec.sh":  KindExec,
				"link":     KindSymlink,
			},
		},
		{
			name:          "excludes hidden files",
			includeHidden: false,
			wantHidden:    false,
			wantKinds: map[string]EntryKind{
				"dir":      KindDir,
				"file.txt": KindFile,
				"exec.sh":  KindExec,
				"link":     KindSymlink,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			items, err := Scan(tmp, tc.includeHidden)
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			got := map[string]EntryKind{}
			hiddenFound := false
			for _, item := range items {
				got[item.Name] = item.Kind
				if item.Name == ".hidden" {
					hiddenFound = true
				}
			}

			if hiddenFound != tc.wantHidden {
				t.Fatalf("hidden presence = %v, want %v", hiddenFound, tc.wantHidden)
			}

			for name, kind := range tc.wantKinds {
				if got[name] != kind {
					t.Fatalf("kind for %q = %v, want %v", name, got[name], kind)
				}
			}
		})
	}
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
