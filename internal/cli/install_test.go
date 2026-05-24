package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/installer"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testRegistry returns a minimal registry catalog for use in tests.
func testRegistry() *registry.Catalog {
	return &registry.Catalog{
		Agents: []registry.Agent{
			{
				ID:      "claude-code",
				Name:    "Claude Code",
				Version: "latest",
				Install: registry.InstallCmd{
					Method:  registry.MethodCurlBash,
					Command: "/bin/true",
				},
				Dependencies: []registry.RuntimeDep{{Runtime: "none"}},
			},
			{
				ID:      "opencode",
				Name:    "OpenCode",
				Version: "latest",
				Install: registry.InstallCmd{
					Method:  registry.MethodCurlBash,
					Command: "/bin/true",
				},
				Dependencies: []registry.RuntimeDep{{Runtime: "none"}},
			},
			{
				ID:   "codex",
				Name: "Codex CLI",
				Install: registry.InstallCmd{
					Method:  registry.MethodNpmInstall,
					Command: "npm install -g @openai/codex",
				},
				Dependencies: []registry.RuntimeDep{{Runtime: "node", MinVersion: "22.0.0"}},
			},
		},
	}
}

func TestInstallCommand_DefaultWithConfig(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
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
			return map[string]bool{"claude-code": false, "opencode": false}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			// Simulate successful installation for both agents
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil, nil}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Claude Code installed")
	assert.Contains(t, output, "OpenCode installed")
}

func TestInstallCommand_DefaultWithConfig_AlreadyInstalled(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
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
			return map[string]bool{"claude-code": true, "opencode": false}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Claude Code already installed")
	assert.Contains(t, output, "OpenCode installed")
}

func TestInstallCommand_AgentsFlag(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				InstallOptions: config.InstallOptions{Silent: true, PreferPnpm: true},
				RegistryKnown:  []string{"claude-code", "opencode", "codex"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--agents", "claude-code"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Claude Code installed")
	assert.NotContains(t, output, "OpenCode")
}

func TestInstallCommand_AllFlag(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				InstallOptions: config.InstallOptions{Silent: true, PreferPnpm: true},
				RegistryKnown:  []string{"claude-code", "opencode", "codex"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil, nil, nil}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool {
			// Only block codex (node dependency)
			for _, d := range deps {
				if d.Runtime != "none" {
					return false
				}
			}
			return true
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"--all"})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Claude Code installed")
	assert.Contains(t, output, "OpenCode installed")
	assert.Contains(t, output, "blocked")
	assert.Contains(t, output, "Codex CLI")
}

func TestInstallCommand_MixedFlagsError(t *testing.T) {
	h := &installHandler{
		registryURL: "",
		loadConfig:    func(path string) (*config.Config, error) { return config.DefaultConfig(), nil },
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) { return testRegistry(), nil },
		detectAll:     func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:    func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		isRuntimeMet:  func(deps []registry.RuntimeDep) bool { return true },
		configPath:    func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetArgs([]string{"--agents", "claude-code", "--all"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mutually exclusive")
}

func TestInstallCommand_NoConfig(t *testing.T) {
	// When config exists but has no selected_agents, install is a no-op.
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				InstallOptions: config.InstallOptions{Silent: true, PreferPnpm: true},
				RegistryKnown:  []string{"claude-code", "opencode", "codex"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll:    func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:   func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No agents to install")
}

func TestInstallCommand_RegistryFetchFailure(t *testing.T) {
	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{SelectedAgents: []string{"claude-code"}}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return nil, errors.New("connection refused")
		},
		detectAll:    func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:   func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestInstallCommand_PartialFailure(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{SelectedAgents: []string{"claude-code", "opencode"}}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil, errors.New("exit code 1")}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err) // overall failure due to partial errors
	assert.Contains(t, err.Error(), "one or more installations failed")

	output := buf.String()
	assert.Contains(t, output, "Claude Code installed")
	assert.Contains(t, output, "OpenCode")
}

func TestInstallCommand_EmptyRegistry(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{SelectedAgents: []string{"claude-code"}}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return &registry.Catalog{Agents: []registry.Agent{}}, nil
		},
		detectAll:    func(agents []registry.Agent) map[string]bool { return map[string]bool{} },
		installAll:   func(agents []registry.Agent, progress installer.ProgressFn) []error { return nil },
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "No agents found")
}

