# Spec: `squad add` Command

## References

- PRD §8 (Comandos del CLI) — `squad add` table entry
- PRD §7.4 (Modo interactivo manual) — add usage context

## Requirements

### Requirement: TUI Interactive Selection

`squad add` in interactive mode (TTY) SHALL launch a Bubbletea TUI for agent selection.

The TUI SHALL display a list of all registry agents with checkboxes. The first row SHALL be a select-all sentinel using the same ◉/○ checkbox style as agent rows, with dynamic text "select all" when no compatible agents are checked or "unselect all" when all compatible agents are checked. A blank line SHALL separate the select-all row from the first agent row.

The last row SHALL be an "Apply" sentinel item. A separator line rendered in the surface color SHALL appear immediately above the Apply item. The Apply row SHALL NOT render a checkbox. The Apply row SHALL be rendered with bold styling.

Agent rows SHALL use ◉ (checked) / ○ (unchecked) as checkbox symbols. Agents with unmet runtime dependencies SHALL render their name followed by `(BlockReason)` in parentheses, styled as blocked (faint/gray). The title SHALL NOT contain emoji characters.

Installed agents with met runtime dependencies SHALL be pre-checked on TUI load. Blocked agents SHALL NOT be pre-checked regardless of install state.

Enter SHALL toggle the currently highlighted agent (same behavior as Space), except on the Apply item where Enter SHALL confirm the selection and exit the TUI. Space SHALL also confirm on the Apply item. The `a` key SHALL also confirm the selection from any cursor position.

The header SHALL display "Squad AI (version 0.15.0)" with the mauve color for "Squad AI" and subdued color for the version.

The apply sentinel SHALL be excluded from `selectedIDs`, `toggleAll`, and SHALL NOT appear in the checkbox count.

(Previously: The last row was "✓ Done". Enter confirmed only on Done. `a` key toggled all agents. No separator existed.)

#### Scenario: Interactive TUI with select-all and Apply rows

- GIVEN a registry with 3 agents and a TTY
- WHEN the user runs `squad add`
- THEN the TUI shows a select-all row as the first item
- AND a blank line follows the select-all row
- AND an "Apply" row appears as the last item
- AND a separator line appears above Apply
- AND the Apply row has no checkbox

#### Scenario: Select-all toggles all compatible agents

- GIVEN the TUI is open with 3 agents and none checked
- WHEN the user toggles the select-all row
- THEN all compatible (non-blocked) agents become checked
- AND the select-all label changes to "unselect all"

#### Scenario: Dynamic label reflects all-checked state

- GIVEN the TUI is open and all compatible agents are checked
- THEN the select-all row shows "○ unselect all" (or "◉ unselect all")
- WHEN one agent is unchecked
- THEN the select-all row shows "○ select all" (or "◉ select all")

#### Scenario: Select-all uses same ◉/○ style

- GIVEN the TUI renders the select-all row
- THEN the checkbox SHALL use ◉ (checked) or ○ (unchecked), NOT `[x]`/`[ ]`

#### Scenario: Enter toggles agent selection

- GIVEN the TUI is open with cursor on a non-blocked agent
- WHEN the user presses Enter
- THEN the agent checkbox toggles (check → uncheck or vice versa)

#### Scenario: Enter on Apply confirms selection

- GIVEN the TUI is open with cursor on the "Apply" row
- WHEN the user presses Enter
- THEN the system returns the selected agent IDs

#### Scenario: Space on Apply confirms selection

- GIVEN the TUI is open with cursor on the "Apply" row
- WHEN the user presses Space
- THEN the system returns the selected agent IDs

#### Scenario: `a` key confirms selection

- GIVEN the TUI is open
- WHEN the user presses `a`
- THEN the selection is confirmed (same as pressing Enter on Apply)

#### Scenario: Apply excluded from selectedIDs

- GIVEN the TUI has agents checked
- WHEN `selectedIDs` is called
- THEN the Apply sentinel SHALL NOT appear in the returned IDs

#### Scenario: Apply excluded from toggleAll

- GIVEN the TUI is open with all compatible agents checked
- WHEN `toggleAll` is called
- THEN the Apply sentinel SHALL remain unchecked and unchanged

#### Scenario: Blocked agents render without emoji

- GIVEN a blocked agent with BlockReason "requires Node.js 22+"
- WHEN the TUI renders
- THEN the agent appears as `Codex CLI (requires Node.js 22+)`
- AND no ⛔ emoji is present
- AND the agent cannot be toggled

#### Scenario: Installed agents are toggleable

- GIVEN an installed agent is shown in the TUI
- WHEN the user toggles it on
- THEN the agent becomes checked
- AND the user can toggle it off again

#### Scenario: Installed compatible agents pre-checked

- GIVEN an agent is installed and has met runtime dependencies
- WHEN the TUI loads
- THEN the agent checkbox SHALL be checked (◉)

#### Scenario: Installed blocked agents NOT pre-checked

- GIVEN an agent is installed but has unmet runtime dependencies
- WHEN the TUI loads
- THEN the agent checkbox SHALL be unchecked (○)

#### Scenario: Header shows version

- GIVEN the TUI is open
- THEN the header contains "Squad AI (version 0.15.0)"
- AND the "Squad AI" part uses mauve color
- AND the version part uses subdued/overlay color

### Requirement: TUI Quit Clean Exit

When the user quits the TUI (via `q`, `Ctrl+C`, or `Escape`), the system MUST exit the interactive flow immediately WITHOUT prompting to uninstall or install any agents.

