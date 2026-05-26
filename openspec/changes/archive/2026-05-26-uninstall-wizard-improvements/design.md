# Design: Uninstall Wizard Improvements

## Technical Approach

Three coordinated changes to `wizardState` and `handleWizardKey` in `model.go`:

1. **Wizard cursor extension**: Cursor range goes from 0-2 (3 radio options) to 0-4 (radios + Back button + Next button). Wrapping behavior extends naturally.
2. **Auto-advance on Enter**: When Enter is pressed on a radio option (cursor 0-2), store the choice AND call the advance logic immediately. No separate `n` press needed.
3. **Summary view**: New `showingSummary bool` field on `wizardState`. After the last step, set `showingSummary = true` instead of calling `completeWizard()`. Summary renders a table of agent → action choices with Apply and Back buttons.

## Architecture Decisions

### Decision: Cursor positions 3=Back, 4=Next

**Choice**: Map Back button to cursor position 3 and Next to position 4, extending the existing 0-2 radio range to 0-4.
**Alternatives considered**: Separate cursor for buttons, tab-based focus.
**Rationale**: Extending the existing cursor index is the minimal change — the existing `handleWizardKey` wrapping logic (decrement/increment with bounds check) naturally extends to a larger range. No new state fields needed.

### Decision: `showingSummary` as wizardState field, not separate mode

**Choice**: Add `bool showingSummary` to `wizardState`. When true, `renderWizardView()` renders the summary instead of the step view.
**Alternatives considered**: Separate view mode flag on model, separate summary state struct.
**Rationale**: The summary is a transient state within the wizard lifecycle. Keeping it on `wizardState` means the View() method only needs one branch (check wizard non-nil then dispatch to step/summary), and the wizard lifecycle (init → steps → summary → complete/cancel) is self-contained.

### Decision: Summary shows action text, not choice codes

**Choice**: Map 0→"Uninstall app only", 1→"Uninstall app + config data", 2→"Keep installed" for display.
**Alternatives considered**: Display raw choice number 0/1/2.
**Rationale**: Human-readable action text is what the user expects from a summary. The map to text is simple and hardcoded.

### Decision: Buttons use inline cursor + styled brackets, not borders

**Choice**: Render buttons as `[ ◄ Back ]` / `[ Next ► ]` with lipgloss styling, not as bordered lipgloss boxes.
**Alternatives considered**: RoundedBorder boxes, pane-style buttons.
**Rationale**: Matches the existing TUI aesthetic (styled text, no heavy box drawing). The `[ ]` bracket convention is universally recognized as a button in terminal UIs.

## Data Flow

```
User presses Enter on Apply
  └─ Wizard init (wizardState, cursor 0-4)

Wizard step view:
  ┌─ j/k ←→ navigate through 5 positions [0-4]
  │
  ├─ Enter on radio [0-2]:
  │   └─ ws.choices[ws.step] = ws.cursor
  │   └─ auto-advance (ws.step++)
  │       ├─ ws.step < ws.total → next step, cursor=0
  │       └─ ws.step >= ws.total → showingSummary=true
  │
  ├─ Enter on Back [3]:
  │   └─ ws.step-- (if > 0), cursor=0
  │
  ├─ Enter on Next [4]:
  │   └─ if ws.choices[ws.step] != -1 → ws.step++
  │       └─ (same advance logic as auto-advance)
  │
  └─ n/b/q shortcuts work as before

Summary view (showingSummary=true):
  ├─ Enter on Apply → completeWizard() → return to agent list
  ├─ Enter on Back → showingSummary=false, step=same, cursor=0
  └─ j/k navigate between Apply/Back buttons
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modify | Extend cursor range, auto-advance, summary view/state, button rendering |
| `internal/tui/model_test.go` | Modify | New tests for buttons, auto-advance, summary, Apply/Back from summary |

## Interfaces / Contracts

```go
// Updated wizardState
type wizardState struct {
    step           int
    total          int
    indices        []int
    choices        []int
    cursor         int     // 0-4: 0/1/2=radio, 3=Back, 4=Next
    showingSummary bool    // true = show summary view instead of step
}

// Cursor position constants (for code clarity, not exported)
const (
    wizardRadio0    = 0
    wizardRadio1    = 1
    wizardRadio2    = 2
    wizardBackBtn   = 3
    wizardNextBtn   = 4
    wizardPositions = 5
)
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Cursor wraps all 5 positions (j/k 5 times → back to 0) | `model.Update()` with key msg, check `wizard.cursor` |
| Unit | Enter on radio auto-advances to next step | Check `wizard.step` incremented after Enter on radio |
| Unit | Enter on Back goes to previous step | Check `wizard.step` decremented |
| Unit | Enter on Next advances (with choice confirmed) | Check `wizard.step` incremented |
| Unit | Summary renders with all agents + choices | Check `View()` contains agent names and action text |
| Unit | Apply on summary completes wizard | Check `wizard == nil` after Apply on summary |
| Unit | Back from summary returns to last step | Check `showingSummary == false` |
| Unit | n/b shortcuts still work in step view | Check step advances/decrements |
| Unit | Existing behavior preserved (quit, cancel, choice selection) | Verify unchanged assertions |
| View | Buttons rendered in wizard view | Check `View()` contains `[ ◄ Back ]` and `[ Next ► ]` |

## Migration / Rollout

No migration required. All changes are within the TUI model. No data format changes.

## Open Questions

None — design is complete based on spec requirements.
