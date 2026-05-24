# Runtime Detection Specification

## Purpose

The runtime-detection domain identifies available language runtimes (Node.js, Go, Python) on the host system. It provides structured info for agent dependency verification — blocking installation when a required runtime is missing.

## Requirements

### Requirement: RuntimeInfo Type

RuntimeInfo SHALL contain fields: Name (string), Installed (bool), Version (string), RawOutput (string), Err (error). Every detect call SHALL return a populated RuntimeInfo — never nil.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Node.js installed | node v22.3.0 in PATH | DetectNode() | Name="node", Installed=true, Version="22.3.0" |
| Go installed | go 1.24.0 in PATH | DetectGo() | Version="1.24.0", Installed=true |
| Python installed | python3 3.12.1 in PATH | DetectPython() | Version="3.12.1", Installed=true |

### Requirement: Runtime Not Installed

When a runtime binary is not found, detect functions SHALL return Installed=false, Version="" and MUST NOT return an error. The caller distinguishes "not installed" from "something broke."

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Node.js missing | no node binary in PATH | DetectNode() | Installed=false, Version="", Err=nil |
| python3 absent | no python3 in PATH | DetectPython() | Installed=false, Err=nil |

### Requirement: Version Parsing

Each detect function SHALL parse version output per runtime format:
- Node: strip leading "v" from `node --version` stdout
- Go: extract `go X.Y.Z` from `go version` output, then strip "go " prefix
- Python: strip "Python " prefix from `python3 --version` stdout

Parsing SHALL use strings.TrimPrefix and strings.Fields — no external semver library.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Node version with v prefix | output "v22.3.0\n" | version parsed | Version="22.3.0" |
| Go version format | output "go version go1.24.0 linux/amd64\n" | version parsed | Version="1.24.0" |
| Python version format | output "Python 3.12.1\n" | version parsed | Version="3.12.1" |
| Unexpected output | command returns garbage text | version parsed | Installed=true, Version="", Err has parse error |

### Requirement: IsCompatible (Optional)

The system SHOULD provide IsCompatible(info RuntimeInfo, minVersion string) bool for future agent dependency checks. It SHALL compare versions numerically as `[]int{major, minor, patch}`.

| Scenario | GIVEN | WHEN | THEN |
|----------|-------|------|------|
| Version meets min | Version="22.3.0", min="18.0.0" | IsCompatible() | true |
| Version below min | Version="16.0.0", min="18.0.0" | IsCompatible() | false |
| Runtime not installed | Installed=false | IsCompatible() with any min | false |
| Malformed version | Version="abc" | IsCompatible() with "18.0.0" | false |
