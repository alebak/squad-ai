# Design: TUI Apply + Dialog + Uninstall Wizard

## Technical Approach

Three coordinated changes to the TUI model, all modifying `model.go`:

1. **#37 (Apply + separator + `a` shortcut)**: Rename the sentinel item from "✓ Done" to "Apply". Add a separator just before it. `a` key in `handleRuneKey` calls `submitSelection()` (same logic as pressing Enter on Apply). Header updated to include version string with mauve/subdued colors.

2. **#38 (No changes dialog)**: Add `showDialog` field to model. When submit is triggered with nothing selected, set `showDialog = "no-changes"`. The view checks this and renders a bordered dialog overlay. Enter dismisses it.

3. **#39 (Inline uninstall wizard)**: Add `wizard *wizardState` to model. When submit is triggered AND there are deselected installed agents, initialize wizard. The view checks for wizard first (highest priority), then dialog, then agent list. Navigation handlers check wizard state and dispatch to wizard-specific key handling.

`RunSelection` signature changes from `([]string, error)` to `([]string, map[string]int, error)`. The wizard choices map (`agentID → choice`) is returned alongside selected IDs. `runAddFlowInteractive` consumes the wizard choices, calling `UninstallAgent`/`UninstallConfig` accordingly.

Colors switch from hardcoded lipgloss color numbers to Catppuccin Mocha hex values.

## Architecture Decisions

### Decision: Wizard state lives inside model, not externally

**Choice**: `wizard *wizardState` field on the `model` struct.
**Alternatives considered**: External wizard controller, separate teaprogram.
**Rationale**: The wizard is a transient view state that affects rendering and key handling. Keeping it inside the model is the simplest approach — no new components needed, and the wizard lifecycle (init → steps → complete/cancel) maps cleanly to model state transitions.

### Decision: `RunSelection` returns wizard choices as second return value

**Choice**: `([]string, map[string]int, error)` signature.
**Alternatives considered**: Dedicated result struct, callback during wizard.
**Rationale**: The wizard collects choices per agent before the final Apply. Returning them alongside the final selected IDs keeps the interface clean. The map is nil when no wizard was needed, simplifying the caller.

### Decision: Separator as plain string in View, not a new AgentItem type

**Choice**: Render the separator as a raw string in the `View()` method, not as a new kind of sentinel item.
**Alternatives considered**: Adding a `IsSeparator` sentinel to AgentItem.
**Rationale**: The separator is purely visual, not interactive. Adding it to AgentItem would require filtering in cursor navigation, toggle logic, and selectedIDs. Rendering it between the last agent and Apply in View() is simpler and less error-prone.

### Decision: Wizard uses internal choice values (0/1/2), not exported constants

**Choice**: The wizard choices are `int` values with documented meaning (0=app, 1=app+config, 2=skip).
**Alternatives considered**: Exported constants with descriptive names.
**Rationale**: The choice values are consumed only in `runAddFlowInteractive` and have a single interpretation. Constants would add boilerplate without clarity benefit. Document meaning at the return site.

## Data Flow

```
User presses Enter on Apply or 'a'
  │
  ├─ wizard != nil → complete wizard step or submit final → wizard = nil → show Apply
  │
  ├─ showDialog != "" → dismiss dialog → showDialog = ""
  │
  ├─ Installed agents deselected? 
  │   └─ Yes → init wizardState → show wizard view (no submit yet)
  │
  ├─ selectedIDs empty AND no deselected installed?
  │   └─ Yes → showDialog = "no-changes" → render dialog overlay
  │
  └─ Normal submit → return selectedIDs + wizardChoices (as map[string]int)
```

After TUI returns, `runAddFlowInteractive`:
```
RunSelection returns (ids, choices, err)
  ├─ err != nil → fatal
  ├─ ids == nil → user quit
  └─ ids != nil → process wizard choices
       ├─ For each agent in choices:
       │   0 → UninstallAgent
       │   1 → UninstallAgent + UninstallConfig
       │   2 → skip (keep installed)
       └─ Proceed to install selected ids
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modify | Add `showDialog`, `wizard`, `wizardState`, Catppuccin styles, Apply item, separator, dialog/wizard rendering, new key handlers |
| `internal/tui/model_test.go` | Modify | Update existing tests for Apply, add tests for dialog, wizard, `a` shortcut, separator |
| `internal/cli/add.go` | Modify | Update `RunSelection` signature, `addHandler.runSelection` type, `runAddFlowInteractive` to consume wizard choices, remove stdin prompt code |

## Interfaces / Contracts

```go
// New RunSelection signature:
func RunSelection(agents []AgentItem) (selectedIDs []string, wizardChoices map[string]int, err error)

// wizardChoices: agentID → choice
//   0 = uninstall app only
//   1 = uninstall app + config
//   2 = keep installed (skip)

// wizardState (unexported, inside model)
type wizardState struct {
    step    int
    total   int
    agents  []registry.Agent
    choices []int
    cursor  int  // which radio option is highlighted
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Apply item renders, separator rendering, `a` submits | Direct `Model.Update()` with `tea.KeyMsg` |
| Unit | Dialog appears when Apply pressed with empty selection | State check after Enter on Apply |
| Unit | Dialog dismisses on Enter | State check after second Enter |
| Unit | Wizard initializes when installed agents deselected | Mock `AgentItem` with `PreChecked` + unchecked |
| Unit | Wizard navigation (j/k/enter/n/b/q) | State checks after each key |
| Unit | Wizard returns choices after completion | Verify `wizard.choices` |
| Unit | `RunSelection` new return signature | Mock tests |
| Integration | `runAddFlowInteractive` consumes wizard choices | Mock `runSelection` returning choices |

## Migration / Rollout

No migration required. The `addHandler` struct fields `uninstallChoiceFn` and `confirmFn` are removed. All consumers update in a single commit.

## Open Questions

None — design is complete based on the spec requirements.
