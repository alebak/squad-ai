# Design: CLI Commands (install + list)

## Architecture Decision

**Use handler structs with injectable function fields** for both commands. This provides:
- Testability: tests can pass mock implementations for config/registry/runtime
- No global mutable state: all dependencies are explicit
- Follows existing patterns: `ProgressFn` in installer already uses function-typed callbacks
- Minimal boilerplate: no interfaces or DI framework needed

## Handler Structure

### installHandler

```go
type installHandler struct {
    loadConfig     func(path string) (*config.Config, error)
    fetchRegistry  func(ctx context.Context, url string) (*registry.Catalog, error)
    detectAll      func(agents []registry.Agent) map[string]bool
    installAll     func(agents []registry.Agent, progress installer.ProgressFn) []error
    isRuntimeMet   func(deps []registry.RuntimeDep) bool
    configPath     func() (string, error)
}
```

### listHandler

```go
type listHandler struct {
    loadConfig     func(path string) (*config.Config, error)
    fetchRegistry  func(ctx context.Context, url string) (*registry.Catalog, error)
    detectAll      func(agents []registry.Agent) map[string]bool
    isRuntimeMet   func(deps []registry.RuntimeDep) bool
    configPath     func() (string, error)
}
```

## Flow Diagrams

### Install Flow

```
squad install [--agents a,b] [--all]
         │
         ├─ Read config (config.Load)
         ├─ Fetch registry (remote or cache)
         ├─ Determine target agents:
         │   ├──agents → parse CSV
         │   ├──all → all from registry
         │   └─ default → config.SelectedAgents
         ├─ Detect installed (installer.DetectAll)
         ├─ Detect runtimes (runtime.Detect*)
         ├─ Filter:
         │   ├─ Skip already installed
         │   └─ Skip blocked (runtime not met)
         ├─ Install each (installer.InstallAll)
         └─ Report results (✅/❌ lines)
```

### List Flow

```
squad list
    │
    ├─ Fetch registry (remote or cache)
    ├─ Read config
    ├─ Detect installed (installer.DetectAll)
    ├─ For each agent:
    │   ├─ Is installed? → Installed=✅, Status=installed
    │   ├─ In selected_agents? → Status=selected
    │   ├─ Runtime met? → Status=blocked
    │   └─ Otherwise → Status=available
    └─ Print header + rows
```

## Runtime Compatibility Check

The `isRuntimeMet` function checks each agent's `Dependencies` list:

```go
func defaultIsRuntimeMet(deps []registry.RuntimeDep) bool {
    for _, dep := range deps {
        switch dep.Runtime {
        case "none":
            continue  // always compatible
        case "node":
            info := runtime.DetectNode()
            if dep.MinVersion != "" {
                if !runtime.IsCompatible(info, dep.MinVersion) {
                    return false
                }
            } else if !info.Installed {
                return false
            }
        case "go":
            info := runtime.DetectGo()
            if dep.MinVersion != "" {
                if !runtime.IsCompatible(info, dep.MinVersion) {
                    return false
                }
            } else if !info.Installed {
                return false
            }
        case "python":
            info := runtime.DetectPython()
            if dep.MinVersion != "" {
                if !runtime.IsCompatible(info, dep.MinVersion) {
                    return false
                }
            } else if !info.Installed {
                return false
            }
        default:
            // Unknown runtime — block to be safe
            return false
        }
    }
    return true
}
```

## Progress Callback

Both commands create a `ProgressFn` that prints to stdout:
```
✅ Claude Code installed
❌ OpenCode — exited with code 1 (see log: ...)
```

## Config Save (install only)

When running `squad install --agents <ids>`, the command updates the config's `selected_agents` with the requested agents. For default and `--all` modes, the config already reflects the desired state.

## File Changes

| File | Action |
|------|--------|
| `internal/cli/install.go` | Create: handler struct + `newInstallCommand()` |
| `internal/cli/install_test.go` | Create: table-driven tests |
| `internal/cli/list.go` | Create: handler struct + `newListCommand()` |
| `internal/cli/list_test.go` | Create: table-driven tests |
| `internal/cli/root.go` | Update: register both commands |
