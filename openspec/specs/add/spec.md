# Spec: `squad add` Command

## References

- PRD §8 (Comandos del CLI) — `squad add` table entry
- PRD §7.4 (Modo interactivo manual) — add usage context

## Requirements

### Requirement: TUI Interactive Selection

`squad add` in interactive mode (TTY) SHALL launch a Bubbletea TUI for agent selection.

The TUI SHALL display a list of all registry agents with checkboxes. The first row SHALL be a select-all sentinel using the same ◉/○ checkbox style as agent rows, with dynamic text "select all" when no compatible agents are checked or "unselect all" when all compatible agents are checked. A blank line SHALL separate the select-all row from the first agent row.

The last row SHALL be a "✓ Done" sentinel item. The Done row SHALL NOT render a checkbox. The Done row SHALL be rendered with bold styling.

Agent rows SHALL use ◉ (checked) / ○ (unchecked) as checkbox symbols. Agents with unmet runtime dependencies SHALL render their name followed by `(BlockReason)` in parentheses, styled as blocked (faint/gray). The title SHALL NOT contain emoji characters.

Installed agents with met runtime dependencies SHALL be pre-checked on TUI load. Blocked agents SHALL NOT be pre-checked regardless of install state.

Enter SHALL toggle the currently highlighted agent (same behavior as Space), except on the Done item where Enter SHALL confirm the selection and exit the TUI. Space SHALL also confirm on the Done item.

The done sentinel SHALL be excluded from `selectedIDs`, `toggleAll`, and SHALL NOT appear in the checkbox count.

(Previously: Enter confirmed the entire selection and exited the TUI from any item. No Done sentinel existed. Before that: select-all used `[x]`/`[ ]` style, no blank line, all agents started unchecked)

#### Scenario: Interactive TUI with select-all and Done rows

- GIVEN a registry with 3 agents and a TTY
- WHEN the user runs `squad add`
- THEN the TUI shows a select-all row as the first item
- AND a blank line follows the select-all row
- AND a "✓ Done" row appears as the last item
- AND the Done row has no checkbox

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

### Requirement: TUI Quit Clean Exit

When the user quits the TUI (via `q`, `Ctrl+C`, or `Escape`), the system MUST exit the interactive flow immediately WITHOUT showing the uninstall prompt and WITHOUT installing any agents.

`RunSelection` SHALL return `nil, nil` when the user quits and `[]string{}, nil` when the user confirms an empty selection with Enter. The `runAddFlowInteractive` function MUST distinguish these cases:
- `nil` → user quit, return `cfg, nil` immediately
- `[]string{}` (non-nil empty slice) → user confirmed nothing selected, proceed to uninstall/restart logic
- non-empty slice → proceed to installation

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
- WHEN the user presses `Enter`
- THEN the system prints "No agents selected. Nothing to install."
- AND the command returns cleanly

### Requirement: 3-Option Uninstall Prompt

When a user deselects an installed agent after confirming a selection with Enter in the interactive TUI flow (`runAddFlowInteractive`), the system SHALL handle the deselection as follows:

(Previously: applied to any RunSelection return value, including nil/quit. The nil/quit case is now handled by the TUI Quit Clean Exit requirement.)

- If **only one** installed agent was deselected, the system SHALL display the existing 3-option prompt (Uninstall app only / Uninstall app + config data / Cancel).
- If **multiple** installed agents were deselected simultaneously, the system SHALL display a SINGLE combined confirmation prompt listing ALL deselected installed agent names.

The combined prompt SHALL use `confirmFn` with a message like: `"Some selected agents are already installed: Claude Code, OpenCode. Uninstall them as well? [y/N]"`.

When the combined prompt is confirmed (y/yes), the system SHALL uninstall ALL listed agents using `UninstallAgent` (app-only) and SHALL restart the TUI selection loop with the updated installed state. When declined, the system SHALL keep all agents in the `installed` map and SHALL restart the TUI selection loop.

The per-agent 3-option prompt SHALL remain unchanged with options:
1. **"Uninstall app only"** — calls `UninstallAgent`, then restarts the TUI selection loop
2. **"Uninstall app + config data"** — calls `UninstallAgent` AND `UninstallConfig`, then restarts the TUI selection loop
3. **"Cancel"** — restarts the TUI selection flow with the cancelled agent restored (no uninstall executed)

After ANY uninstall decision (option 1, 2, 3, or bulk confirm/decline), the system SHALL:
- Rebuild the agent selection items using `buildAgentItemsForAdd` with the updated installed map
- Re-launch the Bubbletea TUI for re-selection
- Agents that were uninstalled SHALL NOT appear pre-checked in the re-launched TUI
- Agents that were cancelled (not uninstalled) SHALL remain pre-checked

If NO installed agents were deselected, the system SHALL skip the uninstall prompt entirely and proceed directly to installation.

(Previously: Each deselected installed agent always prompted individually with the 3-option menu; no bulk combined prompt existed)

#### Scenario: Multiple installed agents deselected, user confirms bulk uninstall

- GIVEN two installed agents (Claude Code, OpenCode) are both deselected in the TUI
- WHEN the user submits the selection
- THEN a single combined prompt SHALL appear listing both names
- WHEN the user types "y"
- THEN `UninstallAgent` SHALL be called for both agents
- AND the TUI SHALL re-launch with both agents no longer pre-checked

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
- AND the TUI SHALL re-launch with the agent no longer pre-checked

#### Scenario: Uninstall app only/app+config restarts TUI

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 1 or 2 at the prompt
- THEN the agent IS uninstalled
- AND the TUI SHALL re-launch with the agent no longer pre-checked

#### Scenario: Invalid input re-prompts

- GIVEN the 3-option prompt is displayed
- WHEN the user enters "4" or "abc"
- THEN the prompt SHALL display an error message
- AND re-display the 3 options

#### Scenario: nil (quit) does not trigger uninstall prompt

- GIVEN the TUI is open with an installed agent deselected
- WHEN the user presses `q` instead of `Enter`
- THEN the uninstall prompt SHALL NOT appear
- AND the command exits cleanly

### Requirement: Registry Fetch Failure

If the registry cannot be fetched and no cache is available, `squad add` SHALL error with a message explaining the failure.

#### Scenario: Registry fetch failure

- GIVEN no internet and no cache
- WHEN the user runs `squad add`
- THEN the command errors explaining the network failure
