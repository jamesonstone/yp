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
	targetRunes := []rune(strings.ToLower(target))
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

		candidateRunes := []rune(strings.ToLower(name))
		lengthDifference := len(targetRunes) - len(candidateRunes)
		if lengthDifference < 0 {
			lengthDifference = -lengthDifference
		}
		if lengthDifference > maximumDistance {
			continue
		}

		distance := editDistance(targetRunes, candidateRunes, maximumDistance)
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

func editDistance(left, right []rune, maximumDistance int) int {
	outsideLimit := maximumDistance + 1
	if len(left)-len(right) > maximumDistance || len(right)-len(left) > maximumDistance {
		return outsideLimit
	}

	previous := make([]int, len(right)+1)
	current := make([]int, len(right)+1)
	for index := range previous {
		previous[index] = outsideLimit
		current[index] = outsideLimit
	}
	for index := 0; index <= min(len(right), maximumDistance); index++ {
		previous[index] = index
	}

	for leftIndex, leftRune := range left {
		row := leftIndex + 1
		start := max(1, row-maximumDistance)
		end := min(len(right), row+maximumDistance)
		rowMinimum := outsideLimit

		if start == 1 {
			current[0] = row
			rowMinimum = row
		} else {
			current[start-1] = outsideLimit
		}

		for column := start; column <= end; column++ {
			replacementCost := 0
			if leftRune != right[column-1] {
				replacementCost = 1
			}
			current[column] = min(
				current[column-1]+1,
				previous[column]+1,
				previous[column-1]+replacementCost,
			)
			rowMinimum = min(rowMinimum, current[column])
		}
		if end < len(right) {
			current[end+1] = outsideLimit
		}
		if rowMinimum > maximumDistance {
			return outsideLimit
		}
		previous, current = current, previous
	}

	if previous[len(right)] > maximumDistance {
		return outsideLimit
	}
	return previous[len(right)]
}
