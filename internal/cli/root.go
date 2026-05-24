package cli

import (
	"fmt"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/spf13/cobra"
)

// version is set at build time via ldflags. Defaults to "0.1.0" for
// development builds.
var version = "0.1.0"

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
		Version: version,
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
	cmd.AddCommand(newResetCommand())
	cmd.AddCommand(newSelfUpdateCommand())
	return cmd
}

// runRootFlow launches the interactive TUI to view and manage all agents.
// It always shows the full agent list regardless of whether config exists.
func runRootFlow(cmd *cobra.Command, args []string) error {
	cfgPath, err := config.ConfigPath()
	if err != nil {
		return fmt.Errorf("determining config path: %w", err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	if len(cfg.SelectedAgents) == 0 {
		cmd.Println("👋 Welcome to Squad AI! Let's set up your coding agents.")
		cmd.Println()
	}

	_, err = runAddFlow(defaultAddHandler(), cmd)
	return err
}
