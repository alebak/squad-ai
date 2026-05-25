# Delta for add

## MODIFIED Requirements

### Requirement: TUI Interactive Selection

`squad add` in interactive mode (TTY) SHALL launch a Bubbletea TUI for agent selection.

The TUI SHALL display a list of all registry agents with checkboxes. The first row SHALL be a select-all sentinel using the same ◉/○ checkbox style as agent rows, with dynamic text "select all" when no compatible agents are checked or "unselect all" when all compatible agents are checked. A blank line SHALL separate the select-all row from the first agent row.

Agent rows SHALL use ◉ (checked) / ○ (unchecked) as checkbox symbols. Agents with unmet runtime dependencies SHALL render their name followed by `(BlockReason)` in parentheses, styled as blocked (faint/gray). The title SHALL NOT contain emoji characters.

Installed agents with met runtime dependencies SHALL be pre-checked on TUI load. Blocked agents SHALL NOT be pre-checked regardless of install state.

(Previously: select-all used `[x]`/`[ ]` style, no blank line, all agents started unchecked)

#### Scenario: Interactive TUI with select-all row

- GIVEN a registry with 3 agents and a TTY
- WHEN the user runs `squad add`
- THEN the TUI shows a select-all row as the first item using ◉/○ style
- AND a blank line follows the select-all row
- AND agents compatible with runtime dependencies start checked if installed

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

## REMOVED Requirements

### Requirement: (none removed)
