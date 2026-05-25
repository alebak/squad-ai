package tui

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// testAgents returns a standard set of agents for testing.
// The first agent is always the select-all sentinel.
func testAgents() []AgentItem {
	return []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "claude-code", Name: "Claude Code", Blocked: false},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
		{ID: "codex", Name: "Codex CLI", Blocked: true, BlockReason: "requires Node.js 22+"},
		{ID: "aider", Name: "Aider", Blocked: false},
	}
}

// updateModel sends a rune key to the model and returns the updated model.
// Use type key (j/k/q) — sent as KeyRunes.
func updateModel(m model, key string) model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(model)
}

// updateModelKey sends a tea.KeyMsg with the given type and returns the updated model.
func updateModelKey(m model, keyType tea.KeyType) model {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated.(model)
}

func TestModel_InitialState(t *testing.T) {
	m := newModel(testAgents())

	assert.Equal(t, 0, m.cursor, "cursor starts at 0")
	assert.True(t, m.agents[0].IsSelectAll, "first agent is select-all sentinel")

	// All agents start unchecked
	for i, a := range m.agents {
		if a.IsSelectAll {
			continue
		}
		assert.False(t, m.checked[i], "agent %d should start unchecked", i)
	}

	assert.False(t, m.isSubmitted)
}

func TestModel_CursorDown(t *testing.T) {
	m := newModel(testAgents())

	assert.Equal(t, 0, m.cursor, "starts at select-all row")
	m = updateModel(m, "j")
	assert.Equal(t, 1, m.cursor, "moves to first agent")

	m = updateModel(m, "j")
	assert.Equal(t, 2, m.cursor)
}

func TestModel_CursorDownWraps(t *testing.T) {
	m := newModel(testAgents())

	for range m.agents {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 0, m.cursor, "wraps to select-all row")
}

func TestModel_CursorUp(t *testing.T) {
	m := newModel(testAgents())

	// Up from select-all row wraps to last
	m = updateModel(m, "k")
	assert.Equal(t, len(m.agents)-1, m.cursor)

	m = updateModel(m, "k")
	assert.Equal(t, len(m.agents)-2, m.cursor)
}

func TestModel_CursorArrows(t *testing.T) {
	m := newModel(testAgents())

	// Up arrow wraps to last
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, len(m.agents)-1, updated.(model).cursor)

	// Down arrow works from last → wraps to 0
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, updated.(model).cursor)
}

func TestModel_ToggleCheck(t *testing.T) {
	m := newModel(testAgents())
	assert.False(t, m.checked[4], "aider starts unchecked")

	// Navigate to aider (index 4 — index 0 is select-all)
	for range 4 {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 4, m.cursor)

	// Toggle with space
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[4], "aider should now be checked")

	// Toggle again → off
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[4], "aider should now be unchecked")
}

func TestModel_BlockedAgentNoToggle(t *testing.T) {
	m := newModel(testAgents())

	// Navigate to codex (index 3, blocked — index 0 is select-all)
	for range 3 {
		m = updateModel(m, "j")
	}

	assert.True(t, m.agents[3].Blocked)
	assert.False(t, m.checked[3])

	// Try to toggle — should be no-op
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[3], "blocked agent should remain unchecked")
}

func TestModel_EnterConfirms(t *testing.T) {
	m := newModel(testAgents())

	// Check aider
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j") // now at index 4 (aider)
	toggled, _ := m.Update(tea.KeyMsg{Type: tea.KeySpace})
	m = toggled.(model)

	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mFinal := mResult.(model)

	assert.True(t, mFinal.isSubmitted)
	assert.ElementsMatch(t, []string{"aider"}, mFinal.selectedIDs())
}

func TestModel_EnterIncludesToggledOn(t *testing.T) {
	m := newModel(testAgents())

	// Toggle on claude-code (index 1)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Toggle on aider (index 4)
	m = updateModel(m, "j") // index 2
	m = updateModel(m, "j") // index 3
	m = updateModel(m, "j") // index 4
	m = updateModelKey(m, tea.KeySpace)

	// Uncheck claude-code (go back to index 1)
	m = updateModel(m, "k") // index 3
	m = updateModel(m, "k") // index 2
	m = updateModel(m, "k") // index 1
	m = updateModelKey(m, tea.KeySpace)

	// Confirm — only aider should remain
	final, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ids := final.(model).selectedIDs()
	assert.ElementsMatch(t, []string{"aider"}, ids)
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
	assert.Contains(t, view, "select all")
	assert.Contains(t, view, "Claude Code")
	assert.Contains(t, view, "OpenCode")
	assert.Contains(t, view, "Codex CLI (requires Node.js 22+)")
	assert.Contains(t, view, "Aider")
	assert.NotContains(t, view, "✅", "no installed emoji")
	assert.NotContains(t, view, "⛔", "no blocked emoji")
	assert.NotContains(t, view, "a select all", "help bar does not mention a key")
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "toggle")
}

