# Exploration: TUI First-Run Agent Selection

## Current State

The CLI package (`internal/cli/`) has a complete set of commands (install, list, add, remove, info, update, version). The `add` command currently has a "TUI coming soon" stub that prints agent IDs in a simple table.

The TUI package (`internal/tui/`) exists as a directory with only a `.gitkeep` file — no code yet.

Existing internal packages:
- `internal/config/` — config read/write with `ConfigPath()`, `Load()`, `Save()`
- `internal/registry/` — `Catalog`, `Agent`, `RuntimeDep` types + `Fetch()`, `LoadCache()`, `SaveCache()`
- `internal/installer/` — `InstallAgent()`, `InstallAll()`, `DetectAgent()`, `DetectAll()`
- `internal/runtime/` — `DetectNode()`, `DetectGo()`, `DetectPython()`, `IsCompatible()`

Dependencies (before this change): cobra, testify
Dependencies (after): + bubbletea, lipgloss, bubbles

All existing CLI commands follow the **handler struct with injectable function fields** pattern for testability. No global mutable state.

## Affected Areas

- `internal/tui/model.go` — **Create**: Bubbletea Model for agent selection list
- `internal/tui/model_test.go` — **Create**: Unit tests for the TUI model
- `internal/cli/add.go` — **Modify**: Wire TUI into add command, replace stub
- `internal/cli/root.go` — **Modify**: Wire the default `squad` root flow for first-run
- `internal/tui/.gitkeep` — **Remove**: No longer needed
- `go.mod` / `go.sum` — **Update**: Added bubbletea, lipgloss, bubbles

## Approaches

1. **Simple tea.Model with manual rendering** (Recommended)
   - Build a custom list UI without Bubbles list component
   - Pro: Minimal deps, full control over checkbox semantics, simpler code for <10 items
   - Pro: Blocked-state rendering is straightforward
   - Con: More manual key handling code
   - Effort: Medium

2. **Use Bubbles list component**
   - Leverage `bubbles/list` for scrolling, filtering, pagination
   - Pro: Built-in keyboard navigation, viewport handling
   - Con: Single-select only, need to hack checkbox state externally
   - Con: Adds complexity for 6-item list that doesn't need filtering
   - Effort: Medium

3. **Hybrid: Bubbles list with custom delegate**
   - Use list for navigation, custom ItemDelegate for checkbox rendering
   - Pro: Bubbles handles viewport/scroll
   - Con: Checkbox toggling fights list's single-select model
   - Effort: High

### Decision

**Approach 1** — Simple manual rendering with Bubbletea. For MVP's 6 agents, a manual list is clearer, more testable, and avoids fighting Bubbles' single-select semantics. The lipgloss dependency is kept for visual polish (colors, borders). Bubbles dependency will be used later for the spinner during installation phase.

## Architecture

### Exported API

```go
package tui

// AgentItem is the display model for one agent in the TUI selection list.
type AgentItem struct {
    ID          string
    Name        string
    Description string
    PreChecked  bool   // pre-select checkbox (compatible agents)
    Blocked     bool   // disabled (missing runtime)
    BlockReason string // e.g. "requires Node.js 18+"
}

// RunSelection launches the Bubbletea TUI and returns the selected agent IDs.
func RunSelection(agents []AgentItem) ([]string, error)
```

### Model internals

```go
type model struct {
    agents    []AgentItem
    cursor    int
    checked   map[int]bool  // index → checked state
    ready     bool
    width     int
    height    int
    submitted bool
    err       error
}
```

### Key mapping

| Key | Action |
|-----|--------|
| ↑/k | Cursor up |
| ↓/j | Cursor down |
| Space | Toggle checkbox (no-op if blocked) |
| Enter | Confirm selection, return checked IDs |
| q/Ctrl+C | Quit without selecting |

### View layout

```
┌──────────────────────────────────────────┐
│  Select AI Coding Agents                 │
│                                          │
│  [x] Claude Code                         │
│  [x] OpenCode                            │
│  [x] gentle-ai                           │
│  [ ] Codex CLI  ⛔ requires Node.js 22+  │
│  [ ] Gemini CLI  ⛔ requires Node.js 18+ │
│  [ ] Aider      ⛔ requires Python 3.10+ │
│                                          │
│  ↑↓ navigate • space toggle • enter done │
└──────────────────────────────────────────┘
```

### Flow (CLI integration)

```
squad add (or squad first-run)
    │
    ├─ Fetch registry
    ├─ Detect installed agents
    ├─ Check runtime compatibility per agent
    ├─ Build []AgentItem slice:
    │   ├─ PreChecked=true for compatible, not-installed agents
    │   ├─ Blocked=true for incompatible agents
    │   └─ Never show already-installed agents (add) or show all (first-run)
    ├─ tui.RunSelection(agents) → selectedIDs
    ├─ Install selected agents sequentially
    └─ Save config.json
```

### Wiring strategy

- `squad add` command: Replace the text table stub with TUI → install → save flow
- Root `squad` command: Detect no config exists → auto-launch TUI first-run flow
- Non-interactive install (`squad install --agents`) stays unchanged

## Risks

- Terminal size handling: If terminal is too small (<40 cols or <10 rows), rendering breaks
- Non-TTY environments: When run inside `postCreateCommand`, the TUI won't render if stdin isn't a TTY — must fall back to non-interactive mode
- First-run detection: The root command needs to detect "no config" vs "config exists" to decide TUI vs silent mode
- Test complexity: Bubbletea TUI requires `teatest` or a mock terminal for full integration tests; unit tests can only verify model state transitions

## Ready for Proposal

Yes — all patterns are well-understood from the existing codebase, and the TUI design follows Bubbletea conventions.
