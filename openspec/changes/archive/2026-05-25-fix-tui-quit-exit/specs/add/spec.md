# Delta for add

## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: 3-Option Uninstall Prompt

When a user deselects an installed agent **after confirming a selection with Enter** (i.e., `RunSelection` returned a non-nil value, including an empty slice), the system SHALL handle the deselection as follows:

(Previously: applied to any RunSelection return value, including nil/quit)

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

#### Scenario: nil (quit) does not trigger uninstall prompt

- GIVEN the TUI is open with an installed agent deselected
- WHEN the user presses `q` instead of `Enter`
- THEN the uninstall prompt SHALL NOT appear
- AND the command exits cleanly
