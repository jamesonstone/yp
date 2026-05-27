# yp

🧲 Yank Path

`yp` copies a resolved absolute path to your clipboard and prints it as `📋 <path>`.

## Install

```bash
go install github.com/jamesonstone/yp@latest
```

## Usage

- `yp` opens the interactive picker in the current directory.
- `yp <dir>` opens the picker in `<dir>`.
- `yp <file>` skips the picker, copies the file path, and prints it.
- `yp --no-hidden` hides dotfiles in the picker.
