package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/jamesonstone/yp/internal/clip"
	"github.com/jamesonstone/yp/internal/picker"
)

func Execute() error {
	return run(os.Args[1:], os.Stdout, clip.Write, picker.Available, picker.Run)
}

func run(
	args []string,
	out io.Writer,
	writeClipboard func(string) error,
	pickerAvailable func() bool,
	pickPath func(string) (string, error),
) error {
	if target, ok, err := directoryListTarget(args); ok || err != nil {
		if err != nil {
			return err
		}
		if err := writeClipboard(target); err != nil {
			return err
		}

		_, err = fmt.Fprintf(out, "📋 %s\n", target)
		return err
	}

	start := "."
	if len(args) > 0 && args[0] != "" {
		start = args[0]
	}

	resolved, err := resolvePath(start)
	if err != nil {
		return err
	}

	target := resolved
	if pickerAvailable() && isDir(resolved) {
		target, err = pickPath(resolved)
		if err != nil {
			return err
		}
	}

	if err := writeClipboard(target); err != nil {
		return err
	}

	_, err = fmt.Fprintf(out, "📋 %s\n", target)
	return err
}

func directoryListTarget(args []string) (string, bool, error) {
	if len(args) == 0 {
		return "", false, nil
	}

	if len(args) == 1 {
		if !isStarPath(args[0]) {
			return "", false, nil
		}

		targets, err := subdirectoriesForStarPath(args[0])
		if err != nil {
			return "", true, err
		}
		if len(targets) == 0 {
			return "", true, fmt.Errorf("no subdirectories found for %s", args[0])
		}
		return strings.Join(targets, "\n"), true, nil
	}

	targets, err := directoryArgs(args)
	if err != nil {
		return "", true, err
	}
	if len(targets) == 0 {
		return "", false, nil
	}
	return strings.Join(targets, "\n"), true, nil
}

func isStarPath(path string) bool {
	return filepath.Base(path) == "*"
}

func subdirectoriesForStarPath(path string) ([]string, error) {
	parent, err := resolvePath(filepath.Dir(path))
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(parent)
	if err != nil {
		return nil, err
	}

	targets := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := filepath.Join(parent, name)
		if isDir(fullPath) {
			targets = append(targets, fullPath)
		}
	}
	return targets, nil
}

func directoryArgs(args []string) ([]string, error) {
	targets := make([]string, 0, len(args))
	for _, arg := range args {
		resolved, err := resolvePath(arg)
		if err != nil {
			return nil, err
		}
		if isDir(resolved) {
			targets = append(targets, resolved)
		}
	}
	return targets, nil
}

func resolvePath(path string) (string, error) {
	if path == "" {
		path = "."
	}

	if isDir(path) {
		return pwdForDir(path)
	}

	dirname := commandOutput("dirname", path)
	basename := commandOutput("basename", path)

	dir, err := pwdForDir(dirname)
	if err != nil {
		return string(os.PathSeparator) + basename, nil
	}
	return dir + string(os.PathSeparator) + basename, nil
}

func pwdForDir(path string) (string, error) {
	if path == "" {
		path = "."
	}

	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if !info.IsDir() {
		return "", fmt.Errorf("%s is not a directory", path)
	}
	if filepath.IsAbs(path) {
		return filepath.Clean(path), nil
	}

	base, err := logicalWorkingDir()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(base, path)), nil
}

func logicalWorkingDir() (string, error) {
	physical, err := os.Getwd()
	if err != nil {
		return "", err
	}

	pwd := os.Getenv("PWD")
	if pwd == "" || !filepath.IsAbs(pwd) {
		return physical, nil
	}

	pwdInfo, err := os.Stat(pwd)
	if err != nil {
		return physical, nil
	}
	physicalInfo, err := os.Stat(physical)
	if err != nil {
		return physical, nil
	}
	if !os.SameFile(pwdInfo, physicalInfo) {
		return physical, nil
	}

	return pwd, nil
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func commandOutput(name, arg string) string {
	cmd := exec.Command(name, arg)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	if err != nil {
		return strings.TrimRight(string(output), "\n")
	}
	return strings.TrimRight(string(output), "\n")
}
