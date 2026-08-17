package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func validatePath(path string) error {
	if path == "" {
		path = "."
	}

	_, err := os.Stat(path)
	if err == nil {
		return nil
	}
	if !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("cannot access %q: %w", path, err)
	}

	if suggestion := closestSiblingFile(path); suggestion != "" {
		return fmt.Errorf("%w: %q; did you mean %q?", fs.ErrNotExist, path, suggestion)
	}
	return fmt.Errorf("%w: %q", fs.ErrNotExist, path)
}

func closestSiblingFile(path string) string {
	if strings.HasSuffix(path, string(os.PathSeparator)) {
		return ""
	}

	parent := filepath.Dir(path)
	entries, err := os.ReadDir(parent)
	if err != nil {
		return ""
	}

	target := filepath.Base(path)
	targetExtension := strings.ToLower(filepath.Ext(target))
	maximumDistance := suggestionDistanceLimit(target)
	bestDistance := maximumDistance + 1
	bestName := ""

	for _, entry := range entries {
		name := entry.Name()
		candidate := filepath.Join(parent, name)
		info, err := os.Stat(candidate)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}

		if strings.ToLower(filepath.Ext(name)) != targetExtension {
			continue
		}

		distance := editDistance(strings.ToLower(target), strings.ToLower(name))
		if distance < bestDistance {
			bestDistance = distance
			bestName = name
		}
	}

	if bestName == "" {
		return ""
	}
	return filepath.Join(parent, bestName)
}

func suggestionDistanceLimit(name string) int {
	switch length := len([]rune(name)); {
	case length >= 12:
		return 3
	case length >= 5:
		return 2
	default:
		return 1
	}
}

func editDistance(left, right string) int {
	rightRunes := []rune(right)
	previous := make([]int, len(rightRunes)+1)
	for index := range previous {
		previous[index] = index
	}

	for leftIndex, leftRune := range []rune(left) {
		current := make([]int, len(rightRunes)+1)
		current[0] = leftIndex + 1
		for rightIndex, rightRune := range rightRunes {
			replacementCost := 0
			if leftRune != rightRune {
				replacementCost = 1
			}
			current[rightIndex+1] = min(
				current[rightIndex]+1,
				previous[rightIndex+1]+1,
				previous[rightIndex]+replacementCost,
			)
		}
		previous = current
	}

	return previous[len(rightRunes)]
}
