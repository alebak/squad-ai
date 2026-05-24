package cli

import (
	"context"
	"fmt"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/installer"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/spf13/cobra"
)

// listHandler holds injectable functions for the list command.
type listHandler struct {
	registryURL   string
	loadConfig    func(path string) (*config.Config, error)
	fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
	detectAll     func(agents []registry.Agent) map[string]bool
	isRuntimeMet  func(deps []registry.RuntimeDep) bool
	configPath    func() (string, error)
}

// defaultListHandler returns a listHandler wired to real implementations.
func defaultListHandler() *listHandler {
	return &listHandler{
		registryURL:   "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json",
		loadConfig:    config.Load,
		fetchRegistry: registry.Fetch,
		detectAll:     installer.DetectAll,
		isRuntimeMet:  defaultIsRuntimeMet,
		configPath:    config.ConfigPath,
	}
}

// newListCommand creates the list subcommand.
func newListCommand() *cobra.Command {
	return newListCommandWithHandler(defaultListHandler())
}

// newListCommandWithHandler creates the list subcommand with a given handler,
// enabling dependency injection for testing.
func newListCommandWithHandler(h *listHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List agents and their installation status",
		Long: `List all agents from the registry and show their status:
installed, selected in config, available, or blocked by runtime requirements.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runListFlow(h, cmd, args)
		},
	}
}

// runListFlow executes the core list logic: read config, fetch registry,
// detect installed agents, and print a status table.
func runListFlow(h *listHandler, cmd *cobra.Command, args []string) error {
	// 1. Read config
	cfgPath, err := h.configPath()
	if err != nil {
		return fmt.Errorf("determining config path: %w", err)
	}
	cfg, err := h.loadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	selected := make(map[string]bool, len(cfg.SelectedAgents))
	for _, id := range cfg.SelectedAgents {
		selected[id] = true
	}

	// 2. Fetch registry (or error)
	catalog, err := h.fetchRegistry(context.Background(), h.registryURL)
	if err != nil {
		return fmt.Errorf("fetching registry: %w\nTry again when online or use a cached registry.", err)
	}

	if len(catalog.Agents) == 0 {
		cmd.Println("No agents found in the registry.")
		return nil
	}

	// 3. Detect installed agents
	installed := h.detectAll(catalog.Agents)

	// 4. Print table header
	cmd.Printf("%-16s %-18s %-11s %s\n", "Agent ID", "Name", "Installed", "Status")
	cmd.Printf("%-16s %-18s %-11s %s\n", "--------", "----", "---------", "------")

	// 5. Print each agent row
	for _, agent := range catalog.Agents {
		isInst := installed[agent.ID]
		installedMark := "❌"
		status := "available"

		if isInst {
			installedMark = "✅"
			status = "installed"
		} else if selected[agent.ID] {
			status = "selected"
		} else if !h.isRuntimeMet(agent.Dependencies) {
			status = "blocked"
		}

		cmd.Printf("%-16s %-18s %-11s %s\n",
			agent.ID, agent.Name, installedMark, status)
	}

	return nil
}
