package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/alebak/squad-ai/internal/registry"
	"github.com/spf13/cobra"
)

// infoHandler holds injectable functions for the info command.
type infoHandler struct {
	registryURL   string
	fetchRegistry func(ctx context.Context, url string) (*registry.Catalog, error)
}

// defaultInfoHandler returns an infoHandler wired to real implementations.
func defaultInfoHandler() *infoHandler {
	return &infoHandler{
		registryURL:   "https://raw.githubusercontent.com/alebak/squad-ai/main/registry/agents.json",
		fetchRegistry: registry.Fetch,
	}
}

// newInfoCommand creates the info subcommand.
func newInfoCommand() *cobra.Command {
	return newInfoCommandWithHandler(defaultInfoHandler())
}

// newInfoCommandWithHandler creates the info subcommand with a given handler,
// enabling dependency injection for testing.
func newInfoCommandWithHandler(h *infoHandler) *cobra.Command {
	return &cobra.Command{
		Use:   "info <agent-id>",
		Short: "Show details about a specific AI coding agent",
		Long: `Display detailed information about an AI coding agent from the
registry, including version, runtime dependencies, install method,
and description.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInfoFlow(h, cmd, args)
		},
	}
}

// runInfoFlow executes the core info logic: fetch registry, find the
// requested agent, and print its details.
func runInfoFlow(h *infoHandler, cmd *cobra.Command, args []string) error {
	agentID := strings.TrimSpace(args[0])

	catalog, err := h.fetchRegistry(context.Background(), h.registryURL)
	if err != nil {
		return fmt.Errorf("fetching registry: %w\nTry again when online or use a cached registry.", err)
	}

	var agent *registry.Agent
	for i, a := range catalog.Agents {
		if a.ID == agentID {
			agent = &catalog.Agents[i]
			break
		}
	}
	if agent == nil {
		return fmt.Errorf("agent %q not found in registry", agentID)
	}

	printAgentDetails(cmd, agent, buildRuntimeDisplay(agent.Dependencies))
	return nil
}

// buildRuntimeDisplay builds a human-readable runtime dependency string.
func buildRuntimeDisplay(deps []registry.RuntimeDep) string {
	var runtimes []string
	for _, dep := range deps {
		switch {
		case dep.Runtime == "none":
			runtimes = append(runtimes, "none (self-contained)")
		case dep.MinVersion != "":
			runtimes = append(runtimes, fmt.Sprintf("%s %s+", dep.Runtime, dep.MinVersion))
		default:
			runtimes = append(runtimes, dep.Runtime)
		}
	}
	runtimeStr := strings.Join(runtimes, ", ")
	if runtimeStr == "" {
		runtimeStr = "none"
	}
	return runtimeStr
}

// printAgentDetails prints formatted agent details to cmd.
func printAgentDetails(cmd *cobra.Command, agent *registry.Agent, runtimeStr string) {
	cmd.Printf("Agent:     %s\n", agent.Name)
	cmd.Printf("ID:        %s\n", agent.ID)
	cmd.Printf("Version:   %s\n", agent.Version)
	cmd.Printf("Runtime:   %s\n", runtimeStr)
	cmd.Printf("Install:   %s\n", agent.Install.Method)
	cmd.Printf("Command:   %s\n", agent.Install.Command)
	cmd.Printf("Detect:    %s\n", agent.DetectCmd)
	if agent.Description != "" {
		cmd.Printf("About:     %s\n", agent.Description)
	}
	if len(agent.Tags) > 0 {
		cmd.Printf("Tags:      %s\n", strings.Join(agent.Tags, ", "))
	}
}
