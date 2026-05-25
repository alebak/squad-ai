# Exploration: TUI Agent Selection Redesign (Issue #21)

## Current State

The TUI agent selection screen in `internal/tui/model.go` uses a `AgentItem` struct with four semantic fields: `PreChecked`, `Blocked`, `BlockReason`, and `IsInstalled`. The Bubbletea model (`model`) manages cursor position and a `checked map[int]bool` for checkbox state.

### Current rendering

```
┌──────────────────────────────────────────────┐
│  🤖 Select Your AI Coding Agents             │
│                                              │
│  ○ Claude Code                               │
│  ○ OpenCode  ✅                              │
│  ○ Codex CLI  ⛔ requires Node.js 22+        │
│  ○ Aider                                     │
│                                              │
│  ↑↓/jk navigate • space toggle • a select all│
│  • enter confirm • q quit                    │
└──────────────────────────────────────────────┘
```

### Current init behavior
- `newModel()` pre-checks agents where `PreChecked && !Blocked`
- `buildAgentItemsForAdd()` sets `PreChecked` for compatible, not-installed, not-already-selected agents
- `IsInstalled` agents cannot be toggled (space is no-op)
- `Blocked` agents cannot be toggled

### Toggle-all (`a` key)
- `toggleAll()` flips all compatible (non-blocked, non-installed) agents
- If any compatible is unchecked → checks all
- If all compatible are checked → unchecks all
- `a` key is documented in the help bar

### Current emoji usage
- ✅ after agent name when `IsInstalled`
- ⛔ + `BlockReason` after agent name when `Blocked`
- 🤖 in title
- ◉ for checked checkbox

## Affected Areas

- `internal/tui/model.go` — Core changes: remove `PreChecked`/`IsInstalled`/emoji rendering, add select-all row, dynamic label, block reason in name
- `internal/tui/model_test.go` — Rewrite tests: new initial state (all unchecked), select-all row tests, no emoji assertions
- `internal/cli/add.go` — Remove `PreChecked` logic from `buildAgentItemsForAdd`, insert select-all Item as first element, remove emoji from `reportAddResults`
- `internal/cli/add_test.go` — Update test expectations matching new init state

## Approaches

### 1. Sentinel AgentItem with IsSelectAll flag (RECOMMENDED)

Add an `IsSelectAll bool` field to `AgentItem`. `buildAgentItemsForAdd()` prepends a sentinel item with `IsSelectAll=true`. The TUI model detects the sentinel in rendering and key handling.

**Changes to `AgentItem`**:
```
type AgentItem struct {
    ID          string
    Name        string
    Description string
    // PreChecked  REMOVED — all unchecked by default
    // IsInstalled REMOVED — no longer rendered
    Blocked     bool       // kept: prevents toggling
    BlockReason string     // kept: "(Node.js is required)" appended to Name
    IsSelectAll bool       // NEW: flags the sentinel row
}
```

