## Exploration: Uninstall Wizard Improvements

### Current State

The uninstall wizard shows one step at a time with 3 radio options ("Uninstall app only", "Uninstall app + config data", "Keep installed (skip)"). Navigation relies on non-obvious key commands: `n` for next, `b` for back, `j`/`k` for up/down. There are no visible Next/Back buttons. After the last step, the wizard immediately completes and returns to the agent list — no summary review. The user must press `n` to advance even after selecting a radio option with Enter.

### Affected Areas

- `internal/tui/model.go` — All wizard state, rendering, and key handling live here
- `internal/tui/model_test.go` — All wizard tests live here
- `openspec/specs/add/spec.md` — Spec for add flow including wizard behavior
- `openspec/specs/uninstall/spec.md` — Spec for uninstall behavior

### Approaches

1. **Inline cursor-based buttons** — Extend wizard cursor to move past radio options onto Next/Back buttons rendered at the bottom. Auto-advance on Enter for radio selection. Add `showingSummary` field to wizard state. Render summary after last step.
   - Pros: Consistent with existing cursor model, no new UI paradigm needed
   - Cons: Cursor logic gets more complex (radio items + buttons), edge cases for disabled states
   - Effort: Medium

2. **Dedicated button handling** — Buttons rendered as visual elements you tab onto. Use Enter on a button to trigger action. Otherwise same as approach 1.
   - Pros: Clearer separation of concerns
   - Cons: More complex implementation, Bubbletea doesn't have native focus management for this pattern
   - Effort: High

### Recommendation

**Approach 1**: Inline cursor-based buttons. The wizard cursor currently wraps among 3 radio options (0-2). Extend the cursor to include positions 3 (Back button) and 4 (Next button). When cursor is on a button, Enter triggers the button action rather than storing a choice. Auto-advance: pressing Enter on a radio option stores the choice AND automatically advances to the next step (or shows summary if last step).

### Risks

- Auto-advance changes existing behavior — users who wanted to review before advancing lose that pause. Mitigated by the summary screen where they can review all choices.
- Cursor positions 3 and 4 for buttons may feel unintuitive for keyboard-first users. Mitigated by still supporting `n`/`b` shortcuts.
- Summary view needs to handle display of potentially long agent names. Mitigated by simple table format with truncation.

### Ready for Proposal

Yes — proceed to proposal phase.
