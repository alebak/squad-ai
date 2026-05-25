# Archive Report: fix-tui-quit-exit

## Summary

Bug fix: pressing `q` in the TUI showed the uninstall prompt instead of exiting. Root cause: `RunSelection` returns `nil` (quit) vs `[]` (confirmed empty), but `runAddFlowInteractive` treated both the same via `len(selectedIDs) == 0`. Fixed by adding a `selectedIDs == nil` guard that returns immediately on quit.

## Artifact Observation IDs (Engram)

| Artifact | Observation ID |
|----------|---------------|
| Proposal | obs-2536e5c842255285 |
| Spec | obs-5b953c0b3dda2f38 |
| Design | obs-c9382e4119469d17 |
| Tasks | obs-f030814f750374a4 |
| Apply Progress | obs-b13352533b82819d |
| Verify Report | obs-6eb8692ad2cf9330 |

## Specs Synced

| Domain | Action | Details |
|--------|--------|---------|
| add | Updated | Added "TUI Quit Clean Exit" requirement (5 scenarios); modified "3-Option Uninstall Prompt" to clarify it applies after Enter, not after quit |

## Archive Contents

- proposal.md ✅
- specs/add/spec.md ✅
- design.md ✅
- tasks.md ✅ (5/5 tasks complete)
- verify-report.md ✅

## Source of Truth Updated

`openspec/specs/add/spec.md` now reflects both the new quit behavior and the clarified uninstall prompt scope.

## SDD Cycle Complete

The change has been fully planned, implemented (3 file edits), verified (125 tests passing), and archived.
