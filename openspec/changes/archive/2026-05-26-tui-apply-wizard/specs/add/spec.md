# Delta for `add` Spec

## ADDED Requirements

### Requirement: Apply Item with Separator

The "✓ Done" sentinel item SHALL be renamed to "Apply". A separator line rendered with the separator style SHALL appear immediately above the Apply item. The Apply item SHALL use the same bold styling as Done.

The `a` key SHALL trigger submission from any cursor position (same as pressing Enter on Apply).

The title header SHALL display "Squad AI (version 0.15.0)" with the mauve color (`#cba6f7`). The version number SHALL use the subdued/overlay color (`#6c7086`).

The separator SHALL be rendered as `"───────────────────────────────────────"` using the surface1 color (`#45475a`).

#### Scenario: Apply item in rendered view

- GIVEN the TUI is open
- THEN "Apply" appears in the rendered view as the last item
- AND a separator line (surface1 color) appears above Apply
- AND "✓ Done" does NOT appear in the rendered view

#### Scenario: `a` key submits selection

- GIVEN the TUI is open with agents checked
- WHEN the user presses `a`
- THEN the selection is submitted (same as pressing Enter on Apply)

#### Scenario: Apply item responds to Enter

- GIVEN the TUI is open with cursor on Apply
- WHEN the user presses Enter
- THEN the selection is submitted

#### Scenario: Header shows version

- GIVEN the TUI is open
- THEN the header contains "Squad AI (version 0.15.0)"
- AND the "Squad AI" part uses mauve color
- AND the version part uses subdued/overlay color

### Requirement: No Changes Dialog

The model SHALL have a `showDialog string` field. Empty string means no dialog is shown.

When Apply is pressed and no agents are selected AND no installed agents were deselected (empty `selectedIDs` and no deselected installed), the system SHALL set `showDialog = "no-changes"`.

When in dialog mode, the view SHALL render a dialog overlay with the text "No changes to apply." and "Press enter to continue..." inside a bordered box.

When Enter is pressed in dialog mode, the system SHALL dismiss the dialog (`showDialog = ""`).

#### Scenario: No changes dialog appears

- GIVEN the TUI is open with no agents checked
- WHEN Apply is pressed (Enter on Apply item or `a` key)
- THEN the "No changes to apply" dialog SHALL appear over the agent list

#### Scenario: Enter dismisses dialog

- GIVEN the "No changes to apply" dialog is shown
- WHEN the user presses Enter
- THEN the dialog SHALL disappear, returning to the agent list

### Requirement: Inline Uninstall Wizard

The model SHALL have a `wizard *wizardState` field. `nil` means not in wizard mode.

The `wizardState` struct SHALL contain:
- `step int` — current step index (0-based)
- `total int` — total number of wizard steps
- `agents []registry.Agent` — the agents being processed
- `choices []int` — user choices per step (0=app only, 1=app+config, 2=skip)

When Apply is pressed and installed agents were deselected, the system SHALL initialize the wizard with those agents and enter wizard mode.

The wizard view SHALL replace the agent list and display:
- Title: "Step X of Y — Agent Name"
- Text: "This agent is currently installed. Choose an action:"
- Radio buttons: "Uninstall app only", "Uninstall app + config data", "Keep installed (skip)"
- Help bar: "enter select • ↑↓ navigate • n next • b back"

Navigation inside the wizard:
- `↑`/`↓` and `j`/`k` navigate between radio options
- `enter` confirms the selection for the current step
- `n` advances to the next step
- `b` returns to the previous step
- `q` cancels the wizard, returns to the agent list (wizard = nil)

After the last step, `wizard` SHALL be set to `nil` and the Apply item SHALL reappear at the bottom.

The cursor (▸) SHALL use the blue color (`#89b4fa`), checked radio SHALL use the green color (`#a6e3a1`), and the help text SHALL use the subdued color (`#6c7086`).

#### Scenario: Wizard opens for deselected installed agents

- GIVEN an installed agent is deselected in the TUI
- WHEN Apply is pressed
- THEN the agent list is replaced by the wizard view
- AND the wizard shows "Step 1 of 1 — Agent Name"

#### Scenario: Wizard navigation with j/k

- GIVEN the wizard is open with 3 options
- WHEN the user presses `j`
- THEN the cursor moves down (wraps at end)
- WHEN the user presses `k`
- THEN the cursor moves up (wraps at start)

#### Scenario: Enter selects option

- GIVEN the wizard is open
- WHEN the user presses Enter
- THEN the current option is stored in `choices`
- AND the option shows as selected

#### Scenario: n advances to next step

- GIVEN the wizard has multiple agents
- WHEN the user presses `n`
- THEN the step index advances

#### Scenario: b goes to previous step

- GIVEN the wizard is on step 2 of 3
- WHEN the user presses `b`
- THEN the step index goes back to 1

#### Scenario: q cancels wizard

- GIVEN the wizard is open
- WHEN the user presses `q`
- THEN `wizard` is set to nil
- AND the agent list view is restored

#### Scenario: Wizard returns choices after completion

- GIVEN the wizard completes all steps
- WHEN Apply is pressed again (after wizard)
- THEN `RunSelection` returns the wizard choices alongside selected IDs

## MODIFIED Requirements

### Requirement: TUI Interactive Selection

`RunSelection` SHALL accept `[]AgentItem` and return `([]string, map[string]int, error)` — adding the wizard choices map.

(Previously: returned `([]string, error)` only)

The following return value contract SHALL apply:
- `nil, nil, nil` — user quit (q, Ctrl+C, Escape)
- `[]string{}, nil, nil` — confirmed empty selection, no deselected installed agents
- `ids, nil, nil` — confirmed with selected IDs but no deselected installed agents
- `ids, choices, nil` — confirmed with deselected installed agents, wizard choices are populated
- `nil, nil, error` — fatal TUI error

The Done item behavior is replaced: Enter on Apply (formerly Done) SHALL submit the selection. The `a` key SHALL also submit from anywhere.

(Previously: Enter confirmed the entire selection only on the Done item; `a` toggled all agents)

The text of the sentinel item SHALL be "Apply" (not "✓ Done"). A separator line SHALL render above the Apply item.

#### Scenario: Renamed to Apply from Done (Updated)

- GIVEN the TUI is open with agents
- THEN the last sentinel item shows "Apply" (not "✓ Done")
- AND a separator line appears above Apply

#### Scenario: `a` key submits (Updated)

- GIVEN the TUI is open with agents checked
- WHEN the user presses `a`
- THEN the selection is submitted (not "toggle all" — that behavior is removed)

#### Scenario: Multiple installed agents deselected with wizard (Updated)

- GIVEN two installed agents are both deselected in the TUI
- WHEN the user presses Enter on Apply
- THEN the inline wizard SHALL appear for each agent
- AND AFTER the wizard completes and Apply is pressed again
- THEN `RunSelection` SHALL return the wizard choices alongside the rest of the selection

(Previously: a combined stdin confirmation prompt appeared)
