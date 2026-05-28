package picker

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func Available() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

func Run(startDir string) (string, error) {
	state, err := os.CreateTemp("", "yp.*")
	if err != nil {
		return "", err
	}
	defer os.Remove(state.Name())
	if _, err := state.WriteString(startDir); err != nil {
		state.Close()
		return "", err
	}
	if err := state.Close(); err != nil {
		return "", err
	}

	helperPath, err := writeHelper(state.Name())
	if err != nil {
		return "", err
	}
	defer os.Remove(helperPath)

	initialList, err := exec.Command(helperPath, "list").Output()
	if err != nil {
		return "", err
	}

	var picked bytes.Buffer
	cmd := exec.Command("fzf",
		"--height", "60%",
		"--reverse",
		"--cycle",
		"--prompt=📂 "+startDir+"/",
		"--pointer=👉",
		"--header=*: copy subdirs · tab/shift-tab: cycle · enter: drill or pick · esc: cancel",
		"--delimiter=\t",
		"--with-nth=1,2",
		"--nth=2",
		"--bind", "tab:down,btab:up",
		"--bind", "enter:transform("+shellQuote(helperPath)+" enter {})",
	)
	cmd.Stdin = bytes.NewReader(initialList)
	cmd.Stdout = &picked
	cmd.Stderr = os.Stderr
	_ = cmd.Run()

	currentDir, err := readState(state.Name())
	if err != nil {
		return "", err
	}

	line := strings.TrimRight(picked.String(), "\n")
	if line == "" {
		return currentDir, nil
	}
	return targetFromPicked(currentDir, line)
}

func writeHelper(statePath string) (string, error) {
	helper, err := os.CreateTemp("", "yph.*")
	if err != nil {
		return "", err
	}
	helperPath := helper.Name()

	script := fmt.Sprintf(`#!/bin/sh
action="$1"
line="$2"
state=%s
self="$0"
case "$action" in
  list)
    dir=$(cat "$state")
    [ "$dir" != "/" ] && printf "📁\t..\n"
    ls -1A "$dir" 2>/dev/null | while IFS= read -r f; do
      if [ -L "$dir/$f" ]; then printf "🔗\t%%s\n" "$f"
      elif [ -d "$dir/$f" ]; then printf "📁\t%%s\n" "$f"
      elif [ -x "$dir/$f" ]; then printf "⚙️\t%%s\n" "$f"
      else printf "📄\t%%s\n" "$f"
      fi
    done
    printf "✳️\t*\n"
    ;;
  prompt)
    printf "📂 %%s/" "$(cat "$state")"
    ;;
  enter)
    type=$(printf "%%s" "$line" | cut -f1)
    if [ "$type" = "📁" ]; then
      name=$(printf "%%s" "$line" | cut -f2-)
      cur=$(cat "$state")
      if [ -d "$cur/$name" ]; then
        newdir=$(cd "$cur/$name" 2>/dev/null && pwd) && printf "%%s" "$newdir" > "$state"
      fi
      echo "reload($self list)+transform-prompt($self prompt)"
    else
      echo "accept"
    fi
    ;;
esac
`, shellQuote(statePath))

	if _, err := helper.WriteString(script); err != nil {
		helper.Close()
		os.Remove(helperPath)
		return "", err
	}
	if err := helper.Close(); err != nil {
		os.Remove(helperPath)
		return "", err
	}
	if err := os.Chmod(helperPath, 0o755); err != nil {
		os.Remove(helperPath)
		return "", err
	}
	return helperPath, nil
}

func readState(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func targetFromPicked(currentDir, line string) (string, error) {
	name := pickedName(line)
	if name == "*" {
		return subdirectoryList(currentDir)
	}
	return currentDir + string(os.PathSeparator) + name, nil
}

func pickedName(line string) string {
	_, name, ok := strings.Cut(line, "\t")
	if !ok {
		return line
	}
	return name
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}

func subdirectoryList(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	targets := make([]string, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}

		fullPath := dir + string(os.PathSeparator) + name
		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			targets = append(targets, fullPath)
		}
	}
	if len(targets) == 0 {
		return "", fmt.Errorf("no subdirectories found for %s", dir)
	}
	return strings.Join(targets, "\n"), nil
}
