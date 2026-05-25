# Tasks: TUI Restart After Uninstall

## Metadata

- **Delivery strategy**: single-pr (small change, <50 lines)
- **Decision needed before apply**: No
- **Chained PRs recommended**: No
- **400-line budget risk**: Low (~20 lines changed)

## Task List

### Task 1: Add `needsRestart = true` in per-agent uninstall cases

**File**: `internal/cli/add.go`
**Lines**: 275-299 in `runAddFlowInteractive`

Add `needsRestart = true` after both `uninstallAppOnly` and `uninstallAppConfig` cases in the per-agent 3-option switch. This is a 2-line change.

**Checklist:**
- [x] Add `needsRestart = true` after `uninstallAppOnly` block (after `delete(installed, agent.ID)`)
- [x] Add `needsRestart = true` after `uninstallAppConfig` block (after `h.uninstallConfig(agent)`)

### Task 2: Update `TestAddCommand_UninstallAppOnly`

**File**: `internal/cli/add_test.go`

The test mock `runSelection` currently returns once. After the fix, `runSelection` will be called twice (first for initial selection, second after TUI restart). Update the mock to return on the second call as well.

**Checklist:**
- [x] Make `runSelection` track call count (like existing `TestAddCommand_UninstallPromptCancel`)
- [x] On first call: return `["opencode"]` (deselect claude-code)
- [x] On second call: return `["opencode", "claude-code"]` (user accepts after restart)
- [x] Assert `runSelection` is called exactly 2 times
- [x] Assert "Installing selected agents" still appears in output

### Task 3: Update `TestAddCommand_UninstallAppAndConfig`

**File**: `internal/cli/add_test.go`

Same pattern as Task 2 — `runSelection` must return twice.

**Checklist:**
- [x] Make `runSelection` track call count
- [x] On first call: return `["opencode"]` (deselect claude-code)
- [x] On second call: return `["opencode", "claude-code"]` (user accepts after restart)
- [x] Assert `runSelection` is called exactly 2 times
- [x] Assert uninstall output still appears

### Task 4: Update `TestAddCommand_BulkUninstallConfirm`

**File**: `internal/cli/add_test.go`

`runSelection` must return twice — first to deselect both agents, second after restart.

**Checklist:**
- [x] Make `runSelection` track call count
- [x] On first call: return `[]` (deselect both)
- [x] On second call: return `["opencode", "claude-code"]` (user accepts after restart)
- [x] Assert `runSelection` is called exactly 2 times
- [x] Assert uninstall output still appears

### Task 5: Run full test suite

**Checklist:**
- [x] `go test ./... -v -count=1` passes
- [x] `go build ./cmd/squad` compiles
