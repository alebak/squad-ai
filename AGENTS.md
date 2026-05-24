# AGENTS.md — Squad AI Code Standards

## Project Context

Squad AI is a Go CLI tool that manages the installation of AI coding agents inside dev containers. It uses Cobra for commands, Bubbletea for TUI, and GoReleaser for distribution.

The main developer is learning Go. Code clarity and idiomatic patterns take priority over clever abstractions.

## Language: Go

### General Rules

- Write idiomatic Go. Follow Effective Go and the Go Code Review Comments guide.
- Prefer simplicity over abstraction. If a function does one thing clearly, it's good enough.
- No global mutable state. Pass dependencies explicitly.
- Every exported function, type, and constant must have a godoc comment starting with the name.
- Keep functions short. If a function exceeds 40 lines, it probably needs to be split.
- Use `errors.New()` or `fmt.Errorf()` for errors. Wrap errors with `%w` for context.
- Handle every error. Never use `_` to discard errors silently.
- No `panic()` in library code. Only acceptable in `main()` for truly unrecoverable situations.

### Naming

- Use MixedCaps (not snake_case). Acronyms are all caps: `URL`, `HTTP`, `ID`.
- Package names are short, lowercase, singular: `registry`, `installer`, `config`.
- Interface names describe behavior: `Reader`, `Installer`, `Detector`.
- Avoid stuttering: `registry.Registry` is bad, `registry.Client` is good.
- Boolean variables and functions read as questions: `isInstalled`, `hasRuntime`.

### Project Structure

- `cmd/squad/main.go` — entry point only. No logic here.
- `internal/cli/` — Cobra command definitions.
- `internal/tui/` — Bubbletea models and views.
- `internal/registry/` — Remote registry fetch, cache, and parsing.
- `internal/installer/` — Agent installation and progress tracking.
- `internal/config/` — User configuration read/write.
- `internal/runtime/` — Runtime dependency detection (Node.js, Go, Python).

### Error Handling

- Always add context when wrapping: `fmt.Errorf("fetching registry: %w", err)`.
- Return errors to the caller. Let `main()` or the CLI layer decide how to present them.
- Use custom error types only when the caller needs to distinguish error kinds.
- Log errors at the boundary (CLI layer), not deep in the code.

### Dependencies

- Minimize external dependencies. stdlib is preferred when it can do the job.
- Approved dependencies: cobra, bubbletea, lipgloss, bubbles, testify.
- Adding a new dependency requires justification. Do not add a library for something Go stdlib handles.

### Testing

- Test files live next to the code: `installer.go` → `installer_test.go`.
- Use table-driven tests for functions with multiple input/output combinations.
- Test names follow `TestFunctionName_Scenario`: `TestDetectRuntime_NodeNotInstalled`.
- Use testify for assertions, but don't overuse mocks. Prefer real implementations with test data.

### Commits

- Follow Conventional Commits: `feat:`, `fix:`, `docs:`, `chore:`, `refactor:`, `test:`.
- Commit messages in English.
- One logical change per commit. Do not mix refactoring with features.

### Security

- Never execute user input as shell commands without validation.
- Validate all data from the remote registry before using it.
- Verify checksums before executing installation scripts.
- No hardcoded secrets or API keys.

## What the reviewer should flag

- Exported symbols without godoc comments.
- Discarded errors (`_`).
- Functions longer than 40 lines.
- New dependencies without clear justification.
- Shell command construction from unvalidated input.
- Global mutable state.
- Logic in `main.go` beyond wiring.
- Package name stuttering.
- Hardcoded paths (should use `os.UserConfigDir()` or similar).
