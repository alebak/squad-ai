package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/installer"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/alebak/squad-ai/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddCommand_NoTTYShowsAgentList(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &addHandler{
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
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			return nil, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool {
			// All runtimes met
			return true
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No interactive terminal detected")
	assert.Contains(t, output, "claude-code")
	assert.Contains(t, output, "opencode")
}

func TestAddCommand_EmptyRegistry(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return &registry.Catalog{Agents: []registry.Agent{}}, nil
		},
		detectAll:    func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:   func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		runSelection: func(items []tui.AgentItem) ([]string, error) { return nil, nil },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:   func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No agents found")
}

func TestAddCommand_RegistryFetchFailure(t *testing.T) {
	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return nil, errors.New("connection refused")
		},
		detectAll:    func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:   func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		runSelection: func(items []tui.AgentItem) ([]string, error) { return nil, nil },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:   func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestAddCommand_AllAgentsAlreadyHandled(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"claude-code", "opencode"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		// All agents are installed — nothing available to add
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{
				"claude-code": true,
				"opencode":    true,
				"codex":       true,
			}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			return nil, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:   func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	// All agents are still listed even when installed
	assert.Contains(t, buf.String(), "claude-code")
	assert.Contains(t, buf.String(), "opencode")
	assert.Contains(t, buf.String(), "codex")
}

func TestAddCommand_TUISuccessFlow(t *testing.T) {
	buf := new(bytes.Buffer)

	var installedIDs []string
	h := &addHandler{
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
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				installedIDs = append(installedIDs, a.ID)
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil, nil}
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			// Simulate selecting both compatible agents
			return []string{"claude-code", "opencode"}, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:   func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Installing selected agents")
	assert.Contains(t, output, "Claude Code installed")
	assert.Contains(t, output, "OpenCode installed")
	assert.ElementsMatch(t, []string{"claude-code", "opencode"}, installedIDs)
}

func TestAddCommand_TUIEmptySelection(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &addHandler{
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
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			return nil, nil // user cancelled
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:   func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No agents selected")
}

func TestAddCommand_BlockedAgentsShownInTTYFallback(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &addHandler{
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
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			return nil, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool {
			// Block agents with non-"none" runtime
			for _, d := range deps {
				if d.Runtime != "none" {
					return false
				}
			}
			return true
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Blocked agents should still appear in the fallback list
	assert.Contains(t, output, "codex")
}
