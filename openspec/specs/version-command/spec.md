# version-command Specification

## Purpose

The `version-command` domain defines the foundational project structure, Go module configuration, binary entry point, and `squad version` subcommand. This is the first deliverable that establishes patterns (Cobra wiring, table-driven tests, package layout) for all subsequent commands.

## Requirements

### Requirement: Go Module Initialization

The project MUST define a Go module at `github.com/alebak/squad-ai` targeting Go 1.24+. All third-party dependencies MUST be pinned to specific versions in `go.mod`.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Module path is correct | no existing `go.mod` | `go mod init github.com/alebak/squad-ai` | `go.mod` module directive is `module github.com/alebak/squad-ai` |
| Go version meets minimum | valid `go.mod` | reading the `go` directive | it is `go 1.24` or higher |
| Dependencies pinned | `go.mod` written | `go mod tidy` runs | `go.sum` is generated; cobra is `v1.10.2`; testify is `v1.11.1` |

### Requirement: Directory Structure

The project MUST contain `cmd/squad/` and all 6 `internal/*` directories: `cli`, `tui`, `registry`, `installer`, `config`, `runtime`. Each empty directory MUST contain a `.gitkeep` file.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| All directories exist | project root | listing `cmd/` and `internal/` | `cmd/squad/` and 6 `internal/*/` dirs exist |
| .gitkeep present | each empty directory | checking for `.gitkeep` | file is present |

### Requirement: Entry Point Wiring

`cmd/squad/main.go` SHALL invoke `cli.Execute()` and MUST NOT contain business logic. It SHALL use `package main` and import the `internal/cli` package.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Entry wires root command | `cmd/squad/main.go` | building with `go build ./cmd/squad` | binary compiles and produces exit code 0 |
| No business logic in main | `cmd/squad/main.go` | inspecting the file | only `package main`, import, and `main()` calling `cli.Execute()` |

### Requirement: Version Output

The `squad version` subcommand MUST output the version string to stdout and exit with code 0. The root command MUST set the `Version` field so `squad --version` produces the same output via Cobra's built-in mechanism. The version string SHALL be `0.1.0`.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| `squad version` subcommand | built binary | `squad version` | stdout contains `squad version 0.1.0`; exit code 0 |
| `squad --version` flag | built binary | `squad --version` | stdout contains `squad version 0.1.0`; exit code 0 |

### Requirement: Version Test (Table-Driven)

The version output MUST be tested using a table-driven test in `internal/cli/version_test.go`. The test SHALL capture stdout via `cmd.SetOut()` and use testify for assertions.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Subcommand output validated | root command with captured output | `ExecuteC` runs the version subcommand | output matches `0.1.0` with testify assertion |
| Both paths produce same output | test table with two rows (subcommand, `--version` flag) | each row is executed | both rows produce identical version string |
