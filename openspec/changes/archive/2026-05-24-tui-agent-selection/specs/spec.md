# Spec: TUI Agent Selection

## References

- PRD §7.1 (Primera ejecución sin manifiesto) — first-run flow
- PRD §7.4 (Modo interactivo manual) — `squad add` TUI context
- PRD §5 (Stack técnico) — Bubbletea, Lipgloss, Bubbles
- PRD §9 (Verificación de dependencias de runtime) — blocked agent semantics

## Requirements

### R1: Agent selection TUI

`squad add` SHALL open a Bubbletea-based TUI that displays a list of agents available from the registry, excluding agents that are already installed or already selected in the config.

### R2: Pre-selection

All agents that are runtime-compatible SHALL appear with their checkbox pre-checked. Agents whose runtime dependencies are not met SHALL appear with their checkbox unchecked and disabled.

### R3: Blocked indicator

Agents blocked by runtime requirements SHALL display the reason visually:
- A blocked icon (⛔) before the name
- A grayed-out or dimmed appearance
- A text explanation, e.g., "requires Node.js 22+"

Blocked agents MUST NOT be toggleable via spacebar.

### R4: Keyboard navigation

| Key | Action |
|-----|--------|
| ↑ / k | Move cursor up |
| ↓ / j | Move cursor down |
| Space | Toggle checkbox on the current item (no-op if blocked) |
| Enter | Confirm selection and exit TUI |
| q / Ctrl+C | Quit without confirming (returns empty selection) |

### R5: Return value

`tui.RunSelection(agents []AgentItem) ([]string, error)` SHALL return the IDs of all checked agents when the user presses Enter. It SHALL return an empty slice and no error if the user presses `q` or `Ctrl+C`.

### R6: CLI integration (squad add)

The `squad add` command SHALL:
1. Read the user config
2. Fetch the registry
3. Detect installed agents
4. Filter: available = not installed AND not already selected AND runtime-compatible
5. Build `[]tui.AgentItem` with compatible agents pre-checked
6. Launch TUI via `tui.RunSelection()`
7. Install the selected agents sequentially
8. Save config.json with updated `selected_agents`

### R7: First-run flow (root squad)

When `squad` is run without arguments and no `config.json` exists, it SHALL detect the missing config and launch the TUI first-run flow. If `config.json` exists, it SHALL run the silent install flow (existing behavior, not part of this change).

### R8: Non-TTY fallback

If stdin is not a TTY, the TUI SHALL NOT be launched. The command SHALL fall back to printing a message instructing the user to use `--agents` for non-interactive selection.

### R9: Styling

The TUI SHALL use lipgloss for basic styling:
- Title in bold/colored text
- Cursor indicator (▸) on the current line
- Blocked agents in dimmed/dark style
- Clean visual separation

### R10: Dependencies

The `go.mod` SHALL include:
- `github.com/charmbracelet/bubbletea` — TUI framework
- `github.com/charmbracelet/lipgloss` — TUI styling
- `github.com/charmbracelet/bubbles` — TUI components (for future spinner/progress)

## Scenarios

### Scenario S1: Basic agent selection
**Given** 3 compatible agents from the registry
**When** the user runs `squad add`
**Then** the TUI displays all 3 agents with checkboxes checked
**And** the user can navigate and toggle

### Scenario S2: Blocked agent displayed
**Given** 1 compatible agent and 1 blocked agent (missing Node.js)
**When** the user runs `squad add`
**Then** the blocked agent shows a ⛔ indicator and runtime warning
**And** pressing space on the blocked agent does NOT toggle it

### Scenario S3: Confirm selection
**Given** 2 agents with both checked
**When** the user presses Enter
**Then** `RunSelection` returns `["claude-code", "opencode"]`

### Scenario S4: Quit without selecting
**Given** the TUI is displayed
**When** the user presses `q` or `Ctrl+C`
**Then** `RunSelection` returns `([], nil)`

### Scenario S5: Blocked agent cannot be toggled
**Given** a blocked agent with cursor on it
**When** the user presses Space
**Then** the agent remains unchecked (no-op)

### Scenario S6: Non-TTY fallback
**Given** stdin is not a terminal
**When** `squad add` is executed
**Then** the command prints a message about using `--agents` instead
