package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFetch_Success(t *testing.T) {
	reg := Catalog{
		Agents: []Agent{
			{
				ID:   "test-agent",
				Name: "Test Agent",
				Install: InstallCmd{
					Method: MethodCurlBash,
					URL:    "https://example.com/install.sh",
				},
				Dependencies: []RuntimeDep{{Runtime: "none"}},
			},
		},
		UpdatedAt: "2026-05-24",
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(reg)
		assert.NoError(t, err)
	}))
	defer srv.Close()

	ctx := context.Background()
	result, err := Fetch(ctx, srv.URL)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "test-agent", result.Agents[0].ID)
	assert.Equal(t, "Test Agent", result.Agents[0].Name)
}

func TestFetch_ContextCancellation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Wait for context cancellation by delaying response
		select {
		case <-r.Context().Done():
			http.Error(w, "cancelled", http.StatusInternalServerError)
		case <-time.After(5 * time.Second):
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := Fetch(ctx, srv.URL)
	require.Error(t, err)
	// The error should wrap the cancellation cause
	assert.ErrorIs(t, err, context.Canceled)
}

func TestFetch_Non200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	ctx := context.Background()
	_, err := Fetch(ctx, srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "404")
}

func TestFetch_NetworkError(t *testing.T) {
	ctx := context.Background()
	// Point to a port where nothing is listening
	_, err := Fetch(ctx, "http://127.0.0.1:1")
	require.Error(t, err)
	// The error should wrap the underlying network error
	assert.ErrorContains(t, err, "connection refused")
}

func TestCacheRoundTrip(t *testing.T) {
	original := &Catalog{
		Agents: []Agent{
			{
				ID:   "agent-1",
				Name: "Agent One",
				Install: InstallCmd{
					Method: MethodCurlBash,
					URL:    "https://example.com/install.sh",
				},
				Dependencies: []RuntimeDep{{Runtime: "none"}},
				Tags:         []string{"test"},
				AddedAt:      "2026-05-24",
			},
			{
				ID:   "agent-2",
				Name: "Agent Two",
				Install: InstallCmd{
					Method:  MethodNpmInstall,
					Command: "npm i -g @example/agent-two",
				},
				Dependencies: []RuntimeDep{{Runtime: "node"}},
				AddedAt:      "2026-05-24",
			},
		},
		UpdatedAt: "2026-05-24",
	}

	cachePath := filepath.Join(t.TempDir(), "cache.json")

	err := SaveCache(cachePath, original)
	require.NoError(t, err)

	loaded, err := LoadCache(cachePath)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, original.UpdatedAt, loaded.UpdatedAt)
	assert.Len(t, loaded.Agents, 2)
	assert.Equal(t, original.Agents[0].ID, loaded.Agents[0].ID)
	assert.Equal(t, original.Agents[0].Name, loaded.Agents[0].Name)
	assert.Equal(t, original.Agents[0].Install.Method, loaded.Agents[0].Install.Method)
	assert.Equal(t, original.Agents[1].ID, loaded.Agents[1].ID)
	assert.Equal(t, original.Agents[1].Install.Command, loaded.Agents[1].Install.Command)
}

func TestIsStale_Fresh(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "fresh.json")
	err := SaveCache(cachePath, &Catalog{UpdatedAt: "2026-05-24"})
	require.NoError(t, err)

	assert.False(t, IsStale(cachePath, 24*time.Hour))
}

func TestIsStale_Stale(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "stale.json")
	err := SaveCache(cachePath, &Catalog{UpdatedAt: "2026-05-24"})
	require.NoError(t, err)

	// Set mtime to 25 hours ago
	oldTime := time.Now().Add(-25 * time.Hour)
	err = os.Chtimes(cachePath, oldTime, oldTime)
	require.NoError(t, err)

	assert.True(t, IsStale(cachePath, 24*time.Hour))
}

func TestIsStale_Missing(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "nonexistent.json")
	assert.True(t, IsStale(cachePath, 24*time.Hour))
}

func TestLoadCache_MalformedJSON(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "bad.json")
	err := os.WriteFile(cachePath, []byte("{invalid json}"), 0644)
	require.NoError(t, err)

	_, err = LoadCache(cachePath)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing cache JSON")
}

func TestLoadCache_EmptyList(t *testing.T) {
	cachePath := filepath.Join(t.TempDir(), "empty.json")
	err := SaveCache(cachePath, &Catalog{Agents: []Agent{}, UpdatedAt: "2026-05-24"})
	require.NoError(t, err)

	loaded, err := LoadCache(cachePath)
	require.NoError(t, err)
	assert.Len(t, loaded.Agents, 0)
}

func TestFetch_EmptyCatalog(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(Catalog{Agents: []Agent{}})
		assert.NoError(t, err)
	}))
	defer srv.Close()

	ctx := context.Background()
	result, err := Fetch(ctx, srv.URL)
	require.NoError(t, err)
	assert.Len(t, result.Agents, 0)
}

func TestBundledAgentsJSON_CurlBashRequiresChecksum(t *testing.T) {
	// Validate the real registry checked into the repo.
	path := filepath.Join("..", "..", "registry", "agents.json")
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var catalog Catalog
	require.NoError(t, json.Unmarshal(data, &catalog))
	require.NotEmpty(t, catalog.Agents)

	for _, agent := range catalog.Agents {
		require.NotEmpty(t, agent.ID)
		require.NotEmpty(t, agent.Install.Method)
		if agent.Install.Method != MethodCurlBash {
			continue
		}
		require.NotEmpty(t, agent.Install.URL, "curl_bash agent %s needs install.url", agent.ID)
		require.True(t, len(agent.Install.URL) >= 8 && agent.Install.URL[:8] == "https://",
			"curl_bash agent %s URL must be https", agent.ID)
		require.NotNil(t, agent.Checksum, "curl_bash agent %s requires checksum", agent.ID)
		require.NotEmpty(t, agent.Checksum.SHA256, "curl_bash agent %s requires sha256", agent.ID)
		require.Len(t, agent.Checksum.SHA256, 64, "curl_bash agent %s sha256 must be 64 hex chars", agent.ID)
	}
}
