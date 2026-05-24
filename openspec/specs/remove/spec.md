# Spec: `squad remove` Command

## References

- PRD §8 (Comandos del CLI) — `squad remove` table entry
- PRD §7.4 (Modo interactivo manual) — remove usage context
- Issue #19 — Agent uninstall support
- `openspec/changes/archive/2026-05-24-agent-uninstall/specs/agent-uninstall/spec.md`

## Requirements

### Requirement: Remove Agent from Config

`squad remove <id>` SHALL read the user config, remove the specified agent ID from `selected_agents`, save the config, and print a confirmation message.

If the agent ID is not in `selected_agents`, the command SHALL print a warning and succeed.

#### Scenario: Removes agent from config

- GIVEN a config with `selected_agents: ["claude-code", "opencode"]`
- WHEN the user runs `squad remove opencode`
- THEN the agent is removed from `selected_agents`
- AND the config is saved
- AND output contains confirmation message
- AND the message notes the agent is still installed on the system

#### Scenario: Agent not in config

- GIVEN a config with `selected_agents: ["claude-code"]`
- WHEN the user runs `squad remove nonexistent`
- THEN the command prints a "not found" message
- AND exits with success (no error)

### Requirement: Argument Validation

`squad remove` SHALL require exactly one argument (the agent ID). If no argument is provided, the command SHALL error.

#### Scenario: Missing argument

- GIVEN no arguments
- WHEN the user runs `squad remove`
- THEN the command errors indicating that an agent ID is required

### Requirement: Config Save Failure

If the config cannot be saved after removal, the command SHALL surface the error.

#### Scenario: Config save failure

- GIVEN a config that loads successfully
- WHEN saving after removal fails
- THEN the command errors indicating the save failure

### Requirement: Uninstall Flag

`squad remove <id> --uninstall` SHALL uninstall the agent binary before removing it from the config. The `--force` flag SHALL skip the confirmation prompt.

Uninstall resolution order:
1. If the agent has an explicit `uninstall` command in the registry → execute it
2. If `npm_install` method → derive `npm uninstall -g <package>` from install command
3. If `curl_bash` method → resolve binary via `exec.LookPath` and delete with `os.Remove`
4. Otherwise → return an error indicating no uninstall method is defined

#### Scenario: Uninstall and remove from config

- GIVEN an agent installed on the system and in `selected_agents`
- WHEN the user runs `squad remove <id> --uninstall --force`
- THEN the agent SHALL be uninstalled
- AND the agent SHALL be removed from `selected_agents`
- AND output SHALL contain both uninstall and removal confirmation messages

#### Scenario: Uninstall when agent not installed

- GIVEN an agent in `selected_agents` but NOT installed (binary missing)
- WHEN the user runs `squad remove <id> --uninstall --force`
- THEN the command SHALL print a notice that the agent is not installed
- AND the agent SHALL still be removed from `selected_agents`

#### Scenario: Uninstall without --uninstall flag (backward compat)

- GIVEN an agent installed on the system and in `selected_agents`
- WHEN the user runs `squad remove <id>` (no `--uninstall`)
- THEN the agent SHALL NOT be uninstalled
- AND the agent SHALL be removed from `selected_agents`
- AND the output SHALL note the agent is still installed

#### Scenario: Confirmation prompt

- GIVEN the user runs `squad remove <id> --uninstall` WITHOUT `--force`
- WHEN the agent is installed
- THEN the command SHALL display the uninstall action
- AND SHALL prompt for confirmation
- AND SHALL only proceed on affirmative response

#### Scenario: Agent not in registry

- GIVEN the user runs `squad remove <id> --uninstall --force`
- AND `<id>` is not in the remote registry
- WHEN the registry is fetched
- THEN the command SHALL print a warning that the agent is not in the registry
- AND the agent SHALL still be removed from `selected_agents`

### Requirement: Registry Uninstall Field

The `InstallCmd` struct in `internal/registry/agent.go` SHALL include an optional `UninstallCmd` string field with JSON tag `"uninstall,omitempty"`. The `registry/agents.json` SHALL include explicit uninstall commands for all `npm_install` agents.
