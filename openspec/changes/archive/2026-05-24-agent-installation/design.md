# Design: Agent Installation with Progress Tracking

## Technical Approach

Add `install.go` to the existing `internal/installer/` package. Three public symbols: `ProgressFn` (callback type), `InstallAgent` (single agent), `InstallAll` (batch). One unexported helper: `logPath` (path generation). Follow existing patterns — godoc on all exported symbols, error wrapping with `%w`, stdlib only.

## Architecture Decisions

### Decision: Single file for install logic

**Choice**: One `install.go` file.
**Alternatives considered**: Splitting into `progress.go`, `log.go`, `install.go`.
**Rationale**: The existing package has one file (`installer.go`). Adding a second file keeps symmetry. The install logic is ~100 lines — splitting would be premature modularization.

### Decision: exec.Command("sh", "-c", command) for execution

**Choice**: Use `exec.Command("sh", "-c", agent.Install.Command)`.
**Alternatives considered**: Parsing command string into argv, using `bash -c`.
**Rationale**: Agents provide install commands as shell strings (e.g., `curl https://... | bash`). `sh` is POSIX-mandated and lighter than `bash`. Direct exec without shell would break pipes and redirects.

### Decision: No checksum verification in MVP

**Choice**: Log a warning when checksums are nil and proceed.
**Alternatives considered**: Blocking installation when no checksum.
**Rationale**: PRD §11.1 says checksums are MVP — but the registry may not populate them yet. Blocking would break the CLI. The placeholder makes the contract clear.

## Data Flow

```
InstallAll(agents, progress)
  │
  └─► for each agent ──► InstallAgent(agent, progress)
       │                      │
       │                      ├─► generate logPath(agent.ID)
       │                      ├─► create log directory
       │                      ├─► exec.Command("sh", "-c", command)
       │                      ├─► capture stdout+stderr to log file
       │                      ├─► call progress(agent.ID, 100) on success
       │                      └─► return error if exit code != 0
       │
       └─► collect []error results
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/installer/install.go` | Create | InstallAgent, InstallAll, ProgressFn, logPath |
| `internal/installer/install_test.go` | Create | Tests for all functions |

## Interfaces / Contracts

```go
// ProgressFn is called during installation to report progress.
// percentage is 0-100. Implementations must handle nil safely.
type ProgressFn func(agentID string, percentage int)

// InstallAgent executes the agent's install command, captures output to a log
// file, and calls progress(agentID, 100) on success.
func InstallAgent(agent registry.Agent, progress ProgressFn) error

// InstallAll installs all agents sequentially and returns per-agent errors.
func InstallAll(agents []registry.Agent, progress ProgressFn) []error
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | InstallAgent success | `exec.Command("/bin/true")` — expect nil |
| Unit | InstallAgent failure | `exec.Command("/bin/false")` — expect error |
| Unit | InstallAgent with progress | Echo command + mock fn checks percentage=100 |
| Unit | InstallAll mixed results | 3 agents (true, false, true) — check 2 nil, 1 error |
| Unit | InstallAll empty | Empty slice — expect empty result |
| Unit | LogPath directory creation | Call logPath, verify dir exists |

## Migration / Rollout

No migration required. New code only, no existing behavior changes.

## Open Questions

None. Requirements are fully specified in PRD §10 and §11.
