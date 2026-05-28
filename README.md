```text
██╗   ██╗██████╗
╚██╗ ██╔╝██╔══██╗
 ╚████╔╝ ██████╔╝
  ╚██╔╝  ██╔═══╝
   ██║   ██║      yank paths
   ╚═╝   ╚═╝
```

**`yp` is a tiny CLI for copying filesystem paths to your clipboard.** Give it
a file or missing path and it copies that path directly. Give it a directory
and, when `fzf` is installed, it opens the same picker workflow as
`/Users/jamesonstone/.config/zsh/functions/yp.zsh`.

No shell integration. No background process. Just `📋 <path>`.

## Install

```sh
git clone https://github.com/jamesonstone/yp.git
cd yp
make build
./bin/yp README.md
```

For a local install:

```sh
make install
yp README.md
```

## Quick Start

```sh
# browse the current directory with fzf when available
yp

# browse a specific directory
yp ~/src

# copy a file path directly
yp README.md

# copy a not-yet-created path
yp future/file.txt

# copy every immediate subdirectory path as a newline-separated list
yp ~/src/*
```

## Behavior

- `yp` uses the current directory.
- `yp <dir>` opens `fzf` when available; without `fzf`, it copies the directory.
- `yp <file>` copies the file path directly.
- `yp <missing-path>` matches the zsh function's path resolution.
- `yp <dir>/*` copies every immediate, non-hidden subdirectory path.
- Shell-expanded directory args are copied as a newline-separated list.
- If multiple args contain no directories, `yp` falls back to the first arg.

The picker uses `ls -1A`, so dotfiles are shown. `enter` drills into
directories or accepts files; `tab` and `shift+tab` move through rows; `esc`
copies the current directory. Type `*` in the picker and press `enter` to copy
all immediate, non-hidden subdirectory paths from the current picker directory.

## Requirements

- Go 1.22+
- `pbcopy`, `xclip`, or `wl-copy`
- `fzf` for interactive directory browsing

## Development

```sh
make build
make test
make vet
```

The project intentionally stays small: standard-library Go plus the same system
tools used by the original zsh function.
