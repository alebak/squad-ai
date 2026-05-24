## Exploration: CLI Commands (install + list)

### Current State

The project has a functional CLI scaffold with a root command (`internal/cli/root.go`) and a `version` subcommand. All core packages are built:

- **config** (`internal/config/config.go`): Load/Save user config from `~/.config/squad-ai/config.json`. Includes `ConfigPath()`, `Load()`, `Save()` with atomic writes.
- **registry** (`internal/registry/`): `Agent`, `Catalog` types, `Fetch()` (HTTP GET + JSON parse), `LoadCache()`, `SaveCache()`, `IsStale()`.
- **runtime** (`internal/runtime/runtime.go`): `DetectNode()`, `DetectGo()`, `DetectPython()`, `IsCompatible()`, version comparison.
- **installer** (`internal/installer/`): `DetectAgent()`/`DetectAll()` (binary lookup via `exec.LookPath`), `InstallAgent()`/`InstallAll()` (shell exec + log capture), `ProgressFn` callback.

Packages have thorough tests using table-driven patterns and `testify` assertions. The entry point (`cmd/squad/main.go`) only calls `cli.NewRootCommand().Execute()`.

### What's Missing

1. **`squad install`** command — needs to wire: read config → fetch registry → detect installed → detect runtimes → filter blocked → install missing
2. **`squad list`** command — needs to wire: fetch registry → read config → detect installed → detect runtimes → print table
3. Registration of both subcommands in `NewRootCommand()`

### Affected Areas

- `internal/cli/install.go` — **new file**: `squad install` command definition
- `internal/cli/install_test.go` — **new file**: tests for install command
- `internal/cli/list.go` — **new file**: `squad list` command definition
- `internal/cli/list_test.go` — **new file**: tests for list command
- `internal/cli/root.go` — **update**: register both new subcommands

### Approaches

1. **Direct function calls in RunE** — simplest, matches `version.go` pattern. Test by setting up real temp files and capturing output. Pro: minimal code. Con: tests require real HTTP server or config files.
   
2. **Interface-based command handlers** — define `RegistryFetcher`, `ConfigLoader`, etc. interfaces, inject into command structs. Pro: clean testability. Con: premature abstraction for MVP.

3. **Package-level var overrides** — define `var loadConfigFn = config.Load` etc., override in tests. Pro: simple test overrides. Con: mutable package state, not goroutine-safe.

4. **Function-field command struct** — command struct with injectable function fields, constructor sets default impls. Pro: testable, no mutable state, explicit. Con: slightly more boilerplate.

### Recommendation

**Approach 4**: Struct with injectable function fields, following the project's existing patterns. The function fields are set to real implementations in production and can be swapped in tests. This is the Go-idiomatic middle ground — no interfaces needed, no mutable package state.

```go
type installHandler struct {
    loadConfig    func(path string) (*config.Config, error)
    fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
    detectAll     func(agents []registry.Agent) map[string]bool
    installAll    func(agents []registry.Agent, progress installer.ProgressFn) []error
    detectRuntime func(name string) runtime.RuntimeInfo
}
```

### Risks

- Registry fetch failure must not crash the commands — graceful fallback to cache
- `--all` flag + empty registry → no-op, must not error
- Runtime detection is real OS calls — tests must not depend on specific runtimes being present
- The PRD says `squad install` reads config by default, but if no config exists, it uses `DefaultConfig()` with empty `SelectedAgents` — which means installing nothing. That's correct per spec (first-run TUI handles initial selection).

### Ready for Proposal

Yes — the scope is clear, the risk is low, and all dependencies exist.
