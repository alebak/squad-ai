// Package tui provides Bubbletea-based terminal UI components for Squad AI,
// including agent selection lists and installation progress views.
package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ──── AgentItem ──────────────────────────────────────────────────────────────

// AgentItem is the display model for one agent in the TUI selection list.
// Blocked agents show their BlockReason as "(reason)" appended to Name.
// IsSelectAll marks the sentinel select-all row (always first element).
// PreChecked marks agents that should start with their checkbox on.
// IsDone marks the sentinel Apply row (always last element) — pressing Enter
// or Space on it confirms the selection and exits.
type AgentItem struct {
	ID          string
	Name        string
	Description string
	Blocked     bool   // disabled because a runtime dependency is missing
	BlockReason string // human-readable reason, e.g. "requires Node.js 22+"
	IsSelectAll bool   // true for the sentinel select-all row
	PreChecked  bool   // true if agent is installed and compatible
	IsDone      bool   // true for the sentinel Apply row
}

// ──── Catppuccin Mocha Styles ─────────────────────────────────────────────────

var (
	styleMauve   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#cba6f7"))
	styleGreen   = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1"))
	styleBlue    = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))
	styleGray    = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))
	styleYellow  = lipgloss.NewStyle().Foreground(lipgloss.Color("#f9e2af"))
	styleRed     = lipgloss.NewStyle().Foreground(lipgloss.Color("#f38ba8"))
	styleWhite   = lipgloss.NewStyle().Foreground(lipgloss.Color("#cdd6f4"))
	styleSurface = lipgloss.NewStyle().Foreground(lipgloss.Color("#45475a"))

	styleTitle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#cba6f7"))

	styleCursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#89b4fa"))

	styleChecked = lipgloss.NewStyle().Foreground(lipgloss.Color("#a6e3a1"))

	styleBlocked = lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("#6c7086"))

	styleHelp = lipgloss.NewStyle().Foreground(lipgloss.Color("#6c7086"))

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

// wizardState holds the state for the inline uninstall wizard.
// When non-nil, the wizard replaces the agent list view.
type wizardState struct {
	step    int     // current step index (0-based)
	total   int     // total number of wizard steps
	indices []int   // indices into m.agents for deselected installed agents
	choices []int   // per-step choices (-1=unset, 0=app, 1=app+config, 2=skip)
	cursor  int     // radio cursor position (0=app only, 1=app+config, 2=skip)
}

// ──── Model ──────────────────────────────────────────────────────────────────

// model is the unexported Bubbletea model for the agent selection TUI.
// It implements tea.Model.
type model struct {
	agents    []AgentItem
	cursor    int          // index of the currently highlighted agent
	checked   map[int]bool // agent index → checkbox state
	isReady   bool         // true after first WindowSizeMsg
	width     int
	height    int
	isSubmitted bool // true when user pressed Enter/Apply to confirm
	spinner   spinner.Model

	showDialog string          // non-empty when a dialog is active ("no-changes")
	wizard     *wizardState    // non-nil when the uninstall wizard is active
	wizardOut  map[string]int  // accumulated wizard choices after completion (agentID → choice)
}

// newModel creates a model from the given agent items.
// Agents with PreChecked=true (and not blocked, not select-all) start checked.
// Blocked agents cannot be toggled.
// A sentinel Apply item is appended at the end for confirming the selection.
func newModel(agents []AgentItem) model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("#cba6f7"))

	// Append the Apply sentinel at the end of the agent list.
	done := AgentItem{ID: "_done", Name: "Apply", IsDone: true}
	agents = append(agents, done)

	m := model{
		agents:  agents,
		cursor:  0,
		checked: make(map[int]bool, len(agents)),
		spinner: s,
	}
	for i, a := range agents {
		if a.PreChecked && !a.Blocked && !a.IsSelectAll && !a.IsDone {
			m.checked[i] = true
		}
	}
	return m
}

// Init implements tea.Model.
func (m model) Init() tea.Cmd {
	return tea.Batch(tea.EnterAltScreen, m.spinner.Tick)
}

// Update implements tea.Model. It handles key events and window resize messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.isReady = true
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyMsg(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
	return m, nil
}

