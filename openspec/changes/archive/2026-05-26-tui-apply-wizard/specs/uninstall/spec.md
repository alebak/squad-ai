# Delta for `uninstall` Spec

## MODIFIED Requirements

### Requirement: UninstallAgent (applies to flow context)

When installed agents are deselected in the TUI, the uninstall choice is now collected via the inline wizard instead of the stdin 3-option prompt. The execution remains the same: `UninstallAgent` for app-only, `UninstallAgent` + `UninstallConfig` for app+config.

(Previously: the 3-option prompt was presented via stdin with `defaultUninstallChoiceFn`)

#### Scenario: Uninstall after wizard choice app-only

- GIVEN the wizard completes with choice 0 for an agent
- WHEN the Apply action processes the wizard choices
- THEN `UninstallAgent` is called with that agent

#### Scenario: Uninstall after wizard choice app+config

- GIVEN the wizard completes with choice 1 for an agent
- WHEN the Apply action processes the wizard choices
- THEN `UninstallAgent` AND `UninstallConfig` are called

#### Scenario: Skip keeps agent installed

- GIVEN the wizard completes with choice 2 for an agent
- WHEN the Apply action processes the wizard choices
- THEN neither `UninstallAgent` nor `UninstallConfig` is called
- AND the agent stays in the installed map

## REMOVED Requirements

### Requirement: 3-Option stdin Uninstall Prompt

The stdin 3-option prompt (`defaultUninstallChoiceFn` with 1/2/3 menu) for deselected installed agents is replaced by the inline TUI wizard.

(Reason: The wizard provides a better UX inside the TUI, removing the stdin interruption)

The combined bulk confirmation prompt (`confirmFn`) for multiple agents is also removed.

The following are no longer used in the interactive flow:
- `uninstallChoiceFn` field in `addHandler`
- `confirmFn` field in `addHandler` (for the uninstall confirm case)
- `defaultUninstallChoiceFn` function
- `uninstallChoice` type and its constants
- The bulk confirmation prompt logic in `runAddFlowInteractive`
