// Package installer handles agent installation and detection operations.
package installer

import (
	"os/exec"

	"github.com/alebak/squad-ai/internal/registry"
)

// IsAgentInstalled checks whether a command binary exists in PATH.
// Returns true if the binary is found and executable, false otherwise.
func IsAgentInstalled(detectCmd string) bool {
	if detectCmd == "" {
		return false
	}
	_, err := exec.LookPath(detectCmd)
	return err == nil
}

// DetectAll checks all agents from a slice and returns a map of agentID
// to installation status. Agents with empty detect_command are reported
// as not installed.
func DetectAll(agents []registry.Agent) map[string]bool {
	result := make(map[string]bool, len(agents))
	for _, a := range agents {
		result[a.ID] = IsAgentInstalled(a.DetectCmd)
	}
	return result
}
