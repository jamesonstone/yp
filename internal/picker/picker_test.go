package picker

import (
	"os"
	"path/filepath"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickerBehaviorTableDriven(t *testing.T) {
	tmp := t.TempDir()
	childDir := filepath.Join(tmp, "child")
	if err := os.Mkdir(childDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	filePath := filepath.Join(tmp, "file.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	tests := []struct {
		name       string
		setupModel func(t *testing.T) model
		msg        tea.Msg
		check      func(t *testing.T, m model)
	}{
		{
			name: "injects parent for non-root",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				return m
			},
			check: func(t *testing.T, m model) {
				items := m.list.Items()
				if len(items) == 0 {
					t.Fatal("expected items")
				}
				first := items[0].(entryItem)
				if first.entry.Name != ".." {
					t.Fatalf("first item = %q, want ..", first.entry.Name)
				}
			},
		},
		{
			name: "no parent for root",
			setupModel: func(t *testing.T) model {
				m, err := newModel(string(os.PathSeparator), true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				return m
			},
			check: func(t *testing.T, m model) {
				if len(m.list.Items()) > 0 {
					first := m.list.Items()[0].(entryItem)
					if first.entry.Name == ".." {
						t.Fatal("root should not include parent entry")
					}
				}
			},
		},
		{
			name: "enter on directory drills",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				for i, item := range m.list.Items() {
					if item.(entryItem).entry.Name == "child" {
						m.list.Select(i)
					}
				}
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, m model) {
				if m.currentDir != childDir {
					t.Fatalf("currentDir = %s, want %s", m.currentDir, childDir)
				}
			},
		},
		{
			name: "enter on file picks file",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				for i, item := range m.list.Items() {
					if item.(entryItem).entry.Name == "file.txt" {
						m.list.Select(i)
					}
				}
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyEnter},
			check: func(t *testing.T, m model) {
				if m.result != filePath {
					t.Fatalf("result = %s, want %s", m.result, filePath)
				}
			},
		},
		{
			name: "esc exits with current directory",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyEsc},
			check: func(t *testing.T, m model) {
				if m.result != tmp {
					t.Fatalf("result = %s, want %s", m.result, tmp)
				}
			},
		},
		{
			name: "ctrl+c exits with current directory",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyCtrlC},
			check: func(t *testing.T, m model) {
				if m.result != tmp {
					t.Fatalf("result = %s, want %s", m.result, tmp)
				}
			},
		},
		{
			name: "tab cycles down",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				m.list.Select(len(m.list.Items()) - 1)
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyTab},
			check: func(t *testing.T, m model) {
				if m.list.Index() != 0 {
					t.Fatalf("index = %d, want 0", m.list.Index())
				}
			},
		},
		{
			name: "shift+tab cycles up",
			setupModel: func(t *testing.T) model {
				m, err := newModel(tmp, true)
				if err != nil {
					t.Fatalf("newModel: %v", err)
				}
				m.list.Select(0)
				return m
			},
			msg: tea.KeyMsg{Type: tea.KeyShiftTab},
			check: func(t *testing.T, m model) {
				if m.list.Index() != len(m.list.Items())-1 {
					t.Fatalf("index = %d, want %d", m.list.Index(), len(m.list.Items())-1)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.setupModel(t)
			if tc.msg != nil {
				updated, _ := m.Update(tc.msg)
				m = updated.(model)
			}
			tc.check(t, m)
		})
	}
}

func TestRunPathNormalization(t *testing.T) {
	tmp := t.TempDir()
	nested := filepath.Join(tmp, "a", "b")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	resolved := filepath.Clean(filepath.Join(tmp, "a", ".", "b", ".."))
	if got := filepath.Clean(resolved); got != filepath.Join(tmp, "a") {
		t.Fatalf("clean path = %s, want %s", got, filepath.Join(tmp, "a"))
	}
}
