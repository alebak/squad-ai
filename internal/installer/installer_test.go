package installer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestDetectAgent_UserBinDirOutsidePATH(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	// Minimal PATH without the user bin dir where we place the fake binary.
	t.Setenv("PATH", "/usr/bin:/bin")

	binDir := filepath.Join(home, ".local", "bin")
	require.NoError(t, os.MkdirAll(binDir, 0o755))
	binPath := filepath.Join(binDir, "fake-agent-bin")
	require.NoError(t, os.WriteFile(binPath, []byte("#!/bin/sh\necho ok\n"), 0o755))

	assert.True(t, IsAgentInstalled("fake-agent-bin"))
}

func TestPathWithUserBins_PrependsMissingDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("PATH", "/usr/bin")

	got := pathWithUserBins()
	assert.Contains(t, got, filepath.Join(home, ".local", "bin"))
	assert.Contains(t, got, filepath.Join(home, "bin"))
	assert.Contains(t, got, "/usr/bin")
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