// handleKeyMsg routes key events based on the current mode (dialog, wizard, or normal).
func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEscape {
		m.isSubmitted = false
		return m, tea.Quit
	}
	if m.isSubmitted {
		return m, nil
	}

	// Dialog mode: only Enter handles (dismiss)
	if m.showDialog != "" {
		if msg.Type == tea.KeyEnter {
			m.showDialog = ""
		}
		return m, nil
	}

	// Wizard mode: dispatch to wizard-specific handling
	if m.wizard != nil {
		return m.handleWizardKey(msg)
	}

	// Normal mode
	updated, quit := m.handleSpecialKey(msg)
	m = updated
	if quit {
		return m, tea.Quit
	}
	return m.handleRuneKey(msg.String())
}

// handleSpecialKey handles special keys (arrows, Enter, Space) in normal mode.
func (m model) handleSpecialKey(msg tea.KeyMsg) (model, bool) {
	switch msg.Type {
	case tea.KeyUp:
		m.cursor--
		if m.cursor < 0 {
			m.cursor = len(m.agents) - 1
		}
	case tea.KeyDown:
		m.cursor++
		if m.cursor >= len(m.agents) {
			m.cursor = 0
		}
	case tea.KeySpace, tea.KeyEnter:
		// Apply item — submit the selection
		if m.agents[m.cursor].IsDone {
			m, _ = m.submitSelection()
			return m, m.isSubmitted
		}
		// Toggle agent or select-all
		if m.agents[m.cursor].IsSelectAll {
			m.toggleAll()
		} else if !m.agents[m.cursor].Blocked {
			m.checked[m.cursor] = !m.checked[m.cursor]
		}
	}
	return m, false
}

// handleRuneKey handles rune key presses in normal mode.
func (m model) handleRuneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "a":
		// 'a' submits the selection (same as pressing Enter on Apply)
		m, quit := m.submitSelection()
		if quit {
			return m, tea.Quit
		}
		return m, nil
	case "k":
		m.cursor--
		if m.cursor < 0 {
			m.cursor = len(m.agents) - 1
		}
	case "j":
		m.cursor++
		if m.cursor >= len(m.agents) {
			m.cursor = 0
		}
	case "q":
		m.isSubmitted = false
		return m, tea.Quit
	}
	return m, nil
}

// handleWizardKey handles keys when the uninstall wizard is active.
// It returns tea.Model directly (not a cmd) since wizard mode does not
// trigger external commands.
func (m model) handleWizardKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	ws := m.wizard

	switch {
	case msg.Type == tea.KeyUp || msg.String() == "k":
		ws.cursor--
		if ws.cursor < 0 {
			ws.cursor = 2
		}
	case msg.Type == tea.KeyDown || msg.String() == "j":
		ws.cursor++
		if ws.cursor > 2 {
			ws.cursor = 0
		}
	case msg.Type == tea.KeyEnter:
		// Confirm the current selection
		ws.choices[ws.step] = ws.cursor
	case msg.String() == "n":
		// Next step — only if the current step has a confirmed choice
		if ws.choices[ws.step] != -1 {
			ws.step++
			ws.cursor = 0
			if ws.step >= ws.total {
				m.completeWizard()
			}
		}
	case msg.String() == "b":
		if ws.step > 0 {
			ws.step--
			ws.cursor = 0
		}
	case msg.String() == "q":
		// Cancel wizard — return to agent list
		m.wizard = nil
	}
	return m, nil
}

