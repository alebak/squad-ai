# Proposal: Enter Toggles, Done Confirms

## Intent

Enter key currently confirms the entire selection and exits the TUI. Users
expect Enter to toggle the current item (like Space), with confirmation
reserved for an explicit "Done" item at the bottom of the list. This removes
the footgun of accidentally confirming with Enter while navigating.

## Scope

### In Scope
- Enter toggles agent selection (same behavior as Space)
- Sentinel "✓ Done" item appended at list bottom
- Both Enter and Space on "Done" confirm and exit
- All toggle/skip logic respects the new Done sentinel
- Help bar updated
- Tests for Enter toggle, Done confirm, Done rendering, selectedIDs exclusion

### Out of Scope
- Keyboard shortcut customization
- Multiple confirmation modes
- Visual themes for the Done item

## Capabilities

### New Capabilities
- `enter-toggle-done-confirm`: Enter toggles items, Done item confirms

### Modified Capabilities
- `add`: TUI Interactive Selection requirement changes — Enter no longer
  confirms, Done item appended

## Approach

Add `IsDone bool` field to `AgentItem`. In `newModel`, append a sentinel
`{ID: "_done", Name: "✓ Done", IsDone: true}`. In `handleSpecialKey`, Enter
toggles the same way Space does, except on the Done item where it confirms.
Skip Done item in `selectedIDs`, `toggleAll`, and `renderAgentRow` checkbox.
Update help bar.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modified | AgentItem, newModel, handleSpecialKey, renderAgentRow, selectedIDs, toggleAll, View |
| `internal/tui/model_test.go` | Modified | New tests for Enter toggle, Done confirm, etc. |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Existing Enter-confirm tests break | Low | Update test assertions; Enter now toggles, Done confirms |
| Callers expect Enter-confirm | Low | `RunSelection` return unchanged — just the trigger changes |

## Rollback Plan

Revert the `handleSpecialKey` and `newModel` changes. Remove `IsDone` field
from `AgentItem`. Restore `KeyEnter` to `m.isSubmitted = true; return m, true`.

## Dependencies

None.

## Success Criteria

- [ ] Enter toggles agent checkboxes (same as Space)
- [ ] "✓ Done" appears at list bottom
- [ ] Enter/Space on Done confirms selection
- [ ] All existing tests pass with updated assertions
- [ ] New tests cover Enter toggle, Done confirm, rendering, and exclusion