**How it works**:
- `buildAgentItemsForAdd()` prepends `{Name: "select all", IsSelectAll: true, Blocked: false}`
- Rendering: select-all row gets dynamic checkbox + label "[ ] select all" / "[x] unselect all"
- Space on select-all row: calls `toggleAll()`
- Cursor wraps including the select-all row (it's item 0)
- Blocked agent's `BlockReason` is appended to `Name` in parentheses (done in `add.go` or in rendering)
- `a` key: removed from help bar text but `handleRuneKey` keeps the case `"a"` handler
- Help bar no longer shows "a select all"

**Display model**:
```
┌───────────────────────────────────────────────┐
│  Select AI Coding Agents                      │
│                                               │
│  [ ] select all                               │
│  ○ Claude Code                                │
│  ○ OpenCode                                   │
│  ○ Codex CLI (Node.js is required)            │
│  ○ Aider                                      │
│                                               │
│  ↑↓/jk navigate • space toggle • enter confirm│
│  • q quit                                     │
└───────────────────────────────────────────────┘
```

**Pros**:
- Minimal changes to existing architecture
- Select-all row participates in cursor navigation naturally
- Render logic is isolated to one conditional branch
- No new composable types or extra model fields

**Cons**:
- AgentItem gets a display-only flag that controls behavior
- Space semantics differ depending on whether cursor is on row 0
- Effort: Low

### 2. Model-level selectAll (no AgentItem flag)

Keep `AgentItem` unchanged. Add a `selectAllChecked` field to the `model` struct and a virtual header row concept. Rendering prepends the select-all row without adding it to `agents`.

**Pros**:
- AgentItem stays clean, no sentinel flag

**Cons**:
- Cursor indexing gets more complex (cursor 0 is virtual, items[0] is at cursor 1)
- `selectedIDs()` must skip the select-all state
- Space toggling needs special-casing for cursor 0
- More coupling in the model — two tracking mechanisms for one concept
- Effort: Medium

### 3. CLI-only approach (TUI doesn't know about select-all)

CLI layer computes select-all state and passes it to TUI. TUI just renders whatever items it receives — the select-all row is just an `AgentItem` that the TUI renders differently.

**Pros**:
- Keeps TUI dumber
- CLI controls all semantics

**Cons**:
- Dynamic label ("select all" ↔ "unselect all") requires TUI to know whether all agents are checked → logic lives in TUI anyway
- Space toggle on select-all row must toggle all other agents → TUI needs select-all awareness
- Pushes display logic to wrong layer (CLI shouldn't format UI strings)
- Effort: Medium

## Recommendation

**Approach 1** — Sentinel AgentItem with `IsSelectAll`. It's the minimal diff approach that keeps the architecture intact. The select-all row is just another agent item with special rendering and toggle semantics — Go sentinel patterns are well-understood and the changes are contained to `renderAgentRow()`, `handleSpecialKey()`, and `buildAgentItemsForAdd()`.

The `a` key shortcut removal from help bar is a one-line change in the `View()` method help string. The `PreChecked` field removal simplifies `newModel()` initialization logic significantly. The `BlockReason` rendering in parentheses is cleaner than the current emoji approach.

### Detailed design sketch

**AgentItem changes**:
```go
type AgentItem struct {
    ID          string
    Name        string
    Description string
    Blocked     bool
    BlockReason string
    IsSelectAll bool   // NEW
}
// PreChecked and IsInstalled removed
```

**newModel()** — simplified: no pre-check loop, just initialize empty `checked` map.

**renderAgentRow()** — for select-all:
```go
if agent.IsSelectAll {
    // Show [ ] select all or [x] unselect all
    // Checked state: true when ALL compatible (non-blocked, non-installed) are checked
    // Or maybe: true when ANY are checked? Need UX clarity.
    // RECOMMENDED: checked when at least one compatible agent is checked
}
```

**handleSpecialKey()** — space on select-all:
```go
case tea.KeySpace:
    if m.agents[m.cursor].IsSelectAll {
        m.toggleAll()
    } else if !m.agents[m.cursor].Blocked {
        m.checked[m.cursor] = !m.checked[m.cursor]
    }
```

**buildAgentItemsForAdd()** — prepend select-all:
```go
items := []tui.AgentItem{{
    Name: "select all",
    IsSelectAll: true,
}}
for _, agent := range catalog.Agents {
    items = append(items, tui.AgentItem{
        ID:   agent.ID,
        Name: agent.Name,
        Blocked: blocked,
        BlockReason: reason,
    })
}
```

**View()** — remove emoji from title and help bar:
```go
// Remove "a select all" from help bar but keep in handleRuneKey
```

## Risks

- **Cursor wrapping with select-all row**: When there are many agents, wrapping from the bottom to row 0 (select-all) is a behavior change. Currently wrapping goes to the last agent. With select-all as first row, it wraps to "select all" which is correct UX but tests need updating.
- **Installed agents still non-togglable**: The `IsInstalled` field is removed from the `AgentItem` struct, so the TUI no longer knows which agents are installed. This means installed agents become togglable (space works on them). **This is a design decision to clarify**: should installed agents be togglable or not? Current behavior blocks it. If we keep the block, we need a way to communicate "already installed" to the TUI without rendering it. Options: (a) keep `IsInstalled` for logic only (not rendered), (b) filter installed agents out before passing to TUI, (c) let them be togglable (user can "select" an installed agent but it won't be installed again — handled in `runAddFlowInteractive` via `filterInstalled`).
- **Dynamic label behavior**: The checked state of the select-all row needs a clear definition. Two options: (a) checked when ALL compatible agents are checked (toggleAll state), (b) checked when AT LEAST ONE compatible agent is checked. Option (a) is more intuitive — it reflects the toggleAll state.
- **Blocked agent toggling**: Currently blocked agents cannot be toggled. This stays the same. The `Blocked` field is preserved.

## Ready for Proposal

Yes — with one clarification needed (see Risk #2 about installed agent togglability). The orchestrator should ask the user: "Should installed agents still be non-togglable (space no-op), or should we allow toggling them (they just won't be re-installed)?" This decision affects whether `IsInstalled` stays on `AgentItem` for logic-only use or gets fully removed.
