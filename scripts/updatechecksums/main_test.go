package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateHTTPSURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		wantErr bool
	}{
		{name: "ok", raw: "https://example.com/install.sh", wantErr: false},
		{name: "http", raw: "http://example.com/install.sh", wantErr: true},
		{name: "localhost", raw: "https://localhost/x", wantErr: true},
		{name: "newline", raw: "https://example.com/x\nCHANGED=true", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateHTTPSURL(tt.raw)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestWriteCatalog_PreservesAgentFields(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "agents.json")

	original := &registry.Catalog{
		UpdatedAt: "2026-01-01",
		Agents: []registry.Agent{
			{
				ID:          "demo",
				Name:        "Demo Agent",
				Description: "keeps fields",
				Version:     "latest",
				DetectCmd:   "demo",
				Install: registry.InstallCmd{
					Method:  registry.MethodCurlBash,
					URL:     "https://example.com/install.sh",
					Command: "curl -fsSL https://example.com/install.sh | bash",
				},
				Dependencies: []registry.RuntimeDep{{Runtime: "none"}},
				ConfigPaths:  []string{"~/.demo"},
				Tags:         []string{"demo"},
				AddedAt:      "2026-01-01",
				Checksum: &registry.Checksum{
					SHA256:           "abc",
					ContentChangedAt: "2026-01-01",
					VerifiedAt:       "2026-01-01",
				},
			},
		},
	}

	// Seed a full document on disk first; writeCatalog patches in place.
	seed, err := json.MarshalIndent(original, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, append(seed, '\n'), 0o644))

	// Mutate only checksum content in memory.
	original.Agents[0].Checksum = &registry.Checksum{
		SHA256:           "def",
		ContentChangedAt: "2026-07-24",
		VerifiedAt:       "2026-07-24",
	}
	require.NoError(t, writeCatalog(path, original, "2026-07-24"))

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	var got registry.Catalog
	require.NoError(t, json.Unmarshal(data, &got))
	require.Len(t, got.Agents, 1)

	a := got.Agents[0]
	assert.Equal(t, "demo", a.ID)
	assert.Equal(t, "Demo Agent", a.Name)
	assert.Equal(t, "keeps fields", a.Description)
	assert.Equal(t, "demo", a.DetectCmd)
	assert.Equal(t, "curl -fsSL https://example.com/install.sh | bash", a.Install.Command)
	assert.Equal(t, "https://example.com/install.sh", a.Install.URL)
	assert.Equal(t, []string{"~/.demo"}, a.ConfigPaths)
	assert.Equal(t, []string{"demo"}, a.Tags)
	assert.Equal(t, "2026-07-24", got.UpdatedAt)
	require.NotNil(t, a.Checksum)
	assert.Equal(t, "def", a.Checksum.SHA256)
}

func TestLoadCatalog_RoundTripRepoRegistry(t *testing.T) {
	// Ensure the real repo registry still unmarshals into the shared type.
	cat, err := loadCatalog(filepath.Join("..", "..", "registry", "agents.json"))
	if err != nil {
		// Allow running tests from module root via go test ./scripts/updatechecksums
		cat, err = loadCatalog(filepath.Join("registry", "agents.json"))
	}
	if err != nil {
		t.Skipf("registry not available from test cwd: %v", err)
	}
	require.NotEmpty(t, cat.Agents)
	for _, a := range cat.Agents {
		assert.NotEmpty(t, a.ID)
		assert.NotEmpty(t, a.Name)
		assert.NotEmpty(t, a.Install.Command)
	}
}
