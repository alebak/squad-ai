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
// The caller populates PreChecked, Blocked, and BlockReason based on runtime
// detection before calling RunSelection.
type AgentItem struct {
	ID          string
	Name        string
	Description string
	PreChecked  bool   // pre-select checkbox — true for compatible agents
	Blocked     bool   // disabled because a runtime dependency is missing
	BlockReason string // human-readable reason, e.g. "requires Node.js 22+"
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
	cursor    int            // index of the currently highlighted agent
	checked   map[int]bool   // agent index → checkbox state
	ready     bool           // true after first WindowSizeMsg
	width     int
	height    int
	submitted bool   // true when user pressed Enter
	err       error  // fatal error, if any
	spinner   spinner.Model
	installing bool
	installMsg string
}

// newModel creates a model from the given agent items.
// PreChecked agents start checked; blocked agents start unchecked and cannot
// be toggled.
func newModel(agents []AgentItem) model {
	checked := make(map[int]bool, len(agents))
	for i, a := range agents {
		if a.PreChecked && !a.Blocked {
			checked[i] = true
		}
	}

	s := spinner.New()
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))

	return model{
		agents:  agents,
		cursor:  0,
		checked: checked,
		spinner: s,
	}
}

// Init implements tea.Model. It requests the alternate screen and starts
// the spinner (for future use during install phase).
func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		m.spinner.Tick,
	)
}

// Update implements tea.Model. It handles key events and window resize messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ready = true
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Global quit keys — always work
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyEscape {
			m.submitted = false
			return m, tea.Quit
		}

		// All other keys only work if we're in selection mode
		if m.submitted {
			return m, nil
		}

		// Handle by key type (arrows, space, enter)
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
			if !m.agents[m.cursor].Blocked {
				m.checked[m.cursor] = !m.checked[m.cursor]
			}
		case tea.KeyEnter:
			m.submitted = true
			return m, tea.Quit
		}

		// Also handle vim-style keys (j/k/q) — these are tea.KeyRunes,
		// not tea.KeyUp/tea.KeyDown, so no double-fire with the above.
		switch msg.String() {
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
			m.submitted = false
			return m, tea.Quit
		}

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

// View implements tea.Model. It renders the selection list.
func (m model) View() string {
	if !m.ready {
		return "Loading..."
	}

	var b strings.Builder

	// ── Title ──
	b.WriteString(styleTitle.Render("🤖 Select Your AI Coding Agents"))
	b.WriteString("\n\n")

	// ── Agent list ──
	for i, agent := range m.agents {
		// Cursor marker
		cursor := "  "
		if i == m.cursor {
			cursor = styleCursor.Render("▸ ")
		}

		// Checkbox
		checked := m.checked[i]
		var checkbox string
		if checked {
			checkbox = styleChecked.Render("◉")
		} else {
			checkbox = "○"
		}

		// Agent name
		name := agent.Name

		// Blocked indicator
		blockedSuffix := ""
		if agent.Blocked {
			blockedSuffix = fmt.Sprintf("  ⛔ %s", agent.BlockReason)
		}

		line := fmt.Sprintf("%s%s %s%s", cursor, checkbox, name, blockedSuffix)

		if agent.Blocked {
			line = styleBlocked.Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")
	}

	// ── Help bar ──
	b.WriteString("\n")
	b.WriteString(styleHelp.Render("↑↓/jk navigate • space toggle • enter confirm • q quit"))

	return styleBorder.Render(b.String())
}

// selectedIDs returns the list of checked agent IDs.
func (m model) selectedIDs() []string {
	var ids []string
	for i, a := range m.agents {
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
// If the user quits with q, Ctrl+C, or Escape, RunSelection returns an
// empty slice and no error. An error is returned only for fatal TUI errors.
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
	if !m.submitted {
		return nil, nil
	}

	return m.selectedIDs(), nil
}
