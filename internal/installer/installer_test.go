package installer

import (
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
)

func TestDetectAgent_RealBinary(t *testing.T) {
	got := IsAgentInstalled("go")
	assert.True(t, got, "go should be installed in the dev container")
}

func TestDetectAgent_NonExistent(t *testing.T) {
	got := IsAgentInstalled("nonexistent-command-xyz")
	assert.False(t, got)
}

func TestDetectAgent_EmptyCommand(t *testing.T) {
	got := IsAgentInstalled("")
	assert.False(t, got)
}

func TestDetectAll_Mixed(t *testing.T) {
	agents := []registry.Agent{
		{ID: "go-agent", DetectCmd: "go"},
		{ID: "dummy-agent", DetectCmd: "nonexistent-command-xyz"},
		{ID: "py-agent", DetectCmd: "python3"},
	}
	result := DetectAll(agents)
	assert.True(t, result["go-agent"], "go should be found in PATH")
	assert.False(t, result["dummy-agent"])
	assert.True(t, result["py-agent"], "python3 should be found in PATH")
	assert.Len(t, result, 3)
}

func TestDetectAll_Empty(t *testing.T) {
	result := DetectAll([]registry.Agent{})
	assert.Empty(t, result)
}

func TestDetectAll_IncludesEmptyCmd(t *testing.T) {
	agents := []registry.Agent{
		{ID: "real", DetectCmd: "go"},
		{ID: "empty-cmd", DetectCmd: ""},
	}
	result := DetectAll(agents)
	assert.True(t, result["real"])
	assert.False(t, result["empty-cmd"])
	assert.Len(t, result, 2)
}
