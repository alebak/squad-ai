# Verify Report: TUI Restart After Uninstall

## Summary

- **Status**: PASS
- **Date**: 2026-05-25
- **Tests run**: `go test ./... -count=1` — 76 tests, all pass
- **Build**: `go build ./cmd/squad` — compiles cleanly

## Verification Against Spec

### Per-agent: Uninstall app only → TUI restarts
- **Test**: `TestAddCommand_UninstallAppOnly` — PASS
- **Assertion**: `runSelectionCallCount == 2`, uninstall output present

### Per-agent: Uninstall app + config → TUI restarts
- **Test**: `TestAddCommand_UninstallAppAndConfig` — PASS
- **Assertion**: `runSelectionCallCount == 2`, uninstall + config cleanup output present

### Per-agent: Cancel → TUI restarts (unchanged)
- **Test**: `TestAddCommand_UninstallPromptCancel` — PASS
- **Assertion**: `runSelectionCallCount == 2`, no uninstall called

### Bulk confirm → TUI restarts
- **Test**: `TestAddCommand_BulkUninstallConfirm` — PASS
- **Assertion**: `runSelectionCallCount == 2`, both agents uninstalled

### Bulk decline → TUI restarts (unchanged)
- **Test**: `TestAddCommand_BulkUninstallDecline` — PASS
- **Assertion**: `runSelectionCallCount == 2`, no uninstall called

## Delivered Changes

| File | Change |
|------|--------|
| `internal/cli/add.go` | `needsRestart = true` after `uninstallAppOnly`, `uninstallAppConfig`, and bulk confirm |
| `internal/cli/add_test.go` | 3 tests updated to expect TUI restart after uninstall |

## Risks

None. All behavior changes are covered by existing test patterns (call counting). The logic change is minimal: 3 lines of `needsRestart = true`.
