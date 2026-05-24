# Spec: `squad install` Command

## References

- PRD §8 (Comandos del CLI) — `squad install` table entry
- PRD §7.3 (Override por parámetros) — `--agents`, `--all` flags

## Requirements

### R1: Default install

`squad install` (no flags) SHALL read the user config, fetch the registry, detect which agents are already installed, detect available runtimes, filter agents whose runtime is missing, and install only the selected agents that are NOT already installed.

If the user config has no `selected_agents`, the command SHALL complete with no installations and no error.

### R2: `--agents` flag

`squad install --agents <ids>` SHALL accept a comma-separated list of agent IDs, install each one if not already installed, skip agents whose runtime is missing (with a warning), and SHALL NOT write to the config.

### R3: `--all` flag

`squad install --all` SHALL install all agents from the registry that are:
- Not already installed
- Runtime-compatible (or have `runtime: none`)

### R4: Registry fetch behavior

The command SHALL attempt a remote registry fetch. If the fetch fails, it SHOULD fall back to the local cache. If no cache exists, it SHALL error with a message explaining the network failure.

### R5: Progress reporting

Installation SHALL print one line per agent as it completes:
```
✅ {Name} installed
❌ {Name} — {error message}
```

### R6: Error resilience

A failed installation SHALL NOT stop other installations. Errors SHALL be collected and reported at the end, with a non-zero exit code if any installation failed.

### R7: Flag exclusivity

`--agents` and `--all` SHALL be mutually exclusive. If both are set, the command SHALL error.

## Scenarios

### Scenario S1: Default install with config
**Given** a config with `selected_agents: ["claude-code", "opencode"]`
**When** the user runs `squad install`
**Then** the command fetches the registry
**And** detects currently installed agents
**And** detects runtimes
**And** installs only the selected agents that are not already installed

### Scenario S2: `--agents` flag
**Given** no config or a config with different agents
**When** the user runs `squad install --agents claude-code,opencode`
**Then** the command installs exactly those two agents

### Scenario S3: `--all` flag
**Given** a registry with 5 agents
**When** the user runs `squad install --all`
**Then** the command attempts to install all 5
**And** skips those already installed
**And** skips those with unmet runtime dependencies

### Scenario S4: No config
**Given** no config file exists
**When** the user runs `squad install`
**Then** the command completes with no installations (empty selected_agents)

### Scenario S5: Mixed flags
**Given** both `--agents` and `--all` are provided
**When** the user runs `squad install --agents a --all`
**Then** the command errors with a message about mutually exclusive flags

### Scenario S6: Registry fetch failure
**Given** no internet and no cache
**When** the user runs `squad install`
**Then** the command errors explaining the network failure
