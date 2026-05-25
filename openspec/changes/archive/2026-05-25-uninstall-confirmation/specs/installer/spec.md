# Delta for Installer

## ADDED Requirements

### Requirement: UninstallConfig

The system SHALL provide `UninstallConfig(agent registry.Agent) error` that removes config directories listed in `agent.ConfigPaths`.

Resolution:
1. For each path in `ConfigPaths`, expand `~` to the user's home directory via `os.UserHomeDir()`
2. Resolve relative paths against the home directory
3. Call `os.RemoveAll(path)` for each resolved path
4. If a path does not exist, skip it (no error)
5. If `os.RemoveAll` fails for any path, return a wrapped error listing all failures

The function SHALL validate each path before deletion: path expansion SHALL NOT produce paths outside the user's home directory.

#### Scenario: Remove single config directory

- GIVEN an agent with `ConfigPaths: ["~/.claude"]`
- WHEN `UninstallConfig` is called
- THEN `~/.claude` is expanded to the home directory
- AND `os.RemoveAll` is called on the expanded path

#### Scenario: Skip non-existent config dir

- GIVEN an agent with `ConfigPaths: ["~/.nonexistent"]`
- WHEN `UninstallConfig` is called
- THEN the non-existent path is skipped without error
- AND the function returns nil

#### Scenario: Multiple config paths

- GIVEN an agent with `ConfigPaths: ["~/.config/opencode", "~/.opencode"]`
- WHEN `UninstallConfig` is called
- THEN both paths are removed

#### Scenario: Empty ConfigPaths is no-op

- GIVEN an agent with nil or empty `ConfigPaths`
- WHEN `UninstallConfig` is called
- THEN the function returns nil immediately

#### Scenario: Path traversal protection

- GIVEN an agent with `ConfigPaths: ["../../etc/passwd"]`
- WHEN `UninstallConfig` is called
- THEN the function SHALL NOT remove paths outside the home directory
- AND SHALL return an error
