# Tasks: Bulk Uninstall Confirmation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~60-80 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

## Phase 1: Core Logic — `internal/cli/add.go`

- [x] 1.1 Collect all deselected installed agents first (before any prompt) into a slice
- [x] 1.2 Branch on count: if ==0 skip, if ==1 existing per-agent 3-option flow, if >1 combined `confirmFn` prompt
- [x] 1.3 Implement bulk yes path: iterate slice, call `UninstallAgent` (app-only) for each, `delete(installed, id)`
- [x] 1.4 Implement bulk no path: keep all in `installed`, set `needsRestart=true`, loop restarts with all pre-checked

## Phase 2: Tests — `internal/cli/add_test.go`

- [x] 2.1 Write `TestAddCommand_BulkUninstallConfirm` — 2 installed agents deselected, confirmFn returns true, both uninstalled
- [x] 2.2 Write `TestAddCommand_BulkUninstallDecline` — 2 installed agents deselected, confirmFn returns false, TUI restarts
- [x] 2.3 Verify existing tests still pass: `TestAddCommand_UninstallAppOnly`, `TestAddCommand_UninstallPromptCancel`, `TestAddCommand_TUISuccessFlow`
