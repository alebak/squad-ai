package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/alebak/squad-ai/internal/config"
	"github.com/alebak/squad-ai/internal/installer"
	"github.com/alebak/squad-ai/internal/registry"
	"github.com/alebak/squad-ai/internal/runtime"
	"github.com/alebak/squad-ai/internal/tui"
	"github.com/spf13/cobra"
)

// runtimeBlockReason checks an agent's runtime dependencies and returns a
// human-readable reason if any dependency is not met. Returns empty string
// if all dependencies are satisfied.
func runtimeBlockReason(deps []registry.RuntimeDep) string {
	for _, dep := range deps {
		switch dep.Runtime {
		case "none":
			continue
		case "node":
			info := runtime.DetectNode()
			if !info.Installed {
				return "requires Node.js"
			}
			if dep.MinVersion != "" && !runtime.IsCompatible(info, dep.MinVersion) {
				return fmt.Sprintf("requires Node.js %s+", dep.MinVersion)
			}
		case "go":
			info := runtime.DetectGo()
			if !info.Installed {
				return "requires Go"
			}
			if dep.MinVersion != "" && !runtime.IsCompatible(info, dep.MinVersion) {
				return fmt.Sprintf("requires Go %s+", dep.MinVersion)
			}
		case "python":
			info := runtime.DetectPython()
			if !info.Installed {
				return "requires Python 3"
			}
			if dep.MinVersion != "" && !runtime.IsCompatible(info, dep.MinVersion) {
				return fmt.Sprintf("requires Python %s+", dep.MinVersion)
			}
		default:
			return fmt.Sprintf("requires unknown runtime: %s", dep.Runtime)
		}
	}
	return ""
}

// isTerminal reports whether stdin is a character device (TTY).
func isTerminal() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// uninstallChoice represents the user's choice in the 3-option uninstall prompt.
type uninstallChoice int

const (
	uninstallAppOnly   uninstallChoice = 1
	uninstallAppConfig uninstallChoice = 2
	uninstallCancel    uninstallChoice = 3
)

// addHandler holds injectable functions for the add command.
type addHandler struct {
	registryURL       string
	loadConfig        func(path string) (*config.Config, error)
	fetchRegistry     func(ctx context.Context, url string) (*registry.Catalog, error)
	detectAll         func(agents []registry.Agent) map[string]bool
	installAll        func(agents []registry.Agent, progress installer.ProgressFn) []error
	runSelection      func(items []tui.AgentItem) ([]string, error)
	isRuntimeMet      func(deps []registry.RuntimeDep) bool
	uninstallAgent    func(agent registry.Agent) error
	uninstallConfig   func(agent registry.Agent) error
	uninstallChoiceFn func(agentName string) uninstallChoice
	confirmFn         func(msg string) bool
	configPath        func() (string, error)
	isTerminal        func() bool
}

// defaultAddHandler returns an addHandler wired to real implementations.
func defaultAddHandler() *addHandler {
	return &addHandler{
		registryURL:       "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json",
		loadConfig:        config.Load,
		fetchRegistry:     registry.Fetch,
		detectAll:         installer.DetectAll,
		installAll:        installer.InstallAll,
		runSelection:      tui.RunSelection,
		isRuntimeMet:      defaultIsRuntimeMet,
		uninstallAgent:    installer.UninstallAgent,
		uninstallConfig:   installer.UninstallConfig,
		uninstallChoiceFn: defaultUninstallChoiceFn,
		confirmFn:         confirmAction,
		configPath:        config.ConfigPath,
		isTerminal:        isTerminal,
	}
}

