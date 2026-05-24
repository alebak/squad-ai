# Tasks: Data Foundation — Registry + Config Models

## Review Workload Forecast

| Field | Value |
|-------|-------|
| Estimated changed lines | ~620 |
| 400-line budget risk | High |
| Chained PRs recommended | Yes |
| Suggested split | PR 1 (registry: data + types + client + tests) → PR 2 (config: types + tests) |
| Delivery strategy | ask-on-risk |
| Chain strategy | pending |

Decision needed before apply: Yes
Chained PRs recommended: Yes
Chain strategy: pending
400-line budget risk: High

### Suggested Work Units

| Unit | Goal | Likely PR | Notes |
|------|------|-----------|-------|
| 1 | Registry: data file, types, HTTP client, local cache, tests | PR 1 | Base = main. Agent type model, fetch with httptest, cache round-trip, staleness check. |
| 2 | Config: types, atomic file I/O, path resolution, tests | PR 2 | Base = main (independent of PR 1). Load/Save with temp+rename, DefaultConfig, ConfigPath. |

## Phase 1: Foundation — Registry Data & Types

- [x] 1.1 Create `registry/agents.json` with 7 MVP agents (IDs: claude-code, opencode, pi, codex, antigravity-cli, gemini-cli, gentle-ai)
- [x] 1.2 Create `internal/registry/agent.go` with Agent, InstallCmd, RuntimeDep, Checksum, Registry types + InstallMethod typed consts
- [x] 1.3 Delete `internal/registry/.gitkeep`

## Phase 2: Registry Client & Tests

- [x] 2.1 Create `internal/registry/client.go` with Fetch(), LoadCache(), SaveCache(), IsStale(), var RegistryURL
- [x] 2.2 Create `internal/registry/registry_test.go`: fetch (success/cancellation/non-200/network error), cache round-trip, IsStale (fresh/stale/missing), malformed JSON, empty list

## Phase 3: Config

- [x] 3.1 Create `internal/config/config.go` with Config, InstallOptions, Load(), Save() (atomic temp+rename), DefaultConfig(), ConfigPath()
- [x] 3.2 Create `internal/config/config_test.go`: round-trip, missing file, malformed JSON, atomic write cleanup, concurrent writes, dir creation, missing HOME
- [x] 3.3 Delete `internal/config/.gitkeep`

## Phase 4: Build & Verify

- [x] 4.1 Run `go build ./...` — fix all compilation errors
- [x] 4.2 Run `go test ./internal/...` — all tests pass
