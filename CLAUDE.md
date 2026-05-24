# CLAUDE.md — Squad AI

Squad AI is a Go CLI tool that manages AI coding agents inside dev containers.

## Project Rules

**Read the full code standards in `AGENTS.md`**. Key rules:

- Go idiomatic, stdlib first, minimize deps (cobra, bubbletea, lipgloss, bubbles, testify)
- Conventional Commits: `feat(scope):`, `fix(scope):`, `docs:`, `refactor:`, etc.
- Branch naming: `feat/`, `fix/`, `docs/`, `chore/`, etc.
- Exported symbols need godoc comments
- Handle every error, no panic in library code
- Table-driven tests with testify

## SDD Workflow

This project uses Spec-Driven Development via OpenSpec.

- Spec files live in `openspec/specs/`
- Active changes in `openspec/changes/`
- Archived changes in `openspec/changes/archive/`
- Session artifact store: OpenSpec (primary) + Engram (copy)

SDD phases: explore → propose → spec → design → tasks → apply → verify → archive

## Project Structure (from PRD)

```
cmd/squad/          Entry point
internal/cli/       Cobra commands
internal/tui/       Bubbletea models and views
internal/registry/  Remote registry fetch, cache, parse
internal/installer/ Agent installation and progress
internal/config/    User configuration read/write
internal/runtime/   Runtime dependency detection
```

## Testing

```bash
go test ./...           # Unit tests
```

## Build

```bash
go build -o squad ./cmd/squad
```
