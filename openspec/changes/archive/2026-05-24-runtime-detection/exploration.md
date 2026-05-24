# Exploration: Runtime Detection

## Current State

The `internal/runtime/` directory exists as an empty placeholder (`.gitkeep`). The project has no runtime detection capability yet. The `registry.Agent` type already defines `RuntimeDep` with `Runtime` and `MinVersion` fields, but no code consumes them yet.

The project has two established package patterns:
- **`internal/config/`**: single-file struct + functions. `Config` struct with `Load()`, `Save()`, `ConfigPath()`. Tests use `t.TempDir()` and table-driven patterns.
- **`internal/registry/`**: split into `agent.go` (types) and `client.go` (operations). Tests use `httptest.Server` for HTTP cases.

Both follow: no global state, explicit dependencies via parameters, no external dependencies beyond stdlib + testify.

## Affected Areas

- `internal/runtime/runtime.go` — New file: RuntimeInfo struct and detector functions
- `internal/runtime/runtime_test.go` — New file: table-driven tests for parsing and detection
- `internal/registry/agent.go` — `RuntimeDep` already exists, no change needed
- `internal/installer/` — Future consumer of runtime detection (not in this change)
- `internal/tui/` — Future consumer for showing blocked agents (not in this change)

## Approaches

### 1. Simple functions with RuntimeInfo struct — RECOMMENDED

One file, one `RuntimeInfo` struct, one function per runtime (`DetectNode`, `DetectGo`, `DetectPython`, `DetectCargo`), plus `DetectAll()` and `VersionSatisfies()`. Version parsing with a private `parseSemver` helper.

- **Pros**: Simple, easy to test, no interfaces to wire, aligns with existing `config` package pattern, zero external deps
- **Cons**: Slightly repetitive per-runtime (acceptable for 4 runtimes)
- **Effort**: Low

### 2. Interface-based detector

Define `Detector` interface with `Detect() (*RuntimeInfo, error)`, implement per runtime, register in a map.

- **Pros**: Extensible for new runtimes without touching existing code
- **Cons**: Over-engineered for 4 runtimes, violates "prefer simplicity" project rule, developer is learning Go
- **Effort**: Medium

### 3. Single generic detect function

One function `Detect(cmd, args, parseFn)` that takes the command and a parsing callback.

- **Pros**: No repetition
- **Cons**: Callback signature gets complex (multiple output formats per runtime), harder to test, harder to read
- **Effort**: Low (implementation) / Medium (testing)

## Recommendation

**Approach 1: Simple functions.** It matches the project's established pattern (see `config.go` — no interfaces, direct functions), keeps cognitive load low for the learner, and is trivial to refactor later if needed. The repetition of 4 similar functions is acceptable for clarity.

**Specific design decisions:**

1. **Single file**: `runtime.go` containing `RuntimeInfo` struct + 4 detect functions + `DetectAll()` + `VersionSatisfies()` + private `parseSemver`. Split later if file exceeds 40 lines per function (shouldn't exceed ~200 lines total).

2. **RuntimeInfo struct**:
```go
type RuntimeInfo struct {
    Name      string // "node", "go", "python", "cargo"
    Installed bool
    Version   string // Parsed semver: "22.3.0" (no leading v)
    RawOutput string // Raw stdout from version command
    Err       error  // exec error if command not found
}
```

3. **Version parsing per runtime**:
   - `node --version` → `v22.3.0` → strip `v`, parse semver
   - `go version` → `go version go1.24.0 linux/amd64` → extract `go1.x.y`, strip `go`, parse semver
   - `python3 --version` → `Python 3.12.1` → strip `Python `, parse semver
   - `cargo --version` → `cargo 1.78.0 (xxx)` → extract first semver-like token

4. **`VersionSatisfies`** compares `[]int{major, minor, patch}` numerically. No external semver library needed.

5. **Edge case strategy**:
   - **Not installed**: `exec.Command` returns `exec.ErrNotFound` → `Installed=false, Err=nil` (not an error to the caller, just not installed)
   - **Unexpected output**: best-effort parse, `Installed=true, Version=""`, `Err` contains parse error
   - **Version below minimum**: `VersionSatisfies` returns `false`

## Risks

- **`python` vs `python3`**: Some systems only have `python` (old), others `python3` (modern). The PRD specifies `python3`. If we check both, complexity increases. Risk: some systems may not have `python3` symlink. **Mitigation**: Follow PRD (python3 only), document as known limitation.
- **Go version format**: `go version` output changes across Go versions. The current format has been stable since Go 1.0, but if it changes, parsing breaks. **Mitigation**: Tests with known output strings.
- **Cargo version format**: Cargo's output may vary with different versions. **Mitigation**: Robust extraction (find first `X.Y.Z` pattern).

## Ready for Apply

Yes. The design is clear, the PRD section 9 is explicit, the package is already allocated in AGENTS.md, and no design decisions need user input.