func TestFindNewAgents(t *testing.T) {
	catalog := testRegistry()

	tests := []struct {
		name       string
		knownIDs   []string
		wantIDs    []string
		wantLen    int
	}{
		{
			name:     "all new when known is empty",
			knownIDs: []string{},
			wantIDs:  []string{"claude-code", "opencode", "codex"},
			wantLen:  3,
		},
		{
			name:     "some new",
			knownIDs: []string{"claude-code"},
			wantIDs:  []string{"opencode", "codex"},
			wantLen:  2,
		},
		{
			name:     "none new when all known",
			knownIDs: []string{"claude-code", "opencode", "codex"},
			wantIDs:  []string{},
			wantLen:  0,
		},
		{
			name:     "unknown ID in config does not affect result",
			knownIDs: []string{"claude-code", "nonexistent"},
			wantIDs:  []string{"opencode", "codex"},
			wantLen:  2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{RegistryKnown: tt.knownIDs}
			got := findNewAgents(cfg, catalog)
			assert.Len(t, got, tt.wantLen)
			gotIDs := make([]string, len(got))
			for i, a := range got {
				gotIDs[i] = a.ID
			}
			assert.Equal(t, tt.wantIDs, gotIDs)
		})
	}
}

func TestInstallCommand_NotifiesNewAgents(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"claude-code"},
				RegistryKnown:  []string{"claude-code"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": false}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "New agents available")
	assert.Contains(t, output, "OpenCode")
	assert.Contains(t, output, "Codex CLI")
	assert.Contains(t, output, "squad add")
}

func TestInstallCommand_NoNotificationWhenAllKnown(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &installHandler{
		registryURL: "",
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"claude-code"},
				RegistryKnown:  []string{"claude-code", "opencode", "codex"},
			}, nil
		},
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		detectAll: func(agents []registry.Agent) map[string]bool {
			return map[string]bool{"claude-code": false}
		},
		installAll: func(agents []registry.Agent, progress installer.ProgressFn) []error {
			for _, a := range agents {
				if progress != nil {
					progress(a.ID, 100)
				}
			}
			return []error{nil}
		},
		isRuntimeMet: func(deps []registry.RuntimeDep) bool { return true },
		configPath:   func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newInstallCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	assert.NoError(t, err)

	output := buf.String()
	assert.NotContains(t, output, "New agents available")
}

func TestParseAgentIDs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{name: "single", input: "claude-code", want: []string{"claude-code"}},
		{name: "multiple", input: "a,b,c", want: []string{"a", "b", "c"}},
		{name: "with spaces", input: " a , b , c ", want: []string{"a", "b", "c"}},
		{name: "empty", input: "", want: []string{}},
		{name: "trailing comma", input: "a,", want: []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseAgentIDs(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFilterAgentsByID(t *testing.T) {
	agents := []registry.Agent{
		{ID: "a", Name: "A"},
		{ID: "b", Name: "B"},
		{ID: "c", Name: "C"},
	}

	tests := []struct {
		name string
		ids  []string
		want []string
	}{
		{name: "all match", ids: []string{"a", "b", "c"}, want: []string{"a", "b", "c"}},
		{name: "subset", ids: []string{"a", "c"}, want: []string{"a", "c"}},
		{name: "no match", ids: []string{"x", "y"}, want: []string{}},
		{name: "empty", ids: []string{}, want: []string{}},
		{name: "partial match", ids: []string{"a", "x"}, want: []string{"a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterAgentsByID(agents, tt.ids)
			gotIDs := make([]string, len(got))
			for i, a := range got {
				gotIDs[i] = a.ID
			}
			assert.Equal(t, tt.want, gotIDs)
		})
	}
}