`RunSelection` SHALL return `nil, nil, nil` when the user quits, `[]string{}, nil, nil` when the user confirms an empty selection with Enter, and `[ids...], choices, nil` when the uninstall wizard was used. The `runAddFlowInteractive` function MUST distinguish these cases:
- `nil` → user quit, return `cfg, nil` immediately
- `[]string{}` (non-nil empty slice) → user confirmed nothing selected, proceed to check wizard flow
- non-empty slice → proceed to check wizard flow or installation

(Previously: `RunSelection` returned `([]string, error)` without wizard choices.)

#### Scenario: User presses q and exits cleanly

- GIVEN the TUI is open with installed agents
- WHEN the user presses `q`
- THEN the command returns to the shell without any prompt
- AND no uninstall is called
- AND no install happens

#### Scenario: User presses Ctrl+C and exits cleanly

- GIVEN the TUI is open
- WHEN the user presses `Ctrl+C`
- THEN the command returns to the shell without any prompt

#### Scenario: User presses Escape and exits cleanly

- GIVEN the TUI is open
- WHEN the user presses `Escape`
- THEN the command returns to the shell without any prompt

#### Scenario: User confirms empty selection (Enter with nothing checked)

- GIVEN the TUI is open with no agents checked
- WHEN the user presses `Enter` on Apply
- THEN the "No changes" dialog appears
- WHEN the user presses Enter again to dismiss
- THEN the system prints "No agents selected. Nothing to install."
- AND the command returns cleanly

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

#### Scenario: Next/Back buttons render at bottom

- GIVEN the wizard is open
- THEN the view contains `[ ◄ Back ]` and `[ Next ► ]` at the bottom
- AND the cursor can navigate to both buttons

#### Scenario: Cursor navigates to 5 positions

- GIVEN the wizard is open
- WHEN the user presses `j` / `↓`
- THEN the cursor cycles through 5 positions: radio 0 → radio 1 → radio 2 → Back → Next → radio 0

#### Scenario: Enter on radio auto-advances

- GIVEN the wizard is open with cursor on a radio option
- WHEN the user presses Enter
- THEN the choice is stored
- AND the wizard advances to the next step (or shows summary)

#### Scenario: Enter on Next button advances

- GIVEN the wizard is open with cursor on the Next button
- WHEN the user presses Enter
- THEN the wizard advances to the next step (only if current step has a confirmed choice)

#### Scenario: Enter on Back button goes back

- GIVEN the wizard is open with cursor on the Back button (step > 0)
- WHEN the user presses Enter
- THEN the wizard goes to the previous step

#### Scenario: Summary view shows after last step

- GIVEN the wizard has completed all agent steps
- THEN the wizard shows a summary view with all agents and their choices
- AND Apply and Back buttons are shown

#### Scenario: Apply on summary exits wizard

- GIVEN the summary view is showing
- WHEN the user selects Apply
- THEN the wizard completes
- AND the agent list view re-appears with Apply item

#### Scenario: Back from summary returns to last step

- GIVEN the summary view is showing
- WHEN the user selects Back
- THEN the wizard returns to the last agent step

#### Scenario: Wizard opens for deselected installed agents

- GIVEN an installed agent is deselected in the TUI
- WHEN the user presses Enter on Apply
- THEN the agent list is replaced by the wizard view
- AND the wizard shows "Step 1 of 1 — Agent Name"

#### Scenario: Wizard navigation with j/k

- GIVEN the wizard is open with 5 cursor positions (3 radios + 2 buttons)
- WHEN the user presses `j`
- THEN the cursor moves down (wraps at end position 4 → position 0)
- WHEN the user presses `k`
- THEN the cursor moves up (wraps at start position 0 → position 4)

#### Scenario: Enter selects option

- GIVEN the wizard is open with cursor on a radio option
- WHEN the user presses Enter
- THEN the current option is stored in `choices`
- AND the option shows as selected
- AND the wizard auto-advances to the next step

#### Scenario: n advances to next step

- GIVEN the wizard has multiple agents and current step has a confirmed choice
- WHEN the user presses `n`
- THEN the step index advances (or shows summary)

#### Scenario: b goes to previous step

- GIVEN the wizard is on step 2 of 3
- WHEN the user presses `b`
- THEN the step index goes back to 1

#### Scenario: q cancels wizard

- GIVEN the wizard is open (step view or summary)
- WHEN the user presses `q`
- THEN `wizard` is set to nil
- AND the agent list view is restored

#### Scenario: Wizard completes and returns choices

- GIVEN the wizard completes all steps through the summary Apply
- WHEN Apply is pressed again
- THEN `RunSelection` returns the wizard choices alongside selected IDs

#### Scenario: Uninstall after wizard choice app-only

- GIVEN the wizard completes with choice 0 for an agent
- WHEN the Apply action processes the wizard choices
- THEN `UninstallAgent` is called with that agent

#### Scenario: Uninstall after wizard choice app+config

- GIVEN the wizard completes with choice 1 for an agent
- WHEN the Apply action processes the wizard choices
- THEN `UninstallAgent` AND `UninstallConfig` are called

#### Scenario: Skip keeps agent installed

- GIVEN the wizard completes with choice 2 for an agent
- WHEN the Apply action processes the wizard choices
- THEN neither `UninstallAgent` nor `UninstallConfig` is called
- AND the agent stays in the installed map

### Requirement: Registry Fetch Failure

If the registry cannot be fetched and no cache is available, `squad add` SHALL error with a message explaining the failure.

#### Scenario: Registry fetch failure

- GIVEN no internet and no cache
- WHEN the user runs `squad add`
- THEN the command errors explaining the network failure
