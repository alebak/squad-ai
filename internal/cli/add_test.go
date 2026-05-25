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
		uninstallAgent: func(agent registry.Agent) error {
			return nil
		},
		confirmFn: func(msg string) bool {
			return false
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
		runSelection:   func(items []tui.AgentItem) ([]string, error) { return nil, nil },
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		confirmFn:      func(msg string) bool { return false },
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
		runSelection:   func(items []tui.AgentItem) ([]string, error) { return nil, nil },
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		confirmFn:      func(msg string) bool { return false },
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
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		confirmFn:      func(msg string) bool { return false },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return false },
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
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		confirmFn:      func(msg string) bool { return false },
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
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			return []string{}, nil // user confirmed empty selection (Enter with nothing checked)
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error { return nil },
		confirmFn:      func(msg string) bool { return false },
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

func TestAddCommand_UninstallPromptCancel(t *testing.T) {
	buf := new(bytes.Buffer)

	var runSelectionCallCount int
	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		// Claude-code is installed but user deselected it on first TUI run
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			// Return nil errors matching the number of agents being installed.
			errs := make([]error, len(agents))
			return errs
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			runSelectionCallCount++
			if runSelectionCallCount == 1 {
				// First TUI launch: user deselected claude-code (installed)
				return []string{"opencode"}, nil
			}
			// Second TUI launch after cancel: claude-code is pre-checked again
			return []string{"opencode", "claude-code"}, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			t.Error("uninstallAgent should NOT be called when user cancels")
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			t.Error("uninstallConfig should NOT be called when user cancels")
			return nil
		},
		uninstallChoiceFn: func(agentName string) uninstallChoice {
			return uninstallCancel
		},
		confirmFn:  func(msg string) bool { return false },
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Cancel should re-launch TUI rather than silently keeping the agent
	assert.Equal(t, 2, runSelectionCallCount, "TUI should have been re-launched after cancel")
	// No uninstall should have happened
	assert.NotContains(t, output, "Uninstalled")
	// After second TUI run, the flow proceeds to installation
	assert.Contains(t, output, "Installing selected agents")
}

func TestAddCommand_UninstallAppOnly(t *testing.T) {
	buf := new(bytes.Buffer)
	var uninstalledAgent string
	var runSelectionCallCount int

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
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			runSelectionCallCount++
			if runSelectionCallCount == 1 {
				// First TUI: user deselects claude-code (installed)
				return []string{"opencode"}, nil
			}
			// Second TUI after restart: claude-code was uninstalled, user selects both
			return []string{"opencode", "claude-code"}, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			uninstalledAgent = agent.ID
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			t.Error("uninstallConfig should NOT be called for app-only")
			return nil
		},
		uninstallChoiceFn: func(agentName string) uninstallChoice {
			return uninstallAppOnly
		},
		confirmFn:  func(msg string) bool { return false },
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Equal(t, "claude-code", uninstalledAgent)
	assert.Equal(t, 2, runSelectionCallCount, "TUI should re-launch after uninstall")
	assert.Contains(t, buf.String(), "Uninstalled Claude Code")
}

func TestAddCommand_UninstallAppAndConfig(t *testing.T) {
	buf := new(bytes.Buffer)
	var uninstalledAgent string
	var configCleanedAgent string
	var runSelectionCallCount int

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
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			runSelectionCallCount++
			if runSelectionCallCount == 1 {
				// First TUI: user deselects claude-code (installed)
				return []string{"opencode"}, nil
			}
			// Second TUI after restart: claude-code was uninstalled, user selects both
			return []string{"opencode", "claude-code"}, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			uninstalledAgent = agent.ID
			return nil
		},
		uninstallConfig: func(agent registry.Agent) error {
			configCleanedAgent = agent.ID
			return nil
		},
		uninstallChoiceFn: func(agentName string) uninstallChoice {
			return uninstallAppConfig
		},
		confirmFn:  func(msg string) bool { return false },
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
	assert.Equal(t, 2, runSelectionCallCount, "TUI should re-launch after uninstall")
	assert.Contains(t, buf.String(), "Uninstalled Claude Code (app)")
	assert.Contains(t, buf.String(), "Cleaned config for Claude Code")
}

func TestAddCommand_BulkUninstallConfirm(t *testing.T) {
	buf := new(bytes.Buffer)
	var uninstalledAgents []string
	var runSelectionCallCount int

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		// Both claude-code and opencode are installed
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": true, "opencode": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			runSelectionCallCount++
			if runSelectionCallCount == 1 {
				// First TUI: user deselected both installed agents
				return []string{}, nil
			}
			// Second TUI after restart: both were uninstalled, user selects them
			return []string{"claude-code", "opencode"}, nil
		},
		isRuntimeMet:   func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			uninstalledAgents = append(uninstalledAgents, agent.ID)
			return nil
		},
		confirmFn:  func(msg string) bool { return true },
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	// Both agents should be uninstalled
	assert.ElementsMatch(t, []string{"claude-code", "opencode"}, uninstalledAgents)
	assert.Equal(t, 2, runSelectionCallCount, "TUI should re-launch after bulk uninstall")
	output := buf.String()
	assert.Contains(t, output, "Uninstalled Claude Code")
	assert.Contains(t, output, "Uninstalled OpenCode")
}

func TestAddCommand_BulkUninstallDecline(t *testing.T) {
	buf := new(bytes.Buffer)
	var runSelectionCallCount int

	h := &addHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return config.DefaultConfig(), nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		// Both claude-code and opencode are installed
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": true, "opencode": true}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			return nil
		},
		runSelection: func(items []tui.AgentItem) ([]string, error) {
			runSelectionCallCount++
			if runSelectionCallCount == 1 {
				// First TUI: user deselected both installed agents
				return []string{}, nil
			}
			// Second TUI after restart: user selected both (they were re-added)
			return []string{"claude-code", "opencode"}, nil
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		uninstallAgent: func(agent registry.Agent) error {
			t.Error("uninstallAgent should NOT be called when user declines bulk uninstall")
			return nil
		},
		confirmFn:  func(msg string) bool { return false },
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal: func() bool { return true },
	}

	cmd := newAddCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	// TUI should have been re-launched after decline
	assert.Equal(t, 2, runSelectionCallCount, "TUI should have been re-launched after bulk decline")
	// No uninstall should have happened
	assert.NotContains(t, buf.String(), "Uninstalled")
	// After second TUI run, the flow should note the agents are already installed
	assert.Contains(t, buf.String(), "Selected agents are already installed")
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
		uninstallAgent: func(agent registry.Agent) error { return nil },
		confirmFn:      func(msg string) bool { return false },
		configPath:     func() (string, error) { return "/tmp/test-config.json", nil },
		isTerminal:     func() bool { return false },
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
