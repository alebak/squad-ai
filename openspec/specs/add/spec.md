# Spec: `squad add` Command

## References

- PRD §8 (Comandos del CLI) — `squad add` table entry
- PRD §7.4 (Modo interactivo manual) — add usage context

## Requirements

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

### Requirement: 3-Option Uninstall Prompt

When a user deselects an installed agent in the interactive TUI flow (`runAddFlowInteractive`), the system SHALL handle the deselection as follows:

- If **only one** installed agent was deselected, the system SHALL display the existing 3-option prompt (Uninstall app only / Uninstall app + config data / Cancel).
- If **multiple** installed agents were deselected simultaneously, the system SHALL display a SINGLE combined confirmation prompt listing ALL deselected installed agent names.

The combined prompt SHALL use `confirmFn` with a message like: `"Some selected agents are already installed: Claude Code, OpenCode. Uninstall them as well? [y/N]"`.

When the combined prompt is confirmed (y/yes), the system SHALL uninstall ALL listed agents using `UninstallAgent` (app-only). When declined, the system SHALL keep all agents in the `installed` map and SHALL restart the TUI selection loop.

The per-agent 3-option prompt SHALL remain unchanged with options:
1. **"Uninstall app only"** — calls `UninstallAgent`
2. **"Uninstall app + config data"** — calls `UninstallAgent` AND `UninstallConfig`
3. **"Cancel"** — restarts the TUI selection flow with the cancelled agent restored

When Cancel is chosen for ANY deselected installed agent (in per-agent mode), the system SHALL:
- NOT execute any uninstall for that agent
- Rebuild the agent selection items using `buildAgentItemsForAdd` with the updated installed map
- Re-launch the Bubbletea TUI for re-selection
- If agents were already uninstalled (via options 1 or 2) before a Cancel choice, they SHALL NOT appear pre-checked in the re-launched TUI

If NO installed agents were deselected, the system SHALL skip the uninstall prompt entirely and proceed directly to installation.

(Previously: Each deselected installed agent always prompted individually with the 3-option menu; no bulk combined prompt existed)

#### Scenario: Multiple installed agents deselected, user confirms bulk uninstall

- GIVEN two installed agents (Claude Code, OpenCode) are both deselected in the TUI
- WHEN the user submits the selection
- THEN a single combined prompt SHALL appear listing both names
- WHEN the user types "y"
- THEN `UninstallAgent` SHALL be called for both agents
- AND the flow proceeds to installation

#### Scenario: Multiple installed agents deselected, user declines bulk uninstall

- GIVEN two installed agents (Claude Code, OpenCode) are both deselected in the TUI
- WHEN the user submits the selection
- THEN a single combined prompt SHALL appear listing both names
- WHEN the user types "n"
- THEN no uninstall SHALL be called for either agent
- AND the TUI SHALL re-launch with both agents pre-checked

#### Scenario: Single installed agent deselected, existing 3-option prompt appears

- GIVEN one installed agent (Claude Code) is deselected in the TUI
- WHEN the user submits the selection
- THEN the existing 3-option per-agent prompt SHALL appear
- AND the flow proceeds according to the existing behavior

#### Scenario: User chooses Cancel, TUI re-launches

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 3 at the prompt
- THEN neither `UninstallAgent` nor `UninstallConfig` is called
- AND the TUI SHALL re-launch with the cancelled agent pre-checked

#### Scenario: User cancels one, confirms another, TUI re-launches correctly

- GIVEN two installed agents (A and B) are both deselected in the TUI
- WHEN the user selects option 1 for agent A and option 3 for agent B
- THEN agent A is uninstalled
- AND agent B is NOT uninstalled
- AND the TUI SHALL re-launch with agent B pre-checked and agent A NOT pre-checked

#### Scenario: User chooses Cancel, then confirms on re-launch

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 3 (Cancel) at the prompt
- AND the TUI re-launches
- AND the user deselects the same agent again
- AND the user selects option 1 (Uninstall app only) at the prompt
- THEN the agent IS uninstalled
- AND the flow proceeds to installation

#### Scenario: No cancels, flow proceeds directly

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 1 or 2 at the prompt
- THEN the agent IS uninstalled
- AND the flow proceeds to installation without re-launching the TUI

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
