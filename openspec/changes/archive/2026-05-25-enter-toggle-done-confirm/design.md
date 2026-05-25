# Design: Enter Toggles, Done Confirms

## Technical Approach

Add `IsDone` field to `AgentItem` to mark the sentinel Done row. Append it in
`newModel` after all real agents. In `handleSpecialKey`, delegate Enter to the
same toggle path as Space, adding a Done-item check. Update all iteration
points (`selectedIDs`, `toggleAll`, `renderAgentRow`, `View`) to skip the Done
item. Update the help bar string.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|--------------|-----------|
| IsDone field | Add bool to AgentItem | Separate struct, magic index | Consistent with existing IsSelectAll pattern; minimal diff |
| Done sentinel position | Last index | First, after select-all | Natural reading order: agents then confirm |
| Done label | "✓ Done" | "Confirm", "Submit" | Checkmark signals completeness; matching existing ◉/○ style language |
| Enter toggle logic | Reuse Space path | Separate case block | Same behavior, zero duplication — only add Done guard |

## Data Flow

```
newModel(agents)
  → append {ID: "_done", Name: "✓ Done", IsDone: true}
  → set checked[Done index] = false (never checkable)

handleSpecialKey / handleRuneKey
  ├── KeyEnter │ KeySpace
  │   ├── IsDone → submit=true, tea.Quit
  │   ├── IsSelectAll → toggleAll()
  │   └── default → toggle m.checked[i]
  └── other keys → unchanged

selectedIDs()
  ├── a.IsSelectAll → skip
  ├── a.IsDone → skip
  └── m.checked[i] → collect ID

toggleAll()
  ├── a.IsSelectAll → skip
  ├── a.IsDone → skip
  ├── a.Blocked → skip
  └── toggle m.checked[i]
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modify | AgentItem: add IsDone; newModel: append sentinel; handleSpecialKey: Enter toggles, Done confirms; handleRuneKey: same; renderAgentRow: Done rendering; selectedIDs: skip Done; toggleAll: skip Done; View: update help |
| `internal/tui/model_test.go` | Modify | Add 5-6 new tests; update Enter-confirm tests to use Done |

## Interfaces / Contracts

```go
// AgentItem — adds IsDone field
type AgentItem struct {
    ID          string
    Name        string
    Description string
    Blocked     bool
    BlockReason string
    IsSelectAll bool
    PreChecked  bool
    IsDone      bool   // NEW: sentinel for the confirm-and-exit row
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Enter toggles agent | UpdateModelKey(tea.KeyEnter), assert checked[i] flipped |
| Unit | Enter on Done confirms | updateModel calls, assert isSubmitted |
| Unit | Done renders correctly | View assertion for "✓ Done" and no checkbox |
| Unit | selectedIDs skips Done | Assert Done ID not in returned list |
| Unit | toggleAll skips Done | Assert Done's checked state unchanged |
| Unit | Cursor wraps with Done | Assert len(m.agents) includes Done |

## Migration / Rollout

No migration required. Change is purely behavioral in the TUI layer.

## Open Questions

None.
