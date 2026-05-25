# Spec Delta: `squad add` — Uninstall Restart TUI

## References

- Base spec: `openspec/specs/add/spec.md`
- Issue #28 — TUI exits after uninstall

## Change

### Requirement: 3-Option Uninstall Prompt — TUI Restart On Any Action

**What changed**: The per-agent 3-option uninstall prompt SHALL restart the TUI loop after ANY choice (Uninstall app only, Uninstall app + config, or Cancel), not only after Cancel.

**Rationale**: Previously, options 1 and 2 proceeded directly to installation after uninstalling, causing the CLI to exit. Users could not continue managing agents without re-running `squad add`. Now the TUI re-launches after every uninstall action, showing the updated installed state (uninstalled agents no longer pre-checked).

#### Updated behavior for options 1 and 2

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 1 (Uninstall app only) or option 2 (Uninstall app + config)
- THEN the agent IS uninstalled
- AND the system SHALL rebuild agent selection items with the updated installed map
- AND the system SHALL re-launch the TUI
- AND the uninstalled agent SHALL NOT appear pre-checked in the re-launched TUI

#### Updated behavior for option 3 (Cancel — unchanged)

- GIVEN an installed agent is deselected in the TUI
- WHEN the user selects option 3 (Cancel)
- THEN no uninstall SHALL be executed
- AND the system SHALL rebuild agent selection items with the unchanged installed map
- AND the system SHALL re-launch the TUI
- AND the cancelled agent SHALL remain pre-checked

#### Multiple agents deselected, per-agent mixed choices

- GIVEN two installed agents (A and B) are both deselected in the TUI
- WHEN the user selects option 1 for agent A and option 3 for agent B
- THEN agent A IS uninstalled
- AND agent B is NOT uninstalled
- AND the TUI SHALL re-launch with agent B pre-checked and agent A NOT pre-checked

#### Bulk uninstall confirm also restarts TUI

- GIVEN two installed agents are both deselected in the TUI
- WHEN the user confirms the bulk uninstall prompt ("y")
- THEN both agents SHALL be uninstalled
- AND the TUI SHALL re-launch (existing code already does this via `continue` after `delete(installed)`)
