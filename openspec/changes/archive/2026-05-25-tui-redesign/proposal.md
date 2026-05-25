# Proposal: TUI Agent Selection Redesign

## Intent

Clean up TUI agent selection: remove emoji clutter, add a visible select-all row, show blocked agents as parenthetical text, and make installed agents toggleable with uninstall prompts. Users found emojis noisy and the `a` key unclear.

## Scope

### In Scope
- Remove `✅` (installed) and `⛔` (blocked) emoji indicators
- Add `[ ] select all` / `[x] unselect all` as first list row
- Show blocked agents as `Name (BlockReason)` parenthetical
- All agents start unchecked (remove PreChecked)
- Dynamic label: "select all" ↔ "unselect all" based on state
- Installed agents toggleable; deselection prompts uninstall
- `a` key works but removed from help bar
- Shorten help bar

### Out of Scope
- Visual theme changes (colors, borders, layout)
- Multi-page or search/filter functionality
- Non-TUI (non-interactive) output format changes
- Uninstall itself (already exists in `squad remove --uninstall`)

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `add`: TUI selection behavior changes — new select-all row, no emojis, all unchecked, dynamic label, installed toggleable

## Approach

Add `IsSelectAll bool` sentinel field to `AgentItem`. `buildAgentItemsForAdd()` prepends a sentinel select-all row with `IsSelectAll=true`. The `newModel()` skips any PreChecked logic. Space on the sentinel row calls `toggleAll()`. Blocked agents render `Name (BlockReason)` instead of `Name  ⛔ BlockReason`. Installed agents lose the `IsInstalled` guard — all agents toggleable.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modified | AgentItem struct, newModel, renderAgentRow, handleSpecialKey, View, toggleAll, selectedIDs |
| `internal/tui/model_test.go` | Modified | All tests: new initial state, select-all, no emojis, dynamic label |
| `internal/cli/add.go` | Modified | buildAgentItemsForAdd, runAddFlowInteractive (uninstall prompt for deselected) |
| `internal/cli/add_test.go` | Modified | TUI mock expectations |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Cursor wrap includes select-all row (index 0) | Low | Natural behavior — part of normal navigation |
| Uninstall prompt in add flow surprises users | Low | Prompt explains action; "y/N" default is no |
| Existing tests break | High | Update all test assertions for new behavior |

## Rollback Plan

Revert `AgentItem` struct changes, restore `PreChecked`/`IsInstalled` fields, restore emoji rendering in `renderAgentRow()`, restore help bar in `View()`.

## Dependencies

None.

## Success Criteria

- [ ] TUI renders with select-all row as first element
- [ ] No emoji characters in TUI output
- [ ] Blocked agents show parenthetical reason instead of ⛔
- [ ] All agents start unchecked
- [ ] Dynamic label changes on toggle
- [ ] `a` key still toggles all but absent from help bar
- [ ] Installed agents toggleable
- [ ] All existing tests updated and pass
