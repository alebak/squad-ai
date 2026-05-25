# Tasks: Fix TUI 'q' exit triggering uninstall prompt

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~15 (3 code + 4 godoc + 8 test) |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |
| Chain strategy | pending |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

## Phase 1: Core Fix

- [x] 1.1 Add nil guard after `RunSelection` in `internal/cli/add.go:runAddFlowInteractive`
- [x] 1.2 Update `RunSelection` godoc in `internal/tui/model.go` to document nil vs empty-slice contract

## Phase 2: Test Fix

- [x] 2.1 Update `TestAddCommand_TUIEmptySelection` in `internal/cli/add_test.go` — change `runSelection` mock from `return nil, nil` to `return []string{}, nil`

## Phase 3: Verification

- [x] 3.1 Run `go test ./internal/cli/` and `go test ./internal/tui/` — all pass (28 + 19 tests)
- [x] 3.2 Run `go build ./cmd/squad` — compiles cleanly
