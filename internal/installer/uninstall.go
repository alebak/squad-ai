package installer

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/alebak/squad-ai/internal/registry"
)

var npmPkgRe = regexp.MustCompile(`^[@]?[a-zA-Z0-9][a-zA-Z0-9._-]*(/[a-zA-Z0-9][a-zA-Z0-9._-]*)?$`)
var safeCmdRe = regexp.MustCompile(`^[a-zA-Z0-9 /@_.:-]+$`)

// UninstallAgent removes an agent from the system.
//
// Resolution order:
//  1. If agent.Install.UninstallCmd is non-empty, execute it via sh -c.
//  2. If method is npm_install, derive "npm uninstall -g <package>" from the
//     install command.
//  3. If method is curl_bash, resolve the binary via exec.LookPath and delete
//     it with os.Remove (no shell involved — prevents injection from detect_cmd).
//  4. Otherwise, return an error indicating no uninstall method is defined.
//
// On success, nil is returned. On failure, the error is wrapped with context.
func UninstallAgent(agent registry.Agent) error {
	cmd := agent.Install.UninstallCmd

	if cmd == "" {
		// No explicit command — derive from the install method.
		switch agent.Install.Method {
		case registry.MethodNpmInstall:
			pkg := ExtractNPMPackage(agent.Install.Command)
			if pkg == "" {
				return fmt.Errorf("uninstalling %s: could not extract npm package name from install command %q", agent.ID, agent.Install.Command)
			}
			if !npmPkgRe.MatchString(pkg) {
				return fmt.Errorf("uninstalling %s: extracted package name %q contains invalid characters", agent.ID, pkg)
			}
			cmd = fmt.Sprintf("npm uninstall -g %s", pkg)

		case registry.MethodCurlBash:
			return uninstallCurlBashFallback(agent)

		case registry.MethodCustom:
			return fmt.Errorf("uninstalling %s: no uninstall command defined for agent with custom install method", agent.ID)

		default:
			return fmt.Errorf("uninstalling %s: unknown install method %q", agent.ID, agent.Install.Method)
		}
	}

	// Validate the (explicit or derived) command before execution.
	if strings.ContainsRune(cmd, '\x00') {
		return fmt.Errorf("uninstalling %s: uninstall command contains null byte", agent.ID)
	}
	if !safeCmdRe.MatchString(cmd) {
		return fmt.Errorf("uninstalling %s: uninstall command contains invalid characters", agent.ID)
	}

	return runAndLog(agent.ID, cmd, nil)
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
			// Race condition: someone else removed it between LookPath and Remove.
			return nil
		}
		return fmt.Errorf("uninstalling %s: removing binary at %s: %w", agent.ID, path, err)
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
