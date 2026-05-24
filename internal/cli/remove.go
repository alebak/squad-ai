package cli

import (
	"fmt"
	"strings"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/spf13/cobra"
)

// removeHandler holds injectable functions for the remove command.
type removeHandler struct {
	loadConfig func(path string) (*config.Config, error)
	saveConfig func(path string, cfg *config.Config) error
	configPath func() (string, error)
}

// defaultRemoveHandler returns a removeHandler wired to real implementations.
func defaultRemoveHandler() *removeHandler {
	return &removeHandler{
		loadConfig: config.Load,
		saveConfig: config.Save,
		configPath: config.ConfigPath,
	}
}

// newRemoveCommand creates the remove subcommand.
func newRemoveCommand() *cobra.Command {
	return newRemoveCommandWithHandler(defaultRemoveHandler())
}

// newRemoveCommandWithHandler creates the remove subcommand with a given handler,
// enabling dependency injection for testing.
func newRemoveCommandWithHandler(h *removeHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <agent-id>",
		Short: "Remove an agent from your config selection",
		Long: `Remove an agent from the selected_agents list in your config.
This does NOT uninstall the agent from your system — it only removes it
from the squad-managed selection.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			agentID := strings.TrimSpace(args[0])

			// 1. Read config
			cfgPath, err := h.configPath()
			if err != nil {
				return fmt.Errorf("determining config path: %w", err)
			}
			cfg, err := h.loadConfig(cfgPath)
			if err != nil {
				return fmt.Errorf("reading config: %w", err)
			}

			// 2. Find and remove the agent from selected_agents
			found := false
			updated := make([]string, 0, len(cfg.SelectedAgents))
			for _, id := range cfg.SelectedAgents {
				if id == agentID {
					found = true
					continue
				}
				updated = append(updated, id)
			}
			cfg.SelectedAgents = updated

			if !found {
				cmd.Printf("⚠️  Agent %q not found in your selected agents.\n", agentID)
				return nil
			}

			// 3. Save config
			if err := h.saveConfig(cfgPath, cfg); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}

			cmd.Printf("✅ Removed %q from your selected agents.\n", agentID)
			cmd.Println("Note: the agent is still installed on your system if present.")
			return nil
		},
	}
}
