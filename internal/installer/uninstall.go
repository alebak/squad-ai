package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alebak/squad-ai/internal/registry"
)

var npmPkgRe = regexp.MustCompile(`^[@]?[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)?$`)
var safeCmdRe = regexp.MustCompile(`^[a-zA-Z0-9 /@_.:-]+$`)

// UninstallAgent removes an agent from the system.
//
// Resolution order:
//  1. If agent.Install.UninstallCmd is non-empty, execute it as argv (no shell).
//  2. If method is npm_install, derive "npm uninstall -g <package>" from the
//     install command and execute it as argv.
//  3. If method is curl_bash, resolve the binary via exec.LookPath and delete
//     it with os.Remove (no shell involved — prevents injection from detect_cmd).
//  4. Otherwise, return an error indicating no uninstall method is defined.
//
// On success, nil is returned. On failure, the error is wrapped with context.
func UninstallAgent(agent registry.Agent) error {
	cmd := agent.Install.UninstallCmd

	if cmd == "" {
		derived, err := deriveUninstallCommand(agent)
		if err != nil {
			return err
		}
		if derived == "" {
			// curl_bash fallback already performed.
			return nil
		}
		cmd = derived
	}

	if strings.ContainsRune(cmd, '\x00') {
		return fmt.Errorf("uninstalling %s: uninstall command contains null byte", agent.ID)
	}
	if !safeCmdRe.MatchString(cmd) {
		return fmt.Errorf("uninstalling %s: uninstall command contains invalid characters", agent.ID)
	}

	fields := strings.Fields(cmd)
	if len(fields) == 0 {
		return fmt.Errorf("uninstalling %s: empty uninstall command", agent.ID)
	}
	switch fields[0] {
	case "npm", "pnpm", "yarn":
		// allowed package managers
	default:
		return fmt.Errorf("uninstalling %s: uninstall binary %q is not in the allowlist", agent.ID, fields[0])
	}
	return runAndLog(agent.ID, fields[0], fields[1:], nil /* progress */)
}

// deriveUninstallCommand returns a derived uninstall command string, or empty
// string when curl_bash fallback already completed successfully.
func deriveUninstallCommand(agent registry.Agent) (string, error) {
	switch agent.Install.Method {
	case registry.MethodNpmInstall:
		pkg := ExtractNPMPackage(agent.Install.Command)
		if pkg == "" {
			return "", fmt.Errorf("uninstalling %s: could not extract npm package name from install command %q", agent.ID, agent.Install.Command)
		}
		if !npmPkgRe.MatchString(pkg) {
			return "", fmt.Errorf("uninstalling %s: extracted package name %q contains invalid characters", agent.ID, pkg)
		}
		return fmt.Sprintf("npm uninstall -g %s", pkg), nil
	case registry.MethodCurlBash:
		return "", uninstallCurlBashFallback(agent)
	case registry.MethodCustom:
		return "", fmt.Errorf("uninstalling %s: no uninstall command defined for agent with custom install method", agent.ID)
	default:
		return "", fmt.Errorf("uninstalling %s: unknown install method %q", agent.ID, agent.Install.Method)
	}
}

// uninstallCurlBashFallback resolves the binary path via exec.LookPath and
// deletes it with os.Remove. This avoids shell injection from detect_cmd.
func uninstallCurlBashFallback(agent registry.Agent) error {
	if agent.DetectCmd == "" {
		return fmt.Errorf("uninstalling %s: detect_command is required for curl_bash fallback uninstall", agent.ID)
	}
	if strings.Contains(agent.DetectCmd, "/") || strings.Contains(agent.DetectCmd, "\\") {
		return fmt.Errorf("uninstalling %s: detect_command must be a bare executable name, not a path", agent.ID)
	}

	path, err := exec.LookPath(agent.DetectCmd)
	if err != nil {
		return fmt.Errorf("uninstalling %s: binary %q not found in PATH: %w", agent.ID, agent.DetectCmd, err)
	}

	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("uninstalling %s: removing binary at %s: %w", agent.ID, path, err)
	}
	return nil
}

// UninstallConfig removes config/data directories for the given agent.
//
// For each path in agent.ConfigPaths:
//  1. Expand "~" to the user's home directory.
//  2. Resolve the path to an absolute path.
//  3. Verify the resolved path is within the user's home directory.
//  4. Call os.RemoveAll on the resolved path.
//
// Non-existent paths are skipped silently. Returns nil on success or if
// ConfigPaths is nil/empty. Returns an error if any path escapes the home
// directory or if os.RemoveAll fails.
func UninstallConfig(agent registry.Agent) error {
	if len(agent.ConfigPaths) == 0 {
		return nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("uninstalling config for %s: cannot determine home dir: %w", agent.ID, err)
	}

	var firstErr error
	for _, p := range agent.ConfigPaths {
		if err := removeConfigPath(agent.ID, home, p); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func removeConfigPath(agentID, home, raw string) error {
	resolved := raw
	if strings.HasPrefix(resolved, "~") {
		resolved = filepath.Join(home, resolved[1:])
	}

	absPath, err := filepath.Abs(resolved)
	if err != nil {
		return fmt.Errorf("uninstalling config for %s: resolving path %q: %w", agentID, raw, err)
	}

	// All registry config paths must stay inside the user's home directory.
	rel, err := filepath.Rel(home, absPath)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("uninstalling config for %s: path %q resolves outside home directory", agentID, raw)
	}

	if _, statErr := os.Stat(absPath); os.IsNotExist(statErr) {
		return nil
	}
	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("uninstalling config for %s: removing %s: %w", agentID, absPath, err)
	}
	return nil
}

// ExtractNPMPackage extracts the npm package name from an install command.
//
// Examples:
//
//	"npm i -g @openai/codex"       → "@openai/codex"
//	"npm install -g @google/gemini-cli" → "@google/gemini-cli"
//
// It finds the last positional argument that is not a flag and comes after the
// npm subcommand (i, install, add). Returns "" if no package name is found.
func ExtractNPMPackage(installCmd string) string {
	parts := strings.Fields(installCmd)
	// Need at least: npm <subcommand> [flags...] <package>
	if len(parts) < 3 {
		return ""
	}
	// Walk backwards from the end, starting after "npm <subcmd>",
	// to find the package name (the first non-flag argument).
	for i := len(parts) - 1; i >= 2; i-- {
		if !strings.HasPrefix(parts[i], "-") {
			return parts[i]
		}
	}
	return ""
}
