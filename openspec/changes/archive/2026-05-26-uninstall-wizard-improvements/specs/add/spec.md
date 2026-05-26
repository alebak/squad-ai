# Delta for `add` Spec

## MODIFIED Requirements

### Requirement: Uninstall Wizard via TUI

When a user deselects an installed agent (PreChecked → unchecked) and presses Apply, the TUI SHALL enter an inline uninstall wizard mode instead of submitting immediately.

The wizard SHALL display a step-by-step interface for each deselected installed agent:
- Title: "Step X of Y — Agent Name"
- Text: "This agent is currently installed. Choose an action:"
- Radio buttons: "Uninstall app only", "Uninstall app + config data", "Keep installed (skip)"
- At the bottom, the wizard SHALL render navigation buttons: `[ ◄ Back ]` and `[ Next ► ]`

The cursor SHALL navigate across radio options AND the navigation buttons. The cursor SHALL have 5 positions: 3 radio options, then Back button, then Next button. Navigation wraps through all 5 positions.

Navigation inside the wizard:
- `↑`/`↓` and `j`/`k` SHALL navigate across all 5 cursor positions (3 radio options + Back + Next), wrapping at each end
- `enter` SHALL:
  - On a radio option: confirm the selection AND auto-advance to the next step (or summary if last step)
  - On the Back button: go to the previous step
  - On the Next button: advance to the next step (only if current step has a confirmed choice)
- `n` SHALL advance to the next step (only when current step has a confirmed choice)
- `b` SHALL return to the previous step
- `q` SHALL cancel the wizard and return to the agent list view

When Enter is pressed on a radio option, the choice SHALL be recorded AND the wizard SHALL automatically advance to the next step without requiring the user to press Next.

After the last step, instead of completing immediately, the wizard SHALL enter a summary view (`showingSummary = true`). The summary view SHALL display:
- Title: "Step X of Y — Summary"
- A table of each agent and its chosen action (e.g., "Claude Code → Uninstall app only")
- A separator line
- An "Apply" button and a "Back" button
- Help text: "↑↓ navigate • enter select • q quit"

On the summary view, Enter on Apply SHALL complete the wizard (same as `completeWizard()`), returning the user to the agent list view with Apply visible. Enter on Back SHALL return to the last wizard step to modify choices. `n`/`b` SHALL NOT work in summary view.

After wizard completion (summary Apply pressed), the agent list view SHALL re-appear with the Apply item at the bottom. The user SHALL press Apply again to submit both the selected agent IDs and the wizard choices.

(Previously: No visible buttons, no auto-advance — user had to press `n` to advance after selecting with Enter. No summary view — wizard completed immediately after the last agent step.)

Choices collected by the wizard:
- 0 (Uninstall app only) → `UninstallAgent` is called
- 1 (Uninstall app + config data) → `UninstallAgent` AND `UninstallConfig` are called
- 2 (Keep installed / skip) → no uninstall is performed

When the Apply is pressed after wizard completion, the TUI SHALL return the wizard choices alongside the selected IDs.

The `runAddFlowInteractive` function SHALL process wizard choices by calling `UninstallAgent` and/or `UninstallConfig` for each agent based on the choice value. After processing uninstalls, the function SHALL rebuild agent items with the updated installed state and re-launch the TUI.

If no uninstalls were performed (all choices were "skip"), the flow SHALL proceed directly to installation.

#### Scenario: Next/Back buttons render at bottom (Added)

- GIVEN the wizard is open
- THEN the view contains `[ ◄ Back ]` and `[ Next ► ]` at the bottom
- AND the cursor can navigate to both buttons

#### Scenario: Cursor navigates to 5 positions (Added)

- GIVEN the wizard is open
- WHEN the user presses `j` / `↓`
- THEN the cursor cycles through 5 positions: radio 0 → radio 1 → radio 2 → Back → Next → radio 0

#### Scenario: Enter on radio auto-advances (Added)

- GIVEN the wizard is open with cursor on a radio option
- WHEN the user presses Enter
- THEN the choice is stored
- AND the wizard advances to the next step (or shows summary)

#### Scenario: Enter on Next button advances (Added)

- GIVEN the wizard is open with cursor on the Next button
- WHEN the user presses Enter
- THEN the wizard advances to the next step (only if current step has a confirmed choice)

#### Scenario: Enter on Back button goes back (Added)

- GIVEN the wizard is open with cursor on the Back button (step > 0)
- WHEN the user presses Enter
- THEN the wizard goes to the previous step

#### Scenario: Summary view shows after last step (Added)

- GIVEN the wizard has completed all agent steps
- THEN the wizard shows a summary view with all agents and their choices
- AND Apply and Back buttons are shown

#### Scenario: Apply on summary exits wizard (Added)

- GIVEN the summary view is showing
- WHEN the user selects Apply
- THEN the wizard completes
- AND the agent list view re-appears with Apply item

#### Scenario: Back from summary returns to last step (Added)

- GIVEN the summary view is showing
- WHEN the user selects Back
- THEN the wizard returns to the last agent step

#### Scenario: Wizard opens for deselected installed agents (Unchanged)

- GIVEN an installed agent is deselected in the TUI
- WHEN the user presses Enter on Apply
- THEN the agent list is replaced by the wizard view
- AND the wizard shows "Step 1 of 1 — Agent Name"

#### Scenario: Wizard navigation with j/k (Updated)

- GIVEN the wizard is open with 5 cursor positions (3 radios + 2 buttons)
- WHEN the user presses `j`
- THEN the cursor moves down (wraps at end position 4 → position 0)
- WHEN the user presses `k`
- THEN the cursor moves up (wraps at start position 0 → position 4)

#### Scenario: Enter selects option (Updated)

- GIVEN the wizard is open with cursor on a radio option
- WHEN the user presses Enter
- THEN the current option is stored in `choices`
- AND the option shows as selected
- AND the wizard auto-advances to the next step

#### Scenario: q cancels wizard (Unchanged)

- GIVEN the wizard is open
- WHEN the user presses `q`
- THEN `wizard` is set to nil
- AND the agent list view is restored

#### Scenario: Wizard completes and returns choices (Updated)

- GIVEN the wizard completes all steps through the summary Apply
- WHEN Apply is pressed again
- THEN `RunSelection` returns the wizard choices alongside selected IDs

#### Scenario: Uninstall after wizard choice app-only (Unchanged)

- GIVEN the wizard completes with choice 0 for an agent
- WHEN the Apply action processes the wizard choices
- THEN `UninstallAgent` is called with that agent

#### Scenario: Uninstall after wizard choice app+config (Unchanged)

- GIVEN the wizard completes with choice 1 for an agent
- WHEN the Apply action processes the wizard choices
- THEN `UninstallAgent` AND `UninstallConfig` are called

#### Scenario: Skip keeps agent installed (Unchanged)

- GIVEN the wizard completes with choice 2 for an agent
- WHEN the Apply action processes the wizard choices
- THEN neither `UninstallAgent` nor `UninstallConfig` is called
- AND the agent stays in the installed map
