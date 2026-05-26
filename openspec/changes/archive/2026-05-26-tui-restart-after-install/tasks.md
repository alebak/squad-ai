# Tasks: TUI Restart After Agent Installation

## Delivery Strategy

- **Single PR**: ~80 lines estimated. Well under 400-line review budget.
- **Decision needed before apply**: No
- **Chained PRs recommended**: No
- **400-line budget risk**: Low

## Task 1: Move installation logic inside the loop in `runAddFlowInteractive`

**File**: `internal/cli/add.go`

1. Remove the `break` at line 253 and the block below it (lines 256–275)
2. Move the installation logic INSIDE the for-loop body, after the `len(selectedIDs) == 0` check
3. After install:
   - Mark succeeded agents in the `installed` map
   - Rebuild `agentItems` with `buildAgentItemsForAdd(h, catalog, installed)`
   - `continue` instead of `break`
4. When `len(toInstall) == 0` (all selected already installed): rebuild `agentItems` and `continue`
5. Keep the `selectedIDs == nil` quit check and empty selection exit as-is

**Verification**: `go build ./...` compiles cleanly

## Task 2: Update existing tests for the restart behavior

**File**: `internal/cli/add_test.go`

Update tests whose `runSelection` mocks only return once:

| Test | Change |
|------|--------|
| `TestAddCommand_TUISuccessFlow` | Add callCount; return nil (quit) on second call; assert callCount == 2 |
| `TestAddCommand_UninstallViaWizardAppOnly` | Add callCount; return nil on third call; assert callCount == 3 |
| `TestAddCommand_UninstallViaWizardAppAndConfig` | Add callCount; return nil on third call; assert callCount == 3 |
| `TestAddCommand_WizardRestartsTUIAfterUninstall` | Add callCount; return nil on third call; assert callCount == 3 |

## Task 3: Add new test for install restart

**File**: `internal/cli/add_test.go`

Add `TestAddCommand_TUIRelaunchAfterInstall`:
- Mock `runSelection` to return IDs on first call, nil on second
- Verify TUI is called exactly 2 times
- Verify output contains install messages
- Verify `installAll` was called with expected agents

## Verification

```bash
go test ./internal/cli/ -run TestAddCommand -v -count=1
go build ./...
```