// submitSelection checks whether the current TUI state needs a wizard or dialog,
// or if it can proceed to submission. It returns the (possibly updated) model
// and a boolean indicating whether the TUI should quit (isSubmitted=true).
func (m model) submitSelection() (model, bool) {
	// If wizard was already completed, submit
	if m.wizardOut != nil {
		m.isSubmitted = true
		return m, true
	}

	// Look for deselected installed agents (PreChecked but now unchecked)
	var deselectedIndices []int
	for i, a := range m.agents {
		if a.IsDone || a.IsSelectAll || a.Blocked {
			continue
		}
		if a.PreChecked && !m.checked[i] {
			deselectedIndices = append(deselectedIndices, i)
		}
	}

	if len(deselectedIndices) > 0 {
		// Start the uninstall wizard
		m.wizard = &wizardState{
			step:    0,
			total:   len(deselectedIndices),
			indices: deselectedIndices,
			choices: make([]int, len(deselectedIndices)),
			cursor:  0,
		}
		for i := range m.wizard.choices {
			m.wizard.choices[i] = -1 // not yet selected
		}
		return m, false
	}

	// Check for empty selection with no deselected agents
	if len(m.selectedIDs()) == 0 {
		m.showDialog = "no-changes"
		return m, false
	}

	// Normal submit
	m.isSubmitted = true
	return m, true
}

// completeWizard builds the wizardOut map from the wizard state and clears the
// wizard, returning the view to the agent list with the Apply item visible.
func (m *model) completeWizard() {
	m.wizardOut = make(map[string]int, len(m.wizard.indices))
	for i, idx := range m.wizard.indices {
		agent := m.agents[idx]
		m.wizardOut[agent.ID] = m.wizard.choices[i]
	}
	m.wizard = nil
}

