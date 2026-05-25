# Proposal: 3-Option Uninstall Confirmation

## Intent

When a user deselects an installed agent in the TUI, they currently get a single yes/no prompt for uninstall. This gives no control over what gets removed. The agent binary and config data should be deletable independently, with a clear cancel option that reselects the agent.

## Scope

### In Scope
- Add `ConfigPaths []string` field to Agent struct / JSON schema
- Add `UninstallConfig(agent)` function to installer package
- Update `UninstallAgent` to indicate what was done
- 3-option prompt in `runAddFlowInteractive` (app only / app+config / cancel)
- Update `registry/agents.json` with config_paths for all 7 agents
- Tests for all new functions

### Out of Scope
- Separate `squad uninstall` command (future)
- Non-interactive uninstall flags for `squad add`
- Cleanup of editor plugins or shell rc modifications
- Uninstall from `squad remove` command (out of scope — existing behavior unchanged)

## Capabilities

### New Capabilities
- `uninstall`: Agent uninstall with config cleanup support

### Modified Capabilities
- `add`: Uninstall prompt changed from yes/no to 3-option — requires delta spec
- `agent-registry`: Agent struct gets new `config_paths` field — requires delta spec
- `installer`: New `UninstallConfig` function — requires delta spec

## Approach

1. Add `ConfigPaths []string` to `registry.Agent` with `json:"config_paths,omitempty"`
2. Create `UninstallConfig(agent) error` in `internal/installer/` that removes each config path via `os.RemoveAll` after expanding `~` and env vars
3. Refactor `runAddFlowInteractive` in `internal/cli/add.go`: replace single `confirmFn` call with 3-option prompt loop
4. Update `registry/agents.json` with `config_paths` entries
5. Write tests for config path removal and prompt behavior

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/registry/agent.go` | Modified | Add `ConfigPaths` field to `Agent` |
| `internal/installer/uninstall.go` | Modified | Add `UninstallConfig` function |
| `internal/cli/add.go` | Modified | 3-option uninstall prompt |
| `registry/agents.json` | Modified | Add `config_paths` entries |
| `internal/cli/add_test.go` | Modified | Test 3-option prompt |
| `internal/installer/uninstall_test.go` | Modified | Test `UninstallConfig` |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| `os.RemoveAll` on wrong path | Low | Expand only well-known paths from registry, validate path safety |
| Prompt text too long/confusing | Low | Keep options short, use numbered selection |

## Rollback Plan

Revert changes to `add.go`, `agent.go`, `uninstall.go`, and `agents.json`. The old yes/no prompt will resume working without data loss.

## Dependencies

- Existing `UninstallAgent` function (unchanged interface)
- Existing `confirmFn` handler field for prompt testing

## Success Criteria

- [ ] Deselecting an installed agent shows 3 options: [1] app only [2] app+config [3] cancel
- [ ] Option 1 removes binary via existing `UninstallAgent`
- [ ] Option 2 removes binary + config directories
- [ ] Option 3 reselects the agent (no data loss)
- [ ] All existing tests pass
