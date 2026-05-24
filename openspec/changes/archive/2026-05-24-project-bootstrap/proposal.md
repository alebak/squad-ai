# Proposal: Project Bootstrap

## Intent

Bootstrap the Go project with module definition, directory structure, entry point, and a working `squad version` command. Establishes foundational patterns (Cobra wiring, test alongside code) for all future development. Covers PRD §8 (Comandos del CLI) and §14 (Fase 1 — MVP).

## Scope

### In Scope
- `go.mod` / `go.sum` — module `github.com/alebak/squad-ai`, Go 1.24+
- `cmd/squad/main.go` — entry point calling `cli.Execute()`
- `internal/cli/root.go` — root `squad` command with Cobra `Version` field
- `internal/cli/version.go` — `squad version` subcommand, output `0.1.0`
- `internal/cli/version_test.go` — test capturing stdout, asserting `0.1.0`
- 7 dirs with `.gitkeep`: `cmd/squad/`, `internal/{cli,tui,registry,installer,config,runtime}/`

### Out of Scope
- Other CLI commands (`install`, `add`, `list`, `remove`, `info`, `update`)
- TUI, registry, config, installer, runtime detection code
- CI/CD, GoReleaser, install.sh (deferred to later phase)

## Capabilities

### New Capabilities
- `version-command`: `squad version` outputs the current version. Cobra `--version` flag also works. Tested by capturing stdout.

### Modified Capabilities
None — greenfield project, no existing specs.

## Approach

Approach 1 from exploration: `go mod init` → create directory tree → implement `squad version` via Cobra (`root.go` sets `Version`, `version.go` adds subcommand) → write one test using `cmd.SetOut()` + `ExecuteC`. No speculative code. Version `0.1.0` matches PRD.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `go.mod` | New | Module definition |
| `go.sum` | New | Dependency lockfile |
| `cmd/squad/main.go` | New | Entry point |
| `internal/cli/root.go` | New | Root Cobra command |
| `internal/cli/version.go` | New | `squad version` subcommand |
| `internal/cli/version_test.go` | New | First test |
| `internal/*/` (6 pkg dirs) | New | Empty with `.gitkeep` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Module path mismatch | Low | Use `github.com/alebak/squad-ai` — verified in PRD |
| go.sum stale before commit | Low | Run `go mod tidy` before final commit |
| Cobra API changes | Low | Pin `v1.10.2` in go.mod |

## Rollback Plan

`git revert` the bootstrap commit. No data migration needed — greenfield project, no production state.

## Dependencies

- `github.com/spf13/cobra` v1.10.2
- `github.com/stretchr/testify` v1.11.1

## Success Criteria

- [ ] `go build ./cmd/squad` produces a working binary
- [ ] `squad version` outputs `0.1.0`
- [ ] `squad --version` outputs `0.1.0` (Cobra built-in)
- [ ] `go test ./internal/cli/...` passes
- [ ] All 7 internal package directories exist with `.gitkeep`
