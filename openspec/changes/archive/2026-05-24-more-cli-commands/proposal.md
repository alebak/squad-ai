# Proposal: Remaining CLI Commands (add, remove, update, info)

## Intent

Complete the MVP CLI by implementing the four remaining commands defined in PRD §8. These commands round out the user-facing CLI for agent management: browsing available agents, removing selections, refreshing the registry, and inspecting agent details.

## Scope

### In Scope
- `squad add` — TUI stub showing available agents not installed
- `squad remove <id>` — remove agent from config `selected_agents`
- `squad update` — force registry re-fetch and cache update
- `squad info <id>` — show agent details from registry
- Table-driven tests for all 4 commands (success + error paths)
- Registration in `NewRootCommand()`

### Out of Scope
- Actual TUI implementation for `add` (future work)
- Uninstalling agents from the system (`remove` only touches config)
- Caching logic changes (`update` reuses existing `registry.SaveCache`)
- Checksum verification (post-MVP)

## Capabilities

### New Capabilities
- `add-command`: TUI stub that lists available agents and announces "TUI coming soon"
- `remove-command`: Remove an agent ID from config selected_agents and save
- `update-command`: Force remote registry re-fetch and cache save
- `info-command`: Display agent details (version, runtime, install method, description)

### Modified Capabilities
- None

## Approach

One handler struct per command in its own file, following the exact pattern from `install.go`/`list.go`:
- `addHandler` with `loadConfig`, `fetchRegistry`, `detectAll`, `configPath`
- `removeHandler` with `loadConfig`, `saveConfig`, `configPath`
- `updateHandler` with `fetchRegistry`, `saveCache`, `cachePath`
- `infoHandler` with `fetchRegistry`, `configPath`

Each has a `default*Handler()` factory and `new*CommandWithHandler()` constructor for testability.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cli/add.go` | New | `squad add` command |
| `internal/cli/add_test.go` | New | Tests for add |
| `internal/cli/remove.go` | New | `squad remove` command |
| `internal/cli/remove_test.go` | New | Tests for remove |
| `internal/cli/update.go` | New | `squad update` command |
| `internal/cli/update_test.go` | New | Tests for update |
| `internal/cli/info.go` | New | `squad info` command |
| `internal/cli/info_test.go` | New | Tests for info |
| `internal/cli/root.go` | Modified | Register 4 new commands |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `add` TUI stub will change later | High | Keep stub minimal — just output text |
| `remove` could confuse with uninstall | Low | Clear messaging: "removed from config" |
| Cache path not centralized | Low | Compute from HOME in handler factory |

## Rollback Plan

Revert the 8 new files and 1 modified file. No data migration needed — only config reads/writes.

## Dependencies

- Existing `internal/config.Load`, `config.Save`, `config.ConfigPath`
- Existing `internal/registry.Fetch`, `registry.SaveCache`
- Existing `internal/installer.DetectAll`

## Success Criteria

- [ ] All 4 commands compile and produce correct output
- [ ] All 4 commands have tests covering success and error paths
- [ ] `go test ./...` passes with no regressions
- [ ] `go vet ./...` passes with no issues
