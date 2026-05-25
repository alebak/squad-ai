# Tasks: TUI Agent Selection Redesign

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 200-280 |
| 400-line budget risk | Medium |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: single-pr
400-line budget risk: Medium

## Phase 1: AgentItem Struct Changes

- [x] 1.1 Remove `PreChecked bool`, `IsInstalled bool` from `AgentItem` in `internal/tui/model.go`
- [x] 1.2 Add `IsSelectAll bool` field to `AgentItem`

## Phase 2: TUI Model Logic

- [x] 2.1 Rewrite `newModel()` — no PreChecked loop, empty checked map
- [x] 2.2 Rewrite `renderAgentRow()` — detect IsSelectAll for dynamic label, blocked agents render as `Name (BlockReason)`, no emoji, installed no longer gets ✅ tag
- [x] 2.3 Rewrite `handleSpecialKey()` — space on select-all row calls `toggleAll()`, remove `!IsInstalled` guard, keep `!Blocked` guard
- [x] 2.4 Rewrite `toggleAll()` — remove `IsInstalled` skip (only skip blocked)
- [x] 2.5 Update `View()` — remove emoji from title, remove "a select all" from help bar
- [x] 2.6 Update `selectedIDs()` — skip select-all sentinel row (index 0)

## Phase 3: CLI Integration

- [x] 3.1 Rewrite `buildAgentItemsForAdd()` in `internal/cli/add.go` — prepend sentinel, no PreChecked/IsInstalled fields
- [x] 3.2 Add uninstall prompt to `runAddFlowInteractive()` — for deselected installed agents, prompt `Uninstall <name>? [y/N]` and call uninstaller

## Phase 4: Tests

- [x] 4.1 Update `TestModel_InitialState` — all unchecked, agents[0] is select-all sentinel
- [x] 4.2 Update `TestModel_ToggleCheck` — toggle works on non-blocked agents
- [x] 4.3 Update `TestModel_BlockedAgentNoToggle` — no ⛔ rendering assertion
- [x] 4.4 Update `TestModel_EnterConfirms` — initial selectedIDs is empty
- [x] 4.5 Update `TestModel_ToggleAll` — only blocked agents skipped
- [x] 4.6 Update `TestModel_ToggleAllDeselect` — only blocked agents skipped
- [x] 4.7 Update `TestModel_ViewRenders` — no emoji, contains select-all row
- [x] 4.8 Remove `TestModel_HelpIncludesToggleAll` — help no longer shows "a select all"
- [x] 4.9 Remove or update `TestModel_InstalledAgentNoToggle` — installed IS toggleable now
- [x] 4.10 Add `TestModel_SelectAllRowToggle` — space on select-all toggles all compatible
- [x] 4.11 Add `TestModel_SelectAllDynamicLabel` — label switches between select/unselect all
- [x] 4.12 Add `TestModel_BlockedAgentNoEmoji` — verify no ⛔ in row output
- [x] 4.13 Add `TestModel_InstalledAgentToggleable` — verify space toggles installed agent
