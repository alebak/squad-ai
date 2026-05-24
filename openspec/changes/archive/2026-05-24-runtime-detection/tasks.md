# Tasks: Runtime Detection

## Forecast

| Guard | Value |
|-------|-------|
| Decision needed before apply | No |
| Chained PRs recommended | No |
| 400-line budget risk | **Low** (~240 lines total: 90 + 150) |

## Tasks

### 1. internal/runtime: Types and helpers

- [x] **1.1** Define `RuntimeInfo` struct in `internal/runtime/runtime.go` with godoc
- [x] **1.2** Implement private `parseNodeVersion`, `parseGoVersion`, `parsePythonVersion` — each takes raw stdout string, returns (version string, err). Table-driven tests for each parser with 3+ cases per runtime.
- [x] **1.3** Implement `compareVersions(a, b string) int` — returns -1, 0, or 1. Table-driven test with 5+ pairs.

**Verify**: `go test ./internal/runtime/ -run TestParse` passes.

### 2. internal/runtime: Detect functions

- [ ] **2.1** Add package-level `var execCommand = exec.Command` for test injection *(skipped — user chose direct exec.Command per apply instruction)*
- [x] **2.2** Implement private `detectRuntime(name, cmd, args string, parseFn func(string) (string, error)) RuntimeInfo` shared helper
- [x] **2.3** Implement `DetectNode()` using helper + `parseNodeVersion`
- [x] **2.4** Implement `DetectGo()` using helper + `parseGoVersion`
- [x] **2.5** Implement `DetectPython()` using helper + `parsePythonVersion`

**Verify**: `go build ./...` compiles.

### 3. internal/runtime: Tests with injection

- [ ] **3.1** Write `TestDetectNode_ParsesVersion` — replace execCommand with fake returning `v22.3.0\n` *(skipped — user chose real-system tests)*
- [ ] **3.2** Write `TestDetectGo_ParsesVersion` — fake returning `go version go1.24.0 linux/amd64\n` *(skipped — real-system)*
- [ ] **3.3** Write `TestDetectPython_ParsesVersion` — fake returning `Python 3.12.1\n` *(skipped — real-system)*
- [ ] **3.4** Write `TestDetectRuntime_NotInstalled` — point PATH to empty dir, verify Installed=false *(skipped — real-system)*
- [ ] **3.5** Write `TestDetectRuntime_UnexpectedOutput` — fake returns garbage, verify Version="" *(skipped — real-system)*

**Verify**: `go test ./internal/runtime/ -v -count=1` all green.

### 4. internal/runtime: IsCompatible

- [x] **4.1** Implement `IsCompatible(info RuntimeInfo, minVersion string) bool` using `compareVersions`
- [x] **4.2** Write table-driven `TestIsCompatible` — covers meets min, below min, not installed, malformed version, empty version

**Verify**: `go test ./internal/runtime/ -v -count=1` — 100% pass, no race conditions.

### Effort

- ~240 lines total across 2 files
- 4 task groups, completable in one session
- No new dependencies
