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
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			return nil, nil, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool {
			return true
		},
		uninstallAgent: func(agent registry.Agent) error {
			return nil
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
		detectAll:      func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:     func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		runSelection:   func(items []tui.AgentItem) ([]string, map[string]int, error) { return nil, nil, nil },
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return false },
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
		detectAll:      func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:     func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		runSelection:   func(items []tui.AgentItem) ([]string, map[string]int, error) { return nil, nil, nil },
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return false },
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
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			return nil, nil, nil
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
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
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			return []string{"claude-code", "opencode"}, nil, nil
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return true },
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
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			return []string{}, nil, nil // confirmed empty
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "No agents selected")
}

func TestAddCommand_UninstallViaWizardAppOnly(t *testing.T) {
	buf := new(bytes.Buffer)
	var uninstalledAgent string
	var callCount int

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			callCount++
			if callCount == 1 {
				// First TUI: wizard completed with choice 0 for claude-code
				return []string{"opencode"}, map[string]int{"claude-code": 0}, nil
			}
			// Second TUI after restart: claude-code is uninstalled, user selects opencode
			return []string{"opencode"}, nil, nil
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			uninstalledAgent = agent.ID
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			t.Error("uninstallConfig should NOT be called for app-only")
			return nil
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "claude-code", uninstalledAgent)
	assert.Contains(t, buf.String(), "Uninstalled Claude Code")
}

func TestAddCommand_UninstallViaWizardAppAndConfig(t *testing.T) {
	buf := new(bytes.Buffer)
	var uninstalledAgent string
	var configCleanedAgent string
	var callCount int

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			callCount++
			if callCount == 1 {
				return []string{"opencode"}, map[string]int{"claude-code": 1}, nil
			}
			return []string{"opencode"}, nil, nil
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			uninstalledAgent = agent.ID
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			configCleanedAgent = agent.ID
			return nil
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "claude-code", uninstalledAgent)
	assert.Equal(t, "claude-code", configCleanedAgent)
	assert.Contains(t, buf.String(), "Uninstalled Claude Code (app)")
	assert.Contains(t, buf.String(), "Cleaned config for Claude Code")
}

func TestAddCommand_UninstallViaWizardSkip(t *testing.T) {
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
			return map[string]bool{"claude-code": true, "opencode": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			// User chose "Keep installed (skip)" for both agents
			return []string{}, map[string]int{
				"claude-code": 2,
				"opencode":    2,
			}, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			t.Error("uninstallAgent should NOT be called when user selects skip")
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			t.Error("uninstallConfig should NOT be called when user selects skip")
			return nil
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	// All skips with empty selection — "No agents selected" since selectedIDs is empty
	assert.Contains(t, buf.String(), "No agents selected")
}

func TestAddCommand_WizardRestartsTUIAfterUninstall(t *testing.T) {
	buf := new(bytes.Buffer)
	var runSelectionCallCount int
	var uninstalledAgent string

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			runSelectionCallCount++
			if runSelectionCallCount == 1 {
				// First TUI: wizard completed with choice 0 for claude-code
				return []string{"opencode"}, map[string]int{"claude-code": 0}, nil
			}
			// Second TUI after restart: claude-code was uninstalled, user selects both
			return []string{"opencode", "claude-code"}, nil, nil
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			uninstalledAgent = agent.ID
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			t.Error("uninstallConfig should NOT be called for app-only")
			return nil
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "claude-code", uninstalledAgent)
	// TUI should have been re-launched after wizard uninstall
	assert.Equal(t, 2, runSelectionCallCount, "TUI should re-launch after wizard uninstall")
	assert.Contains(t, buf.String(), "Uninstalled Claude Code")
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
		runSelection: func(items []tui.AgentItem) ([]string, map[string]int, error) {
			return nil, nil, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool {
			for _, d := range deps {
				if d.Runtime != "none" {
					return false
				}
			}
			return true
		},
		uninstallAgent: func(agent registry.Agent) error { return nil },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return false },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "codex")
}
