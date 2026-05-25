# Proposal: TUI Checkbox Polish & Pre-Checked Agents

## Intent

Two related UX polish issues discovered during the `squad add` TUI:
1. The select-all row uses `[x]`/`[ ]` while agents use `◉`/`○` — inconsistent checkbox style
2. Installed agents start unchecked, requiring manual selection each visit

Fix both to improve UX consistency and reduce friction for repeat `squad add` visits.

## Scope

### In Scope
- Select-all row uses same ◉/○ checkbox style as agents
- Dynamic "select all" / "unselect all" label on select-all row (already present, preserved)
- Blank separator line between select-all row and first agent
- `PreChecked` field on `AgentItem` struct
- `newModel` checks `PreChecked` for compatible agents on init
- `buildAgentItemsForAdd` sets `PreChecked=true` for installed + compatible agents
- `detectAll` call moved earlier in `runAddFlow` to feed `buildAgentItemsForAdd`
- All existing tests updated to match new behavior

### Out of Scope
- Non-interactive (no-TTY) path — no checkboxes to style
- Vertical separator or other visual decorations beyond blank line
- Agent reordering or filtering

## Capabilities

### Modified Capabilities
- `add`: Select-all checkbox style changes (◉/○), separator line, pre-checked agents

## Approach

Two independent code changes (same files) that merge cleanly:
1. Replace `[x]`/`[ ]` with `◉`/`○` in `renderSelectAllRow`, add `\n` after select-all row in `View()`
2. Add back `PreChecked bool` field to `AgentItem`, wire it in `newModel` and `buildAgentItemsForAdd`

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modified | Add PreChecked, update newModel, renderSelectAllRow, View |
| `internal/cli/add.go` | Modified | buildAgentItemsForAdd signature + logic, detectAll moved |
| `internal/tui/model_test.go` | Modified | Update assertions for new checkbox style + PreChecked |
| `internal/cli/add_test.go` | Modified | Update test handlers for new buildAgentItemsForAdd signature |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| PreChecked interacts with select-all toggle logic | Low | toggleAll skips blocked + select-all already, PreChecked just initial state |
| Blank line breaks visual layout | Low | Single `\n` after sentinel row, standard pattern |

## Rollback Plan

Revert four files: `model.go`, `add.go`, `model_test.go`, `add_test.go`. Single-commit revert.

## Dependencies

None.

## Success Criteria

- [ ] All existing tests pass without modification to test logic (only assertion values)
- [ ] Select-all row renders `◉ select all` or `○ unselect all`
- [ ] Blank line exists between select-all and first agent
- [ ] Installed + compatible agents start checked in TUI
- [ ] Blocked agents start unchecked regardless of install state