// runAddFlow executes the core agent selection and installation flow:
// 1. Fetch registry, detect installed, check runtimes
// 2. Build AgentItems for available agents
// 3. Dispatch to interactive or non-interactive path
// 4. Save updated config
//
// Returns the updated config so callers can inspect the result.
func runAddFlow(h *addHandler, cmd *cobra.Command) (*config.Config, error) {
	cfgPath, err := h.configPath()
	if err != nil {
		return nil, fmt.Errorf("determining config path: %w", err)
	}
	cfg, err := h.loadConfig(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	catalog, err := h.fetchRegistry(context.Background(), h.registryURL)
	if err != nil {
		return nil, fmt.Errorf("fetching registry: %w\nTry again when online or use a cached registry.", err)
	}
	if len(catalog.Agents) == 0 {
		cmd.Println("No agents found in the registry.")
		return cfg, nil
	}

	// Detect installed agents early — needed for pre-checked rendering and uninstall prompt.
	installed := h.detectAll(catalog.Agents)

	agentItems := buildAgentItemsForAdd(h, catalog, installed)
	if len(agentItems) == 0 {
		cmd.Println("No agents found in the registry.")
		return cfg, nil
	}

	if !h.isTerminal() {
		return runAddFlowNonInteractive(h, cmd, agentItems, cfg)
	}
	return runAddFlowInteractive(h, cmd, agentItems, catalog, cfg, cfgPath, installed)
}

// buildAgentItemsForAdd builds TUI AgentItems for ALL registry agents.
// The first item is always the select-all sentinel row.
// Each agent includes runtime compatibility info.
// PreChecked=true for agents that are installed AND not blocked.
func buildAgentItemsForAdd(h *addHandler, catalog *registry.Catalog, installed map[string]bool) []tui.AgentItem {
	agentItems := []tui.AgentItem{
		{Name: "select all", IsSelectAll: true},
	}
	for _, agent := range catalog.Agents {
		blocked := !h.isRuntimeMet(agent.Dependencies)
		reason := ""
		if blocked {
			reason = runtimeBlockReason(agent.Dependencies)
		}

		agentItems = append(agentItems, tui.AgentItem{
			ID:          agent.ID,
			Name:        agent.Name,
			Description: agent.Description,
			Blocked:     blocked,
			BlockReason: reason,
			PreChecked:  installed[agent.ID] && !blocked,
		})
	}
	return agentItems
}

// runAddFlowNonInteractive prints available agents and exits when running
// without a TTY.
func runAddFlowNonInteractive(h *addHandler, cmd *cobra.Command, agentItems []tui.AgentItem, cfg *config.Config) (*config.Config, error) {
	cmd.Println("No interactive terminal detected.")
	cmd.Println("Use 'squad install --agents <ids>' to install agents non-interactively.")
	cmd.Println("Available agents:")
	for _, a := range agentItems {
		blocked := ""
		if a.Blocked {
			blocked = fmt.Sprintf(" [blocked: %s]", a.BlockReason)
		}
		cmd.Printf("  - %s (%s)%s\n", a.ID, a.Name, blocked)
	}
	return cfg, nil
}

// defaultUninstallChoiceFn reads a 3-option uninstall choice from stdin.
//
// Options:
//
//	1 — Uninstall app only
//	2 — Uninstall app + config data
//	3 — Cancel (keep agent installed)
//
// Invalid input re-prompts until a valid choice is made.
func defaultUninstallChoiceFn(agentName string) uninstallChoice {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Printf("\nUninstall %s?\n", agentName)
		fmt.Println("1) Uninstall app only")
		fmt.Println("2) Uninstall app + config data")
		fmt.Println("3) Cancel")
		fmt.Print("Choose (1-3): ")

		if !scanner.Scan() {
			// EOF or error — treat as cancel.
			return uninstallCancel
		}

		input := strings.TrimSpace(scanner.Text())
		choice, err := strconv.Atoi(input)
		if err != nil || choice < 1 || choice > 3 {
			fmt.Println("Invalid choice. Please enter 1, 2, or 3.")
			continue
		}
		return uninstallChoice(choice)
	}
}

