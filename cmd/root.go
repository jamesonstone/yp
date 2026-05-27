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
