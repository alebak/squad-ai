# Delta for add

## MODIFIED Requirements

### Requirement: 3-Option Uninstall Prompt

When a user deselects an installed agent in the interactive TUI flow (`runAddFlowInteractive`), the system SHALL display a 3-option prompt instead of a binary yes/no confirmation.

The options SHALL be:
1. **"Uninstall app only"** — calls the existing `UninstallAgent` to remove the binary
2. **"Uninstall app + config data"** — calls `UninstallAgent` AND `UninstallConfig` to remove binary + config directories
3. **"Cancel"** — does NOT call any uninstall function and SHALL restart the TUI selection flow with the cancelled agent restored to its pre-checked state

The prompt SHALL accept numeric input (1, 2, 3). Any invalid input SHALL re-prompt.

When Cancel is chosen for ANY deselected installed agent, the system SHALL:
- NOT execute any uninstall for that agent
- Rebuild the agent selection items using `buildAgentItemsForAdd` with the updated installed map
- Re-launch the Bubbletea TUI for re-selection
- If agents were already uninstalled (via options 1 or 2) before a Cancel choice, they SHALL NOT appear pre-checked in the re-launched TUI

If NO installed agents were deselected, the system SHALL skip the uninstall prompt entirely and proceed directly to installation.

(Previously: Cancel kept the agent selected and skipped TUI restart; flow proceeded to installation)

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
