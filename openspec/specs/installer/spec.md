# Installer Specification

## Purpose

The installer domain executes agent installation commands and tracks progress. It runs shell commands via `exec.Command("sh", "-c", command)`, captures stdout/stderr to log files, and provides a progress callback for TUI integration.

## Requirements

### Requirement: ProgressFn Type

The system SHALL define `type ProgressFn func(agentID string, percentage int)` for reporting installation progress. This SHALL be a callback type — the caller provides the implementation.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Type defined | package installer | imported | ProgressFn is available as a function type |
| Zero value | var fn ProgressFn | fn called | fn is nil-callable (caller checks nil) |

### Requirement: Log Path Generation

The system SHALL generate log paths in the format `~/.config/squad-ai/logs/<agentID>-<timestamp>.log`. The timestamp SHALL use `time.Now().Format(time.RFC3339)` with colons replaced by hyphens for filesystem safety.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Agent ID "opencode" | agentID="opencode" | LogPath() called | Path ends with opencode-<timestamp>.log |
| Timestamp safety | time containing ":" | LogPath() called | Colons replaced with "-" |
| Logs dir | ~/.config/squad-ai/logs/ absent | LogPath() called | Directory created with 0755 |

### Requirement: InstallAgent

InstallAgent(agent registry.Agent, progress ProgressFn) error SHALL execute agent.Install.Command via `exec.Command("sh", "-c", command)`. It SHALL capture combined stdout+stderr and write to a log file.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Command succeeds | Install.Command="echo ok" | InstallAgent() | Returns nil, log file created |
| Command fails | Install.Command="/bin/false" | InstallAgent() | Returns wrapped error with exit code |
| Progress called | non-nil ProgressFn | InstallAgent() runs | ProgressFn called with 100% on success |

### Requirement: Checksum Verification Placeholder

The system SHOULD verify agent.Checksum.SHA256 before executing the install command. When checksums are nil/MVP mode, it SHALL log a warning to stderr and proceed.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Checksum nil | agent.Checksum==nil | InstallAgent() | Warning logged, installation proceeds |
| Checksum not nil | agent.Checksum.SHA256 set | InstallAgent() | Verification attempted (post-MVP) |

### Requirement: UninstallConfig

The system SHALL provide `UninstallConfig(agent registry.Agent) error` that removes config directories listed in `agent.ConfigPaths`.

Resolution:
1. For each path in `ConfigPaths`, expand `~` to the user's home directory via `os.UserHomeDir()`
2. Resolve the path to an absolute path via `filepath.Abs`
3. For paths starting with `~`, verify the resolved path stays within the home directory
4. Call `os.RemoveAll` for each resolved path
5. Non-existent paths SHALL be skipped (no error)

#### Scenario: Remove single config directory

- GIVEN an agent with `ConfigPaths: ["~/.claude"]`
- WHEN `UninstallConfig` is called
- THEN the expanded path is removed with `os.RemoveAll`

#### Scenario: Skip non-existent config dir

- GIVEN an agent with `ConfigPaths: ["~/.nonexistent"]`
- WHEN `UninstallConfig` is called
- THEN the non-existent path is skipped without error

#### Scenario: Empty ConfigPaths is no-op

- GIVEN an agent with nil or empty `ConfigPaths`
- WHEN `UninstallConfig` is called
- THEN the function returns nil immediately

#### Scenario: Path traversal protection

- GIVEN an agent with `ConfigPaths: ["~/../etc"]`
- WHEN `UninstallConfig` is called
- THEN the function returns an error about path outside home directory

### Requirement: InstallAll

InstallAll(agents []registry.Agent, progress ProgressFn) []error SHALL iterate agents sequentially, call InstallAgent for each, and collect results. It MUST NOT stop on first failure — it SHALL attempt all agents.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| All succeed | 3 agents, all pass | InstallAll() | Returns []error{nil, nil, nil} |
| Mixed results | 1 fails, 2 pass | InstallAll() | Returns []error with one non-nil |
| Empty slice | 0 agents | InstallAll() | Returns empty slice |
