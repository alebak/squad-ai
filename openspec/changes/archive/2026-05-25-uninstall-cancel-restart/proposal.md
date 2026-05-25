# Proposal: Uninstall Cancel Returns to TUI

## Intent

When a user deselects an installed agent in the TUI and chooses [3] Cancel at the uninstall prompt, squad prints "Selected agents are already installed." and exits to shell. The user expects to return to the TUI to continue editing selections.

## Scope

### In Scope
- Modify `runAddFlowInteractive` to wrap TUI selection + uninstall prompts in a restart loop
- When Cancel is chosen for any deselected installed agent, rebuild agent selection and re-launch the TUI
- Update the existing "Cancel" test to match new behavior (multiple `runSelection` calls)

### Out of Scope
- Changing the Bubbletea TUI itself — TUI is fine, the issue is in the post-TUI flow
- Changing the uninstall prompt UI — it stays as stdin text prompt
- New TUI features or design changes

## Capabilities

### Modified Capabilities
- `add`: The "Cancel" behavior in the 3-Option Uninstall Prompt is changing — instead of silently keeping the agent selected and proceeding to install, it re-launches the TUI so the user can continue editing

## Approach

Wrap the TUI `RunSelection` call + uninstall prompt loop in an outer `for` loop. When user chooses Cancel for any deselected installed agent, rebuild the `agentItems` slice via `buildAgentItemsForAdd` (updated `installed` map) and `continue` to re-launch the TUI. Only break out of the loop when no cancels occurred.

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modified | Wrap uninstall flow in restart loop |
| `internal/cli/add_test.go` | Modified | Update `TestAddCommand_UninstallPromptCancel` to handle TUI restart |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Test mock `runSelection` loops infinitely | Low | Use call counter in test to return different results on subsequent calls |
| User confusion on TUI restart | Low | Restored selection matches what user expects (cancelled agent re-checked) |

## Rollback Plan

Revert `internal/cli/add.go` and `internal/cli/add_test.go` to their previous state. No schema or config changes.

## Dependencies

None.

## Success Criteria

- [ ] When Cancel is chosen for ALL deselected installed agents, the TUI re-launches instead of exiting
- [ ] When Cancel is chosen but some agents were confirmed for uninstall, the TUI re-launches with only the kept agents pre-checked
- [ ] When NO installed agents are deselected, the flow proceeds to installation without looping
- [ ] All existing tests pass
