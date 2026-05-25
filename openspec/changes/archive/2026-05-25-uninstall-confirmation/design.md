# Design: 3-Option Uninstall Confirmation

## Technical Approach

Extend the `Agent` struct with config paths, add a new `UninstallConfig` function, and replace the single `confirmFn` call in `runAddFlowInteractive` with a 3-option prompt loop. The prompt is plain `fmt` + `bufio.Scanner` (no Bubbletea) since it's a post-TUI CLI interaction.

## Architecture Decisions

### Decision: Plain CLI prompt vs Bubbletea sub-view

**Choice**: `fmt.Printf` + `bufio.Scanner` text prompt
**Alternatives**: Launch a second Bubbletea program for the uninstall prompt
**Rationale**: The TUI already finished (user pressed Enter). A text prompt is simpler, testable via stdin mocking, and avoids the complexity of composing two Bubbletea programs. The 3-option list is short enough for a text interface.

### Decision: ConfigPaths as registry data vs hardcoded map

**Choice**: Add `ConfigPaths []string` to the Agent struct in `agents.json`
**Alternatives**: Hardcode a mapping in Go code
**Rationale**: Registry data keeps it declarative and allows updating paths without recompiling. The path is a property of the agent, not of Squad AI.

### Decision: Path traversal safety

**Choice**: Expand paths relative to `$HOME`, reject any path that escapes via `..` or absolute paths
**Alternatives**: Use `filepath.Abs` and check prefix against `os.UserHomeDir`
**Rationale**: `filepath.Abs` + `strings.HasPrefix` against home dir catches both `../../etc` and absolute paths like `/etc/passwd`.

## Data Flow

```
TUI (selection) → selectedIDs []string
    ↓
For each deselected installed agent:
    ┌─── 3-option prompt ─────────────────────┐
    │  [1] Uninstall app only                 │
    │  [2] Uninstall app + config data        │
    │  [3] Cancel                             │
    └─────────────────────────────────────────┘
         │        │              │
         ▼        ▼              ▼
    Uninstall  Uninstall     Reselect
    Agent      Agent +       agent in
               Uninstall     result set
               Config
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/registry/agent.go` | Modify | Add `ConfigPaths []string` field with `json:"config_paths,omitempty"` |
| `internal/installer/uninstall.go` | Modify | Add `UninstallConfig` function |
| `internal/cli/add.go` | Modify | Replace confirmFn block with 3-option prompt |
| `registry/agents.json` | Modify | Add `config_paths` to all 7 agents |
| `internal/installer/uninstall_test.go` | Modify | Add tests for `UninstallConfig` |
| `internal/cli/add_test.go` | Modify | Update tests for new prompt behavior |

## Interfaces / Contracts

```go
// In internal/registry/agent.go:
type Agent struct {
    // ... existing fields ...
    ConfigPaths []string `json:"config_paths,omitempty"`
}

// In internal/installer/uninstall.go:
// UninstallConfig removes config directories for the given agent.
func UninstallConfig(agent registry.Agent) error { ... }

// In internal/cli/add.go - handler changes:
type addHandler struct {
    // ... existing fields remain the same ...
    // confirmFn stays for backward compat, but uninstall logic changes:
    // instead of calling confirmFn, we show a 3-option prompt
}
```

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | `UninstallConfig` | Create temp dirs, verify `os.RemoveAll` called on correct paths |
| Unit | Path traversal protection | Paths like `../../etc` are rejected |
| Unit | Empty/nil ConfigPaths | No-op returns nil |
| Unit | 3-option prompt logic | Mock stdin with "1\n", "2\n", "3\n", "invalid\n" |
| Integration | Full flow | Test that deselected installed agents trigger new prompt |

## Migration / Rollout

No migration required. Prompt change is purely behavioral — no data format changes. The new `config_paths` field in JSON is optional (`omitempty`), so existing `agents.json` without it still works (config cleanup is simply skipped).

## Open Questions

None.
