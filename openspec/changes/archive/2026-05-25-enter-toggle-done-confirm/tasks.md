# Tasks: Enter Toggles, Done Confirms

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 80-120 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

## Phase 1: Model Changes

- [x] 1.1 Add `IsDone bool` field to `AgentItem` struct
- [x] 1.2 In `newModel`, append Done sentinel `{ID: "_done", Name: "✓ Done", IsDone: true}` after all agents

## Phase 2: Key Handling

- [x] 2.1 In `handleSpecialKey`: Enter toggles like Space, Done item submits
- [x] 2.2 In `handleRuneKey`: same Enter behavior for rune-equivalent

## Phase 3: Rendering & Iteration

- [x] 3.1 In `renderAgentRow`: Done row shows "✓ Done" with bold, no checkbox
- [x] 3.2 In `selectedIDs`: skip Done item (last index)
- [x] 3.3 In `toggleAll`: skip Done item
- [x] 3.4 Update help bar to `"↑↓/jk navigate • space/enter toggle • q quit"`

## Phase 4: Tests

- [x] 4.1 TestEnterTogglesAgent — Enter flips checkbox
- [x] 4.2 TestEnterOnDoneConfirms — Done + Enter submits
- [x] 4.3 TestDoneItemRenders — View contains "✓ Done" with bold
- [x] 4.4 TestSelectedIDsExcludesDone — Done ID not in return
- [x] 4.5 TestToggleAllSkipsDone — Done unchecked state unchanged
- [x] 4.6 Update existing TestModel_EnterConfirms to use Done

## Phase 5: Verify

- [x] 5.1 Run `go test ./internal/tui/` — all tests pass
- [x] 5.2 Run `go build ./cmd/squad` — compiles clean
