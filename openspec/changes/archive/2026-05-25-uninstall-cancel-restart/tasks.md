# Tasks: Uninstall Cancel Returns to TUI

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~60-80 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | single-pr |

```
Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: single-pr
400-line budget risk: Low
```

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Fix cancel behavior & update tests | PR 1 | Single self-contained change |

## Phase 1: Core Implementation

- [x] 1.1 `internal/cli/add.go`: Wrap TUI selection + uninstall in outer `for` loop
- [x] 1.2 `internal/cli/add.go`: Track `needsRestart` flag when Cancel is chosen
- [x] 1.3 `internal/cli/add.go`: On restart, rebuild `agentItems` via `buildAgentItemsForAdd` with updated `installed` map
- [x] 1.4 `internal/cli/add.go`: Delete uninstalled agents from `installed` map immediately after successful uninstall
- [x] 1.5 `internal/cli/add.go`: Remove the `selectedIDs`/`selectedSet` restoration from `uninstallCancel` case (no longer needed — restart gives fresh selection)

## Phase 2: Test Updates

- [x] 2.1 `internal/cli/add_test.go`: Update `TestAddCommand_UninstallPromptCancel` — use call counter for `runSelection` mock to return different results on 2nd invocation
- [x] 2.2 `internal/cli/add_test.go`: Verify that when cancel is chosen and user re-selects, flow proceeds past uninstall to installation
- [x] 2.3 `internal/cli/add_test.go`: Verify no uninstall messages appear when all choices are Cancel

## Phase 3: Verify

- [ ] 3.1 Run `go test ./internal/cli/ -v -count=1` — all tests pass
- [ ] 3.2 Run `go test ./... -count=1` — all project tests pass
- [ ] 3.3 Run `go build ./cmd/squad` — builds clean
