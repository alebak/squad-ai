# Design: Runtime Detection

## Package

`internal/runtime/` — single file `runtime.go`, test file `runtime_test.go`.

## Data Model

```go
// RuntimeInfo describes a language runtime found (or not) on the system.
type RuntimeInfo struct {
    Name      string // "node", "go", "python"
    Installed bool
    Version   string // Semver without prefix, e.g. "22.3.0". Empty if not installed or parse failed.
    RawOutput string // Raw stdout from the version command, for diagnostics.
    Err       error  // exec error. nil when not installed — caller checks Installed instead.
}
```

## Detect Functions

Three exported functions following an identical pattern:

```go
func DetectNode() RuntimeInfo    // exec "node", "--version"
func DetectGo() RuntimeInfo      // exec "go", "version"
func DetectPython() RuntimeInfo  // exec "python3", "--version"
```

**Internals pattern** (shared private helper `detectRuntime`):

```
1. LookPath(cmd) — if exec.ErrNotFound, return RuntimeInfo{Name, Installed: false}
2. exec.Command(cmd, args...), run with Output()
3. If exec fails (binary found but crashed), return Installed=true, Err=exec error
4. Parse stdout with runtime-specific parser
5. If parse fails, return Installed=true, Version="" (best-effort)
```

### Version Parsing

Each runtime has a private parse function. All use `strings` stdlib only — no semver library.

| Runtime | Command | Raw Output | Parser |
|---------|---------|------------|--------|
| Node | `node --version` | `v22.3.0\n` | `strings.TrimPrefix(out, "v")` |
| Go | `go version` | `go version go1.24.0 linux/amd64\n` | Split fields, find token matching `goX.Y.Z`, then `strings.TrimPrefix(token, "go")` |
| Python | `python3 --version` | `Python 3.12.1\n` | `strings.TrimPrefix(out, "Python ")` |

### IsCompatible

```go
func IsCompatible(info RuntimeInfo, minVersion string) bool
```

- If `!info.Installed` → return false
- Parse both `info.Version` and `minVersion` into `[]int{major, minor, patch}` using `strings.Split` + `strconv.Atoi`
- Compare numerically: major first, then minor, then patch
- If either version fails to parse → return false
- Returns true if `info.Version >= minVersion`

## Error Handling

| Scenario | Behavior |
|----------|----------|
| Binary not in PATH | LookPath returns error → RuntimeInfo{Installed: false}, no error returned |
| Binary exists but fails | exec.Command error → RuntimeInfo{Installed: true, Err: wrapped error} |
| Unexpected output format | Parse fails → RuntimeInfo{Installed: true, Version: "", Err: parse error} |
| stdout empty | Parse fails (same as unexpected) → RuntimeInfo{Installed: true, Version: ""} |

## Table-Driven Tests

Test file `runtime_test.go` with these test functions:

```
TestDetectNode_ParsesVersion     — mock node --version, verify "22.3.0"
TestDetectGo_ParsesVersion       — mock go version, verify "1.24.0"
TestDetectPython_ParsesVersion   — mock python3 --version, verify "3.12.1"
TestDetectRuntime_NotInstalled   — PATH without binary, verify Installed=false
TestDetectRuntime_UnexpectedOutput — garbage stdout, verify Version=""
TestParseVersion_Node            — table: various raw outputs → expected versions
TestParseVersion_Go              — table: various go version outputs
TestParseVersion_Python          — table: various python version outputs
TestIsCompatible                 — table: version pairs → expected bool
```

Because actual `exec.Command` calls depend on the host, tests use a **command-dependency injection**: a package-level var `var execCommand = exec.Command` lets tests replace `exec.Command` with a fake function that returns known output without running real binaries.

## Files

| File | Lines (est.) | Purpose |
|------|-------------|---------|
| `internal/runtime/runtime.go` | ~90 | RuntimeInfo, detect functions, parsers, IsCompatible |
| `internal/runtime/runtime_test.go` | ~150 | Table-driven tests with command injection |

## Decision: Command Injection Over Interface

Tests replace `exec.Command` via a package var rather than defining a `Detector` interface. Rationale: 3 functions with identical wiring don't justify an interface. This matches the `config` package pattern (direct functions, no interfaces) and keeps the code accessible for a Go learner. If tests for 4+ runtimes become unwieldy, extract an interface later.
