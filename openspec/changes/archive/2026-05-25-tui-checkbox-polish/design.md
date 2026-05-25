# Design: TUI Checkbox Polish & Pre-Checked Agents

## Technical Approach

Four-file change package with two concerns that share a common struct:

1. **Checkbox consistency**: Replace `[x]`/`[ ]` in `renderSelectAllRow` with `◉`/`○`, matching `renderAgentRow`. Add `\n` after select-all row in `View()`.
2. **Pre-checked agents**: Add `PreChecked bool` to `AgentItem`. `newModel` reads it to set initial `checked` state (only for compatible, non-select-all agents). `buildAgentItemsForAdd` sets it from `detectAll` results.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| How to pass installed info to `buildAgentItemsForAdd` | Replace unused `_ map[string]bool` param with `installed map[string]bool` | Re-detect inside function, add new param | No extra detection call, param was already dead |
| Where to detect installed agents | Move `detectAll` from `runAddFlowInteractive` to `runAddFlow` | Detect again in `buildAgentItemsForAdd` | Single detection, share with uninstall prompt |
| Pre-checked logic | `PreChecked = installed[id] && !blocked` | Always pre-check installed | Blocked agents can't toggle — checking them would confuse |

## Data Flow

```
runAddFlow()
  ├── fetchRegistry() → catalog
  ├── detectAll(catalog.Agents) → installed map  ← MOVED EARLIER
  ├── buildAgentItemsForAdd(h, catalog, installed) → agentItems[]
  │     └── for each agent:
  │           blocked = !isRuntimeMet
  │           PreChecked = installed[id] && !blocked  ← NEW
  ├── runAddFlowInteractive(h, cmd, agentItems, catalog, cfg, cfgPath, installed)
  │     └── (no longer re-detects — receives installed from caller)
  │
  └── runAddFlowNonInteractive(h, cmd, agentItems, cfg)

newModel(agents)
  └── for each agent:
        if PreChecked && !Blocked && !IsSelectAll:
          checked[i] = true

renderSelectAllRow(cursor)  ← UPDATED
  └── ◉ (checked) / ○ (unchecked) instead of [x] / [ ]

View()
  └── after select-all row: write "\n\n" instead of "\n"
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modify | +PreChecked, newModel init, renderSelectAllRow ◉/○, View separator |
| `internal/cli/add.go` | Modify | buildAgentItemsForAdd signature+logic, detectAll moved |
| `internal/tui/model_test.go` | Modify | Update assertions for ◉/○, add PreChecked, test initial state |
| `internal/cli/add_test.go` | Modify | Update test handlers for new buildAgentItemsForAdd signature |

## Interfaces / Contracts

```go
// AgentItem — added field
type AgentItem struct {
    // ...existing fields...
    PreChecked bool   // true if agent is installed and compatible
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Checkbox style | Render test asserts ◉/○ in view, not [x]/[ ] |
| Unit | PreChecked init | newModel + testAgents with PreChecked checks initial state |
| Unit | Blank separator | View contains two newlines after select-all |
| Unit | Blocked not pre-checked | Test agent with Blocked=true + PreChecked=false |
| Integration | buildAgentItemsForAdd | Existing testRegistry + handler mocks verify flow |

## Migration / Rollout

No migration required. Single atomic change.

## Open Questions

None.
