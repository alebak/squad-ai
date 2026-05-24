# Tasks: TUI Agent Selection

## Task 1: Create `internal/tui/model.go`

**Files:** `internal/tui/model.go`

- Define `AgentItem` struct: ID string, Name string, Description string, PreChecked bool, Blocked bool, BlockReason string
- Define `model` struct (unexported): agents []AgentItem, cursor int, checked map[int]bool, ready bool, width int, height int, submitted bool, err error
- Implement `tea.Model` interface: `Init()` → `tea.Batch(tea.EnterAltScreen)`, `Update(msg tea.Msg)`, `View() string`
- Implement `RunSelection(agents []AgentItem) ([]string, error)` — creates `tea.NewProgram(model{...})` and runs it
- Key handling:
  - Up/k: cursor--, wrap to bottom if at top
  - Down/j: cursor++, wrap to top if at bottom
  - Space: toggle checked[cursor] (no-op if Blocked)
  - Enter: submitted=true, return tea.Quit
  - q/Ctrl+C: return tea.Quit (no submission)
- View layout:
  - Lipgloss box border around everything
  - Bold title line: "🤖 Select Your AI Coding Agents"
  - Agent lines: `▸` cursor + `●`/`○` checkbox + agent name + optional `⛔ reason`
  - Help bar at bottom with key hints
  - Blocked agents rendered with lipgloss `Style.Dim(true)` or subdued color
- Non-TTY check: not here (handled in CLI layer)

**Status:** [x]

## Task 2: Create `internal/tui/model_test.go`

**Files:** `internal/tui/model_test.go`

- Test initial state: cursor at 0, compatible pre-checked, blocked unchecked
- Test cursor navigation: up, down, wrapping
- Test toggle: space toggles checked state
- Test blocked no-toggle: space on blocked is no-op
- Test enter: returns checked IDs
- Test quit: q and Ctrl+C return empty slice
- Use simple unit tests (model state assertions), NOT teatest (no terminal needed)

**Status:** [x]

## Task 3: Remove `.gitkeep` and verify compilation

- Remove `internal/tui/.gitkeep`
- Run `go build ./cmd/squad` — must compile
- Run `go vet ./internal/tui/` — no issues

**Status:** [x]

## Task 4: Update `internal/cli/add.go` to wire TUI

**Files:** `internal/cli/add.go`

- Add `runSelection` field to `addHandler` struct
- Default wired to `tui.RunSelection`
- Add `installAll` field to `addHandler`
- Update `defaultAddHandler()` with real implementations
- Replace the "TUI coming soon" stub body with full TUI flow
- Extract logic into `runAddFlow` for reuse by root command
- Non-TTY fallback with agent list

**Status:** [x]

## Task 5: Update `internal/cli/root.go` for first-run flow

**Files:** `internal/cli/root.go`

- Set `RunE` on root command to detect first-run
- Logic: load config, if DefaultConfig → run TUI flow, else print status
- Keep existing subcommands unchanged

**Status:** [x]

## Task 6: Test and verify

- Run `go test ./... -v -count=1` — all tests pass
- Run `go vet ./...` — no issues
- Run `go build ./cmd/squad` — compiles cleanly

**Status:** [x]
