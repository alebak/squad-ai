# Tasks: Fix TUI version header hardcoding and wizard submit flow

## Phase 1: Version Propagation — `internal/tui/model.go`

### 1.1 Add version field to model struct
- [ ] Add `version string` field to model struct
- Location: `internal/tui/model.go` line ~99 (inside model struct)

### 1.2 Update newModel signature
- [ ] Change `func newModel(agents []AgentItem) model` to `func newModel(agents []AgentItem, version string) model`
- [ ] Store `version` in the model struct inside newModel

### 1.3 Update RunSelection signature
- [ ] Change `func RunSelection(agents []AgentItem)` to `func RunSelection(agents []AgentItem, version string)`
- [ ] Pass version to newModel: `newModel(agents, version)`

### 1.4 Update View to render dynamic version
- [ ] Replace hardcoded `" (version 0.15.0)"` with `fmt.Sprintf(" (version %s)", m.version)`

## Phase 2: Wizard Summary Submit — `internal/tui/model.go`

### 2.1 Fix handleSummaryKey Apply handler
- [ ] In `handleSummaryKey()`, when `ws.cursor == 0` (Apply):
  - [ ] Call `m.completeWizard()` (to populate wizardOut)
  - [ ] Set `m.isSubmitted = true`
  - [ ] Set `m.wizard = nil`
  - [ ] Return `m, tea.Quit`
- Note: `handleSummaryKey` must return `(tea.Model, tea.Cmd)` — currently it returns `(model, tea.Cmd)` but handleWizardKey's caller casts to `tea.Model`. Ensure return type matches.

## Phase 3: CLI Wiring — `internal/cli/add.go`

### 3.1 Pass version to RunSelection in runAddFlowInteractive
- [ ] In `runAddFlowInteractive()`, change `h.runSelection(agentItems)` to `h.runSelection(agentItems, version)`
- The `version` var is a package-level variable in `cli` package (defined in root.go)

## Phase 4: Update Tests — `internal/tui/model_test.go`

### 4.1 Update all newModel calls
- [ ] Update every `newModel(agents)` call to `newModel(agents, "0.1.0")`
- Search for all occurrences and update

### 4.2 Update RunSelection calls in tests
- [ ] Update `RunSelection(nil)` to `RunSelection(nil, "0.1.0")`
- [ ] Update `RunSelection([]AgentItem{})` to `RunSelection([]AgentItem{}, "0.1.0")`

### 4.3 Update version assertions
- [ ] `TestModel_ViewRenders`: Change `assert.Contains(t, view, "version 0.15.0")` to `assert.Contains(t, view, "version 0.1.0")`
- [ ] `TestModel_HeaderShowsVersion`: Change assertion to check for dynamic version

### 4.4 Add wizard submit assertions
- [ ] `TestModel_SummaryApplyCompletesWizard`: Add assertions for `m.isSubmitted` and verify `wizard` is nil
- [ ] `TestModel_WizardCompleteThroughSummary`: Verify isSubmitted after Apply on summary

## Phase 5: Update Tests — `internal/cli/add_test.go`

### 5.1 Update all runSelection mock signatures
- [ ] Every `runSelection: func(items []tui.AgentItem)` → `func(items []tui.AgentItem, version string)`
- [ ] The implementation bodies stay the same — just discard the version parameter
- Occurrences: 8 fields in add_test.go

## Phase 6: Verify

### 6.1 Run all tests
- [ ] `go test ./internal/tui/... -v -count=1`
- [ ] `go test ./internal/cli/... -v -count=1`
- [ ] `go build ./cmd/squad`

### 6.2 Code review
- [ ] Verify no remaining hardcoded "0.15.0" in tui package
- [ ] Verify wizard submit returns tea.Quit immediately

## Guard: Delivery Strategy

**Decision needed before apply**: No
**Chained PRs recommended**: No (under 200 lines, single concern)
**400-line budget risk**: Low

Grouped by package and logical dependency — `tui/model.go` first, then `cli/add.go`, then tests in parallel.
