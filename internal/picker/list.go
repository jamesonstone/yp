package picker

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"

	"github.com/jamesonstone/yp/internal/fsutil"
)

type entryItem struct {
	entry fsutil.Entry
}

func (e entryItem) Title() string {
	return fmt.Sprintf("%s %s", iconForKind(e.entry.Kind), e.entry.Name)
}

func (e entryItem) Description() string {
	return ""
}

func (e entryItem) FilterValue() string {
	return e.entry.Name
}

func toListItems(entries []fsutil.Entry) []list.Item {
	items := make([]list.Item, 0, len(entries))
	for _, entry := range entries {
		items = append(items, entryItem{entry: entry})
	}
	return items
}
