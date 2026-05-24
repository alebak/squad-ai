package cli

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInfoCommand_ShowsDetails(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &infoHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
	}

	cmd := newInfoCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"claude-code"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Claude Code")
	assert.Contains(t, output, "claude-code")
	assert.Contains(t, output, "latest")
	assert.Contains(t, output, "none")
	assert.Contains(t, output, "curl_bash")
}

func TestInfoCommand_AgentNotFound(t *testing.T) {
	h := &infoHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
	}

	cmd := newInfoCommandWithHandler(h)
	cmd.SetArgs([]string{"nonexistent"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInfoCommand_MissingArg(t *testing.T) {
	cmd := newInfoCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "accepts")
}

func TestInfoCommand_RegistryFetchFailure(t *testing.T) {
	h := &infoHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return nil, errors.New("connection refused")
		},
	}

	cmd := newInfoCommandWithHandler(h)
	cmd.SetArgs([]string{"claude-code"})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestInfoCommand_AgentWithRuntimeDep(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &infoHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
	}

	cmd := newInfoCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"codex"})

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Codex CLI")
	assert.Contains(t, output, "node")
	assert.Contains(t, output, "npm_install")
}
