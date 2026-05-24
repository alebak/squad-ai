# Tasks: Project Bootstrap

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~128 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: pending
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Full bootstrap: module, dirs, CLI, version, test | PR 1 (single) | Base: main. All code + test + verify in one PR. |

## Phase 1: Foundation

- [x] 1.1 Create `go.mod` — `module github.com/alebak/squad-ai`, `go 1.24`, require `cobra v1.10.2` and `testify v1.11.1`
- [x] 1.2 Run `go mod tidy` to generate `go.sum` with dependency checksums
- [x] 1.3 Create directory tree: `cmd/squad/`, `internal/{cli,tui,registry,installer,config,runtime}/` each with `.gitkeep`

## Phase 2: Root Command

- [x] 2.1 Create `internal/cli/root.go` — `var rootCmd` with `Use:"squad"`, `Short`, `Version:"0.1.0"`. Exported `Execute()` calling `rootCmd.Execute()`. Godoc on `Execute()`.

## Phase 3: Version Command

- [x] 3.1 Create `internal/cli/version.go` — `var versionCmd` with `Use:"version"`, `Run` printing `"squad version {version}"`. Register via `init()`. No exported symbols beyond `init()`.
- [x] 3.2 Create `internal/cli/version_test.go` — table-driven `TestVersionCommand_Output` with two cases (subcommand, flag). Capture stdout via `rootCmd.SetOut(buf)`. Assert `"squad version 0.1.0"`. Use `require.NoError` + `assert.Contains`.

## Phase 4: Entry Point

- [x] 4.1 Create `cmd/squad/main.go` — `package main`, import `squad-ai/internal/cli`, `main()` calls `cli.Execute()`. No business logic.

## Phase 5: Build & Verify

- [x] 5.1 Build: `go build ./cmd/squad` — binary compiles without errors
- [x] 5.2 Smoke test: `./squad version` prints `squad version 0.1.0`, exit code 0
- [x] 5.3 Smoke test: `./squad --version` prints `squad version 0.1.0`, exit code 0
- [x] 5.4 Unit test: `go test ./internal/cli/...` — all tests pass
- [x] 5.5 Lint: `go vet ./...` — no issues
