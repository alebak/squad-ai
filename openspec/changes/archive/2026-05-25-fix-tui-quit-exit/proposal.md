# Proposal: Fix TUI 'q' exit triggering uninstall prompt

## Intent

Bug fix: pressing `q` in the TUI selection screen shows the uninstall prompt instead of exiting back to the shell. `RunSelection` returns `nil` when the user quits (via `q`, Ctrl+C, Escape) and `[]` when the user confirms a genuinely empty selection — but `runAddFlowInteractive` treats both identically via `len(selectedIDs) == 0`.

## Scope

### In Scope
- Add nil-check guard after `RunSelection` returns in `runAddFlowInteractive`
- Update `TestAddCommand_TUIEmptySelection` to use `[]string{}` (empty slice) instead of `nil`
- Update `RunSelection` godoc to document the nil/empty-slice contract

### Out of Scope
- Non-interactive mode — unaffected
- Other TUI key bindings — unaffected
- Other callers of `RunSelection` — only `add.go` uses it

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `add`: The `RunSelection` return contract is refined — callers MUST distinguish `nil` (user quit) from `[]` (confirmed empty selection)

## Approach

Add a `selectedIDs == nil` check immediately after the `RunSelection` call in `runAddFlowInteractive`. If nil, return `cfg, nil` (clean exit). Empty slice continues to the uninstall/restart logic.

Single file change: `internal/cli/add.go`, line ~240.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modified | Add nil guard after RunSelection |
| `internal/tui/model.go` | Modified (doc) | Update RunSelection godoc to specify nil vs empty contract |
| `internal/cli/add_test.go` | Modified | Fix test to use empty slice (not nil) for "confirmed empty" case |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Existing tests simulate quit with `nil` — might mask regression | Low | Fix the test to use `[]string{}` for empty selection; keep a separate test for quit returning nil |

## Rollback Plan

Revert the nil check and godoc changes. Tests will pass as before.

## Dependencies

None.

## Success Criteria

- [ ] Pressing `q` in TUI exits cleanly without uninstall prompt
- [ ] Confirming empty selection (Enter with nothing checked) still shows "No agents selected"
- [ ] Existing uninstall flow (deselect installed agent) continues to work
- [ ] All existing tests pass
