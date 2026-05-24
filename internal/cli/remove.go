package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/installer"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/spf13/cobra"
)

// registryURL is the default URL for fetching the agent registry.
const registryURL = "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json"

// removeHandler holds injectable functions for the remove command.
type removeHandler struct {
	loadConfig       func(path string) (*config.Config, error)
	saveConfig       func(path string, cfg *config.Config) error
	configPath       func() (string, error)
	registryURL      string
	fetchRegistry    func(ctx context.Context, url string) (*registry.Catalog, error)
	uninstallAgent   func(agent registry.Agent) error
	isAgentInstalled func(detectCmd string) bool
	confirmFn        func(msg string) bool
}

// defaultRemoveHandler returns a removeHandler wired to real implementations.
func defaultRemoveHandler() *removeHandler {
	return &removeHandler{
		loadConfig:       config.Load,
		saveConfig:       config.Save,
		configPath:       config.ConfigPath,
		registryURL:      registryURL,
		fetchRegistry:    registry.Fetch,
		uninstallAgent:   installer.UninstallAgent,
		isAgentInstalled: installer.IsAgentInstalled,
		confirmFn:        confirmAction,
	}
}

// newRemoveCommand creates the remove subcommand.
func newRemoveCommand() *cobra.Command {
	return newRemoveCommandWithHandler(defaultRemoveHandler())
}

// newRemoveCommandWithHandler creates the remove subcommand with a given handler,
// enabling dependency injection for testing.
func newRemoveCommandWithHandler(h *removeHandler) *cobra.Command {
	var uninstall bool
	var force bool

	cmd := &cobra.Command{
		Use:   "remove <agent-id>",
		Short: "Remove an agent from your config selection",
		Long: `Remove an agent from the selected_agents list in your config.
By default, this does NOT uninstall the agent from your system — it only removes it
from the squad-managed selection.

Use --uninstall to also remove the agent binary from your system.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRemoveFlow(h, cmd, args, uninstall, force)
		},
	}

	cmd.Flags().BoolVar(&uninstall, "uninstall", false, "Also uninstall the agent binary from the system")
	cmd.Flags().BoolVar(&force, "force", false, "Skip confirmation prompt")

	return cmd
}

// confirmAction prompts the user for confirmation and returns true if they
// affirm (y/yes). Reads from stdin so it works in interactive terminals.
func confirmAction(msg string) bool {
	fmt.Fprint(os.Stderr, msg+" [y/N]: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	input := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return input == "y" || input == "yes"
}

// runRemoveFlow executes the core remove logic: optionally uninstall the agent,
// then remove it from selected_agents, and save config.
func runRemoveFlow(h *removeHandler, cmd *cobra.Command, args []string, uninstall bool, force bool) error {
	agentID := strings.TrimSpace(args[0])

	cfgPath, err := h.configPath()
	if err != nil {
		return fmt.Errorf("determining config path: %w", err)
	}
	cfg, err := h.loadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	if uninstall {
		if err := runUninstallFlow(h, cmd, agentID, force); err != nil {
			return err
		}
	}

	ok, err := extractFromConfig(h, cfg, cfgPath, agentID)
	if err != nil {
		return fmt.Errorf("saving config: %w", err)
	}
	if !ok {
		cmd.Printf("⚠️  Agent %q not found in your selected agents.\n", agentID)
		return nil
	}

	cmd.Printf("✅ Removed %q from your selected agents.\n", agentID)
	if !uninstall {
		cmd.Println("Note: the agent is still installed on your system if present.")
	}
	return nil
}

// extractFromConfig removes agentID from cfg.SelectedAgents and saves the config.
// Returns false if the agent wasn't found.
func extractFromConfig(h *removeHandler, cfg *config.Config, cfgPath, agentID string) (bool, error) {
	found := false
	updated := make([]string, 0, len(cfg.SelectedAgents))
	for _, id := range cfg.SelectedAgents {
		if id == agentID {
			found = true
			continue
		}
		updated = append(updated, id)
	}
	if !found {
		return false, nil
	}
	cfg.SelectedAgents = updated
	return true, h.saveConfig(cfgPath, cfg)
}

// runUninstallFlow handles the uninstall portion of the remove command.
// It fetches the registry, finds the agent, checks installation status,
// prompts for confirmation (unless --force), and calls UninstallAgent.
func runUninstallFlow(h *removeHandler, cmd *cobra.Command, agentID string, force bool) error {
	catalog, err := h.fetchRegistry(context.Background(), h.registryURL)
	if err != nil {
		return fmt.Errorf("fetching registry for uninstall: %w", err)
	}

	// Find the agent in the registry.
	var target *registry.Agent
	for i, a := range catalog.Agents {
		if a.ID == agentID {
			target = &catalog.Agents[i]
			break
		}
	}
	if target == nil {
		cmd.Printf("⚠️  Agent %q not found in the registry. Skipping uninstall.\n", agentID)
		return nil
	}

	// Check if the agent is actually installed.
	if !h.isAgentInstalled(target.DetectCmd) {
		cmd.Printf("ℹ️  %q is not installed on your system. Skipping uninstall.\n", agentID)
		return nil
	}

	// Build a human-readable description of what will be done.
	action := describeUninstallAction(target)
	if !force && !h.confirmFn(action) {
		cmd.Println("Uninstall cancelled.")
		return nil
	}

	if err := h.uninstallAgent(*target); err != nil {
		return fmt.Errorf("uninstalling %s: %w", agentID, err)
	}

	cmd.Printf("✅ Uninstalled %q\n", agentID)
	return nil
}

// describeUninstallAction returns a human-readable description of the
// uninstall operation that will be performed.
func describeUninstallAction(agent *registry.Agent) string {
	if agent.Install.UninstallCmd != "" {
		return fmt.Sprintf("This will run: %s", agent.Install.UninstallCmd)
	}
	switch agent.Install.Method {
	case registry.MethodNpmInstall:
		pkg := installer.ExtractNPMPackage(agent.Install.Command)
		if pkg != "" {
			return fmt.Sprintf("This will run: npm uninstall -g %s", pkg)
		}
		return fmt.Sprintf("This will uninstall %q", agent.Name)
	case registry.MethodCurlBash:
		return fmt.Sprintf("This will delete the %q binary from your system", agent.DetectCmd)
	default:
		return fmt.Sprintf("This will uninstall %q", agent.Name)
	}
}
