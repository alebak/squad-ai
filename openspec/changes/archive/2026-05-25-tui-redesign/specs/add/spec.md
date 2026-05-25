# Delta for add

## MODIFIED Requirements

### Requirement: TUI Interactive Selection

`squad add` in interactive mode (TTY) SHALL launch a Bubbletea TUI for agent selection.

The TUI SHALL display a list of all registry agents with checkboxes. The first row SHALL be a select-all sentinel. Agents with unmet runtime dependencies SHALL render their name followed by `(BlockReason)` in parentheses, styled as blocked (faint/gray). The title SHALL NOT contain emoji characters.

(Previously: displayed "TUI coming soon" stub message — replaced with real interactive selection)

#### Scenario: Interactive TUI with select-all row

- GIVEN a registry with 3 agents and a TTY
- WHEN the user runs `squad add`
- THEN the TUI shows a select-all row as the first item
- AND all agents start unchecked

#### Scenario: Select-all toggles all compatible agents

- GIVEN the TUI is open with 3 agents and none checked
- WHEN the user toggles the select-all row
- THEN all compatible (non-blocked) agents become checked
- AND the select-all label changes to "unselect all"

#### Scenario: Dynamic label reflects all-checked state

- GIVEN the TUI is open and all compatible agents are checked
- THEN the select-all row shows "[x] unselect all"
- WHEN one agent is unchecked
- THEN the select-all row shows "[ ] select all"

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

## REMOVED Requirements

### Requirement: TUI Stub Output

(Reason: replaced by real interactive TUI behavior above — TUI is no longer a stub)
