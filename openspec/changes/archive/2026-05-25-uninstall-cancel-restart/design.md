# Design: Uninstall Cancel Returns to TUI

## Technical Approach

Wrap the TUI `RunSelection` + uninstall prompt in a `for` loop inside `runAddFlowInteractive`. When user picks Cancel for any deselected installed agent, set `needsRestart = true`. After processing ALL deselected agents, if `needsRestart` is true, rebuild `agentItems` with `buildAgentItemsForAdd` (using updated `installed` map) and `continue` — re-launching the TUI. Only `break` when no cancels occurred.

Key insight: agents confirmed for uninstall are removed from the `installed` map so they won't be pre-checked on TUI relaunch. Cancelled agents remain in the map and get re-pre-checked.

## Architecture Decisions

### Decision: Outer loop wraps selection + uninstall

**Choice**: Single `for` loop calling `runSelection` repeatedly
**Alternatives considered**: Keep the TUI running and show uninstall prompt as a Bubbletea overlay
**Rationale**: The TUI (Bubbletea) completes when user presses Enter — we can't inject a prompt into a finished program. Re-launching is the cleanest approach without rewriting the TUI. The loop is contained within `runAddFlowInteractive`, so no other callers are affected.

### Decision: Rebuild agentItems from scratch on restart

**Choice**: Call `buildAgentItemsForAdd(h, catalog, installed)` to rebuild `agentItems`
**Alternatives considered**: Mutate individual PreChecked fields in the existing slice
**Rationale**: The `installed` map may have changed (agents uninstalled via options 1/2 before a cancel). Rebuilding from the canonical function ensures consistency. The function already exists and is idempotent.

### Decision: Remove agents from `installed` map immediately on uninstall

**Choice**: `delete(installed, agent.ID)` after successful uninstall
**Alternatives considered**: Deferred batch uninstall
**Rationale**: If cancel is chosen after some agents were already uninstalled, the `installed` map must reflect the current state for the TUI relaunch. Deleting immediately keeps a single source of truth.

## Data Flow

```
┌──────────┐    selectedIDs    ┌──────────────────┐
│ TUI      │ ───────────────→  │ Uninstall check   │
│ RunSel   │                   │ for each installed│
│          │ ←── re-launch ─── │ agent NOT selected │
└──────────┘    (continue)     │                   │
                                │ Choice?           │
                                │  1 → uninstall    │
                                │  2 → uninstall+cfg│
                                │  3 → needsRestart │
                                └────────┬─────────┘
                                         │
                          ┌──────────────┴──────────────┐
                          │ needsRestart?                │
                          │  YES → rebuild agentItems    │
                          │         continue (re-launch) │
                          │  NO  → break (proceed to     │
                          │         install)             │
                          └─────────────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modify | Wrap uninstall flow in restart loop, rebuild agentItems on cancel |
| `internal/cli/add_test.go` | Modify | Update `TestAddCommand_UninstallPromptCancel` to handle multiple `runSelection` calls |

## Interfaces / Contracts

No new types or interfaces. The `installed map[string]bool` parameter in `runAddFlowInteractive` is already a reference type — mutations inside the function are visible to the caller.

Key contract change: `runSelection` MAY be called multiple times per `runAddFlowInteractive` invocation. The mock in tests must handle this.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Cancel restarts TUI | Update `TestAddCommand_UninstallPromptCancel` — mock `runSelection` to return different results on 2nd call |
| Unit | Cancel + confirm combo | Test user cancels one agent, confirms another, TUI re-launches correctly |
| Unit | No cancel = no loop | Existing test `TestAddCommand_TUISuccessFlow` verifies normal flow continues without looping |
| Unit | All cancels = no uninstalls | Verify uninstall functions never called when all choices are Cancel |

## Migration / Rollout

No migration required. The change is entirely in-memory flow control.

## Open Questions

None.
