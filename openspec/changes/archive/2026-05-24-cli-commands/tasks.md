# Tasks: CLI Commands (install + list)

## Task 1: Create `internal/cli/install.go`

**Files:** `internal/cli/install.go`

- Define `installHandler` struct with injectable function fields
- Implement `defaultInstallHandler()` wired to real implementations
- Implement `newInstallCommand()` returning `*cobra.Command`
- Flow: read config → fetch registry → determine targets → detect installed → check runtimes → filter → install → report
- Print one-line-per-agent progress via ProgressFn
- Handle `--agents` (comma-separated string flag) and `--all` (bool flag)
- Validate flag exclusivity (`--agents` + `--all` → error)
- On `--agents`: update config's selected_agents
- On registry fetch failure: fall back to cache, error if no cache

**Status:** [x]

## Task 1: Create `internal/cli/install.go`

**Files:** `internal/cli/install.go`

- Define `installHandler` struct with injectable function fields
- Implement `defaultInstallHandler()` wired to real implementations
- Implement `newInstallCommand()` returning `*cobra.Command`
- Flow: read config → fetch registry → determine targets → detect installed → check runtimes → filter → install → report
- Print one-line-per-agent progress via ProgressFn
- Handle `--agents` (comma-separated string flag) and `--all` (bool flag)
- Validate flag exclusivity (`--agents` + `--all` → error)
- On `--agents`: update config's selected_agents
- On registry fetch failure: error with user-friendly message

**Status:** [x]

## Task 2: Create `internal/cli/list.go`

**Files:** `internal/cli/list.go`

- Define `listHandler` struct with injectable function fields
- Implement `defaultListHandler()` wired to real implementations
- Implement `newListCommand()` returning `*cobra.Command`
- Flow: fetch registry → read config → detect installed → check runtimes → format table → print
- Determine status per agent (installed/selected/blocked/available)
- Print fixed-width table without external libraries
- On registry fetch failure: same fallback as install

**Status:** [x]

## Task 3: Create `internal/cli/install_test.go`

**Files:** `internal/cli/install_test.go`

- Table-driven tests for install command scenarios:
  - Default install with config (selects matching agents from registry)
  - `--agents` flag parses and filters correctly
  - `--all` flag installs all compatible
  - No config → no-op
  - Mixed flags error
  - Registry fetch failure
  - Skip already installed agents
  - Skip blocked agents (runtime not met)
  - Partial failures don't stop others

**Status:** [x]

## Task 4: Create `internal/cli/list_test.go`

**Files:** `internal/cli/list_test.go`

- Table-driven tests for list command scenarios:
  - Normal table output with installed/selected/blocked agents
  - Empty registry
  - Registry fetch failure
  - Verify output format matches expected columns

**Status:** [x]

## Task 5: Update `internal/cli/root.go`

**Files:** `internal/cli/root.go`

- Add `cmd.AddCommand(newInstallCommand())`
- Add `cmd.AddCommand(newListCommand())`

**Status:** [x]

## Task 6: Build and verify compilation

- Run `go build ./cmd/squad` to verify compilation
- Run `go vet ./...` to catch issues

**Status:** [x]

## Task 3: Create `internal/cli/install_test.go`

**Files:** `internal/cli/install_test.go`

- Table-driven tests for install command scenarios:
  - Default install with config (selects matching agents from registry)
  - `--agents` flag parses and filters correctly
  - `--all` flag installs all compatible
  - No config → no-op
  - Mixed flags error
  - Registry fetch failure
  - Skip already installed agents
  - Skip blocked agents (runtime not met)
  - Partial failures don't stop others

**Status:** [ ]

## Task 4: Create `internal/cli/list_test.go`

**Files:** `internal/cli/list_test.go`

- Table-driven tests for list command scenarios:
  - Normal table output with installed/selected/blocked agents
  - Empty registry
  - Registry fetch failure
  - Verify output format matches expected columns

**Status:** [ ]

## Task 5: Update `internal/cli/root.go`

**Files:** `internal/cli/root.go`

- Add `cmd.AddCommand(newInstallCommand())`
- Add `cmd.AddCommand(newListCommand())`

**Status:** [ ]

## Task 6: Build and verify compilation

- Run `go build ./cmd/squad` to verify compilation
- Run `go vet ./...` to catch issues

**Status:** [ ]
