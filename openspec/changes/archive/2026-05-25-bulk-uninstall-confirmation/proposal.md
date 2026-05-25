# Proposal: Bulk Uninstall Confirmation

## Intent

When user clicks "unselect all" (toggleAll or pressing `a`), ALL installed agents get deselected at once. Currently each deselected installed agent prompts individually with the 3-option menu. User expects a single confirmation when batch-deselecting installed agents: "Some selected agents are already installed: Claude Code, OpenCode. Uninstall them as well?"

## Scope

### In Scope
- Collect ALL deselected installed agents after TUI returns
- When multiple installed agents are deselected, show a single combined y/N prompt
- If confirmed → uninstall all listed agents (app-only)
- If declined → re-add all to selection, restart TUI loop
- When only ONE installed agent is deselected, keep existing per-agent 3-option prompt
- Tests for bulk uninstall flow

### Out of Scope
- Changing the TUI itself — the fix is in the post-TUI flow in `add.go`
- Per-agent uninstall choice (app-only vs app+config) for bulk case — bulk uses app-only
- Removing the existing per-agent 3-option prompt for single-agent cases

## Capabilities

### Modified Capabilities
- `add`: The 3-Option Uninstall Prompt requirement changes — when MULTIPLE installed agents are deselected simultaneously, the system shows a combined confirmation instead of per-agent prompts

## Approach

In `runAddFlowInteractive`, restructure the deselection handling: first collect ALL installed agents not in the selection set. If >1 found, show a combined confirmation via `confirmFn`. If confirmed, uninstall each (app-only). If declined, re-add all to the `installed` map and restart the TUI loop. For single-agent deselection, keep the existing per-agent 3-option prompt.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modified | Restructure deselection handling to combine multiple prompts |
| `internal/cli/add_test.go` | Modified | Add tests for bulk uninstall flow |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Bulk confirm uninstalls agents user wanted to keep | Low | Combined prompt lists ALL agents by name; user can say no and re-select |

## Rollback Plan

Revert `internal/cli/add.go` and `internal/cli/add_test.go`. No schema or config changes.

## Dependencies

None.

## Success Criteria

- [ ] When 2+ installed agents are deselected, a single combined y/N prompt appears listing all names
- [ ] When user says yes, ALL listed agents are uninstalled (app-only)
- [ ] When user says no, ALL deselected agents are re-added to selection and TUI re-launches
- [ ] When only 1 installed agent is deselected, the existing 3-option prompt appears unchanged
- [ ] All existing tests pass
