# Delta: `squad info` Command

## ADDED Requirements

### Requirement: Show Agent Details

`squad info <id>` SHALL fetch the registry, find the agent by ID, and display its details: version, runtime dependencies, install method, and description.

#### Scenario: Agent found

- GIVEN a registry containing `claude-code` with version "latest", runtime "none", install method "curl_bash", and a description
- WHEN the user runs `squad info claude-code`
- THEN the output contains the agent's ID, name, version, runtime, install method, and description

### Requirement: Agent Not Found

If the agent ID does not exist in the registry, `squad info` SHALL error with a "not found" message.

#### Scenario: Agent not found

- GIVEN a registry that does not contain `nonexistent`
- WHEN the user runs `squad info nonexistent`
- THEN the command errors indicating the agent was not found

### Requirement: Argument Validation

`squad info` SHALL require exactly one argument (the agent ID). If no argument is provided, the command SHALL error.

#### Scenario: Missing argument

- GIVEN no arguments
- WHEN the user runs `squad info`
- THEN the command errors indicating that an agent ID is required

### Requirement: Registry Fetch Failure

If the registry cannot be fetched and no cache is available, `squad info` SHALL error.

#### Scenario: Registry fetch failure

- GIVEN the registry is unreachable and no cache exists
- WHEN the user runs `squad info claude-code`
- THEN the command errors explaining the network failure
