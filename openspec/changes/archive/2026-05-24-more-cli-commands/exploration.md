## Exploration: Remaining CLI Commands (add, remove, update, info)

### Current State

The CLI has 4 working commands (`root`, `version`, `install`, `list`) following the handler-struct-with-injectable-functions pattern:

- `internal/cli/install.go` — `installHandler` struct, `defaultInstallHandler()`, `newInstallCommandWithHandler()`
- `internal/cli/list.go` — `listHandler` struct, `defaultListHandler()`, `newListCommandWithHandler()`
- `internal/cli/root.go` — registers all commands in `NewRootCommand()`
- Tests: table-driven, mock handlers injected, output captured via `bytes.Buffer`

Each handler struct holds function fields (`loadConfig`, `fetchRegistry`, etc.) set to real implementations in production, swapped in tests.

### What's Missing

Four commands from PRD §8:

| Command | Description |
|---------|-------------|
| `squad add` | TUI stub: shows "TUI coming soon", lists available agents not installed |
| `squad remove <id>` | Removes agent ID from config `selected_agents` (does NOT uninstall) |
| `squad update` | Forces registry re-fetch, saves to cache, prints confirmation |
| `squad info <id>` | Shows agent details from registry (version, runtime, install method, description) |

### Affected Areas

- `internal/cli/add.go` — NEW: `squad add` command
- `internal/cli/add_test.go` — NEW: tests for add command
- `internal/cli/remove.go` — NEW: `squad remove` command
- `internal/cli/remove_test.go` — NEW: tests for remove command
- `internal/cli/update.go` — NEW: `squad update` command
- `internal/cli/update_test.go` — NEW: tests for update command
- `internal/cli/info.go` — NEW: `squad info` command
- `internal/cli/info_test.go` — NEW: tests for info command
- `internal/cli/root.go` — MODIFY: register 4 new subcommands

### Approaches

1. **Direct handler struct per command** — each gets its own handler+test file. Pro: follows existing pattern exactly, clean separation. Con: some duplication across handlers (shared deps). Effort: Low

2. **Generic command registry** — one handler struct with all possible deps, factory functions. Pro: less boilerplate. Con: violates separation of concerns, unused deps in each command. Effort: Med

3. **Inline in root.go** — no handlers, direct cobra RunE with inline calls. Pro: simplest. Con: untestable, breaks pattern. Effort: Low

### Recommendation

**Approach 1**: One handler struct per command, each in its own file. Matches existing pattern exactly, keeps each command independently testable.

- `add` handler: `loadConfig`, `fetchRegistry`, `detectAll`, `configPath`
- `remove` handler: `loadConfig`, `saveConfig`, `configPath`
- `update` handler: `fetchRegistry`, `saveCache`, `cachePath`
- `info` handler: `fetchRegistry`, `configPath`

### Risks

- `add` is a stub — behavior will change when TUI is implemented. Keep it simple.
- `remove` only touches config, never uninstalls — must be clear in messaging.
- `update` needs cache path — derive from config dir or pass as handler dep.
- `info` with unknown agent ID must error gracefully, not panic.

### Ready for Proposal

Yes — scope is clear, patterns are established, risk is low.
