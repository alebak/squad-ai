# Proposal: TUI Restart After Uninstall

## Intent

Bug fix: after successfully uninstalling an agent (options [1] or [2]), `squad add` exits to the terminal instead of returning to the TUI. Users are forced to re-run the command to continue managing agents. The TUI loop must restart after any uninstall action so users see the updated state and can continue.

## Scope

### In Scope
- Set `needsRestart = true` after `uninstallAppOnly` and `uninstallAppConfig` cases in `runAddFlowInteractive`
- Update per-agent 3-option prompt handling so the loop restarts after any uninstall
- Update existing tests to expect restart after uninstall
- Remove outdated spec scenario that says "flow proceeds without re-launching"

### Out of Scope
- Bulk uninstall prompt (already sets `needsRestart` after decline but not after confirm — fix included)
- Non-interactive flow (no TUI to restart)
- UI improvements beyond the restart fix

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `add`: Requirement for 3-Option Uninstall Prompt — restart TUI after ANY uninstall decision, not just Cancel

## Approach

Add `needsRestart = true` in both `uninstallAppOnly` and `uninstallAppConfig` switch cases in `runAddFlowInteractive`. The loop already rebuilds `agentItems` via `buildAgentItemsForAdd(h, catalog, installed)` and re-launches the TUI when `needsRestart` is true. Same pattern used for the bulk confirm path (lines 308-316).

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modified | Add `needsRestart = true` in 2 switch cases |
| `internal/cli/add_test.go` | Modified | 3 tests need `runSelection` to return twice |
| `openspec/specs/add/spec.md` | Modified | Remove outdated scenario "proceeds without re-launching" |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Infinite loop if user keeps uninstalling | Low | Loop terminates when `deselected` is empty. Uninstalled agents won't appear installed on re-launch. |
| Test regression | Low | Update affected mocks to expect 2 TUI calls |

## Rollback Plan

Revert the git changes to `internal/cli/add.go` and `internal/cli/add_test.go`. Revert spec changes via git checkout.

## Dependencies

None.

## Success Criteria

- [ ] After uninstall [1] or [2], TUI re-launches instead of exiting
- [ ] All existing tests pass after updating for new behavior
- [ ] Spec accurately describes the new behavior
