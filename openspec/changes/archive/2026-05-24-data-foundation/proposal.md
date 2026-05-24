# Proposal: Data Foundation — Registry + Config Models

## Intent

Establish the data layer for Squad AI: the agent registry (source of truth for available agents) and user config (persistent agent selection). This enables the first-run TUI flow and silent install mode defined in PRD §6.2–6.4.

## Scope

### In Scope
- `registry/agents.json` — 7 MVP agent entries matching PRD §13
- `internal/registry/` — Agent types, HTTP fetch, local cache (read/write), staleness check
- `internal/config/` — Config types, atomic file read/write (temp + rename), default config

### Out of Scope
- TUI selection UI (future — sdd-design for install command)
- Runtime detection (future — separate `internal/runtime/`)
- Checksum verification (Fase 2 per PRD §11.2)
- Agent installer execution (future — `internal/installer/`)
- Enterprise/custom registry URLs (Fase 3 per PRD §14)

## Capabilities

### New Capabilities
- **agent-registry**: Registry file format, remote HTTP fetch, local cache with 24h TTL, and Go types (Agent, InstallCmd, RuntimeDep, Checksum, Registry)
- **user-config**: User config file format, load/save with atomic writes, default config, config path resolution

### Modified Capabilities
None — both capabilities are greenfield.

## Approach

- **Registry types**: Flat structs mirroring JSON, one-to-one. `InstallMethod` as typed string const. `Registry` wrapper struct for the agent list + updated_at. Package-level `var RegistryURL` for future override.
- **Registry fetch + cache**: `net/http` GET from raw GitHub URL. Cache to `$HOME/.config/squad-ai/registry.cache.json`. Stale after 24h (check via `registry_last_checked`). Fall back to cache on fetch failure.
- **Config**: `json.Unmarshal` on load, `json.Marshal` on save. Atomic write via temp file + `os.Rename`. Config path hardcoded to `$HOME/.config/squad-ai/config.json` (resolved from user — no XDG abstraction). `DefaultConfig()` returns sensible defaults (no agents selected, zero timestamps).
- **Config path**: `$HOME/.config/squad-ai/` — same pattern as `$HOME/.gentle-ai/`. Create dir on first write if missing.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `registry/agents.json` | New | Source-of-truth with 7 MVP agents |
| `internal/registry/agent.go` | New | Agent, InstallCmd, RuntimeDep, Checksum, Registry types |
| `internal/registry/client.go` | New | Fetch(), LoadCache(), SaveCache(), IsStale() |
| `internal/registry/registry_test.go` | New | Parse/round-trip tests |
| `internal/config/config.go` | New | Config, InstallOptions types. Load(), Save(), DefaultConfig() |
| `internal/config/config_test.go` | New | Read/write round-trip tests |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Hardcoded config path breaks on non-Linux | Low | MVP targets Linux containers only. Revisit at macOS/Windows support. |
| Remote registry fetch fails | Med | Cache fallback on error. First-run without network shows clear error. |
| Config.json corruption on crash | Low | Atomic write (temp + rename) prevents partial writes. |

## Rollback Plan

1. Remove or revert `registry/agents.json` to empty array
2. Delete `internal/registry/` and `internal/config/` packages
3. No migration needed — config is ephemeral (recreated on next `squad` run)

## Dependencies

- Go stdlib: `net/http`, `encoding/json`, `os`, `io`, `time`, `path/filepath`
- No external dependencies

## Success Criteria

- [ ] `go build ./...` compiles with both new packages
- [ ] `registry/agents.json` parses into `[]Agent` with all 7 entries valid
- [ ] Cache round-trip: fetch → save → load → verify identical agents
- [ ] Config round-trip: write config → read back → fields match
- [ ] Atomic write: temp file cleaned up on success, no partial writes
- [ ] Staleness check: cache <24h returns false, >24h returns true
- [ ] Fallback: when fetch fails but cache exists, LoadCache returns cached data
