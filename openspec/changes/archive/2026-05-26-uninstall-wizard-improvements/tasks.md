# Tasks: Uninstall Wizard Improvements

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~210 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

## Phase 1: Core Logic — wizardState and handleWizardKey

- [x] 1.1 Add `showingSummary bool` field to `wizardState` struct
- [x] 1.2 Define cursor position constants: `wizardBackIdx = 3`, `wizardNextIdx = 4`
- [x] 1.3 Update `handleWizardKey`: extend cursor wrapping from 3 to 5 positions
- [x] 1.4 Update `handleWizardKey`: Enter on radio (cursor 0-2) stores choice AND auto-advances
- [x] 1.5 Update `handleWizardKey`: Enter on cursor 3 triggers Back navigation
- [x] 1.6 Update `handleWizardKey`: Enter on cursor 4 triggers Next navigation (if choice confirmed)
- [x] 1.7 Update advance logic: after last step, set `showingSummary = true` instead of `completeWizard()`
- [x] 1.8 Update summary Apply to call `completeWizard()`; summary Back sets `showingSummary = false, step = last`

## Phase 2: Rendering — Wizard and Summary Views

- [x] 2.1 Update `renderWizardView()`: add navigation buttons at bottom (`[ ◄ Back ]` / `[ Next ► ]`)
- [x] 2.2 Style buttons with proper cursor highlighting and disabled state for Back (step 0) / Next (no choice)
- [x] 2.3 Create `renderSummaryView()`: table of agent → action text, separator, Apply/Back buttons, help text
- [x] 2.4 Update `View()`/`renderWizardView()`: dispatch to summary when `showingSummary` is true

## Phase 3: Testing

- [x] 3.1 Test cursor wraps through 5 positions
- [x] 3.2 Test Enter on radio auto-advances
- [x] 3.3 Test Enter on Back button goes to previous step
- [x] 3.4 Test Enter on Next button advances (with and without confirmed choice)
- [x] 3.5 Test summary renders with all agents and action text
- [x] 3.6 Test Apply on summary completes wizard (wizard = nil, wizardOut populated)
- [x] 3.7 Test Back from summary returns to last step (showingSummary = false)
- [x] 3.8 Test n/b shortcuts still work in step view
- [x] 3.9 Verify all existing wizard tests still pass
