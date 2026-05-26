# Design: Fix TUI version header hardcoding and wizard submit flow

## Architecture Decision: Version Propagation

**Decision**: Pass `version` string from `cli` package directly to `tui.RunSelection` as a parameter.

**Alternatives considered**:
1. **Package-level export**: Export `cli.Version` as a public constant. Rejected because `version` is intentionally unexported (set via ldflags only).
2. **Config struct**: Add version to `config.Config`. Over-engineered — version is a build-time constant, not config.
3. **Parameter**: Simple, explicit, no coupling. Accepted.

**Flow**: `cli.version` → `addHandler.runSelection(... version)` → `tui.RunSelection(agents, version)` → `newModel(agents, version)` → stored in `model.version` → used in `View()`.

## Architecture Decision: Wizard Summary Submit

**Decision**: In `handleSummaryKey()`, when Enter is pressed on Apply (cursor=0), directly mutate the model and return `tea.Quit` instead of calling `completeWizard()` which returns to agent list.

**Rationale**: The existing `handleSummaryKey` returns `(tea.Model, tea.Cmd)` via `handleWizardKey` → `handleKeyMsg`. To quit, the function needs to return `tea.Quit` as the `Cmd`. This requires modifying the call chain.

**Alternatives considered**:
- Reuse `completeWizard()` + add a flag to trigger quit next Update. Rejected because it adds unnecessary state.
- Modify `handleSummaryKey` to set `submitted=true`, clear `wizard`, and call `tea.Quit`. This is the minimal change.

## Sequence: Bug 1 — Version Header

```
cli/add.go: runAddFlowInteractive()
  │  h.runSelection(agentItems)  
  │    └─ tui.RunSelection(agents, version)  // ← NEW: version param
  │         └─ newModel(agents, version)      // ← NEW: version param
  │              └─ model{ ..., version: version }
  │                   └─ View(): fmt.Sprintf("Squad AI (version %s)", m.version)
  ▼
Renders dynamic version instead of hardcoded "0.15.0"
```

## Sequence: Bug 2 — Wizard Submit

Before (broken):
```
handleSummaryKey() → Enter on Apply
  → completeWizard()  // sets wizardOut, wizard=nil
  → returns (m, nil)  // returns to agent list
  → user must press Apply again
  → handleSpecialKey() → submitSelection() → isSubmitted=true
```

After (fixed):
```
handleSummaryKey() → Enter on Apply
  → completeWizard()   // sets wizardOut
  → m.isSubmitted = true
  → m.wizard = nil
  → return m, tea.Quit  // QUITS IMMEDIATELY
```

## Interface Changes

### `internal/tui/model.go`

```go
type model struct {
    // ... existing fields ...
    version string  // NEW: build version from cli package
}

// BEFORE: func newModel(agents []AgentItem) model
// AFTER:  func newModel(agents []AgentItem, version string) model

// BEFORE: func RunSelection(agents []AgentItem) ([]string, map[string]int, error)
// AFTER:  func RunSelection(agents []AgentItem, version string) ([]string, map[string]int, error)
```

### `internal/cli/add.go`

```go
type addHandler struct {
    // ... existing fields ...
    // runSelection signature changes from:
    //   func(items []tui.AgentItem) ([]string, map[string]int, error)
    // to:
    //   func(items []tui.AgentItem, version string) ([]string, map[string]int, error)
}

// In buildAgentItemsForAdd or runAddFlowInteractive:
//   h.runSelection(agentItems, version)  // was: h.runSelection(agentItems)
```

### `handleSummaryKey` change

```go
// BEFORE — Enter on Apply (cursor == 0):
//   m.completeWizard()  // returns to agent list

// AFTER:
//   m.completeWizard()  // still populate wizardOut
//   m.isSubmitted = true
//   m.wizard = nil
//   return m, tea.Quit  // immediately quit
```

## Test Changes

### `internal/tui/model_test.go`
- Every call to `newModel(testAgents())` → `newModel(testAgents(), "0.1.0")`
- `TestModel_ViewRenders`: assert.Contains(view, "version 0.1.0") instead of "version 0.15.0"
- `TestModel_HeaderShowsVersion`: same
- `TestModel_WizardCompleteThroughSummary`: add assertion that Apply on summary sets isSubmitted
- `TestModel_SummaryApplyCompletesWizard`: verify isSubmitted and no wizard after Apply
- `TestModel_EmptyAgents`: `RunSelection(nil, "0.1.0")`

### `internal/cli/add_test.go`
- All `runSelection` mock fields update signature to `func(items []tui.AgentItem, version string) ([]string, map[string]int, error)`
- The mock implementations remain functionally identical — just accept and ignore the new `version` parameter

## Affected Files Summary

| File | Changes |
|------|---------|
| `internal/tui/model.go` | Add version field, update newModel/RunSelection signatures, fix View(), fix handleSummaryKey |
| `internal/tui/model_test.go` | Update all newModel calls, version assertions, wizard submit test |
| `internal/cli/add.go` | Pass `version` var to RunSelection calls |
| `internal/cli/add_test.go` | Update all runSelection mock signatures |
