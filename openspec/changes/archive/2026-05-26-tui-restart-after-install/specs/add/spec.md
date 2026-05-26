# Delta Spec: TUI Interactive Selection

## Requirements

Requirements use RFC 2119 keywords.

### R-TUI-RESTART-1: TUI SHALL relaunch after successful installation

After agents are installed through the interactive TUI (`squad add`), the application SHALL return to the TUI showing the updated agent state — with the newly installed agents displayed as checked and installed (◉).

When the user quits the relaunched TUI (q, Ctrl+C, Escape), the application SHALL exit cleanly.

### R-TUI-RESTART-2: TUI SHALL relaunch after partial installation

If some agents fail to install and others succeed, the application SHALL still relaunch the TUI showing the updated state for succeeded agents. Error messages SHALL be printed to the terminal before relaunch.

### R-TUI-RESTART-3: No agents to install SHALL exit

If the user confirms an empty selection (no agents selected), the application SHALL print "No agents selected. Nothing to install." and exit — the same behavior as before.

### R-TUI-RESTART-4: Already installed agents SHALL loop back

If the user selects agents that are already installed, the application SHALL print "Selected agents are already installed." and loop back to the TUI, allowing the user to quit or make a different selection.

### R-TUI-RESTART-5: Uninstall wizard restart SHALL remain unchanged

The existing behavior after uninstall wizard completion (restart TUI to show updated state) SHALL continue to work. The uninstall restart AND the install restart SHALL coexist.

## Scenarios

### Scenario: Install one agent and see updated state

```
Given the target has no agents installed
And the TUI is launched
When the user selects "claude-code"
And confirms installation
And installation succeeds
Then "Installing selected agents..." is printed
And "Claude Code installed" is printed
And the TUI is relaunched
And "claude-code" is shown as checked (PreChecked=true)
```

### Scenario: Install then quit

```
Given a relaunched TUI after successful installation
When the user presses q
Then the application exits with code 0
```

### Scenario: Empty selection exits

```
Given the TUI is launched
When the user confirms with no agents selected
Then "No agents selected. Nothing to install." is printed
And the application exits with code 0
```

### Scenario: All selected agents already installed

```
Given "claude-code" is already installed on the target
And the TUI is launched
When the user selects "claude-code"
And confirms
Then "Selected agents are already installed." is printed
And the TUI is relaunched
```

### Scenario: Uninstall then install — both restarts work

```
Given "claude-code" is installed on the target
And the TUI is launched
When the user deselects "claude-code" (wizard triggers)
And chooses to uninstall
Then "Uninstalled Claude Code" is printed
And the TUI is relaunched (uninstall restart)
When the user selects "opencode"
And confirms installation
And installation succeeds
Then "Installing selected agents..." is printed
And the TUI is relaunched (install restart)
When the user quits
Then the application exits with code 0
```

### Scenario: Partial installation failure

```
Given agents "claude-code" and "opencode" are not installed
When the user selects both
And "claude-code" installs successfully but "opencode" fails
Then "Claude Code installed" is printed
And "<error message for opencode>" is printed
And the TUI is relaunched
And "claude-code" is shown as checked
And "opencode" is shown as unchecked
```

### Scenario: Install restart after uninstall wizard skip

```
Given "claude-code" is installed on the target
And the user deselects "claude-code"
And the wizard shows
When the user selects "Keep installed (skip)" for claude-code
And selects "opencode" for installation
And "opencode" installation succeeds
Then the TUI is relaunched
And both "claude-code" and "opencode" are shown as checked
```
