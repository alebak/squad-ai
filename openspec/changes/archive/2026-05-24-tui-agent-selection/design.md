# Design: TUI Agent Selection

## Architecture Decision

**Use a simple Bubbletea Model without Bubbles list component.** For MVP's ~6 agents, a manual list is clearer, more testable, and avoids fighting Bubbles' single-select semantics. Bubbles is kept as a dependency for future spinner/progress components; lipgloss is used for visual styling.

**Decouple TUI selection from installation.** The TUI (`tui.RunSelection`) returns selected agent IDs to the caller. The caller (CLI handler) manages installation and config persistence. This keeps the TUI focused, testable, and reusable.

## Exported API

```go
package tui

// AgentItem is the display model for one agent in the TUI selection list.
type AgentItem struct {
    ID          string
    Name        string
    Description string
    PreChecked  bool   // pre-select checkbox (compatible agents)
    Blocked     bool   // disabled (runtime not met)
    BlockReason string // e.g. "requires Node.js 18+"
}

// RunSelection launches the Bubbletea program and returns the IDs of
// all checked agents when the user presses Enter. Returns empty slice
// and nil error if the user quits (q/Ctrl+C). Returns error only on
// fatal TUI errors.
func RunSelection(agents []AgentItem) ([]string, error)
```

## Model

```go
type model struct {
    agents    []AgentItem
    cursor    int         // current list position
    checked   map[int]bool // agent index → checkbox state
    ready     bool        // first render complete
    width     int
    height    int
    submitted bool       // Enter was pressed — exit
    err       error
}
```

### Init()

Returns `tea.Batch(tea.EnterAltScreen)` to enter alternate screen buffer.

### Update(msg tea.Msg)

| Message | Action |
|---------|--------|
| `tea.WindowSizeMsg` | Store width/height, set ready=true |
| `tea.KeyMsg{Type: tea.KeyUp}` or `rune: 'k'` | Cursor up (wrap around) |
| `tea.KeyMsg{Type: tea.KeyDown}` or `rune: 'j'` | Cursor down (wrap around) |
| `tea.KeyMsg{Type: tea.KeySpace}` | Toggle agent at cursor (no-op if blocked) |
| `tea.KeyMsg{Type: tea.KeyEnter}` | Set submitted=true, return `tea.Quit` |
| `tea.KeyMsg{Type: tea.KeyCtrlC}` or `rune: 'q'` | Set submitted=false, return `tea.Quit` |

### View()

Renders:
```
┌──────────────────────────────────────────────┐
│  🤖 Select Your AI Coding Agents             │
│                                              │
│  ● Claude Code                               │
│  ● OpenCode                                  │
│  ● gentle-ai                                 │
│  ○ Codex CLI          ⛔ requires Node.js 22+│
│  ○ Gemini CLI         ⛔ requires Node.js 18+│
│  ○ Aider              ⛔ requires Python 3.10│
│                                              │
│  ↑↓ navigate • space toggle • enter confirm  │
│  q to quit                                   │
└──────────────────────────────────────────────┘
```

- Checked: `●` (green/bright)
- Unchecked: `○` (dim)
- Blocked: `○` + ⛔ icon + reason (dimmed)
- Cursor: `▸` prefix on the current line (colored)

The title is rendered in bold via lipgloss. The border is a lipgloss-boxed frame.

## Flow Diagram

### `squad add` with TUI

```
squad add
    │
    ├─ 1. Read config (config.Load)
    ├─ 2. Fetch registry (registry.Fetch)
    ├─ 3. Detect installed (installer.DetectAll)
    ├─ 4. For each agent:
    │   ├─ Already installed? → skip
    │   ├─ Already selected? → skip  
    │   ├─ Runtime not met? → AgentItem{Blocked: true, BlockReason: "..."}
    │   └─ Compatible → AgentItem{PreChecked: true}
    ├─ 5. tui.RunSelection(agents) → selectedIDs
    │   └─ (if no agents to show → print message, skip TUI)
    ├─ 6. For each selectedID:
    │   ├─ Find agent in catalog
    │   ├─ installer.InstallAgent(agent, progressFn)
    │   └─ Print ✅/❌ per agent
    └─ 7. Save config with updated SelectedAgents
```

### Root `squad` first-run detection

```
squad (no subcommand)
    │
    ├─ config.Load(path)
    ├─ If err or config is default (empty selected_agents):
    │   └─ Run first-run flow (same as add but shows ALL non-installed agents)
    └─ If config exists with agents:
        └─ Run silent install (existing behavior)
```

## Non-TTY Detection

```go
func isTerminal() bool {
    stat, _ := os.Stdin.Stat()
    return (stat.Mode() & os.ModeCharDevice) != 0
}
```

If `!isTerminal()`, the handler prints a message and returns early instead of launching the TUI.

## File Changes

| File | Action |
|------|--------|
| `internal/tui/model.go` | Create: Model, AgentItem, RunSelection |
| `internal/tui/model_test.go` | Create: unit tests for model |
| `internal/tui/.gitkeep` | Remove |
| `internal/cli/add.go` | Modify: replace text table stub with TUI flow |
| `internal/cli/root.go` | Modify: add first-run detection logic |
| `go.mod` | Updated (already via go get) |

## Testing Strategy

| Test | Type | What it verifies |
|------|------|-----------------|
| TestModel_InitialState | Unit | Cursor at 0, compatible agents pre-checked, blocked unchecked |
| TestModel_CursorNavigation | Unit | Up/down moves cursor correctly, wraps |
| TestModel_ToggleCheck | Unit | Space toggles checked state for unblocked agent |
| TestModel_BlockedAgentNoToggle | Unit | Space on blocked agent is no-op |
| TestModel_EnterConfirms | Unit | Enter sets submitted=true, returns checked IDs |
| TestModel_QuitReturnsEmpty | Unit | q / Ctrl+C returns empty slice |
| TestAddCommand_TUIFlow | Integration | Handler builds AgentItems, calls RunSelection, installs, saves |
| TestAddCommand_NoTTY | Unit | Fallback message when stdin not a TTY |
| TestAddCommand_NoAvailableAgents | Unit | Skips TUI when no agents to show |
