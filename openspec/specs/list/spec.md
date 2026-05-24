# Spec: `squad list` Command

## References

- PRD §8 (Comandos del CLI) — `squad list` table entry
- PRD §7.4 (Modo interactivo manual) — list usage context

## Requirements

### R1: Table output

`squad list` SHALL display a table with columns: Agent ID, Name, Installed (✅/❌), Status (selected/available/blocked).

### R2: Data sources

The command SHALL:
- Fetch the registry (remote or cache)
- Read the user config for selected agents
- Detect currently installed agents via `command -v`
- Detect runtime availability for each agent

### R3: Status definitions

- **installed**: agent's `detect_command` binary found in PATH. The Installed column shows ✅.
- **selected**: agent is in the config's `selected_agents` list but NOT installed.
- **available**: agent is in the registry, not installed, not selected, not blocked.
- **blocked**: agent requires a runtime that is not available or below minimum version.

### R4: Registry fetch behavior

Same as install: attempt remote fetch, fall back to cache, error if neither is available.

### R5: Simple text format (MVP)

No table library. Use `fmt.Printf` with fixed-width columns. Output looks like:

```
Agent ID       Name              Installed   Status
claude-code    Claude Code       ✅          selected
opencode       OpenCode          ❌          available
codex          Codex CLI         ❌          blocked
```

## Scenarios

### Scenario S1: Normal list output
**Given** a registry with 3 agents, config with 1 selected
**When** the user runs `squad list`
**Then** the output shows 3 rows with correct installed/status columns

### Scenario S2: Empty registry
**Given** an empty registry
**When** the user runs `squad list`
**Then** the output shows only the header row

### Scenario S3: Registry fetch failure
**Given** no internet and no cache
**When** the user runs `squad list`
**Then** the command errors explaining the network failure
