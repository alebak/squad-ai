# Design: TUI Agent Selection Redesign

## Technical Approach

Sentinel `IsSelectAll` field on `AgentItem` prepended as first row by `buildAgentItemsForAdd()`. Model's `newModel()` drops all `PreChecked` logic. `renderAgentRow()` detects the sentinel and renders dynamic label. Space on sentinel row calls `toggleAll()`. Blocked agents render `Name (BlockReason)` in `renderAgentRow()` instead of emoji. Installed agents lose the toggle guard. `add.go`'s interactive flow checks for deselected installed agents and prompts uninstall.

## Architecture Decisions

| Decision | Choice | Alternatives | Rationale |
|----------|--------|-------------|-----------|
| Select-all sentinel | `IsSelectAll bool` on `AgentItem` | Virtual header row in model | Sentinel keeps cursor indexing clean; row 0 is real, no offset math |
| Installed togglability | Remove `IsInstalled` from toggle guard | Keep `IsInstalled` for logic only | Spec requires togglability; `add.go` handles uninstall prompt post-TUI |
| PreChecked removal | Delete `PreChecked` field entirely | Keep field as no-op | Dead fields confuse; Go has no unused field warning but it's tech debt |
| Dynamic label logic | Checked when ALL compatible checked | Checked when ANY checked | Mirrors `toggleAll()` semantics; consistent UX |

## Data Flow

```
buildAgentItemsForAdd()
  ├── Prepend select-all sentinel (IsSelectAll=true)
  ├── For each agent: set Blocked/BlockReason (NO PreChecked, NO IsInstalled)
  └── Return items to TUI

TUI (model.go)
  ├── newModel()   → checked = empty{} (no PreChecked loop)
  ├── handleKeyMsg → space on index 0 → toggleAll()
  │                 → space on other indices → toggle if !Blocked
  ├── renderAgentRow() → IsSelectAll → "[ ]/ [x] select all" / "unselect all"
  │                     → Blocked → "Name (reason)" in faint style
  │                     → default → "○/◉ Name"
  └── View() → title without emoji, help without "a select all"

runAddFlowInteractive()
  ├── After TUI returns selectedIDs
  ├── For each previously-installed agent NOT in selectedIDs
  │   └── Prompt: "Uninstall <name>? [y/N]" → if yes, call UninstallAgent
  └── Install newly selected agents
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modify | AgentItem: remove `PreChecked`, `IsInstalled`; add `IsSelectAll`. Rewrite newModel, renderAgentRow, toggleAll, handleSpecialKey, View |
| `internal/tui/model_test.go` | Modify | All tests: remove PreChecked assertions, add select-all tests, remove emoji tests, add toggle tests |
| `internal/cli/add.go` | Modify | buildAgentItemsForAdd: prepend sentinel, no PreChecked. runAddFlowInteractive: add uninstall prompt for deselected installed agents |
| `internal/cli/add_test.go` | Modify | Update TUI mock expectations if needed |

## Interfaces / Contracts

```go
// AgentItem — modified struct
type AgentItem struct {
    ID          string
    Name        string
    Description string
    Blocked     bool
    BlockReason string
    IsSelectAll bool   // NEW: sentinel for select-all row
    // PreChecked  REMOVED
    // IsInstalled REMOVED
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | newModel initial state | All unchecked, cursor at 0, agents[0] is select-all |
| Unit | Select-all toggle | Space on index 0 toggles all compatible |
| Unit | Dynamic label | Label is "select all" / "unselect all" based on checked state |
| Unit | Blocked agent rendering | No ⛔ emoji, "(reason)" appended |
| Unit | No emoji in title/help | View() output assertions |
| Unit | Installed toggleable | Space toggles installed agent on/off |

## Migration / Rollout

No migration required. Feature toggle not needed — this is a pure UI change.

## Open Questions

- None — all design decisions resolved in exploration and user requirements.
