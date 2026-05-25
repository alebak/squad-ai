# Design: Bulk Uninstall Confirmation

## Technical Approach

In `runAddFlowInteractive`, restructure the deselection handling loop:

1. **Collect** all deselected installed agents first into a slice
2. **Branch** on count: if >1, show combined `confirmFn` prompt; if ==1, show existing per-agent 3-option prompt
3. **Bulk yes** → iterate slice, call `UninstallAgent` (app-only) for each, delete from `installed` map
4. **Bulk no** → keep all in `installed` map, set `needsRestart = true` (re-launch TUI with all pre-checked)
5. **Single agent** → unchanged 3-option flow from #25

The key structural change: extract detection (which agents are deselected) from the prompt loop.

## Architecture Decisions

### Decision: Branch on count of deselected installed agents

**Choice**: If count > 1 → combined prompt; if count == 1 → existing per-agent prompt; if count == 0 → skip
**Alternatives considered**: Always use combined prompt; remove per-agent prompt entirely
**Rationale**: Bulk uninstall with granular choice (app-only vs app+config vs cancel per agent) would be complex UX. Combined prompt is simpler and addresses the actual pain point (unselect all with zero confirmation). The per-agent 3-option prompt is retained for single-agent edge cases where granularity matters.

### Decision: Bulk uses app-only uninstall only

**Choice**: `UninstallAgent` but not `UninstallConfig` for bulk case
**Alternatives considered**: Include a 2-option sub-prompt; always do app+config
**Rationale**: App-only is the safest default — it removes the binary but preserves config. Users who need config cleanup can use the per-agent path. This keeps the combined prompt simple (one y/N question).

### Decision: Use existing `confirmFn` for combined prompt

**Choice**: Reuse `h.confirmFn(msg)` — the same function used in `remove.go`
**Alternatives considered**: Create a new dedicated prompt function
**Rationale**: `confirmFn` already formats a `[y/N]` prompt and reads from stdin. The message parameter makes it flexible. No new interface method needed.

## Data Flow

```
TUI returns selectedIDs
       │
       ▼
Build selectedSet map
       │
       ▼
Collect deselectedInstalled = []
for each agent in catalog:
  if installed[agent.ID] && !selectedSet[agent.ID]:
    append to deselectedInstalled
       │
       ▼
┌─── count(deselectedInstalled) ───┐
│                                  │
│ 0                                │
│ → skip prompt,                   │
│   proceed to install             │
│                                  │
│ 1                                │
│ → existing per-agent 3-option    │
│   flow (unchanged from #25)      │
│                                  │
│ >1                               │
│ → build names list               │
│ → h.confirmFn("Some selected...  │
│    Uninstall them as well?")     │
│   ├─ y/yes → uninstallAll()     │
│   │          app-only for each   │
│   └─ n/no → needsRestart=true   │
│             (re-add all)         │
└──────────────────────────────────┘
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modify | Restructure lines 255-286: collect all deselected agents first, branch on count for combined vs per-agent prompt |
| `internal/cli/add_test.go` | Modify | Add `TestAddCommand_BulkUninstallConfirm` and `TestAddCommand_BulkUninstallDecline` tests |

## Interfaces / Contracts

No new types or interfaces. The `confirmFn func(string) bool` in `addHandler` is already defined.

New helper (inline in `runAddFlowInteractive`): no new exported functions.

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Bulk confirm yes — both agents uninstalled | Mock `detectAll` returns 2 installed agents; `runSelection` returns 0 (deselect all); `confirmFn` returns true; verify both `UninstallAgent` calls |
| Unit | Bulk confirm no — TUI restarts with all pre-checked | Same setup but `confirmFn` returns false; verify no uninstall calls; TUI called twice |
| Unit | Single agent deselected — existing 3-option flow | Existing `TestAddCommand_UninstallAppOnly` and `TestAddCommand_UninstallPromptCancel` still pass |
| Unit | No installed agents deselected — skip prompt | Existing `TestAddCommand_TUISuccessFlow` still passes |

## Migration / Rollout

No migration required. In-memory flow control change only.

## Open Questions

None.
