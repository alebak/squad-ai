# Tasks: TUI Apply + Dialog + Uninstall Wizard

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~400-500 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Delivery strategy | single-pr |
| Decision needed before apply | No |

```
Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Medium
```

### Suggested Work Units

Single PR — all changes are in the same 3 files and cannot be split independently.

## Phase 1: Model extensions

- [x] 1.1 Add `showDialog string` and `wizard *wizardState` fields to `model` struct
- [x] 1.2 Define `wizardState` struct with `step, total int`, `indices, choices, cursor int`
- [x] 1.3 Update Catppuccin Mocha color styles (mauve, green, blue, subdued/overlay, yellow, red, surface, white)
- [x] 1.4 Add `submitSelection()` helper method on model

## Phase 2: Apply + separator + `a` shortcut (#37)

- [x] 2.1 Rename Done sentinel "✓ Done" → "Apply" in `newModel()`, update submit logic
- [x] 2.2 Render separator line above Apply item in `renderAgentRow()`
- [x] 2.3 Update header in `View()` to include "Squad AI (version 0.15.0)" with mauve + subdued styling
- [x] 2.4 In `handleRuneKey`: change `a` from toggleAll to submitSelection
- [x] 2.5 In `handleSpecialKey`: Enter on Apply triggers submit, not toggle

## Phase 3: No changes dialog (#38)

- [x] 3.1 In `submitSelection`: when nothing selected and no deselected installed, set `showDialog = "no-changes"`
- [x] 3.2 Render dialog overlay in `View()` when showDialog != ""
- [x] 3.3 In `handleKeyMsg`: when dialog is active, only Enter is handled (dismiss)

## Phase 4: Inline uninstall wizard (#39)

- [x] 4.1 In `submitSelection`: when installed agents are deselected, init wizardState
- [x] 4.2 Render wizard view when `wizard != nil` (replaces agent list)
- [x] 4.3 Wizard key handling: ↑↓/jk for radio, enter to confirm, n next, b back, q cancel
- [x] 4.4 After last wizard step completes (all agents done), set wizard = nil, show Apply again
- [x] 4.5 On wizard completion + final Apply, populate `wizardOut` map from `wizard.choices`

## Phase 5: Update `RunSelection` return

- [x] 5.1 Change `RunSelection` signature to `([]string, map[string]int, error)`
- [x] 5.2 Update return to pass wizardOut alongside selectedIDs
- [x] 5.3 Update `addHandler.runSelection` type to match new signature

## Phase 6: Wire wizard choices in add.go

- [x] 6.1 In `runAddFlowInteractive`: consume wizard choices map after TUI returns
- [x] 6.2 For each wizard choice: call UninstallAgent (0), UninstallAgent+UninstallConfig (1), or skip (2)
- [x] 6.3 Remove `uninstallChoiceFn`, `confirmFn`, `defaultUninstallChoiceFn`, `uninstallChoice` type
- [x] 6.4 Remove bulk confirmation prompt logic (now handled by wizard)

## Phase 7: Tests for all changes

- [x] 7.1 Add test: `TestModel_ApplyItemRenders` — separator line appears, shows "Apply" not "✓ Done"
- [x] 7.2 Add test: `TestModel_AKeySubmits` — `a` key triggers submit, not toggleAll
- [x] 7.3 Add test: `TestModel_NoChangesDialog` — dialog appears when Apply pressed with empty selection
- [x] 7.4 Add test: `TestModel_NoChangesDismiss` — Enter dismisses dialog
- [x] 7.5 Add test: `TestModel_WizardInit` — wizard initializes when installed agents deselected
- [x] 7.6 Add tests: `TestModel_WizardNavigationJK` + `TestModel_WizardEnterSelects` + `TestModel_WizardNextBack` + `TestModel_WizardCancel` — j/k/enter/n/b/q in wizard mode
- [x] 7.7 Add tests: `TestModel_WizardCompletesAndSubmits` + `TestModel_SelectedIDsAfterWizard` — choices map populated correctly
- [x] 7.8 Add test: `TestModel_HeaderShowsVersion` — header includes mauve title + subdued version
- [x] 7.9 Update existing tests: rename Done→Apply, a-key behavior change, color assertions

## Phase 8: Verify and finalize

- [x] 8.1 Run `go test ./internal/tui/...` — all pass
- [x] 8.2 Run `go test ./internal/cli/...` — all pass
- [x] 8.3 Run `go build ./cmd/squad` — compiles
