package cli

import (
	"bytes"
	"errors"
	"testing"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRemoveCommand_RemovesAgent(t *testing.T) {
	var savedCfg *config.Config
	buf := new(bytes.Buffer)

	h := &removeHandler{
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"claude-code", "opencode"},
			}, nil
		},
		saveConfig: func(path string, cfg *config.Config) error {
			savedCfg = cfg
			return nil
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newRemoveCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"opencode"})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), `Removed "opencode"`)
	assert.Contains(t, buf.String(), "still installed")

	require.NotNil(t, savedCfg)
	assert.Equal(t, []string{"claude-code"}, savedCfg.SelectedAgents)
}

func TestRemoveCommand_AgentNotInConfig(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &removeHandler{
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"claude-code"},
			}, nil
		},
		saveConfig: func(path string, cfg *config.Config) error {
			return nil
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newRemoveCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "not found")
}

func TestRemoveCommand_MissingArg(t *testing.T) {
	cmd := newRemoveCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts")
}

func TestRemoveCommand_SaveFailure(t *testing.T) {
	h := &removeHandler{
		loadConfig: func(path string) (*config.Config, error) {
			return &config.Config{
				SelectedAgents: []string{"claude-code"},
			}, nil
		},
		saveConfig: func(path string, cfg *config.Config) error {
			return errors.New("permission denied")
		},
		configPath: func() (string, error) { return "/tmp/test-config.json", nil },
	}

	cmd := newRemoveCommandWithHandler(h)
	cmd.SetArgs([]string{"claude-code"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "permission denied")
}
