# Delta for Add Command

## ADDED Requirements

### Requirement: 3-Option Uninstall Prompt

When a user deselects an installed agent in the interactive TUI flow (`runAddFlowInteractive`), the system SHALL display a 3-option prompt instead of a binary yes/no confirmation.

The options SHALL be:
1. **"Uninstall app only"** — calls the existing `UninstallAgent` to remove the binary
2. **"Uninstall app + config data"** — calls `UninstallAgent` AND `UninstallConfig` to remove binary + config directories
3. **"Cancel"** — does nothing, and the agent SHALL remain selected (checked) in the TUI result

The prompt SHALL accept numeric input (1, 2, 3) or arrow keys + enter. Any invalid input SHALL re-prompt.

#### Scenario: User chooses "Uninstall app only"

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 1 at the prompt
- THEN `UninstallAgent` is called for that agent
- AND `UninstallConfig` is NOT called
- AND the agent is removed from `selectedAgents`

#### Scenario: User chooses "Uninstall app + config data"

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 2 at the prompt
- THEN `UninstallAgent` is called for that agent
- AND `UninstallConfig` is called for that agent
- AND the agent is removed from `selectedAgents`

#### Scenario: User chooses "Cancel"

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 3 at the prompt
- THEN neither `UninstallAgent` nor `UninstallConfig` is called
- AND the agent SHALL NOT be removed from `selectedAgents`
- AND the agent remains installed on the system

#### Scenario: Invalid input re-prompts

- GIVEN the 3-option prompt is displayed
- WHEN the user enters "4" or "abc"
- THEN the prompt SHALL display an error message
- AND re-display the 3 options

#### Scenario: Uninstall failure is reported

- GIVEN the user chooses option 1 or 2
- WHEN `UninstallAgent` returns an error
- THEN the error SHALL be printed as a warning
- AND the flow continues to the next deselected agent (no crash)
