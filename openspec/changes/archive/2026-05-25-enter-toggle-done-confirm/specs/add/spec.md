# Delta for add (TUI Interactive Selection)

## MODIFIED Requirements

### Requirement: TUI Interactive Selection

`squad add` in interactive mode (TTY) SHALL launch a Bubbletea TUI for agent selection.

The TUI SHALL display a list of all registry agents with checkboxes. The first row SHALL be a select-all sentinel using the same ◉/○ checkbox style as agent rows, with dynamic text "select all" when no compatible agents are checked or "unselect all" when all compatible agents are checked. A blank line SHALL separate the select-all row from the first agent row.

The last row SHALL be a "✓ Done" sentinel item. The Done row SHALL NOT render a checkbox. The Done row SHALL be rendered with bold styling.

Agent rows SHALL use ◉ (checked) / ○ (unchecked) as checkbox symbols. Agents with unmet runtime dependencies SHALL render their name followed by `(BlockReason)` in parentheses, styled as blocked (faint/gray). The title SHALL NOT contain emoji characters.

Installed agents with met runtime dependencies SHALL be pre-checked on TUI load. Blocked agents SHALL NOT be pre-checked regardless of install state.

Enter SHALL toggle the currently highlighted agent (same behavior as Space), except on the Done item where Enter SHALL confirm the selection and exit the TUI. Space SHALL also confirm on the Done item.

The done sentinel SHALL be excluded from `selectedIDs`, `toggleAll`, and SHALL NOT appear in the checkbox count.

(Previously: Enter confirmed the entire selection and exited the TUI from any item. No Done sentinel existed.)

#### Scenario: Interactive TUI with select-all and Done rows

- GIVEN a registry with 3 agents and a TTY
- WHEN the user runs `squad add`
- THEN the TUI shows a select-all row as the first item
- AND a blank line follows the select-all row
- AND a "✓ Done" row appears as the last item
- AND the Done row has no checkbox

#### Scenario: Enter toggles agent selection

- GIVEN the TUI is open with cursor on a non-blocked agent
- WHEN the user presses Enter
- THEN the agent checkbox toggles (check → uncheck or vice versa)

#### Scenario: Enter on Done confirms selection

- GIVEN the TUI is open with cursor on the "✓ Done" row
- WHEN the user presses Enter
- THEN the system returns the selected agent IDs

#### Scenario: Space on Done confirms selection

- GIVEN the TUI is open with cursor on the "✓ Done" row
- WHEN the user presses Space
- THEN the system returns the selected agent IDs

#### Scenario: Done excluded from selectedIDs

- GIVEN the TUI has agents checked
- WHEN `selectedIDs` is called
- THEN the Done sentinel SHALL NOT appear in the returned IDs

#### Scenario: Done excluded from toggleAll

- GIVEN the TUI is open with all compatible agents checked
- WHEN `toggleAll` is called
- THEN the Done sentinel SHALL remain unchecked and unchanged
