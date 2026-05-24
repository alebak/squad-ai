package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListCommand_NormalOutput(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &listHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"opencode"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{
				"claude-code": true,
				"opencode":    false,
				"codex":       false,
			}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool {
			for _, d := range deps {
				if d.Runtime != "none" {
					return false
				}
			}
			return true
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newListCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()

	// Check header
	assert.Contains(t, output, "Agent ID")
	assert.Contains(t, output, "Name")
	assert.Contains(t, output, "Installed")
	assert.Contains(t, output, "Status")

	// claude-code is installed → ✅ + installed
	assert.Contains(t, output, "claude-code")
	assert.Contains(t, output, "Claude Code")
	assert.Contains(t, output, "✅")
	assert.Contains(t, output, "installed")

	// opencode is selected but not installed → ❌ + selected
	assert.Contains(t, output, "opencode")
	assert.Contains(t, output, "OpenCode")
	assert.Contains(t, output, "❌")
	assert.Contains(t, output, "selected")

	// codex is blocked (runtime not met) → ❌ + blocked
	assert.Contains(t, output, "codex")
	assert.Contains(t, output, "Codex CLI")
	assert.Contains(t, output, "blocked")
}

func TestListCommand_EmptyRegistry(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &listHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return &registry.Catalog{Agents: []registry.Agent{}}, nil
		},
		detectAll:   func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:  func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newListCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No agents found")
}

func TestListCommand_RegistryFetchFailure(t *testing.T) {
	h := &listHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return nil, errors.New("connection refused")
		},
		detectAll:   func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:  func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newListCommandWithHandler(h)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestListCommand_AllAvailable(t *testing.T) {
	// When no agents are installed, none selected, all runtimes met → all "available".
	buf := new(bytes.Buffer)

	h := &listHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newListCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// All three should be "available" (not installed, not selected, not blocked)
	assert.Contains(t, output, "available")
	// Should contain "Codex CLI" even though node is not installed,
	// because our mock isRuntimeMet returns true
	assert.Contains(t, output, "Codex CLI")
}
