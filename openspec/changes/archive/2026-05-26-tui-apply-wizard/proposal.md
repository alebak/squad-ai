# Proposal: TUI Apply + Dialog + Uninstall Wizard

## Intent

Three UX improvements to the `squad add` TUI: rename "Done" to "Apply" with separator and `a` shortcut, show a "No changes" dialog when nothing is selected, and replace the stdin uninstall prompt with an inline wizard. All three changes modify the same files (`model.go`, `model_test.go`, `add.go`) and ship together.

## Scope

### In Scope
- #37: Rename "✓ Done" → "Apply", add separator line, `a` key submits, header shows version in mauve
- #38: "No changes to apply" dialog when Apply pressed with empty selection
- #39: Inline uninstall wizard replacing stdin 3-option prompt for deselected installed agents
- Update `RunSelection` return signature to carry wizard choices
- Update `runAddFlowInteractive` to consume wizard choices and execute uninstalls
- Tests for all new and modified behavior

### Out of Scope
- Non-interactive (`--agents`) flow changes
- Actual uninstall execution code (already exists in `installer` package)
- TUI rendering of installation progress

## Capabilities

### New Capabilities
- `tui-apply-wizard`: TUI Apply button, no-changes dialog, and inline uninstall wizard

### Modified Capabilities
- `add`: The TUI selection spec changes — Done renamed to Apply, `a` shortcut, separator, dialog, and wizard replace stdin prompts
- `uninstall`: The uninstall specification changes — wizard replaces the stdin 3-option prompt flow

## Approach

Extend the existing `model` struct with `showDialog` and `wizard` fields. The Done item becomes Apply with separator. `handleRuneKey` gets `a` for submit. When Apply is pressed, the flow checks for deselected installed agents and enters wizard mode. After wizard completes, `RunSelection` returns wizard choices alongside selected IDs. `runAddFlowInteractive` reads the wizard choices and calls existing `UninstallAgent`/`UninstallConfig`. Colors switch to Catppuccin Mocha palette.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modified | New fields, Apply item, separator, wizard view, dialog view, Catppuccin colors |
| `internal/tui/model_test.go` | Modified | Tests for Apply, dialog, wizard, `a` shortcut, separator rendering |
| `internal/cli/add.go` | Modified | `RunSelection` type updated, wizard choices consumed in `runAddFlowInteractive` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Wizard state transitions break nav | Low | Table-driven tests for every key in wizard mode |
| Backward compat: `RunSelection` signature change | Medium | Update all callers in add.go and tests at once |
| Color change breaks existing tests | Low | Update assertion strings in test expectations |

## Rollback Plan

Revert all three changes atomically — they touch the same files. No data migration needed.

## Dependencies

- Existing `installer.UninstallAgent` and `installer.UninstallConfig` for execution
- Catppuccin Mocha colors are lipgloss constants only

## Success Criteria

- [ ] Apply item renders with separator, `a` key submits
- [ ] "No changes" dialog appears when Apply pressed without selection
- [ ] Inline wizard replaces stdin 3-option prompt for deselected installed agents
- [ ] Wizard collects choices per agent, returns them via `RunSelection`
- [ ] `runAddFlowInteractive` executes uninstalls based on wizard choices
- [ ] All existing tests pass (with updated assertions)
- [ ] New tests cover Apply, dialog, wizard, and `a` shortcut
