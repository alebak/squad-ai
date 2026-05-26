# Proposal: Fix TUI version header hardcoding and wizard submit flow

## Intent

Fix two bugs in the TUI that affect correctness and user experience:

1. The TUI header hardcodes version string "0.15.0" instead of using the actual build version from the `cli.version` variable (set via ldflags at build time). This means every build shows "0.15.0" regardless of the actual release version.

2. Pressing Apply on the wizard summary view returns to the agent list instead of submitting and quitting the TUI. The user must press Apply a second time, which is confusing and inconsistent with the non-wizard flow where Apply quits immediately.

## Scope

### In Scope
- Pass version string from `cli` package to `tui.RunSelection` and `newModel`
- Store version in `model` struct and render dynamically in `View()`
- When wizard is on summary and user presses Apply, submit immediately (set `submitted=true`, clear wizard, return `tea.Quit`)
- Update all tests (both `internal/tui/` and `internal/cli/`)
- Update spec for version header requirement to use dynamic version

### Out of Scope
- Changing how the version variable is set or propagated elsewhere
- Refactoring the wizard flow beyond the bug fix
- Adding new tests unrelated to the two bugs

## Capabilities

### New Capabilities
- None

### Modified Capabilities
- `add`: Update spec requirement for "Header shows version" to use build-time version instead of hardcoded "0.15.0". Add new spec requirement for wizard submit behavior.

## Approach

### Bug 1 — Hardcoded Version
1. Add `version string` field to `model` struct in `internal/tui/model.go`
2. Update `newModel` to accept a `version` parameter
3. Update `RunSelection` to accept a `version` parameter and pass it to `newModel`
4. In `View()`, render `fmt.Sprintf("Squad AI (version %s)", m.version)` instead of hardcoded string
5. In `internal/cli/add.go`, pass the package-level `version` var to `tui.RunSelection`

### Bug 2 — Wizard Summary Submit
1. In `handleSummaryKey()`, when `ws.cursor == 0` (Apply) and Enter is pressed: set `wizard = nil`, `submitted = true`, and return `tea.Quit` immediately instead of just calling `completeWizard()`
2. Since `handleSummaryKey` is called via `handleWizardKey` which returns `(tea.Model, tea.Cmd)`, modify the return path to support quitting

## Affected Areas

| Area | Impact | Description |
|------|--------|-------------|
| `internal/tui/model.go` | Modified | Add `version` field, update `newModel`/`RunSelection` signatures, fix View render, fix wizard submit |
| `internal/tui/model_test.go` | Modified | Update all `newModel(testAgents())` calls to pass version, update version assertions, add test for wizard submit |
| `internal/cli/add.go` | Modified | Pass `version` to `tui.RunSelection` |
| `internal/cli/add_test.go` | Modified | All `runSelection` mock signatures updated; no functional change needed since it's an interface |
| `openspec/specs/add/spec.md` | Modified | Update version header requirement to use dynamic version; add wizard submit scenario |

## Risks

| Risk | Likelihood | Mitigation |
|------|------------|------------|
| Breaking `RunSelection` API signature affects callers | Low | Only two callers: `tui.RunSelection` in tests and `addHandler.runSelection` — both easy to update |
| Test assertions on hardcoded string need update | Low | Search for "0.15.0" in test files and update to match new dynamic format |

## Rollback Plan

Revert the changes to the four files (model.go, model_test.go, add.go, add_test.go) and the spec update. The bugs are cosmetic/UX, not data-loss, so rollback is safe at any point.

## Dependencies

- None. All changes are within the existing codebase.

## Success Criteria

- [ ] `go test ./internal/tui/...` passes
- [ ] `go test ./internal/cli/...` passes
- [ ] `go build ./cmd/squad` compiles
- [ ] View shows actual version (from `cli.version`), not hardcoded "0.15.0"
- [ ] Pressing Apply on wizard summary immediately quits and delivers choices
