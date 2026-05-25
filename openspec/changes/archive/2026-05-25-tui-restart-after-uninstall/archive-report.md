# Archive Report: TUI Restart After Uninstall

## Change Summary

**Issue**: #28 — After successfully uninstalling an agent (choosing [1] or [2]), squad exits to terminal instead of returning to TUI.

**Root Cause**: In `runAddFlowInteractive`, the TUI restart loop only set `needsRestart = true` for the `uninstallCancel` case. Options 1 and 2 (and bulk confirm) performed the uninstall but fell through the loop to installation, then exit.

**Fix**: Added `needsRestart = true` after every uninstall action:
- `uninstallAppOnly` (per-agent)
- `uninstallAppConfig` (per-agent)  
- Bulk uninstall confirmation (multi-agent)

## Delivered

| File | Lines Changed | Description |
|------|---------------|-------------|
| `internal/cli/add.go` | +6 | `needsRestart = true` in 3 locations |
| `internal/cli/add_test.go` | +15 | 3 tests updated to expect TUI restart |
| `openspec/specs/add/spec.md` | ~15 | Updated spec to reflect new restart behavior |

## Verification

- **Tests**: All 76 tests pass across 6 packages
- **Build**: `go build ./cmd/squad` compiles cleanly
- **Coverage**: All uninstall paths (per-agent app-only, per-agent app+config, per-agent cancel, bulk confirm, bulk decline) covered by existing tests with updated assertions

## Lineage

- Proposal: `sdd/tui-restart-after-uninstall/proposal` (#95)
- Spec: `sdd/tui-restart-after-uninstall/spec` (#96)
- Design: `sdd/tui-restart-after-uninstall/design` (#97)
- Tasks: `sdd/tui-restart-after-uninstall/tasks` (#98)
- Apply: `sdd/tui-restart-after-uninstall/apply-progress` (#99)
- Verify: `sdd/tui-restart-after-uninstall/verify-report` (#100)
