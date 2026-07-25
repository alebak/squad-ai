// Package installer handles agent installation and detection operations.
package installer

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/alebak/squad-ai/internal/registry"
)

// userBinDirs returns common directories where package managers and installers
// place user-local binaries. These are often missing from non-interactive PATH.
func userBinDirs() []string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return nil
	}
	return []string{
		filepath.Join(home, ".local", "bin"),
		filepath.Join(home, "bin"),
		filepath.Join(home, ".npm-global", "bin"),
	}
}

// agentHomeBinDir returns $HOME/.<agentID>/bin when agentID is a safe id.
// OpenCode and similar installers drop binaries there and only append the dir
// to interactive shell rc files (e.g. ~/.bashrc), which non-interactive
// squad runs never source.
func agentHomeBinDir(agentID string) string {
	if agentID == "" || strings.Contains(agentID, "..") || strings.ContainsRune(agentID, os.PathSeparator) {
		return ""
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, "."+agentID, "bin")
}

// pathWithUserBins returns PATH with common user binary directories prepended
// when they are not already present. Used for child process environments only.
func pathWithUserBins(extraDirs ...string) string {
	path := os.Getenv("PATH")
	parts := filepath.SplitList(path)
	present := make(map[string]bool, len(parts))
	for _, p := range parts {
		present[p] = true
	}

	var prefix []string
	add := func(dir string) {
		if dir == "" || present[dir] {
			return
		}
		prefix = append(prefix, dir)
		present[dir] = true
	}
	for _, dir := range userBinDirs() {
		add(dir)
	}
	for _, dir := range extraDirs {
		add(dir)
	}
	if len(prefix) == 0 {
		return path
	}
	if path == "" {
		return strings.Join(prefix, string(os.PathListSeparator))
	}
	return strings.Join(prefix, string(os.PathListSeparator)) + string(os.PathListSeparator) + path
}

// isExecutable reports whether info refers to a regular executable file.
func isExecutable(info fs.FileInfo) bool {
	mode := info.Mode()
	if !mode.IsRegular() {
		return false
	}
	return mode&0o111 != 0
}

// lookInPath searches for detectCmd across PATH plus common user bin dirs
// without mutating process-global environment state.
func lookInPath(detectCmd string, extraDirs ...string) bool {
	if detectCmd == "" {
		return false
	}
	// Absolute/relative paths: check directly.
	if strings.Contains(detectCmd, string(os.PathSeparator)) {
		info, err := os.Stat(detectCmd)
		return err == nil && isExecutable(info)
	}

	for _, dir := range filepath.SplitList(pathWithUserBins(extraDirs...)) {
		if dir == "" {
			continue
		}
		candidate := filepath.Join(dir, detectCmd)
		info, err := os.Stat(candidate)
		if err == nil && isExecutable(info) {
			return true
		}
	}
	return false
}

// IsAgentInstalled checks whether a command binary exists in PATH.
// Common user install directories are included even when the current shell
// PATH is minimal (non-interactive containers, postCreate hooks).
// Returns true if the binary is found and executable, false otherwise.
//
// Prefer AgentIsInstalled when the agent ID is known so $HOME/.<id>/bin is
// searched (needed for OpenCode and similar installers).
func IsAgentInstalled(detectCmd string) bool {
	return lookInPath(detectCmd)
}

// AgentIsInstalled reports whether agent's detect command is available,
// including agent-specific home bin dirs that interactive installers add
// only to shell rc files.
func AgentIsInstalled(agent registry.Agent) bool {
	extra := agentHomeBinDir(agent.ID)
	if extra == "" {
		return lookInPath(agent.DetectCmd)
	}
	return lookInPath(agent.DetectCmd, extra)
}

// DetectAll checks all agents from a slice and returns a map of agentID
// to installation status. Agents with empty detect_command are reported
// as not installed.
func DetectAll(agents []registry.Agent) map[string]bool {
	result := make(map[string]bool, len(agents))
	for _, a := range agents {
		result[a.ID] = AgentIsInstalled(a)
	}
	return result
}