// View implements tea.Model. It renders the selection list, dialog, or wizard.
func (m model) View() string {
	if !m.isReady {
		return "Loading..."
	}

	var b strings.Builder

	// Header with version info: "Squad AI (version 0.15.0)"
	b.WriteString(styleMauve.Render("Squad AI"))
	b.WriteString(styleGray.Render(" (version 0.15.0)"))
	b.WriteString("\n\n")

	// Wizard mode replaces the agent list
	if m.wizard != nil {
		b.WriteString(m.renderWizardView())
		return styleBorder.Render(b.String())
	}

	// Agent list + Apply
	b.WriteString(styleMauve.Render("Select Your AI Coding Agents"))
	b.WriteString("\n\n")

	for i, agent := range m.agents {
		b.WriteString(m.renderAgentRow(i, agent))
		b.WriteString("\n")
		if i == 0 {
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(styleHelp.Render(
		"↑↓/jk navigate • space/enter toggle • a apply • q quit",
	))

	rendered := styleBorder.Render(b.String())

	// Dialog overlay on top of the agent list
	if m.showDialog != "" {
		rendered = m.renderDialogOverlay(rendered)
	}

	return rendered
}

// renderAgentRow renders a single agent row or the Apply sentinel.
func (m model) renderAgentRow(i int, agent AgentItem) string {
	cursor := "  "
	if i == m.cursor {
		cursor = styleCursor.Render("▸ ")
	}

	if agent.IsSelectAll {
		return m.renderSelectAllRow(cursor)
	}

	if agent.IsDone {
		// Apply sentinel: separator above, bold label, no checkbox.
		sep := styleSurface.Render("───────────────────────────────────────")
		label := styleTitle.Render(agent.Name)
		return fmt.Sprintf("%s\n%s%s", sep, cursor, label)
	}

	var checkbox string
	if m.checked[i] {
		checkbox = styleGreen.Render("◉")
	} else {
		checkbox = styleGray.Render("○")
	}

	name := agent.Name
	if agent.Blocked && agent.BlockReason != "" {
		name = fmt.Sprintf("%s (%s)", agent.Name, agent.BlockReason)
	}

	line := fmt.Sprintf("%s%s %s", cursor, checkbox, name)
	if agent.Blocked {
		line = styleBlocked.Render(line)
	}
	return line
}

// renderSelectAllRow renders the sentinel select-all row.
func (m model) renderSelectAllRow(cursor string) string {
	allChecked := true
	for i, a := range m.agents {
		if a.IsSelectAll || a.IsDone || a.Blocked {
			continue
		}
		if !m.checked[i] {
			allChecked = false
			break
		}
	}

	var checkbox string
	var label string
	if allChecked {
		checkbox = styleGreen.Render("◉")
		label = "unselect all"
	} else {
		checkbox = styleGray.Render("○")
		label = "select all"
	}

	return fmt.Sprintf("%s%s %s", cursor, checkbox, label)
}

// renderWizardView renders the inline uninstall wizard.
func (m model) renderWizardView() string {
	var b strings.Builder

	ws := m.wizard
	agent := m.agents[ws.indices[ws.step]]

	// Title: "Step X of Y — Agent Name"
	b.WriteString(fmt.Sprintf("  %s %s %s\n\n",
		styleMauve.Render(fmt.Sprintf("Step %d of %d", ws.step+1, ws.total)),
		styleGray.Render("—"),
		styleWhite.Render(agent.Name),
	))

	b.WriteString(styleGray.Render("  This agent is currently installed."))
	b.WriteString("\n")
	b.WriteString(styleGray.Render("  Choose an action:"))
	b.WriteString("\n\n")

	// Radio options
	options := []string{
		"Uninstall app only",
		"Uninstall app + config data",
		"Keep installed (skip)",
	}

	choicesMade := false
	for optIdx, label := range options {
		prefix := "  "
		if ws.cursor == optIdx {
			prefix = styleBlue.Render("▸ ")
		}
		// Show checked state
		if ws.choices[ws.step] == optIdx {
			prefix = styleGreen.Render("◉ ")
			choicesMade = true
		}
		b.WriteString(fmt.Sprintf("  %s%s%s\n", prefix, "", label))
		_ = choicesMade
	}

	b.WriteString("\n\n")
	b.WriteString(styleHelp.Render(
		"enter select • ↑↓ navigate • n next • b back • q quit",
	))

	return b.String()
}

// renderDialogOverlay renders the "No changes" dialog overlay on top of the
// existing agent list view.
func (m model) renderDialogOverlay(view string) string {
	dialogContent := fmt.Sprintf(
		"No changes to apply.\n\n%s",
		styleHelp.Render("Press enter to continue..."),
	)
	dialogBox := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Padding(1, 2).
		Width(38).
		Render(dialogContent)

	// Center the dialog over the existing view by placing it on top.
	return lipgloss.JoinVertical(lipgloss.Center,
		view,
		"\n"+dialogBox,
	)
}

// toggleAll flips all compatible agents: if any are unchecked, check all;
// if all are checked, uncheck all. Blocked agents are never affected.
// The select-all sentinel (index 0) and Apply sentinel (last) are skipped
// since they are not real agents.
func (m *model) toggleAll() {
	allChecked := true
	for i, a := range m.agents {
		if a.IsSelectAll || a.IsDone || a.Blocked {
			continue
		}
		if !m.checked[i] {
			allChecked = false
			break
		}
	}
	for i, a := range m.agents {
		if a.IsSelectAll || a.IsDone || a.Blocked {
			continue
		}
		m.checked[i] = !allChecked
	}
}

// selectedIDs returns the list of checked agent IDs.
// The select-all sentinel and Apply sentinel are skipped.
func (m model) selectedIDs() []string {
	var ids []string
	for i, a := range m.agents {
		if a.IsSelectAll || a.IsDone {
			continue
		}
		if m.checked[i] {
			ids = append(ids, a.ID)
		}
	}
	return ids
}

// ──── RunSelection ───────────────────────────────────────────────────────────

// RunSelection launches the Bubbletea TUI and returns the IDs of all
// agents the user selected (checked boxes when Apply was pressed), plus
// wizard choices if the uninstall wizard was used.
//
// Return values encode outcomes:
//   - nil, nil, nil       — user quit (q, Ctrl+C, Escape)
//   - []string{}, nil, nil — confirmed empty selection, no wizard
//   - [ids...], nil, nil  — confirmed with selected IDs, no wizard
//   - [ids...], choices, nil — confirmed with wizard choices for deselected agents
//   - nil, nil, error     — fatal TUI error
//
// Callers MUST distinguish nil from an empty slice: nil means the user
// wants to exit without taking any action.
func RunSelection(agents []AgentItem) ([]string, map[string]int, error) {
	if len(agents) == 0 {
		return nil, nil, nil
	}

	m := newModel(agents)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, nil, fmt.Errorf("running TUI: %w", err)
	}

	m = finalModel.(model)
	if !m.isSubmitted {
		return nil, nil, nil
	}

	return m.selectedIDs(), m.wizardOut, nil
}
