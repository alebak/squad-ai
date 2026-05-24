# Design: Agent Uninstall Support

## Architecture Decision

**Use the same handler injection pattern as install/list** for the remove command. The `removeHandler` struct gains new injectable fields for registry fetching and uninstall execution, enabling testability without globals.

**Use direct `os.Remove` for curl_bash fallback** instead of shell commands to prevent shell injection from `detect_cmd` values.

## Registry Schema Change

```go
// internal/registry/agent.go — InstallCmd struct
type InstallCmd struct {
    Method         InstallMethod `json:"method"`
    URL            string        `json:"url"`
    Command        string        `json:"command"`
    NonInteractive bool          `json:"non_interactive"`
    UninstallCmd   string        `json:"uninstall,omitempty"`  // NEW
}
```

`UninstallCmd` is a simple string (not a nested struct) to keep the schema flat and forward-compatible. When empty, `UninstallAgent` derives the command from the install method.

## UninstallAgent Flow

```
UninstallAgent(agent)
    │
    ├─ agent.Install.UninstallCmd != ""? ────→ validate → sh -c <cmd> → done
    │
    └─ Derive from method:
        ├─ npm_install → extract package name → npm uninstall -g <pkg>
        ├─ curl_bash → exec.LookPath → os.Remove(binary)
        └─ custom → return error
```

### npm Package Extraction

```
Input:  "npm i -g @openai/codex"
Output: "@openai/codex"

Input:  "npm install -g @google/gemini-cli"
Output: "@google/gemini-cli"

Logic:  Find the last token in the command that
        doesn't start with '-'
```

### curl_bash Binary Removal

```go
func uninstallCurlBashFallback(agent registry.Agent) error {
    path, err := exec.LookPath(agent.DetectCmd)
    if err != nil {
        return fmt.Errorf("binary %q not found", agent.DetectCmd)
    }
    // Log the removal path
    logPath, _ := logPath(agent.ID)
    os.WriteFile(logPath, []byte("Removing: "+path+"\n"), 0644)
    // Remove — handles gone-after-check gracefully
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
        return fmt.Errorf("removing %s: %w", path, err)
    }
    return nil
}
```

Uses `os.Remove` directly instead of `rm -f $(which ...)` to eliminate shell injection risk.

## Remove Handler Extension

```go
type removeHandler struct {
    loadConfig       func(path string) (*config.Config, error)
    saveConfig       func(path string, cfg *config.Config) error
    configPath       func() (string, error)
    fetchRegistry    func(ctx context.Context, url string) (*registry.Catalog, error)
    uninstallAgent   func(agent registry.Agent) error
    isAgentInstalled func(detectCmd string) bool
    confirmFn        func(msg string) bool
}
```

### Default Wiring

```go
func defaultRemoveHandler() *removeHandler {
    return &removeHandler{
        loadConfig:       config.Load,
        saveConfig:       config.Save,
        configPath:       config.ConfigPath,
        fetchRegistry:    registry.Fetch,
        uninstallAgent:   installer.UninstallAgent,
        isAgentInstalled: installer.IsAgentInstalled,
        confirmFn:        confirmAction,
    }
}
```

### Confirm Action

```go
func confirmAction(msg string) bool {
    fmt.Fprint(os.Stderr, msg+" [y/N]: ")
    var input string
    fmt.Scanln(&input)
    input = strings.ToLower(strings.TrimSpace(input))
    return input == "y" || input == "yes"
}
```

## CLI Changes

### Command Registration

```go
// internal/cli/remove.go — newRemoveCommandWithHandler
func newRemoveCommandWithHandler(h *removeHandler) *cobra.Command {
    var uninstall bool
    var force bool

    cmd := &cobra.Command{
        Use:   "remove <agent-id>",
        Short: "Remove an agent from your config selection",
        Long: `Remove an agent from the selected_agents list in your config.
By default, this does NOT uninstall the agent from your system.
Use --uninstall to also remove the agent binary from your system.`,
        Args: cobra.ExactArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            return runRemoveFlow(h, cmd, args, uninstall, force)
        },
    }

    cmd.Flags().BoolVar(&uninstall, "uninstall", false,
        "Also uninstall the agent binary from the system")
    cmd.Flags().BoolVar(&force, "force", false,
        "Skip confirmation prompt")

    return cmd
}
```

### Remove Flow (updated)

```
squad remove [--uninstall] [--force] <agent-id>
    │
    ├─ Read config
    │
    ├─ --uninstall? ─────────────────────────→
    │   ├─ Fetch registry
    │   ├─ Find agent in registry
    │   ├─ Agent in registry? ──no──→ warn, continue
    │   ├─ Agent installed? ──no──→ notice, skip
    │   ├─ --force? ──no──→ prompt → declined? → abort
    │   └─ UninstallAgent(agent)
    │
    └─ Remove agent from config (always)
```

## Registry Data Updates (agents.json)

| Agent | Method | Uninstall Command |
|-------|--------|-------------------|
| pi | npm | `npm uninstall -g @earendil-works/pi-coding-agent` |
| codex | npm | `npm uninstall -g @openai/codex` |
| gemini-cli | npm | `npm uninstall -g @google/gemini-cli` |
| claude-code | curl_bash | _(fallback: rm binary)_ |
| opencode | curl_bash | _(fallback: rm binary)_ |
| antigravity-cli | curl_bash | _(fallback: rm binary)_ |
| gentle-ai | curl_bash | _(fallback: rm binary)_ |

Note: pi's install method changed from curl_bash to npm in the uninstall data. This is because the user's instructions explicitly specify an npm uninstall command for pi. The agent entry in `agents.json` still shows `method: "curl_bash"` — the explicit `uninstall` field overrides the method-derived fallback.

## File Changes

| File | Action |
|------|--------|
| `internal/registry/agent.go` | Add `UninstallCmd` field to `InstallCmd` |
| `internal/installer/uninstall.go` | **Create**: `UninstallAgent`, `extractNPMPackage`, `uninstallCurlBashFallback` |
| `internal/installer/uninstall_test.go` | **Create**: tests for all uninstall scenarios |
| `internal/cli/remove.go` | Update handler, add flags, extend flow |
| `internal/cli/remove_test.go` | Add tests for `--uninstall` and `--force` |
| `registry/agents.json` | Add `uninstall` fields for npm agents |

## Risks and Mitigations

| Risk | Mitigation |
|------|------------|
| Shell injection via detect_cmd in rm fallback | Use `exec.LookPath` + `os.Remove` (no shell) |
| `os.Remove` fails on permission denied | Error surfaces to user; they can retry with `sudo` |
| npm package name extraction fails | Returns error; user can file registry fix |
| User accidentally uninstalls | `--uninstall` must be explicitly set; confirmation prompt |
| Registry fetch fails during uninstall | Error message guides user; agent still removed from config |
