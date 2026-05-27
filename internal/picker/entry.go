package picker

import "github.com/jamesonstone/yp/internal/fsutil"

func iconForKind(kind fsutil.EntryKind) string {
	switch kind {
	case fsutil.KindDir:
		return "📁"
	case fsutil.KindSymlink:
		return "🔗"
	case fsutil.KindExec:
		return "⚙️"
	default:
		return "📄"
	}
}
