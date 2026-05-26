# Design: TUI Restart After Agent Installation

## Current Flow

```
runAddFlowInteractive()
│
├── for {                          // outer loop
│     selectedIDs, wizardChoices ← runSelection(agentItems)
│     if selectedIDs == nil → return cfg, nil         // user quit
│     if wizardChoices != nil → process uninstalls
│       if restart → rebuild agentItems; continue     // ← uninstall restart exists
│     if selectedIDs == [] → "No agents selected"; return cfg, nil
│     break                                            // ← exits loop here
│   }
│
├── toInstall ← findAgentsByIDs(catalog, selectedIDs)
├── toInstall ← filterInstalled(h, toInstall)
├── installAll(toInstall)
├── report results
├── save config
└── return cfg, nil                 // ← exits to terminal, BUG
```

The problem: after `break`, the installation logic runs OUTSIDE the loop, and then `return cfg, nil` exits the function entirely. The user never sees the updated TUI state.

## Target Flow

```
runAddFlowInteractive()
│
└── for {                          // outer loop
      selectedIDs, wizardChoices ← runSelection(agentItems)
      if selectedIDs == nil → return cfg, nil
      if wizardChoices != nil → process uninstalls
        if restart → rebuild agentItems; continue
      
      // Moved INSIDE loop:
      if selectedIDs == [] → "No agents selected"; return cfg, nil
      
      toInstall ← findAgentsByIDs(catalog, selectedIDs)
      toInstall ← filterInstalled(h, toInstall)
      
      if toInstall == [] → "Selected agents already installed"
        rebuild agentItems; continue        // ← NEW: loop back
      
      installAll(toInstall)
      report results
      save config
      update installed map                   // ← NEW
      rebuild agentItems                     // ← NEW
      continue                               // ← instead of break
    }
```

## Code Change Detail

### `internal/cli/add.go` — `runAddFlowInteractive`

Move the installation block from AFTER the for-loop INTO the for-loop body, replacing the single `break`:

```go
// Before (current):
if len(selectedIDs) == 0 {
    cmd.Println("No agents selected. Nothing to install.")
    return cfg, nil
}
break

// After (new):
if len(selectedIDs) == 0 {
    cmd.Println("No agents selected. Nothing to install.")
    return cfg, nil
}

toInstall := findAgentsByIDs(catalog, selectedIDs)
toInstall = filterInstalled(h, toInstall)
if len(toInstall) == 0 {
    cmd.Println("Selected agents are already installed.")
    agentItems = buildAgentItemsForAdd(h, catalog, installed)
    continue
}

cmd.Println("Installing selected agents...")
results := h.installAll(toInstall, makeProgressFn(cmd, toInstall))

succeeded, hasErrors := reportAddResults(cmd, toInstall, results)
cfg.SelectedAgents = append(cfg.SelectedAgents, succeeded...)
if saveErr := config.Save(cfgPath, cfg); saveErr != nil {
    cmd.Printf("Warning: failed to save config: %v\n", saveErr)
}

// Mark as installed and rebuild TUI items
for _, id := range succeeded {
    installed[id] = true
}
agentItems = buildAgentItemsForAdd(h, catalog, installed)

if hasErrors {
    cmd.Println("Some installations failed. You can retry or continue.")
}
continue
```

The key changes:
1. Move `findAgentsByIDs` + `filterInstalled` + install + save INSIDE the loop
2. `break` → `continue` after install
3. Update `installed` map and rebuild `agentItems` after install
4. When all selected agents are already installed → rebuild and `continue` (not exit)
5. Error messaging remains, but TUI relaunches instead of returning

### `internal/cli/add_test.go` — Test Updates

**Existing tests that need updating:**

| Test | Change Needed |
|------|---------------|
| `TestAddCommand_TUISuccessFlow` | Add second `runSelection` return (nil to quit) | 
| `TestAddCommand_TUIEmptySelection` | No change (empty selection exits) |
| `TestAddCommand_UninstallViaWizardAppOnly` | Add third `runSelection` return (nil to quit after install) |
| `TestAddCommand_UninstallViaWizardAppAndConfig` | Add third `runSelection` return (nil to quit after install) |
| `TestAddCommand_UninstallViaWizardSkip` | No change (empty selection after skips exits) |
| `TestAddCommand_WizardRestartsTUIAfterUninstall` | Add third `runSelection` return (nil to quit after install) |

**New test to add:**

`TestAddCommand_TUIRelaunchAfterInstall` — verifies that `runSelection` is called twice (first for selection, second after install relaunch) and that the second call exits cleanly.

## Rationale

1. **Mirrors uninstall pattern**: The uninstall path already uses `continue` to restart the TUI. This applies the same pattern to the install path.
2. **Minimal diff**: No new concepts, no new dependencies. The change is a restructuring of existing code within the same function.
3. **Backward compatible**: The external API (`runAddFlow` signature, `Config` structure, CLI output) doesn't change.
