package clip

import (
	"errors"
	"os/exec"
)

var errNoClipboardTool = errors.New("❌ no clipboard tool found")

func Write(value string) error {
	if commandExists("pbcopy") {
		return writeWithCommand(value, "pbcopy")
	}
	if commandExists("xclip") {
		return writeWithCommand(value, "xclip", "-selection", "clipboard")
	}
	if commandExists("wl-copy") {
		return writeWithCommand(value, "wl-copy")
	}
	return errNoClipboardTool
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func writeWithCommand(value, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}
	if _, err := stdin.Write([]byte(value)); err != nil {
		_ = stdin.Close()
		_ = cmd.Wait()
		return err
	}
	if err := stdin.Close(); err != nil {
		_ = cmd.Wait()
		return err
	}
	return cmd.Wait()
}
