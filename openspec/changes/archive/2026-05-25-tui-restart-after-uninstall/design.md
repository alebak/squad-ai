# Design: TUI Restart After Uninstall

## Problem

In `runAddFlowInteractive` (`internal/cli/add.go`), the TUI selection+uninstall loop only restarts when `uninstallCancel` is chosen. When the user chooses `uninstallAppOnly` or `uninstallAppConfig`, the code performs the uninstall (deleting from the `installed` map) but falls through the loop without setting `needsRestart = true`. The loop exits via `break`, proceeding to installation and eventually exiting the CLI. The user never sees the updated TUI state.

## Solution

Set `needsRestart = true` after every uninstall action in the per-agent 3-option switch (both `uninstallAppOnly` and `uninstallAppConfig` cases). The existing restart machinery (lines 322-327 of `add.go`) already handles rebuilding `agentItems` via `buildAgentItemsForAdd(h, catalog, installed)` and re-running the loop.

## Flow Diagram

```
TUI Launch → User Selection → Deselected installed agents?
  ├── No → Proceed to installation → break → install → exit
  └── Yes → Show prompt
       ├── Bulk confirm → uninstall all → needsRestart=true → continue loop
       ├── Bulk decline → needsRestart=true → continue loop
       └── Per-agent 3-option
            ├── [1] Uninstall app → needsRestart=true ← FIX
            ├── [2] Uninstall app+config → needsRestart=true ← FIX
            └── [3] Cancel → needsRestart=true (existing)

Loop continues → rebuild agentItems → TUI re-launches
Loop breaks when deselected is empty (user accepted state)
```

## Code Change

### `internal/cli/add.go` — `runAddFlowInteractive`

**Before (lines 275-299):**
```go
case uninstallAppOnly:
    // ... uninstall logic ...
    // MISSING: needsRestart = true
case uninstallAppConfig:
    // ... uninstall logic ...
    // MISSING: needsRestart = true
case uninstallCancel:
    needsRestart = true
```

**After:**
```go
case uninstallAppOnly:
    // ... uninstall logic ...
    needsRestart = true  // ← ADDED
case uninstallAppConfig:
    // ... uninstall logic ...
    needsRestart = true  // ← ADDED
case uninstallCancel:
    needsRestart = true
```

### `internal/cli/add_test.go` — 3 tests need updating

1. **`TestAddCommand_UninstallAppOnly`**: `runSelection` must return twice. First call returns `["opencode"]` (deselecting "claude-code"). After restart, `installed` no longer has "claude-code", so no agents are deselected. Second call can return `["opencode", "claude-code"]` or `["opencode"]` — nothing matters since no agents are installed.

2. **`TestAddCommand_UninstallAppAndConfig`**: Same pattern as above.

3. **`TestAddCommand_BulkUninstallConfirm`**: `runSelection` must return twice. First call returns `[]` (deselecting both). After confirm and restart, no agents are installed (both uninstalled). Second call needs to not deselect anything.
