# Spec: `squad add` Command

## References

- PRD §8 (Comandos del CLI) — `squad add` table entry
- PRD §7.4 (Modo interactivo manual) — add usage context

## Requirements

### Requirement: TUI Stub Output

`squad add` SHALL display a "TUI coming soon" message and list agents that are available for installation.

The system SHALL fetch the registry, read the config, detect which agents are installed, and print only agents that are:
- Not already installed
- Not blocked by runtime requirements
- Not already selected in config

#### Scenario: Shows available agents

- GIVEN a registry with 3 agents where one is installed, one is selected, and one is blocked
- WHEN the user runs `squad add`
- THEN the output contains "TUI coming soon"
- AND the output lists only agents that are available (not installed, not selected, not blocked)

#### Scenario: Empty registry

- GIVEN an empty registry
- WHEN the user runs `squad add`
- THEN the output contains "TUI coming soon"
- AND the output contains "No agents" or equivalent message

### Requirement: Registry Fetch Failure

If the registry cannot be fetched and no cache is available, `squad add` SHALL error with a message explaining the failure.

#### Scenario: Registry fetch failure

- GIVEN no internet and no cache
- WHEN the user runs `squad add`
- THEN the command errors explaining the network failure
