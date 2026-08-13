# External Systems Reference

## Purpose

- Record durable notes about external systems, APIs, providers, or design sources that recur across features
- Keep feature-specific reference details in feature docs as canonical front matter references

## Current State

- `yp` does not call external services or APIs.
- Interactive selection delegates to the local `fzf` executable when it is available.
- Clipboard delivery delegates to the first supported local clipboard command for the operating system: `pbcopy`, `xclip`, or `wl-copy`.
- Tests replace local command execution with fakes; they do not require those executables to be installed.
