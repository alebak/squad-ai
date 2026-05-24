package cli

import "github.com/spf13/cobra"

// NewRootCommand builds the root Cobra command tree for Squad AI.
func NewRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "squad",
		Short:   "Squad AI manages AI coding agents inside dev containers",
		Version: "0.1.0",
	}
	cmd.SetVersionTemplate("Squad AI version {{.Version}}\n")
	cmd.AddCommand(newVersionCommand())
	return cmd
}