// runAddFlowInteractive launches the TUI for agent selection, installs the
// chosen agents, prompts to uninstall deselected installed agents,
// and saves the updated config.
// installed is pre-computed by runAddFlow to enable pre-checked rendering.
// The TUI selection + uninstall prompt loop restarts when the user chooses
// Cancel, re-launching the TUI with the latest installed state.
func runAddFlowInteractive(h *addHandler, cmd *cobra.Command, agentItems []tui.AgentItem, catalog *registry.Catalog, cfg *config.Config, cfgPath string, installed map[string]bool) (*config.Config, error) {

	var selectedIDs []string
	for {
		var err error
		selectedIDs, err = h.runSelection(agentItems)
		if err != nil {
			return nil, fmt.Errorf("TUI selection failed: %w", err)
		}

		// Collect deselected installed agents BEFORE the empty-check so that
		// the "unselect all" case (selectedIDs is empty but installed agents
		// were deselected) is handled with a confirmation prompt.
		selectedSet := make(map[string]bool, len(selectedIDs))
		for _, id := range selectedIDs {
			selectedSet[id] = true
		}

		var deselected []registry.Agent
		for _, agent := range catalog.Agents {
			if installed[agent.ID] && !selectedSet[agent.ID] {
				deselected = append(deselected, agent)
			}
		}

		if len(selectedIDs) == 0 && len(deselected) == 0 {
			cmd.Println("No agents selected. Nothing to install.")
			return cfg, nil
		}

		if len(deselected) == 0 {
			// No installed agents deselected — proceed to installation.
			break
		}

		var needsRestart bool
		if len(deselected) == 1 {
			// Single agent — existing per-agent 3-option flow.
			agent := deselected[0]
			choice := h.uninstallChoiceFn(agent.Name)
			switch choice {
			case uninstallAppOnly:
				if err := h.uninstallAgent(agent); err != nil {
					cmd.Printf("Warning: failed to uninstall %s: %v\n", agent.Name, err)
				} else {
					cmd.Printf("Uninstalled %s\n", agent.Name)
					delete(installed, agent.ID)
				}
			case uninstallAppConfig:
				if err := h.uninstallAgent(agent); err != nil {
					cmd.Printf("Warning: failed to uninstall %s: %v\n", agent.Name, err)
				} else {
					cmd.Printf("Uninstalled %s (app)\n", agent.Name)
					delete(installed, agent.ID)
				}
				if err := h.uninstallConfig(agent); err != nil {
					cmd.Printf("Warning: failed to clean config for %s: %v\n", agent.Name, err)
				} else {
					cmd.Printf("Cleaned config for %s\n", agent.Name)
				}
			case uninstallCancel:
				// Cancel — re-launch TUI so the user can continue editing.
				// The agent stays in the installed map, so it will be
				// pre-checked when the TUI re-launches.
				needsRestart = true
			}
		} else {
			// Multiple agents — combined confirmation prompt.
			names := make([]string, len(deselected))
			for i, a := range deselected {
				names[i] = a.Name
			}
			msg := fmt.Sprintf("Some selected agents are already installed: %s. Uninstall them as well?", strings.Join(names, ", "))
			if h.confirmFn(msg) {
				for _, agent := range deselected {
					if err := h.uninstallAgent(agent); err != nil {
						cmd.Printf("Warning: failed to uninstall %s: %v\n", agent.Name, err)
					} else {
						cmd.Printf("Uninstalled %s\n", agent.Name)
						delete(installed, agent.ID)
					}
				}
			} else {
				needsRestart = true
			}
		}

		if needsRestart {
			// Rebuild agent selection items with updated installed state
			// and re-launch the TUI.
			agentItems = buildAgentItemsForAdd(h, catalog, installed)
			continue
		}

		// No cancels — proceed to installation.
		break
	}

	toInstall := findAgentsByIDs(catalog, selectedIDs)
	toInstall = filterInstalled(h, toInstall)
	if len(toInstall) == 0 {
		cmd.Println("Selected agents are already installed.")
		return cfg, nil
	}

	cmd.Println("Installing selected agents...")
	results := h.installAll(toInstall, makeProgressFn(cmd, toInstall))

	succeeded, hasErrors := reportAddResults(cmd, toInstall, results)
	cfg.SelectedAgents = append(cfg.SelectedAgents, succeeded...)
	if saveErr := config.Save(cfgPath, cfg); saveErr != nil {
		cmd.Printf("Warning: failed to save config: %v\n", saveErr)
	}

	if hasErrors {
		return cfg, fmt.Errorf("one or more installations failed")
	}
	return cfg, nil
}

// filterInstalled removes agents that are already installed from the slice.
func filterInstalled(h *addHandler, agents []registry.Agent) []registry.Agent {
	installed := h.detectAll(agents)
	var filtered []registry.Agent
	for _, a := range agents {
		if !installed[a.ID] {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

// findAgentsByIDs locates agents in the catalog matching the given IDs.
func findAgentsByIDs(catalog *registry.Catalog, ids []string) []registry.Agent {
	var agents []registry.Agent
	for _, id := range ids {
		for _, a := range catalog.Agents {
			if a.ID == id {
				agents = append(agents, a)
				break
			}
		}
	}
	return agents
}

// reportAddResults reports installation results to cmd and returns the
// IDs of successfully installed agents and whether any failures occurred.
func reportAddResults(cmd *cobra.Command, toInstall []registry.Agent, results []error) ([]string, bool) {
	var hasErrors bool
	var succeeded []string
	for i, err := range results {
		if err != nil {
			hasErrors = true
			cmd.Printf("❌ %s — %v\n", toInstall[i].Name, err)
		} else {
			succeeded = append(succeeded, toInstall[i].ID)
		}
	}
	return succeeded, hasErrors
}

// newAddCommand creates the add subcommand.
func newAddCommand() *cobra.Command {
	return newAddCommandWithHandler(defaultAddHandler())
}

// newAddCommandWithHandler creates the add subcommand with a given handler,
// enabling dependency injection for testing.
func newAddCommandWithHandler(h *addHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Browse and add AI coding agents",
		Long: `Browse available AI coding agents from the registry using an
interactive TUI. Select agents to install, then confirm to begin installation.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := runAddFlow(h, cmd)
			return err
		},
	}
}

// formatAgentIDs formats a slice of agent IDs for user-friendly display.
func formatAgentIDs(ids []string) string {
	if len(ids) == 0 {
		return "(none)"
	}
	return strings.Join(ids, ", ")
}
