// Package config manages user configuration for squad-ai, including agent
// selection and installation preferences. Config is persisted as JSON in
// $HOME/.config/squad-ai/config.json.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// InstallOptions holds installation preferences.
type InstallOptions struct {
	Silent     bool `json:"silent"`
	PreferPnpm bool `json:"prefer_pnpm"`
}

// Config represents the user's configuration for squad-ai.
type Config struct {
	SelectedAgents    []string       `json:"selected_agents"`
	RegistryLastCheck string         `json:"registry_last_checked"`
	RegistryKnown     []string       `json:"registry_known_agents"`
	InstallOptions    InstallOptions `json:"install_options"`
}

// DefaultConfig returns a Config with sensible defaults: Silent=true,
// PreferPnpm=true, empty slices, and zero-value timestamp.
func DefaultConfig() *Config {
	return &Config{
		InstallOptions: InstallOptions{
			Silent:     true,
			PreferPnpm: true,
		},
	}
}

// ConfigPath returns $HOME/.config/squad-ai/config.json and ensures the
// parent directories exist with 0755 permissions.
func ConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	path := filepath.Join(home, ".config", "squad-ai", "config.json")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", fmt.Errorf("creating config directory: %w", err)
	}
	return path, nil
}

// Load reads a Config from a JSON file at path. If the file does not exist,
// it returns DefaultConfig with no error. Returns a wrapped error on
// malformed JSON or read failure.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config JSON: %w", err)
	}

	return &cfg, nil
}

// Save writes cfg as JSON to path atomically: it creates a temporary file in
// the same directory, marshals the config, sets 0644 permissions, then
// renames to the target path. Parent directories are created with 0755 if
// missing.
func Save(path string, cfg *Config) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	tmp, err := os.CreateTemp(dir, "config.json.*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return fmt.Errorf("writing temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("closing temp file: %w", err)
	}

	if err := os.Chmod(tmp.Name(), 0644); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("setting temp file permissions: %w", err)
	}

	if err := os.Rename(tmp.Name(), path); err != nil {
		os.Remove(tmp.Name())
		return fmt.Errorf("renaming temp file: %w", err)
	}

	return nil
}
