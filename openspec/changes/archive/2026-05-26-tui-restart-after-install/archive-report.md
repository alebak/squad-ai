# Archive Report: TUI Restart After Agent Installation

## Closure Summary

Change `tui-restart-after-install` archived 2026-05-26. SDD cycle complete.

**Root cause**: In `runAddFlowInteractive`, the installation logic was placed AFTER the outer for-loop. After install completed, the function returned `cfg, nil` — dropping the user to the terminal instead of looping back to the TUI.

**Fix**: Moved the installation logic inside the for-loop body. After install, update the `installed` map, rebuild `agentItems` with `buildAgentItemsForAdd`, and `continue` back to the top of the loop. This mirrors the existing uninstall restart pattern.

## Specs Synced

- **add**: Updated — Added "TUI Restart After Installation" requirement with 5 new scenarios

## Archive Contents

All artifacts present:
- proposal.md ✅
- specs/add/spec.md ✅
- design.md ✅
- tasks.md ✅
- verify-report.md ✅
- archive-report.md ✅

## Engram Observation IDs (lineage)

| Artifact | Observation ID |
|----------|---------------|
| proposal | obs-8fd6b8516b3e8abb |
| spec | obs-a3ab9177430a1ee8 |
| design | obs-debba0db08bc57c0 |
| tasks | obs-b93af70a6faf2b2e |
| verify-report | obs-fc1a483ddc8fc588 |
| archive-report | (current) |

## Changed Files

- `internal/cli/add.go` — Moved install logic inside loop, added `continue` path
- `internal/cli/add_test.go` — Updated 4 existing tests, added `TestAddCommand_TUIRelaunchAfterInstall`
- `openspec/specs/add/spec.md` — Merged TUI Restart requirement and 5 scenarios
