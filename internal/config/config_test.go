package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.True(t, cfg.InstallOptions.Silent)
	assert.True(t, cfg.InstallOptions.PreferPnpm)
	assert.Empty(t, cfg.SelectedAgents)
	assert.Empty(t, cfg.RegistryKnown)
	assert.Empty(t, cfg.RegistryLastCheck)
}

func TestConfigPath(t *testing.T) {
	path, err := ConfigPath()
	require.NoError(t, err)
	assert.Contains(t, path, ".config/squad-ai/config.json")
}

func TestConfigPath_MissingHOME(t *testing.T) {
	t.Setenv("HOME", "")
	_, err := ConfigPath()
	require.Error(t, err)
}

func TestSaveLoad_RoundTrip(t *testing.T) {
	original := &Config{
		SelectedAgents:    []string{"claude-code", "opencode"},
		RegistryLastCheck: "2026-05-24T12:00:00Z",
		RegistryKnown:     []string{"claude-code", "opencode", "pi"},
		InstallOptions: InstallOptions{
			Silent:     true,
			PreferPnpm: false,
		},
	}

	path := filepath.Join(t.TempDir(), "config.json")

	err := Save(path, original)
	require.NoError(t, err)

	loaded, err := Load(path)
	require.NoError(t, err)
	require.NotNil(t, loaded)

	assert.Equal(t, original.SelectedAgents, loaded.SelectedAgents)
	assert.Equal(t, original.RegistryLastCheck, loaded.RegistryLastCheck)
	assert.Equal(t, original.RegistryKnown, loaded.RegistryKnown)
	assert.Equal(t, original.InstallOptions, loaded.InstallOptions)
}

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(t *testing.T, dir string) string
		wantErr bool
		check   func(t *testing.T, cfg *Config)
	}{
		{
			name: "missing file returns defaults",
			setup: func(t *testing.T, dir string) string {
				return filepath.Join(dir, "nonexistent.json")
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				assert.True(t, cfg.InstallOptions.Silent)
				assert.True(t, cfg.InstallOptions.PreferPnpm)
				assert.Empty(t, cfg.SelectedAgents)
			},
		},
		{
			name: "malformed JSON returns error",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "bad.json")
				err := os.WriteFile(p, []byte("{invalid}"), 0644)
				require.NoError(t, err)
				return p
			},
			wantErr: true,
			check: func(t *testing.T, cfg *Config) {
				assert.Nil(t, cfg)
			},
		},
		{
			name: "valid config loads correctly",
			setup: func(t *testing.T, dir string) string {
				p := filepath.Join(dir, "valid.json")
				err := Save(p, &Config{
					SelectedAgents: []string{"agent-1"},
					InstallOptions: InstallOptions{Silent: true},
				})
				require.NoError(t, err)
				return p
			},
			wantErr: false,
			check: func(t *testing.T, cfg *Config) {
				require.NotNil(t, cfg)
				assert.Equal(t, []string{"agent-1"}, cfg.SelectedAgents)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := tt.setup(t, dir)
			cfg, err := Load(path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "parsing config JSON")
			} else {
				assert.NoError(t, err)
			}
			tt.check(t, cfg)
		})
	}
}

func TestSave_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	err := Save(path, DefaultConfig())
	require.NoError(t, err)

	// Verify file exists with content.
	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "install_options")

	// Verify no temp file remains in the directory.
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, "config.json", entries[0].Name())
}

func TestSave_CreatesParentDirs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "a", "b", "config.json")

	err := Save(path, DefaultConfig())
	require.NoError(t, err)

	_, err = os.Stat(path)
	assert.NoError(t, err)
}

func TestConfig_ZeroValueJSON(t *testing.T) {
	var cfg Config
	err := json.Unmarshal([]byte("{}"), &cfg)
	require.NoError(t, err)
	assert.Nil(t, cfg.SelectedAgents)
	assert.Nil(t, cfg.RegistryKnown)
	assert.Empty(t, cfg.RegistryLastCheck)
	assert.False(t, cfg.InstallOptions.Silent)
	assert.False(t, cfg.InstallOptions.PreferPnpm)
}

func TestConfigPath_GOOS(t *testing.T) {
	// Verify the path ends with the expected suffix.
	path, err := ConfigPath()
	require.NoError(t, err)
	assert.Contains(t, path, "/.config/squad-ai/config.json")
}

func TestConfigPath_DirectoryCreation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	path, err := ConfigPath()
	require.NoError(t, err)
	assert.Contains(t, path, ".config/squad-ai/config.json")

	info, err := os.Stat(filepath.Dir(path))
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestSave_ReadOnlyDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ro", "config.json")

	// Create dir first, then make it read-only.
	err := os.MkdirAll(filepath.Dir(path), 0755)
	require.NoError(t, err)
	err = os.Chmod(filepath.Dir(path), 0444)
	require.NoError(t, err)

	err = Save(path, DefaultConfig())
	require.Error(t, err)
}

func TestSave_ConcurrentWrites(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	done := make(chan struct{})

	for i := 0; i < 5; i++ {
		go func() {
			_ = Save(path, DefaultConfig())
			done <- struct{}{}
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), `"install_options"`)
}
