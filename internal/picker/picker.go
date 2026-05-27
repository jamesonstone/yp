package picker

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/jamesonstone/yp/internal/fsutil"
)

type model struct {
	list          list.Model
	currentDir    string
	includeHidden bool
	result        string
	err           error
}

func Run(startDir string, includeHidden bool) (string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}
	absStart = filepath.Clean(absStart)

	m, err := newModel(absStart, includeHidden)
	if err != nil {
		return "", err
	}

	final, err := tea.NewProgram(m).Run()
	if err != nil {
		return "", err
	}
	finished := final.(model)
	if finished.err != nil {
		return "", finished.err
	}
	if finished.result == "" {
		return finished.currentDir, nil
	}
	return filepath.Clean(finished.result), nil
}

func newModel(startDir string, includeHidden bool) (model, error) {
	m := model{
		currentDir:    startDir,
		includeHidden: includeHidden,
	}

	delegate := list.NewDefaultDelegate()
	l := list.New(nil, delegate, 0, 0)
	l.SetFilteringEnabled(true)
	l.SetShowStatusBar(false)
	l.SetShowHelp(true)
	l.Title = fmt.Sprintf("📂 %s/", startDir)
	m.list = l

	if err := m.reload(); err != nil {
		return model{}, err
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.result = m.currentDir
			return m, tea.Quit
		case "tab":
			m.cycleDown()
			return m, nil
		case "shift+tab":
			m.cycleUp()
			return m, nil
		case "enter":
			selected, ok := m.list.SelectedItem().(entryItem)
			if !ok {
				m.result = m.currentDir
				return m, tea.Quit
			}
			if selected.entry.Kind == fsutil.KindDir {
				m.currentDir = filepath.Clean(selected.entry.Path)
				if err := m.reload(); err != nil {
					m.err = err
					return m, tea.Quit
				}
				return m, nil
			}

			m.result = filepath.Clean(selected.entry.Path)
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return m.list.View()
}

func (m *model) cycleDown() {
	count := len(m.list.Items())
	if count == 0 {
		return
	}
	m.list.Select((m.list.Index() + 1) % count)
}

func (m *model) cycleUp() {
	count := len(m.list.Items())
	if count == 0 {
		return
	}
	m.list.Select((m.list.Index() - 1 + count) % count)
}

func (m *model) reload() error {
	items, err := fsutil.Scan(m.currentDir, m.includeHidden)
	if err != nil {
		return err
	}

	entries := make([]fsutil.Entry, 0, len(items)+1)
	if !isRoot(m.currentDir) {
		entries = append(entries, fsutil.Entry{
			Name: "..",
			Kind: fsutil.KindDir,
			Path: filepath.Dir(m.currentDir),
		})
	}
	entries = append(entries, items...)

	m.list.SetItems(toListItems(entries))
	m.list.Title = fmt.Sprintf("📂 %s/", m.currentDir)
	return nil
}

func isRoot(path string) bool {
	clean := filepath.Clean(path)
	return filepath.Dir(clean) == clean || clean == string(os.PathSeparator)
}
