# Tasks: Agent Uninstall Support

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | 200-250 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Delivery strategy | single-pr |

## Phase 1: Registry Schema

### Task 1.1: Add UninstallCmd field to InstallCmd

**Files:** `internal/registry/agent.go`

- Add `UninstallCmd string \`json:"uninstall,omitempty"\`` to the `InstallCmd` struct
- Existing fields remain unchanged
- Verify `go build` still compiles

**Status:** [x]

## Phase 2: Installer Package

### Task 2.1: Create `internal/installer/uninstall.go`

**Files:** `internal/installer/uninstall.go`

- Implement `UninstallAgent(agent registry.Agent) error`:
  - If `agent.Install.UninstallCmd` is non-empty, validate and run via `sh -c`
  - If method is `MethodNpmInstall`, extract package and derive `npm uninstall -g <pkg>`
  - If method is `MethodCurlBash`, call `exec.LookPath` + `os.Remove`
  - If method is `MethodCustom`, return error
- Implement `ExtractNPMPackage(installCmd string) string`
- Implement `uninstallCurlBashFallback(agent registry.Agent) error`

**Status:** [x]

### Task 2.2: Create `internal/installer/uninstall_test.go`

**Files:** `internal/installer/uninstall_test.go`

- `TestUninstallAgent_ExplicitCommand` — explicit UninstallCmd executes
- `TestUninstallAgent_NpmFallback` — derives npm uninstall from install command
- `TestUninstallAgent_CurlBashFallback` — resolves and removes binary
- `TestUninstallAgent_CustomNoUninstall` — returns error
- `TestUninstallAgent_NullByte` — rejects null byte in command
- `TestUninstallAgent_CurlBashEmptyDetectCmd` — returns error
- `TestExtractNPMPackage` — various npm install command formats
- `TestUninstallAgent_NpmFallback_NoPackage` — edge case: no package found

**Status:** [x]

## Phase 3: CLI

### Task 3.1: Update `internal/cli/remove.go`

**Files:** `internal/cli/remove.go`

- Add `registryURL`, `fetchRegistry`, `uninstallAgent`, `isAgentInstalled`, `confirmFn` fields to `removeHandler`
- Update `defaultRemoveHandler()` to wire real implementations
- Add `--uninstall` and `--force` flags to `newRemoveCommandWithHandler`
- Update `runRemoveFlow` signature to accept `uninstall`, `force` bool params
- Implement uninstall flow:
  1. Fetch registry (if `--uninstall`)
  2. Find agent in registry
  3. Check if installed
  4. Prompt for confirmation (unless `--force`)
  5. Call `uninstallAgent`
  6. Always remove from config
- Update command `Long` description to document new flags

**Status:** [x]

### Task 3.2: Update `internal/cli/remove_test.go`

**Files:** `internal/cli/remove_test.go`

- `TestRemoveCommand_UninstallSuccess` — agent uninstalled then removed from config
- `TestRemoveCommand_UninstallWithForceSkipsConfirm` — force skips confirmation
- `TestRemoveCommand_UninstallNotInstalled` — skip uninstall for missing binary
- `TestRemoveCommand_UninstallNotInRegistry` — warn and still remove from config
- `TestRemoveCommand_UninstallConfirmCancels` — cancelled confirm doesn't uninstall
- `TestRemoveCommand_UninstallClaudeCode` — curl_bash fallback without explicit uninstall
- Existing tests must still pass unchanged

**Status:** [x]

## Phase 4: Registry Data

### Task 4.1: Update `registry/agents.json`

**Files:** `registry/agents.json`

- Add `"uninstall": "npm uninstall -g @earendil-works/pi-coding-agent"` to pi entry
- Add `"uninstall": "npm uninstall -g @openai/codex"` to codex entry
- Add `"uninstall": "npm uninstall -g @google/gemini-cli"` to gemini-cli entry
- Leave curl_bash agents without uninstall field (fallback applies)

**Status:** [x]

## Phase 5: Verification

### Task 5.1: Build and test

- Run `go build ./...` to verify compilation
- Run `go test ./... -v -count=1` to verify all tests pass
- Run `go vet ./...` to catch issues

**Status:** [x]

### Task 5.2: Verify spec compliance

- All scenarios from `spec.md` are covered by tests or verified manually
- `squad remove --help` shows new flags
- `squad remove <id>` still works without flags (backward compat)

**Status:** [x]
