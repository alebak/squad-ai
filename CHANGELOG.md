# Changelog

## [0.1.0] — 2026-05-24

### Features

- Bootstrap Go project with `squad version` command
- Agent registry with HTTP fetch and local cache (7 MVP agents)
- User config with atomic read/write (`~/.config/squad-ai/config.json`)
- Runtime detection: Node.js, Go, Python version parsing
- Agent detection via `exec.LookPath`
- Installation pipeline with checksum verification and log capture
- CLI commands: install, list, add, remove, update, info
- Interactive TUI for agent selection (Bubbletea)
- GoReleaser config and install script for distribution
- Release-please for automated changelog and versioning

[0.1.0]: https://github.com/alebak/squad-ai/releases/tag/v0.1.0
