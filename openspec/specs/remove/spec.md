# Spec: `squad remove` Command

## References

- PRD §8 (Comandos del CLI) — `squad remove` table entry
- PRD §7.4 (Modo interactivo manual) — remove usage context

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
