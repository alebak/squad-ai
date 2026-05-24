package cli

import "github.com/spf13/cobra"

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the version number of Squad AI",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Printf("Squad AI version %s\n", cmd.Root().Version)
		},
	}
}
