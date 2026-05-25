# Design: Fix TUI 'q' exit triggering uninstall prompt

## Technical Approach

Add a single nil-guard after `RunSelection` returns in `runAddFlowInteractive`. The guard checks if `selectedIDs` is `nil` (indicating user quit via `q`, `Ctrl+C`, or `Escape`) and returns `cfg, nil` immediately, bypassing the uninstall/restart logic.

## Architecture Decisions

### Decision: Nil-as-quit, empty-slice-as-confirmed-empty

**Choice**: Use Go's nil vs empty-slice distinction to encode two different user intents.
**Alternatives considered**: Return a sentinel error, add a bool param to `RunSelection`, use a custom return type.
**Rationale**: Minimal diff, zero new types, no API breakage. Go callers already know `len(nil) == 0` — the fix makes the distinction explicit at the call site. Other callers (tests, non-interactive path) are unaffected because they either use both return values correctly or don't run the TUI.

### Decision: Guard at call site, not inside RunSelection

**Choice**: Add the nil check in `runAddFlowInteractive`, not inside `RunSelection`.
**Alternatives considered**: Change `RunSelection` to return an enum or separate quit flag.
**Rationale**: `RunSelection` already returns nil for quit — the existing behavior is correct and tested. The bug is that the CALLER doesn't distinguish. Fixing the caller is the smallest, safest change.

## Data Flow

```
RunSelection(agentItems)
  │
  ├─ returns (nil, nil) ──→ nil check → return cfg, nil (clean exit)
  │                            [NEW CODE]
  ├─ returns ([], nil)  ──→ len(selectedIDs) == 0 → uninstall/restart loop
  │
  └─ returns ([a,b], nil) ──→ len(selectedIDs) > 0 → install flow
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Modify | Add nil check after RunSelection at line ~241 |
| `internal/tui/model.go` | Modify | Update RunSelection godoc to document nil vs empty-slice contract |
| `internal/cli/add_test.go` | Modify | Update `TestAddCommand_TUIEmptySelection` to use `[]string{}` not `nil` |

## Interfaces / Contracts

```go
// RunSelection returns:
//   nil, nil       — user quit (q, Ctrl+C, Escape)
//   []string{}, nil — user confirmed with nothing checked
//   [ids...], nil  — user confirmed with selected agent IDs
//   nil, error     — fatal TUI error
func RunSelection(agents []AgentItem) ([]string, error)
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Quit returns nil | `TestModel_QQuits` already verifies `isSubmitted == false` |
| Unit | Empty selection returns empty slice | Update `TestAddCommand_TUIEmptySelection` to use `[]string{}` return — should still see "No agents selected" |
| Unit | Nil guard exits cleanly | No explicit test needed — the existing test uses nil and it should still pass (just with different flow) |
| Integration | Full add flow with nil return | Run `go test ./internal/cli/` — all existing tests must pass |

## Migration / Rollout

No migration required. This is a behavioral bug fix — the old behavior (showing uninstall on quit) was wrong.

## Open Questions

None.
