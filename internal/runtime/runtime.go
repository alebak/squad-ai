// Package runtime detects language runtimes (Node.js, Go, Python) available
// on the host system. Each detect function runs the runtime's version command
// and parses stdout into a structured RuntimeInfo value.
package runtime

import (
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// RuntimeInfo describes a language runtime found (or not) on the system.
type RuntimeInfo struct {
	Name      string
	Installed bool
	Version   string // Semver without prefix, e.g. "22.3.0". Empty if not installed or parse failed.
	RawOutput string // Raw stdout from the version command, for diagnostics.
	Err       error  // exec error. nil when not installed — caller checks Installed instead.
}

// detectRuntime is a shared helper that locates a binary, runs it with args,
// and parses its stdout through parseFn. Returns a populated RuntimeInfo.
func detectRuntime(name, cmdName string, args []string, parseFn func(string) (string, error)) RuntimeInfo {
	if _, err := exec.LookPath(cmdName); err != nil {
		return RuntimeInfo{Name: name, Installed: false}
	}

	out, err := exec.Command(cmdName, args...).Output()
	if err != nil {
		return RuntimeInfo{
			Name:      name,
			Installed: true,
			RawOutput: string(out),
			Err:       fmt.Errorf("executing %s: %w", cmdName, err),
		}
	}

	raw := string(out)
	version, parseErr := parseFn(raw)
	if parseErr != nil {
		return RuntimeInfo{
			Name:      name,
			Installed: true,
			Version:   "",
			RawOutput: raw,
			Err:       fmt.Errorf("parsing %s version: %w", cmdName, parseErr),
		}
	}

	return RuntimeInfo{
		Name:      name,
		Installed: true,
		Version:   version,
		RawOutput: raw,
	}
}

// DetectNode checks for Node.js by running "node --version".
func DetectNode() RuntimeInfo {
	return detectRuntime("node", "node", []string{"--version"}, parseNodeVersion)
}

// DetectGo checks for Go by running "go version".
func DetectGo() RuntimeInfo {
	return detectRuntime("go", "go", []string{"version"}, parseGoVersion)
}

// DetectPython checks for Python 3 by running "python3 --version".
func DetectPython() RuntimeInfo {
	return detectRuntime("python", "python3", []string{"--version"}, parsePythonVersion)
}

// parseNodeVersion strips the leading "v" from "node --version" output.
// Expects raw stdout such as "v22.3.0\n".
func parseNodeVersion(raw string) (string, error) {
	v := strings.TrimSpace(raw)
	if !strings.HasPrefix(v, "v") {
		return "", errors.New("unexpected node version format: missing 'v' prefix")
	}
	return strings.TrimPrefix(v, "v"), nil
}

// parseGoVersion extracts the Go version from "go version" output.
// Expects raw stdout such as "go version go1.24.0 linux/amd64\n".
// Finds the token starting with "go" (e.g. "go1.24.0") and strips the "go" prefix.
func parseGoVersion(raw string) (string, error) {
	fields := strings.Fields(raw)
	for _, f := range fields {
		if strings.HasPrefix(f, "go") && len(f) > 2 {
			return f[2:], nil
		}
	}
	return "", errors.New("go version not found in output")
}

// parsePythonVersion strips the "Python " prefix from "python3 --version" output.
// Expects raw stdout such as "Python 3.12.1\n".
func parsePythonVersion(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if !strings.HasPrefix(trimmed, "Python ") {
		return "", errors.New("unexpected python version format: missing 'Python ' prefix")
	}
	return strings.TrimPrefix(trimmed, "Python "), nil
}

// compareVersions compares two semver strings numerically.
// Returns -1 if a < b, 0 if equal, 1 if a > b.
// Returns 0 if either version fails to parse as a three-part semver.
func compareVersions(a, b string) int {
	aParts := parseVersionParts(a)
	bParts := parseVersionParts(b)
	if aParts == nil || bParts == nil {
		return 0
	}
	for i := 0; i < 3; i++ {
		if aParts[i] < bParts[i] {
			return -1
		}
		if aParts[i] > bParts[i] {
			return 1
		}
	}
	return 0
}

// parseVersionParts splits a semver string into [major, minor, patch] integers.
// Returns nil if the string is not a valid three-part semver.
func parseVersionParts(v string) []int {
	parts := strings.Split(v, ".")
	if len(parts) != 3 {
		return nil
	}
	nums := make([]int, 3)
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil {
			return nil
		}
		nums[i] = n
	}
	return nums
}

// IsCompatible returns true if the runtime is installed and its version is
// greater than or equal to minVersion. Both versions are parsed as
// [major, minor, patch] tuples. Returns false if either version is malformed
// or the runtime is not installed.
func IsCompatible(info RuntimeInfo, minVersion string) bool {
	if !info.Installed {
		return false
	}
	return parseVersionParts(info.Version) != nil &&
		parseVersionParts(minVersion) != nil &&
		compareVersions(info.Version, minVersion) >= 0
}
