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
type AgentItem struct {
	ID          string
	Name        string
	Description string
	Blocked     bool   // disabled because a runtime dependency is missing
	BlockReason string // human-readable reason, e.g. "requires Node.js 22+"
	IsSelectAll bool   // true for the sentinel select-all row
	PreChecked  bool   // true if agent is installed and compatible
}

// ──── Styles ─────────────────────────────────────────────────────────────────

var (
	styleTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("212")) // hot pink

	styleCursor = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")) // cyan

	styleChecked = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")) // green

	styleBlocked = lipgloss.NewStyle().
			Faint(true).
			Foreground(lipgloss.Color("243")) // dark gray

	styleHelp = lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")) // medium gray

	styleBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
)

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
	isSubmitted bool // true when user pressed Enter
	spinner   spinner.Model
}

// newModel creates a model from the given agent items.
// Agents with PreChecked=true (and not blocked, not select-all) start checked.
// Blocked agents cannot be toggled.
func newModel(agents []AgentItem) model {
	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	m := model{
		agents:  agents,
		cursor:  0,
		checked: make(map[int]bool, len(agents)),
		spinner: s,
	}
	for i, a := range agents {
		if a.PreChecked && !a.Blocked && !a.IsSelectAll {
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

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEscape {
		m.isSubmitted = false
		return m, tea.Quit
	}
	if m.isSubmitted {
		return m, nil
	}
	updated, quit := m.handleSpecialKey(msg)
	m = updated
	if quit {
		return m, tea.Quit
	}
	return m.handleRuneKey(msg.String())
}

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
	case tea.KeySpace:
		if m.agents[m.cursor].IsSelectAll {
			m.toggleAll()
		} else if !m.agents[m.cursor].Blocked {
			m.checked[m.cursor] = !m.checked[m.cursor]
		}
	case tea.KeyEnter:
		m.isSubmitted = true
		return m, true
	}
	return m, false
}

func (m model) handleRuneKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "a":
		m.toggleAll()
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

// View implements tea.Model. It renders the selection list.
func (m model) View() string {
	if !m.isReady {
		return "Loading..."
	}

	var b strings.Builder

	b.WriteString(styleTitle.Render("Select Your AI Coding Agents"))
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
		"↑↓/jk navigate • space toggle • enter confirm • q quit",
	))

	return styleBorder.Render(b.String())
}

func (m model) renderAgentRow(i int, agent AgentItem) string {
	cursor := "  "
	if i == m.cursor {
		cursor = styleCursor.Render("▸ ")
	}

	if agent.IsSelectAll {
		return m.renderSelectAllRow(cursor)
	}

	var checkbox string
	if m.checked[i] {
		checkbox = styleChecked.Render("◉")
	} else {
		checkbox = "○"
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

// renderSelectAllRow renders the sentinel select-all row using the same
// ◉/○ checkbox style as agent rows. Label is "select all" when any
// compatible agent is unchecked, "unselect all" when all are checked.
func (m model) renderSelectAllRow(cursor string) string {
	allChecked := true
	for i, a := range m.agents {
		if a.IsSelectAll || a.Blocked {
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
		checkbox = styleChecked.Render("◉")
		label = "unselect all"
	} else {
		checkbox = "○"
		label = "select all"
	}

	return fmt.Sprintf("%s%s %s", cursor, checkbox, label)
}

// toggleAll flips all compatible agents: if any are unchecked, check all;
// if all are checked, uncheck all. Blocked agents are never affected.
// The select-all sentinel (index 0) is skipped since it's not a real agent.
func (m *model) toggleAll() {
	allChecked := true
	for i, a := range m.agents {
		if a.IsSelectAll || a.Blocked {
			continue
		}
		if !m.checked[i] {
			allChecked = false
			break
		}
	}
	for i, a := range m.agents {
		if a.IsSelectAll || a.Blocked {
			continue
		}
		m.checked[i] = !allChecked
	}
}

// selectedIDs returns the list of checked agent IDs.
// The select-all sentinel is skipped since it has no real agent ID.
func (m model) selectedIDs() []string {
	var ids []string
	for i, a := range m.agents {
		if a.IsSelectAll {
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
// agents the user selected (checked boxes when Enter was pressed).
//
// Return values encode three outcomes:
//   - nil, nil       — user quit (q, Ctrl+C, Escape) — no selection made
//   - []string{}, nil — user pressed Enter with nothing checked — confirmed empty
//   - [ids...], nil  — user pressed Enter with checked agents
//   - nil, error     — fatal TUI error
//
// Callers MUST distinguish nil from an empty slice: nil means the user
// wants to exit without taking any action.
func RunSelection(agents []AgentItem) ([]string, error) {
	if len(agents) == 0 {
		return nil, nil
	}

	m := newModel(agents)
	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		return nil, fmt.Errorf("running TUI: %w", err)
	}

	m = finalModel.(model)
	if !m.isSubmitted {
		return nil, nil
	}

	return m.selectedIDs(), nil
}