func TestModel_NotReady(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = false

	view := m.View()
	assert.Contains(t, view, "Loading...")
}

func TestModel_InstalledAgentToggleable(t *testing.T) {
	m := newModel(testAgents())

	// Navigate to opencode (index 2)
	for range 2 {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 2, m.cursor)
	assert.Equal(t, "OpenCode", m.agents[2].Name)
	assert.False(t, m.checked[2], "starts unchecked")

	// Toggle on — should work
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[2], "should toggle on")

	// Toggle off — should work
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[2], "should toggle off")
}

func TestModel_ToggleAll(t *testing.T) {
	m := newModel(testAgents())

	// toggleAll should check all compatible (non-blocked, non-select-all).
	m = updateModel(m, "a")
	assert.True(t, m.checked[1], "claude-code should be checked")
	assert.True(t, m.checked[2], "opencode should be checked")
	assert.True(t, m.checked[4], "aider should be checked")
	assert.False(t, m.checked[3], "codex (blocked) should remain unchecked")
	assert.False(t, m.checked[0], "select-all sentinel should stay unchecked")
}

func TestModel_ToggleAllDeselect(t *testing.T) {
	m := newModel(testAgents())

	// First check all
	m = updateModel(m, "a")
	assert.True(t, m.checked[1], "claude-code should be checked")
	assert.True(t, m.checked[2], "opencode should be checked")

	// toggleAll again should uncheck all compatible
	m = updateModel(m, "a")
	assert.False(t, m.checked[1], "claude-code should be unchecked")
	assert.False(t, m.checked[2], "opencode should be unchecked")
	assert.False(t, m.checked[4], "aider should be unchecked")
	assert.False(t, m.checked[3], "codex (blocked) should remain unchecked")
}

func TestModel_SelectAllRowToggle(t *testing.T) {
	m := newModel(testAgents())

	// Space on select-all row should toggle all compatible
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[1], "claude-code should be checked")
	assert.True(t, m.checked[2], "opencode should be checked")
	assert.True(t, m.checked[4], "aider should be checked")
	assert.False(t, m.checked[3], "codex (blocked) should remain unchecked")
	assert.False(t, m.checked[0], "select-all stays unchecked")

	// Space again on select-all row should uncheck all
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[1], "claude-code should be unchecked")
	assert.False(t, m.checked[2], "opencode should be unchecked")
	assert.False(t, m.checked[4], "aider should be unchecked")
	assert.False(t, m.checked[3], "codex (blocked) should remain unchecked")
}

func TestModel_SelectAllDynamicLabel(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true

	// Initially unchecked: should show "select all"
	view := m.View()
	assert.Contains(t, view, "[ ] select all")

	// Check one agent — label should still be "select all" (not ALL checked)
	m = updateModelKey(m, tea.KeyDown) // move to claude-code (index 1)
	m = updateModelKey(m, tea.KeySpace)
	m.isReady = true

	view2 := m.View()
	assert.Contains(t, view2, "[ ] select all", "still select all — not all checked")

	// Check all compatible agents
	m = updateModel(m, "a")
	m.isReady = true

	view3 := m.View()
	assert.Contains(t, view3, "[x] unselect all", "all checked → unselect all")
}

func TestModel_BlockedAgentNoEmoji(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true

	view := m.View()
	assert.NotContains(t, view, "⛔", "no blocked emoji in view")
	assert.Contains(t, view, "Codex CLI (requires Node.js 22+)", "block reason as parenthetical")
}

func TestModel_NoEmojiInView(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true

	view := m.View()
	assert.NotContains(t, view, "🤖", "no title emoji")
	assert.NotContains(t, view, "✅", "no installed emoji")
	assert.NotContains(t, view, "⛔", "no blocked emoji")
}
