# Tooling Reference

## Purpose

- Record durable repo-wide tooling notes, command references, and local development expectations
- Keep short-lived implementation notes in feature docs instead of here

## Current State

- Go 1.22 or newer is the only language toolchain; dependencies are limited to the standard library.
- `make build`, `make test`, and `make vet` wrap the canonical build and validation commands.
- `make install` installs `yp` with the active Go toolchain.
- `make fmt` applies `gofmt` to the Go entry points and package directories; use `test -z "$(gofmt -l .)"` for a read-only formatting check.
- `fzf` is optional and enables the interactive picker.
- One of `pbcopy`, `xclip`, or `wl-copy` is required at runtime to write to the system clipboard.
