package cmd

import (
	"fmt"
	"strings"
)

type parsedArgs struct {
	help         bool
	copyContents bool
	paths        []string
}

const usageText = `yp copies filesystem paths to the clipboard.

Usage:
  yp [flags] [path...]

Flags:
  -c, --copy    copy one file's contents instead of its path
  -h, --help    show this help

Without -c, yp copies an existing file or directory path. No arguments,
or a directory with a trailing separator, opens fzf when it is available.
A name that starts with "-" is a flag. Pass that path as ./name or after --.

With -c or --copy, yp copies one regular file's bytes through pbcopy,
xclip, or wl-copy, the same way as pbcopy < /path/to/file. yp -c <dir>/
copies the file you pick. A directory, a missing path, or more than one
path is an error and leaves the clipboard unchanged. A missing path
still suggests a close sibling file. The confirmation shows the file
and its size, not the contents. Non-regular files and files that
contain a NUL byte are rejected.
`

func parseArgs(args []string) (parsedArgs, error) {
	parsed := parsedArgs{}
	for index := 0; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--":
			parsed.paths = append(parsed.paths, args[index+1:]...)
			return parsed, nil
		case "-h", "--help":
			parsed.help = true
			return parsed, nil
		case "-c", "--copy":
			parsed.copyContents = true
		default:
			if isUnknownFlag(arg) {
				return parsedArgs{}, fmt.Errorf("unknown flag: %s", arg)
			}
			parsed.paths = append(parsed.paths, arg)
		}
	}
	return parsed, nil
}

func isUnknownFlag(arg string) bool {
	return arg != "-" && strings.HasPrefix(arg, "-")
}
