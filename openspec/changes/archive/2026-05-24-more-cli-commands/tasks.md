# Tasks: Remaining CLI Commands (add, remove, update, info)

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~700-850 (8 new files + 1 modified) |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1: add + remove (simpler), PR 2: update + info |
| Delivery strategy | ask-on-risk |
| Chain strategy | stacked-to-main |

Decision needed before apply: No (user explicitly requested all phases)
Chained PRs recommended: Yes
Chain strategy: stacked-to-main
400-line budget risk: High

## Phase 1: Add + Remove Commands

- [ ] 1.1 Create `internal/cli/add.go` — addHandler, defaultAddHandler, newAddCommandWithHandler
- [ ] 1.2 Create `internal/cli/add_test.go` — tests: shows available, empty registry, fetch failure
- [ ] 1.3 Create `internal/cli/remove.go` — removeHandler, defaultRemoveHandler, newRemoveCommandWithHandler
- [ ] 1.4 Create `internal/cli/remove_test.go` — tests: removes agent, not in config, missing arg, save failure

## Phase 2: Update + Info Commands

- [ ] 2.1 Create `internal/cli/update.go` — updateHandler, defaultUpdateHandler, newUpdateCommandWithHandler
- [ ] 2.2 Create `internal/cli/update_test.go` — tests: successful update, fetch failure
- [ ] 2.3 Create `internal/cli/info.go` — infoHandler, defaultInfoHandler, newInfoCommandWithHandler
- [ ] 2.4 Create `internal/cli/info_test.go` — tests: shows details, not found, missing arg, fetch failure

## Phase 3: Wire and Verify

- [ ] 3.1 Modify `internal/cli/root.go` — register add, remove, update, info commands
- [ ] 3.2 Build: `go build ./cmd/squad`
- [ ] 3.3 Vet: `go vet ./...`
- [ ] 3.4 Test: `go test ./... -v -count=1`
