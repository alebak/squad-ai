package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/installer"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/alebak/squad-ai/internal/runtime"
	"github.com/spf13/cobra"
)

// installHandler holds injectable functions for the install command.
// In production, defaultInstallHandler() wires real implementations.
// In tests, fields are replaced with mock functions.
type installHandler struct {
	registryURL   string
	loadConfig    func(path string) (*config.Config, error)
	fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
	detectAll     func(agents []registry.Agent) map[string]bool
	installAll    func(agents []registry.Agent, progress installer.ProgressFn) []error
	isRuntimeMet  func(deps []registry.RuntimeDep) bool
	configPath    func() (string, error)
}

// defaultIsRuntimeMet checks all runtime dependencies for an agent.
// Returns true only if every dependency is satisfied (installed + version OK).
func defaultIsRuntimeMet(deps []registry.RuntimeDep) bool {
	for _, dep := range deps {
		switch dep.Runtime {
		case "none":
			continue
		case "node":
			info := runtime.DetectNode()
			if dep.MinVersion != "" {
				if !runtime.IsCompatible(info, dep.MinVersion) {
					return false
				}
			} else if !info.Installed {
				return false
			}
		case "go":
			info := runtime.DetectGo()
			if dep.MinVersion != "" {
				if !runtime.IsCompatible(info, dep.MinVersion) {
					return false
				}
			} else if !info.Installed {
				return false
			}
		case "python":
			info := runtime.DetectPython()
			if dep.MinVersion != "" {
				if !runtime.IsCompatible(info, dep.MinVersion) {
					return false
				}
			} else if !info.Installed {
				return false
			}
		default:
			// Unknown runtime — block to be safe
			return false
		}
	}
	return true
}

// defaultInstallHandler returns an installHandler wired to real implementations.
func defaultInstallHandler() *installHandler {
	return &installHandler{
		registryURL:   "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json",
		loadConfig:    config.Load,
		fetchRegistry: registry.Fetch,
		detectAll:     installer.DetectAll,
		installAll:    installer.InstallAll,
		isRuntimeMet:  defaultIsRuntimeMet,
		configPath:    config.ConfigPath,
	}
}

