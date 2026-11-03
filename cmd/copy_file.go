package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
)

func copyFile(
	args []string,
	out io.Writer,
	writeClipboard func(string) error,
	pickerAvailable func() bool,
	pickPath func(string) (string, error),
) error {
	if len(args) > 1 || (len(args) == 1 && isStarPath(args[0])) {
		return fmt.Errorf("-c/--copy expects exactly one file")
	}

	start := "."
	if len(args) == 1 && args[0] != "" {
		start = args[0]
	}
	if err := validatePath(start); err != nil {
		return err
	}

	resolved, err := resolvePath(start)
	if err != nil {
		return err
	}

	target := resolved
	if isDir(resolved) && shouldUsePicker(args) && pickerAvailable() {
		target, err = pickPath(resolved)
		if err != nil {
			return err
		}
	}
	if strings.Contains(target, "\n") {
		return fmt.Errorf("-c/--copy expects exactly one file")
	}
	return copyFileContents(target, out, writeClipboard)
}

func copyFileContents(target string, out io.Writer, writeClipboard func(string) error) error {
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", target)
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", target)
	}

	data, err := os.ReadFile(target)
	if err != nil {
		return err
	}
	if bytes.Contains(data, []byte{0}) {
		return fmt.Errorf("%s is a binary file", target)
	}
	if err := writeClipboard(string(data)); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "📋 %s (%s)\n", target, formatCopySize(data))
	return err
}

func formatCopySize(data []byte) string {
	byteCount := len(data)
	lineCount := countLines(data)
	return fmt.Sprintf("%d %s, %d %s",
		byteCount, pluralWord(byteCount, "byte", "bytes"),
		lineCount, pluralWord(lineCount, "line", "lines"),
	)
}

func countLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

func pluralWord(count int, singular, plural string) string {
	if count == 1 {
		return singular
	}
	return plural
}
