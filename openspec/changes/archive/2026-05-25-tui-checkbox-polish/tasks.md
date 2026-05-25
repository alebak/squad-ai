# Tasks: TUI Checkbox Polish & Pre-Checked Agents

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~80 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Single PR covering all changes | PR 1 | ~80 lines, single concern |

## Phase 1: Core Logic

- [ ] 1.1 `internal/tui/model.go`: Add `PreChecked bool` to `AgentItem`
- [ ] 1.2 `internal/tui/model.go`: `newModel` — iterate agents, check `PreChecked && !Blocked && !IsSelectAll`
- [ ] 1.3 `internal/tui/model.go`: `renderSelectAllRow` — replace `[x]`/`[ ]` with `◉`/`○`
- [ ] 1.4 `internal/tui/model.go`: `View` — add blank line after select-all row (use `\n\n`)

## Phase 2: Wiring

- [ ] 2.1 `internal/cli/add.go`: Change `buildAgentItemsForAdd` signature — replace `_ map[string]bool` with `installed map[string]bool`
- [ ] 2.2 `internal/cli/add.go`: `buildAgentItemsForAdd` — set `PreChecked: installed[agent.ID] && !blocked`
- [ ] 2.3 `internal/cli/add.go`: `runAddFlow` — call `detectAll` before `buildAgentItemsForAdd`, pass `installed` to both
- [ ] 2.4 `internal/cli/add.go`: `runAddFlowInteractive` — accept `installed map[string]bool` param, skip re-detection

## Phase 3: Tests

- [ ] 3.1 `internal/tui/model_test.go`: Update `testAgents` — add `PreChecked` for some agents
- [ ] 3.2 `internal/tui/model_test.go`: Update `TestModel_InitialState` — PreChecked agents start checked
- [ ] 3.3 `internal/tui/model_test.go`: Update `TestModel_ViewRenders` — assert ◉/○ not [x]/[ ]
- [ ] 3.4 `internal/tui/model_test.go`: Update `TestModel_SelectAllDynamicLabel` — assert ◉/○ style
- [ ] 3.5 `internal/tui/model_test.go`: Add blank-line assertion to View tests
- [ ] 3.6 `internal/cli/add_test.go`: Update handlers for new signature
