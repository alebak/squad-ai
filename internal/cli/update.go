package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/spf13/cobra"
)

// updateHandler holds injectable functions for the update command.
type updateHandler struct {
	registryURL   string
	fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
	saveCache     func(path string, reg *registry.Catalog) error
	cachePath     func() (string, error)
}

// defaultCachePath returns the path to the registry cache file under
// the user's config directory.
func defaultCachePath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine config directory: %w", err)
	}
	return filepath.Join(dir, "squad-ai", "registry.cache.json"), nil
}

// defaultUpdateHandler returns an updateHandler wired to real implementations.
func defaultUpdateHandler() *updateHandler {
	return &updateHandler{
		registryURL:   "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json",
		fetchRegistry: registry.Fetch,
		saveCache:     registry.SaveCache,
		cachePath:     defaultCachePath,
	}
}

// newUpdateCommand creates the update subcommand.
func newUpdateCommand() *cobra.Command {
	return newUpdateCommandWithHandler(defaultUpdateHandler())
}

// newUpdateCommandWithHandler creates the update subcommand with a given handler,
// enabling dependency injection for testing.
func newUpdateCommandWithHandler(h *updateHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Force update of the local agent registry cache",
		Long: `Fetch the latest agent registry from GitHub and update the local
cache. This is useful when you know new agents have been added to the
registry and want to see them without waiting for the automatic refresh.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// 1. Fetch registry from remote
			cmd.Println("Fetching latest registry from GitHub...")
			catalog, err := h.fetchRegistry(context.Background(), h.registryURL)
			if err != nil {
				return fmt.Errorf("fetching registry: %w", err)
			}

			// 2. Save to cache
			cachePath, err := h.cachePath()
			if err != nil {
				return fmt.Errorf("determining cache path: %w", err)
			}

			if err := h.saveCache(cachePath, catalog); err != nil {
				return fmt.Errorf("saving cache: %w", err)
			}

			cmd.Printf("✅ Registry updated — %d agents cached.\n", len(catalog.Agents))
			return nil
		},
	}
}
