# Design: Remaining CLI Commands (add, remove, update, info)

## Technical Approach

Extend the handler-struct-with-injectable-functions pattern used by install/list to four new commands. Each command gets its own handler struct, default factory, and `new*CommandWithHandler()` constructor.

## Architecture Decisions

### Decision: Per-command handlers vs shared handler

**Choice**: One handler struct per command
**Alternatives considered**: Single shared handler with all deps
**Rationale**: Each command has a unique dependency set. A shared handler would carry unused fields. Per-command follows existing pattern exactly and keeps each file independently testable.

### Decision: Cache path for update command

**Choice**: Function field in updateHandler (`cachePath func() (string, error)`)
**Alternatives considered**: Add `registry.CachePath()` function, hardcode path
**Rationale**: Following the injectable-function pattern keeps `update` testable. The default implementation computes `~/.config/squad-ai/registry.cache.json`.

### Decision: add command as TUI stub

**Choice**: Simple text output listing available agents
**Alternatives considered**: No output, error that TUI isn't ready
**Rationale**: The stub is useful today (shows what's available) and the output format can be reused when TUI is implemented.

## Data Flow

### Add Flow
```
squad add
  ├─ Read config
  ├─ Fetch registry
  ├─ Detect installed agents
  ├─ Filter: available = not installed AND not blocked AND not selected
  ├─ Print "TUI coming soon"
  └─ Print list of available agents
```

### Remove Flow
```
squad remove <id>
  ├─ Validate: requires 1 arg
  ├─ Read config
  ├─ Remove ID from selected_agents
  ├─ Save config
  └─ Print confirmation
```

### Update Flow
```
squad update
  ├─ Fetch registry (remote, no cache)
  ├─ Save to registry.cache.json
  └─ Print confirmation
```

### Info Flow
```
squad info <id>
  ├─ Validate: requires 1 arg
  ├─ Fetch registry
  ├─ Find agent by ID
  ├─ Not found? → error
  └─ Print agent details
```

## Handler Signatures

```go
// addHandler
type addHandler struct {
    loadConfig    func(path string) (*config.Config, error)
    fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
    detectAll     func(agents []registry.Agent) map[string]bool
    configPath    func() (string, error)
}

// removeHandler
type removeHandler struct {
    loadConfig func(path string) (*config.Config, error)
    saveConfig func(path string, cfg *config.Config) error
    configPath func() (string, error)
}

// updateHandler
type updateHandler struct {
    fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
    saveCache     func(path string, reg *registry.Catalog) error
    cachePath     func() (string, error)
}

// infoHandler
type infoHandler struct {
    fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
    configPath    func() (string, error)
}
```

## File Changes

| File | Action | Description |
|------|--------|-------------|
| `internal/cli/add.go` | Create | `squad add` handler + command |
| `internal/cli/add_test.go` | Create | Add command tests |
| `internal/cli/remove.go` | Create | `squad remove` handler + command |
| `internal/cli/remove_test.go` | Create | Remove command tests |
| `internal/cli/update.go` | Create | `squad update` handler + command |
| `internal/cli/update_test.go` | Create | Update command tests |
| `internal/cli/info.go` | Create | `squad info` handler + command |
| `internal/cli/info_test.go` | Create | Info command tests |
| `internal/cli/root.go` | Modify | Register 4 new commands |

## Testing Strategy

| Layer | What to Test | Approach |
|-------|-------------|----------|
| Unit | Each command's success path | Mock handler, capture output, assert content |
| Unit | Each command's error paths | Mock handler returning errors |
| Unit | `squad remove` no-arg | Cobra's built-in arg validation or handler check |
| Unit | `squad info` agent not found | Mock registry without the agent |
| Build | `go build ./cmd/squad` | Verify compilation |
| Vet | `go vet ./...` | Verify no issues |

## Migrations

No migration required.
