# Proposal: Uninstall Wizard Improvements

## Intent

Improve the uninstall wizard UX with visible Next/Back buttons, auto-advance on selection, and a summary view before final apply. Currently users must memorize `n`/`b` shortcuts and can't review their choices before confirming.

## Scope

### In Scope

- Visible `[ ◄ Back ]` / `[ Next ► ]` buttons at the bottom of each wizard step
- Auto-advance: pressing Enter on a radio option stores the choice and advances automatically
- Summary view after the last step showing all agents and their chosen actions, with Apply and Back buttons
- `n`/`b` shortcuts continue to work alongside buttons
- Tests for buttons, auto-advance, summary rendering, Apply workflow

### Out of Scope

- Changing the radio option values (0/1/2) — same choices
- Modifying the execution of uninstall choices in `runAddFlowInteractive`
- Non-wizard views (agent list, dialogs)
- Styling changes outside wizard/summary

## Capabilities

### New Capabilities

- None

### Modified Capabilities

- `add`: Wizard behavior changes (buttons, auto-advance, summary view)

## Approach

Extend `wizardState` with `showingSummary bool`. Extend cursor range from 3 (radio options) to 5 (3 radios + Back button + Next button). On Enter at a radio position, store choice and auto-advance. On Enter at Back/Next position, trigger the action. After last step, set `showingSummary = true` instead of completing immediately. Summary renders a table of agent → choice with Apply and Back buttons.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modified | wizardState new field, cursor range extension, auto-advance, summary rendering, button rendering |
| `internal/tui/model_test.go` | Modified | New tests for buttons, auto-advance, summary, Apply from summary |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Auto-advance skips review | Low | Summary screen provides review opportunity |
| Cursor extending to 5 positions breaks nav | Low | Existing j/k + arrow key logic naturally extends |
| Summary table width with long names | Low | Consistent with existing fixed-width rendering |

## Rollback Plan

Revert changes to `model.go` and `model_test.go` — single file change boundary.

## Dependencies

- Existing `wizardState` struct and `handleWizardKey` method

## Success Criteria

- [ ] Next/Back buttons render at bottom of wizard step with proper styling
- [ ] Cursor navigates to buttons (j/k wrap includes positions 3=Back, 4=Next)
- [ ] Enter on radio option stores choice and auto-advances
- [ ] Summary view shows all agent choices with agent name and action text
- [ ] Apply on summary exits wizard and submits; Back returns to last step
- [ ] `n`/`b` keyboard shortcuts still work
- [ ] All existing tests pass
