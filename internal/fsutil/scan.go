package fsutil

import (
	"os"
	"path/filepath"
	"strings"
)

type EntryKind int

const (
	KindDir EntryKind = iota
	KindFile
	KindSymlink
	KindExec
)

type Entry struct {
	Name string
	Kind EntryKind
	Path string
}

func Scan(dir string, includeHidden bool) ([]Entry, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	absDir = filepath.Clean(absDir)

	entries, err := os.ReadDir(absDir)
	if err != nil {
		return nil, err
	}

	items := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if !includeHidden && strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(absDir, name)
		info, err := os.Lstat(fullPath)
		if err != nil {
			continue
		}

		items = append(items, Entry{
			Name: name,
			Kind: kindFromMode(info.Mode()),
			Path: fullPath,
		})
	}

	return items, nil
}

func kindFromMode(mode os.FileMode) EntryKind {
	if mode&os.ModeSymlink != 0 {
		return KindSymlink
	}
	if mode.IsDir() {
		return KindDir
	}
	if mode&0o111 != 0 {
		return KindExec
	}
	return KindFile
}
