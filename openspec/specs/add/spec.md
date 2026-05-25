# Spec: `squad add` Command

## References

- PRD §8 (Comandos del CLI) — `squad add` table entry
- PRD §7.4 (Modo interactivo manual) — add usage context

## Requirements

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

### Requirement: 3-Option Uninstall Prompt

When a user deselects an installed agent in the interactive TUI flow (`runAddFlowInteractive`), the system SHALL display a 3-option prompt instead of a binary yes/no confirmation.

The options SHALL be:
1. **"Uninstall app only"** — calls the existing `UninstallAgent` to remove the binary
2. **"Uninstall app + config data"** — calls `UninstallAgent` AND `UninstallConfig` to remove binary + config directories
3. **"Cancel"** — does nothing, and the agent SHALL remain selected in the TUI result

The prompt SHALL accept numeric input (1, 2, 3). Any invalid input SHALL re-prompt.

#### Scenario: User chooses "Uninstall app only"

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 1 at the prompt
- THEN `UninstallAgent` is called for that agent
- AND `UninstallConfig` is NOT called

#### Scenario: User chooses "Uninstall app + config data"

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 2 at the prompt
- THEN `UninstallAgent` is called for that agent
- AND `UninstallConfig` is called for that agent

#### Scenario: User chooses "Cancel"

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 3 at the prompt
- THEN neither `UninstallAgent` nor `UninstallConfig` is called
- AND the agent remains selected

#### Scenario: Invalid input re-prompts

- GIVEN the 3-option prompt is displayed
- WHEN the user enters "4" or "abc"
- THEN the prompt SHALL display an error message
- AND re-display the 3 options

### Requirement: Registry Fetch Failure

If the registry cannot be fetched and no cache is available, `squad add` SHALL error with a message explaining the failure.

#### Scenario: Registry fetch failure

- GIVEN no internet and no cache
- WHEN the user runs `squad add`
- THEN the command errors explaining the network failure