// runInstallFlow executes the core install logic:
// 1. Read config, 2. Fetch registry, 3. Determine targets,
// 4. Detect installed, 5. Filter & install, 6. Report results,
// 7. Update config.
func runInstallFlow(h *installHandler, cmd *cobra.Command, args []string, agentsFlag string, allFlag bool) error {
	if agentsFlag != "" && allFlag {
		return fmt.Errorf("--agents and --all are mutually exclusive")
	}

	// 1. Read config
	cfgPath, err := h.configPath()
	if err != nil {
		return fmt.Errorf("determining config path: %w", err)
	}
	cfg, err := h.loadConfig(cfgPath)
	if err != nil {
		return fmt.Errorf("reading config: %w", err)
	}

	// 2. Fetch registry (or cache)
	catalog, err := h.fetchRegistry(context.Background(), h.registryURL)
	if err != nil {
		return fmt.Errorf("fetching registry: %w\nTry again when online or use a cached registry.", err)
	}
	if len(catalog.Agents) == 0 {
		cmd.Println("No agents found in the registry.")
		return nil
	}

	// 3. Determine target agents
	var targets []registry.Agent
	switch {
	case agentsFlag != "":
		ids := parseAgentIDs(agentsFlag)
		targets = filterAgentsByID(catalog.Agents, ids)
	case allFlag:
		targets = catalog.Agents
	default:
		targets = filterAgentsByID(catalog.Agents, cfg.SelectedAgents)
	}

	if len(targets) == 0 {
		cmd.Println("No agents to install.")
		return nil
	}

	// 4. Detect already installed
	installed := h.detectAll(targets)

	// 5. Filter and install
	var toInstall []registry.Agent
	for _, agent := range targets {
		if installed[agent.ID] {
			cmd.Printf("✅ %s already installed\n", agent.Name)
			continue
		}
		if !h.isRuntimeMet(agent.Dependencies) {
			cmd.Printf("⏭️  %s — blocked (runtime requirements not met)\n", agent.Name)
			continue
		}
		toInstall = append(toInstall, agent)
	}

	if len(toInstall) == 0 {
		return nil
	}

	// 6. Install
	progress := func(agentID string, pct int) {
		if pct == 100 {
			for _, a := range toInstall {
				if a.ID == agentID {
					cmd.Printf("✅ %s installed\n", a.Name)
					break
				}
			}
		}
	}

	results := h.installAll(toInstall, progress)

	// 7. Report failures
	var hasErrors bool
	for i, err := range results {
		if err != nil {
			hasErrors = true
			cmd.Printf("❌ %s — %v\n", toInstall[i].Name, err)
		}
	}

	// 8. Update selected agents if --agents flag was used
	if agentsFlag != "" {
		cfg.SelectedAgents = parseAgentIDs(agentsFlag)
	}

	// 9. Notify about new agents in the registry
	newAgents := findNewAgents(cfg, catalog)
	if len(newAgents) > 0 {
		cmd.Println("ℹ️  New agents available in the registry:")
		for _, a := range newAgents {
			cmd.Printf("   • %s — %s\n", a.Name, a.Description)
		}
		cmd.Println("Run 'squad add' to explore them.")
	}

	// 10. Update registry tracking and save config
	allIDs := make([]string, len(catalog.Agents))
	for i, a := range catalog.Agents {
		allIDs[i] = a.ID
	}
	cfg.RegistryKnown = allIDs
	cfg.RegistryLastCheck = time.Now().UTC().Format(time.RFC3339)
	if saveErr := config.Save(cfgPath, cfg); saveErr != nil {
		cmd.Printf("Warning: failed to save config: %v\n", saveErr)
	}

	if hasErrors {
		return fmt.Errorf("one or more installations failed")
	}
	return nil
}

// newInstallCommand creates the install subcommand.
func newInstallCommand() *cobra.Command {
	return newInstallCommandWithHandler(defaultInstallHandler())
}

// newInstallCommandWithHandler creates the install subcommand with a given
// handler, enabling dependency injection for testing.
func newInstallCommandWithHandler(h *installHandler) *cobra.Command {
	var agentsFlag string
	var allFlag bool

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install AI coding agents",
		Long: `Install AI coding agents from the squad registry.

Without flags, installs agents listed in the config file (selected_agents).
With --agents, installs the specified comma-separated agent IDs.
With --all, installs every compatible agent from the registry.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstallFlow(h, cmd, args, agentsFlag, allFlag)
		},
	}

	cmd.Flags().StringVar(&agentsFlag, "agents", "", "Comma-separated agent IDs to install")
	cmd.Flags().BoolVar(&allFlag, "all", false, "Install all compatible agents from the registry")

	return cmd
}

// parseAgentIDs splits a comma-separated string into trimmed agent ID slices.
func parseAgentIDs(raw string) []string {
	parts := strings.Split(raw, ",")
	ids := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			ids = append(ids, trimmed)
		}
	}
	return ids
}

// filterAgentsByID returns agents from the catalog whose ID is in the ids list.
func filterAgentsByID(agents []registry.Agent, ids []string) []registry.Agent {
	idSet := make(map[string]bool, len(ids))
	for _, id := range ids {
		idSet[id] = true
	}

	filtered := make([]registry.Agent, 0, len(ids))
	for _, a := range agents {
		if idSet[a.ID] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// findNewAgents returns agents from the catalog whose IDs are not present
// in cfg.RegistryKnown.
func findNewAgents(cfg *config.Config, catalog *registry.Catalog) []registry.Agent {
	known := make(map[string]bool, len(cfg.RegistryKnown))
	for _, id := range cfg.RegistryKnown {
		known[id] = true
	}

	var newAgents []registry.Agent
	for _, a := range catalog.Agents {
		if !known[a.ID] {
			newAgents = append(newAgents, a)
		}
	}
	return newAgents
}

