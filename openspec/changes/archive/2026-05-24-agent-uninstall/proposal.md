# Proposal: Agent Uninstall Support

## Intent

Add the ability to actually uninstall an agent's binary from the system when removing it from the squad config, using a hybrid registry-driven + method-derived fallback approach.

## Scope

- **Registry schema**: Add optional `uninstall` string field to `InstallCmd` struct
- **Installer package**: New `UninstallAgent(agent) error` function with method-derived fallbacks
- **CLI**: `squad remove --uninstall` flag with confirmation prompt
- **Registry data**: Add explicit uninstall commands for npm-based agents (codex, gemini-cli, pi)
- **Tests**: Unit tests for all new functions

Out of scope:
- Batch uninstall (`--all` flag)
- Uninstall from the TUI
- Config file / cache cleanup

## References

- Issue #19 — Agent uninstall support
- PRD §8 — `squad remove <id>` "Remueve un agente del manifiesto (no lo desinstala del sistema)"
- `openspec/changes/agent-uninstall/exploration.md` — Hybrid approach (Approach 3)

## Approach

**Hybrid: registry-driven with method-derived fallback** (Approach 3 from exploration):

1. Add `UninstallCmd string` field to `InstallCmd` in the registry model (`internal/registry/agent.go`)
2. Implement `UninstallAgent(agent registry.Agent) error` in `internal/installer/uninstall.go`:
   - If `agent.Install.UninstallCmd` is non-empty → execute it via `sh -c` (same as install)
   - Else if `MethodNpmInstall` → derive `npm uninstall -g <package>` from install command
   - Else if `MethodCurlBash` → use `exec.LookPath` + `os.Remove` (safe, no shell injection)
   - Else → return descriptive error
3. Update `squad remove` with `--uninstall` flag:
   - Without flag → current behavior (config-only removal)
   - With flag → fetch registry → detect if installed → confirm → uninstall → remove from config
   - With `--force` → skip confirmation
4. Update `registry/agents.json` with explicit uninstall commands for known npm agents
5. Tests for all new code paths

## Key Decisions

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Fallback strategy | Derive from method | Works for all agents without registry changes |
| curl_bash fallback | Go `LookPath` + `os.Remove` | Avoids shell injection from detect_cmd |
| Confirmation | Prompt before destructive ops | Safety; user must explicitly consent to uninstall |
| `--force` flag | Skip confirmation | Allows scripting / non-interactive use |
| Backward compat | `--uninstall` defaults to false | Existing `squad remove` behavior unchanged |
| `IsAgentInstalled` check | Skip uninstall if not installed | No-op for already-removed binaries |

## Rollback Plan

1. Revert changes to `internal/registry/agent.go` (remove `UninstallCmd` field)
2. Delete `internal/installer/uninstall.go` and `internal/installer/uninstall_test.go`
3. Revert `internal/cli/remove.go` to original handler + flow
4. Revert `internal/cli/remove_test.go`
5. Revert `registry/agents.json`

All changes are additive and backward-compatible. No migration needed.
