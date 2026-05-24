package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// testAgents returns a standard set of agents for testing.
func testAgents() []AgentItem {
	return []AgentItem{
		{ID: "claude-code", Name: "Claude Code", PreChecked: true, Blocked: false},
		{ID: "opencode", Name: "OpenCode", PreChecked: true, Blocked: false},
		{ID: "codex", Name: "Codex CLI", PreChecked: false, Blocked: true, BlockReason: "requires Node.js 22+"},
		{ID: "aider", Name: "Aider", PreChecked: false, Blocked: false},
	}
}

// updateModel sends a rune key to the model and returns the updated model.
// Use typ key (j/k/q/space) — sent as KeyRunes.
func updateModel(m model, key string) model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(model)
}

func TestModel_InitialState(t *testing.T) {
	m := newModel(testAgents())

	assert.Equal(t, 0, m.cursor, "cursor starts at 0")

	// Compatible agents (PreChecked=true, Blocked=false) are checked
	assert.True(t, m.checked[0], "claude-code should be checked")
	assert.True(t, m.checked[1], "opencode should be checked")

	// Blocked agent is NOT checked despite PreChecked=true
	assert.False(t, m.checked[2], "codex should NOT be checked (blocked)")

	// Agent with PreChecked=false is not checked
	assert.False(t, m.checked[3], "aider should NOT be checked initially")

	assert.False(t, m.isSubmitted)
}

func TestModel_CursorDown(t *testing.T) {
	m := newModel(testAgents())

	m = updateModel(m, "j")
	assert.Equal(t, 1, m.cursor)

	m = updateModel(m, "j")
	assert.Equal(t, 2, m.cursor)
}

func TestModel_CursorDownWraps(t *testing.T) {
	m := newModel(testAgents())

	for range m.agents {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 0, m.cursor)
}

func TestModel_CursorUp(t *testing.T) {
	m := newModel(testAgents())

	m = updateModel(m, "k")
	assert.Equal(t, len(m.agents)-1, m.cursor)

	m = updateModel(m, "k")
	assert.Equal(t, len(m.agents)-2, m.cursor)
}

func TestModel_CursorArrows(t *testing.T) {
	m := newModel(testAgents())

	// Up arrow works
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, len(m.agents)-1, updated.(model).cursor)

	// Down arrow works from last → wraps to 0
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, updated.(model).cursor)
}

func TestModel_ToggleCheck(t *testing.T) {
	m := newModel(testAgents())
	assert.False(t, m.checked[3], "aider starts unchecked")

	// Navigate to aider (index 3)
	for range 3 {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 3, m.cursor)

	// Toggle with space (real KeyMsg Type)
	toggled, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = toggled.(model)
	assert.True(t, m.checked[3], "aider should now be checked")

	// Toggle again → off
	toggled, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = toggled.(model)
	assert.False(t, m.checked[3], "aider should now be unchecked")
}

func TestModel_BlockedAgentNoToggle(t *testing.T) {
	m := newModel(testAgents())

	// Navigate to codex (index 2, blocked)
	for range 2 {
		m = updateModel(m, "j")
	}

	assert.True(t, m.agents[2].Blocked)
	assert.False(t, m.checked[2])

	// Try to toggle — should be no-op
	toggled, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	assert.False(t, toggled.(model).checked[2], "blocked agent should remain unchecked")
}

func TestModel_EnterConfirms(t *testing.T) {
	m := newModel(testAgents())

	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mFinal := mResult.(model)

	assert.True(t, mFinal.isSubmitted)
	assert.ElementsMatch(t, []string{"claude-code", "opencode"}, mFinal.selectedIDs())
}

func TestModel_EnterIncludesToggledOn(t *testing.T) {
	m := newModel(testAgents())

	// Navigate to aider (index 3) and toggle it on
	for range 3 {
		m = updateModel(m, "j")
	}
	toggled, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = toggled.(model)

	// Navigate back to claude-code (index 0) and toggle it off
	for range 3 {
		m = updateModel(m, "k")
	}
	toggled, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = toggled.(model)

	// Now navigated to opencode — toggle it too, just for fun
	assert.True(t, m.checked[1], "opencode should remain checked")

	// Confirm
	final, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ids := final.(model).selectedIDs()

	assert.ElementsMatch(t, []string{"opencode", "aider"}, ids)
}

func TestModel_CtrlCQuits(t *testing.T) {
	m := newModel(testAgents())

	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.False(t, mResult.(model).isSubmitted)
}

func TestModel_EscapeQuits(t *testing.T) {
	m := newModel(testAgents())

	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, mResult.(model).isSubmitted)
}

func TestModel_QQuits(t *testing.T) {
	m := newModel(testAgents())

	m = updateModel(m, "q")
	assert.False(t, m.isSubmitted)
}

func TestModel_EmptyAgents(t *testing.T) {
	ids, err := RunSelection(nil)
	assert.Nil(t, err)
	assert.Empty(t, ids)

	ids, err = RunSelection([]AgentItem{})
	assert.Nil(t, err)
	assert.Empty(t, ids)
}

func TestModel_ViewRenders(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true
	m.width = 60

	view := m.View()
	assert.Contains(t, view, "Select Your AI Coding Agents")
	assert.Contains(t, view, "Claude Code")
	assert.Contains(t, view, "OpenCode")
	assert.Contains(t, view, "Codex CLI")
	assert.Contains(t, view, "Aider")
	assert.Contains(t, view, "requires Node.js 22+")
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "toggle")
}

func TestModel_NotReady(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = false

	view := m.View()
	assert.Contains(t, view, "Loading...")
}

func TestModel_ToggleAll(t *testing.T) {
	m := newModel(testAgents())

	// All compatible unchecked? toggleAll should check all.
	m.checked[0] = false
	m.checked[1] = false

	m = updateModel(m, "a")
	assert.True(t, m.checked[0], "claude-code should be checked")
	assert.True(t, m.checked[1], "opencode should be checked")
	assert.True(t, m.checked[3], "aider should be checked")
	assert.False(t, m.checked[2], "codex (blocked) should remain unchecked")
}

func TestModel_ToggleAllDeselect(t *testing.T) {
	m := newModel(testAgents())
	m.checked[3] = true // manually check aider too

	// All compatible checked? toggleAll should uncheck all.
	m = updateModel(m, "a")
	assert.False(t, m.checked[0], "claude-code should be unchecked")
	assert.False(t, m.checked[1], "opencode should be unchecked")
	assert.False(t, m.checked[3], "aider should be unchecked")
	assert.False(t, m.checked[2], "codex (blocked) should remain unchecked")
}

func TestModel_HelpIncludesToggleAll(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true

	view := m.View()
	assert.Contains(t, view, "a select all")
}
