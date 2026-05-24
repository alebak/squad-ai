package cli

import (
	"fmt"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the root Cobra command tree for Squad AI.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "squad",
		Short: "Squad AI manages AI coding agents inside dev containers",
		Long: `Squad AI installs and manages AI coding agents in dev containers.

On first run (no config file found), it launches an interactive TUI to
select agents. On subsequent runs, it silently installs any missing agents
from the saved configuration.

Use 'squad add' to browse and add new agents at any time.
Use 'squad install --agents <ids>' for non-interactive installation.`,
		Version: "0.1.0",
		RunE:    runRootFlow,
	}
	cmd.SetVersionTemplate("Squad AI version {{.Version}}\n")
	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newInstallCommand())
	cmd.AddCommand(newListCommand())
	cmd.AddCommand(newAddCommand())
	cmd.AddCommand(newRemoveCommand())
	cmd.AddCommand(newUpdateCommand())
	cmd.AddCommand(newInfoCommand())
	return cmd
}

// runRootFlow implements the root command's RunE. It detects whether this is
// a first run (no config.json found) and launches the interactive TUI flow
// if so. If config exists, it prints a brief message and exits.
func runRootFlow(cmd *cobra.Command, args []string) error {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return fmt.Errorf("determining config path: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	// If config has agents, this is not a first run. Print status and exit.
	if len(cfg.SelectedAgents) > 0 {
		cmd.Printf("✅ Squad AI is configured with %d agent(s).\n", len(cfg.SelectedAgents))
		cmd.Println("Run 'squad install' to sync, or 'squad add' to browse new agents.")
		return nil
	}

	// First run — launch the interactive selection flow.
	cmd.Println("👋 Welcome to Squad AI! Let's set up your coding agents.")
	cmd.Println()
	_, err = runAddFlow(defaultAddHandler(), cmd)
	return err
}
