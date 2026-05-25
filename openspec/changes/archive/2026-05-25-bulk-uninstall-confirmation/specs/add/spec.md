# Delta for add

## MODIFIED Requirements

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

#### Scenario: Mixed deselect — one agent already uninstalled before decline

- GIVEN two installed agents (A and B) are both deselected in the TUI
- WHEN the user confirms the combined prompt (yes)
- AND agent A is successfully uninstalled
- AND the user then cancels at some point (agent B was NOT uninstalled)
- THEN agent A remains uninstalled
- AND the TUI re-launches with agent B pre-checked

#### Scenario: No installed agents deselected

- GIVEN the TUI selection does not deselect any installed agents
- WHEN the user submits the selection
- THEN no uninstall prompt SHALL appear
- AND the flow proceeds directly to installation
