# Tasks: 3-Option Uninstall Confirmation

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~150-250 |
| 400-line budget risk | Low |
| Chained PRs recommended | No |
| Suggested split | Single PR |
| Delivery strategy | single-pr |

Decision needed before apply: No
Chained PRs recommended: No
Chain strategy: size-exception
400-line budget risk: Low

## Phase 1: Registry Model

- [x] 1.1 Add `ConfigPaths []string` field with JSON tag to `internal/registry/agent.go`
- [x] 1.2 Add `config_paths` entries to `registry/agents.json` for all 7 agents

## Phase 2: Installer

- [x] 2.1 Add `UninstallConfig(agent registry.Agent) error` to `internal/installer/uninstall.go`
- [x] 2.2 Add `UninstallConfig` tests to `internal/installer/uninstall_test.go`

## Phase 3: CLI Uninstall Prompt

- [x] 3.1 Add `defaultUninstallChoiceFn` handler field to `internal/cli/add.go`
- [x] 3.2 Add `uninstallConfig` handler field for config cleanup
- [x] 3.3 Replace the confirmFn block in `runAddFlowInteractive` with 3-option prompt
- [x] 3.4 Add tests for all 3 choices (cancel, app-only, app+config)

## Phase 4: Verify

- [x] 4.1 Run `go build ./cmd/squad` — build passes
- [x] 4.2 Run `go test ./...` — all tests pass
