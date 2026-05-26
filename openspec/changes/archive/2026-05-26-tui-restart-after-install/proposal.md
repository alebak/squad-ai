# Proposal: TUI Restart After Agent Installation

## Intent

After installing agents through the interactive TUI (`squad add`), the user is dropped back to the terminal instead of returning to the TUI showing the updated agent state. This forces the user to re-run `squad add` to see their changes or install additional agents.

## Scope

**In scope:**
- `internal/cli/add.go` — modify `runAddFlowInteractive` to loop back to TUI after install
- `internal/cli/add_test.go` — update existing tests and add new test for the relaunch behavior

**Out of scope:**
- Non-interactive flow (`runAddFlowNonInteractive`) — remains unchanged
- Install with errors — error reporting stays, but TUI relaunches instead of exiting
- Any changes to the TUI model or Bubbletea components

## Approach

Move the installation logic INSIDE the outer `for` loop in `runAddFlowInteractive`, then `continue` back to the top instead of `break`/`return`. This mirrors the existing uninstall restart pattern (lines 239-244 in add.go).

After install completes:
1. Mark succeeded agents in the `installed` map
2. Rebuild `agentItems` with `buildAgentItemsForAdd`
3. `continue` — TUI relaunches showing updated state

## Rollback Plan

Low risk — revert the two file changes (add.go, add_test.go). The change is entirely additive to the loop structure; no data migrations or config changes.

## Delivery Strategy

Single PR. Estimated ~60-80 lines changed. Fits well under the 400-line review budget.
