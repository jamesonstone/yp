package clip

import "github.com/atotto/clipboard"

var writeAll = clipboard.WriteAll

func Write(value string) error {
	return writeAll(value)
}
