# Verification Report: TUI Apply + Dialog + Uninstall Wizard

## Change
`tui-apply-wizard` — Three UX improvements to the `squad add` TUI.

## Mode
Standard (no Strict TDD)

## Completeness

| Phase | Tasks | Complete | Status |
|-------|-------|----------|--------|
| Model extensions | 4 | 4 | ✅ |
| Apply + separator (#37) | 5 | 5 | ✅ |
| No changes dialog (#38) | 3 | 3 | ✅ |
| Inline wizard (#39) | 5 | 5 | ✅ |
| RunSelection return | 3 | 3 | ✅ |
| Wire wizard in add.go | 4 | 4 | ✅ |
| Tests | 9 | 9 | ✅ |
| Verify | 3 | 3 | ✅ |
| **Total** | **36** | **36** | **✅ ALL COMPLETE** |

## Build Evidence

```
$ go build ./cmd/squad
→ exit code 0, no errors
```

## Test Evidence

```
$ go test ./... -short
→ all 6 packages: OK
→ internal/tui: 0.013s, all pass
→ internal/cli: 0.024s, all pass
```

## Spec Compliance Matrix

| Spec Requirement | Status | Covering Test |
|---|---|---|
| Apply item renders with separator | ✅ | TestModel_ApplyItemRenders, TestModel_SeparatorRendering |
| `a` key submits selection | ✅ | TestModel_AKeySubmits |
| Header shows "Squad AI (version 0.15.0)" | ✅ | TestModel_HeaderShowsVersion, TestModel_ViewRenders |
| No changes dialog on empty selection | ✅ | TestModel_NoChangesDialog, TestModel_NoChangesDialogViaAKey |
| Enter dismisses dialog | ✅ | TestModel_NoChangesDismiss |
| Dialog blocks other keys | ✅ | TestModel_DialogBlocksOtherKeys |
| Wizard init on deselected installed agents | ✅ | TestModel_WizardInit |
| Wizard navigation (j/k) | ✅ | TestModel_WizardNavigationJK |
| Enter selects wizard option | ✅ | TestModel_WizardEnterSelects |
| n/b wizard navigation | ✅ | TestModel_WizardNextBack |
| q cancels wizard | ✅ | TestModel_WizardCancel |
| Wizard completes and submits | ✅ | TestModel_WizardCompletesAndSubmits |
| RunSelection returns wizard choices | ✅ | (add_test.go wizard tests) |
| add.go processes wizard choices (app only) | ✅ | TestAddCommand_UninstallViaWizardAppOnly |
| add.go processes wizard choices (app+config) | ✅ | TestAddCommand_UninstallViaWizardAppAndConfig |
| add.go processes wizard choices (skip) | ✅ | TestAddCommand_UninstallViaWizardSkip |
| TUI restarts after wizard uninstall | ✅ | TestAddCommand_WizardRestartsTUIAfterUninstall |

## Design Coherence

| Decision | Status | Implementation |
|---|---|---|
| Wizard state lives inside model | ✅ | `wizard *wizardState` field on `model` |
| `RunSelection` returns wizard choices as map | ✅ | `([]string, map[string]int, error)` |
| Separator as plain string in View | ✅ | Rendered in `renderAgentRow()` for IsDone |
| Wizard uses int choice values (0/1/2) | ✅ | 0=app, 1=app+config, 2=skip |

## Issues Found

**None.** All tests pass, build succeeds, and spec requirements are covered.

## Verdict

**PASS** — Implementation matches spec requirements and design decisions. All 36 tasks complete.
