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

func TestUpdateCommand_Success(t *testing.T) {
	var cached *registry.Catalog
	buf := new(bytes.Buffer)

	h := &updateHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		saveCache: func(path string, reg *registry.Catalog) error {
			cached = reg
			return nil
		},
		cachePath: func() (string, error) { return "/tmp/test-cache.json", nil },
	}

	cmd := newUpdateCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.NoError(t, err)

	assert.Contains(t, buf.String(), "Registry updated")
	require.NotNil(t, cached)
	assert.Len(t, cached.Agents, 3)
}

func TestUpdateCommand_FetchFailure(t *testing.T) {
	h := &updateHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return nil, errors.New("connection refused")
		},
		saveCache: func(path string, reg *registry.Catalog) error {
			return nil
		},
		cachePath: func() (string, error) { return "/tmp/test-cache.json", nil },
	}

	cmd := newUpdateCommandWithHandler(h)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestUpdateCommand_CacheSaveFailure(t *testing.T) {
	buf := new(bytes.Buffer)

	h := &updateHandler{
		registryURL: "",
		fetchRegistry: func(ctx context.Context, url string) (*registry.Catalog, error) {
			return testRegistry(), nil
		},
		saveCache: func(path string, reg *registry.Catalog) error {
			return errors.New("disk full")
		},
		cachePath: func() (string, error) { return "/tmp/test-cache.json", nil },
	}

	cmd := newUpdateCommandWithHandler(h)
	cmd.SetOut(buf)
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "disk full")
}
