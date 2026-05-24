package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newResetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset Squad AI to a clean state",
		Long: `Delete Squad AI configuration and cache files, restoring the tool to its
first-run state. This does NOT uninstall any agents already on the system.`,
		RunE: runReset,
	}
}

func runReset(cmd *cobra.Command, args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	base := filepath.Join(home, ".config", "squad-ai")

	files := []string{
		filepath.Join(base, "config.json"),
		filepath.Join(base, "registry.cache.json"),
	}

	removed := 0
	for _, f := range files {
		if _, statErr := os.Stat(f); statErr == nil {
			if rmErr := os.Remove(f); rmErr != nil {
				cmd.Printf("⚠️  Could not remove %s: %v\n", f, rmErr)
			} else {
				cmd.Printf("🗑️  Removed %s\n", f)
				removed++
			}
		}
	}

	if removed == 0 {
		cmd.Println("Nothing to reset — Squad AI is already in a clean state.")
	} else {
		cmd.Printf("\n✅ Reset complete. Run 'squad' to configure your agents.\n")
	}
	return nil
}
