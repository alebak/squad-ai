package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"
)

// testAgents returns a standard set of agents for testing.
// The first agent is always the select-all sentinel.
// Some agents have PreChecked=true (simulating installed + compatible).
func testAgents() []AgentItem {
	return []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "claude-code", Name: "Claude Code", Blocked: false, PreChecked: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
		{ID: "codex", Name: "Codex CLI", Blocked: true, BlockReason: "requires Node.js 22+"},
		{ID: "aider", Name: "Aider", Blocked: false, PreChecked: true},
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

	// PreChecked agents start checked; others start unchecked
	for i, a := range m.agents {
		if a.IsSelectAll {
			assert.False(t, m.checked[i], "select-all sentinel should not be in checked map")
			continue
		}
		if a.PreChecked && !a.Blocked {
			assert.True(t, m.checked[i], "agent %s (PreChecked) should start checked", a.ID)
		} else {
			assert.False(t, m.checked[i], "agent %s should start unchecked", a.ID)
		}
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
	assert.True(t, m.checked[4], "aider starts checked (PreChecked)")

	// Navigate to aider (index 4 — index 0 is select-all)
	for range 4 {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 4, m.cursor)

	// Toggle with space — uncheck
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[4], "aider should now be unchecked")

	// Toggle again → on
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[4], "aider should now be checked")
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

func TestModel_EnterOnDoneConfirms(t *testing.T) {
	m := newModel(testAgents())

	// Check opencode (index 2, not PreChecked)
	m = updateModel(m, "j") // index 1 (claude-code, PreChecked)
	m = updateModel(m, "j") // index 2 (opencode)
	m = updateModelKey(m, tea.KeySpace)

	// Uncheck aider (index 4, PreChecked) so only claude-code + opencode remain
	m = updateModel(m, "j") // index 3 (codex, blocked)
	m = updateModel(m, "j") // index 4 (aider, PreChecked)
	m = updateModelKey(m, tea.KeySpace)

	// Navigate to Done (last item) and press Enter
	m = updateModel(m, "j") // index 5 (✓ Done)
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mFinal := mResult.(model)

	assert.True(t, mFinal.isSubmitted)
	// claude-code (PreChecked) + opencode (manually checked) = both selected
	assert.ElementsMatch(t, []string{"claude-code", "opencode"}, mFinal.selectedIDs())
}

func TestModel_EnterOnDoneReturnsSelection(t *testing.T) {
	m := newModel(testAgents())

	// claude-code (index 1) starts PreChecked — uncheck it
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// opencode (index 2) starts unchecked — check it
	m = updateModel(m, "j") // index 2
	m = updateModelKey(m, tea.KeySpace)

	// aider (index 4) starts PreChecked — uncheck it
	m = updateModel(m, "j") // index 3
	m = updateModel(m, "j") // index 4
	m = updateModelKey(m, tea.KeySpace)

	// Navigate to Done (last item) and press Enter
	m = updateModel(m, "j") // index 5 (✓ Done)
	final, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ids := final.(model).selectedIDs()
	assert.ElementsMatch(t, []string{"opencode"}, ids)
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
	// Select-all uses ◉/○ not [x]/[ ]
	assert.Contains(t, view, "○", "select-all uses ○ for unchecked style")
	assert.NotContains(t, view, "[ ]", "no bracket-style checkboxes")
	assert.NotContains(t, view, "[x]", "no bracket-style checkboxes")
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

	// Initially: claude-code and aider checked, opencode unchecked
	// Not all compatible checked → "select all"
	view := m.View()
	assert.Contains(t, view, "○ select all")

	// Uncheck claude-code — still not all checked → "select all"
	m = updateModelKey(m, tea.KeyDown) // move to claude-code (index 1)
	m = updateModelKey(m, tea.KeySpace)
	m.isReady = true

	view2 := m.View()
	assert.Contains(t, view2, "○ select all", "still select all — not all checked")

	// Check all compatible agents
	m = updateModel(m, "a")
	m.isReady = true

	view3 := m.View()
	assert.Contains(t, view3, "◉ unselect all", "all checked → unselect all")
}

func TestModel_BlankLineAfterSelectAll(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true
	m.width = 60

	view := m.View()
	// The select-all row is followed by a blank line before the first agent.
	// In the bordered view, split by newline to verify the blank line exists.
	lines := strings.Split(view, "\n")
	selectAllIdx := -1
	claudeIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "select all") {
			selectAllIdx = i
		}
		if strings.Contains(line, "Claude Code") {
			claudeIdx = i
		}
	}
	assert.GreaterOrEqual(t, selectAllIdx, 0, "select-all line should be found")
	assert.GreaterOrEqual(t, claudeIdx, 0, "Claude Code line should be found")
	// There should be exactly one line between select-all and Claude Code
	// (the blank separator). If claudeIdx == selectAllIdx+2, there's 1 blank line.
	assert.Equal(t, selectAllIdx+2, claudeIdx,
		"should be exactly one blank line between select-all and first agent")
}

func TestModel_PreCheckedSetsInitialState(t *testing.T) {
	m := newModel(testAgents())

	// claude-code (index 1) and aider (index 4) have PreChecked=true
	assert.True(t, m.checked[1], "claude-code should start checked")
	assert.True(t, m.checked[4], "aider should start checked")

	// opencode (index 2) has PreChecked=false
	assert.False(t, m.checked[2], "opencode should start unchecked")

	// codex (index 3) is blocked with PreChecked=false
	assert.False(t, m.checked[3], "blocked agent should start unchecked")
}

func TestModel_PreCheckedBlockedNotChecked(t *testing.T) {
	agents := []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "blocked-but-installed", Name: "Blocked But Installed", Blocked: true, PreChecked: true, BlockReason: "requires Node.js"},
	}
	m := newModel(agents)

	// Blocked agent should NOT be checked even if PreChecked=true
	assert.False(t, m.checked[1], "blocked agent should not be checked")
}

func TestModel_PreCheckedToggleWorks(t *testing.T) {
	m := newModel(testAgents())

	// claude-code (index 1) starts PreChecked — toggle should uncheck
	m = updateModel(m, "j") // cursor 1
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[1], "pre-checked agent should uncheck on toggle")

	// Toggle again — should re-check
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[1], "pre-checked agent should re-check on second toggle")
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

func TestModel_EnterTogglesAgent(t *testing.T) {
	m := newModel(testAgents())

	// Navigate to claude-code (index 1, PreChecked)
	m = updateModel(m, "j")
	assert.True(t, m.checked[1], "claude-code starts checked")

	// Enter should toggle (same as Space)
	m = updateModelKey(m, tea.KeyEnter)
	assert.False(t, m.checked[1], "claude-code should be unchecked after Enter toggle")

	// Enter toggles again
	m = updateModelKey(m, tea.KeyEnter)
	assert.True(t, m.checked[1], "claude-code should be re-checked after second Enter")
}

func TestModel_DoneItemRenders(t *testing.T) {
	m := newModel(testAgents())
	m.isReady = true
	m.width = 60

	view := m.View()
	assert.Contains(t, view, "✓ Done", "Done item should appear in the rendered view")
	assert.Contains(t, view, "space/enter toggle", "help bar should mention space/enter toggle")
}

func TestModel_SelectedIDsExcludesDone(t *testing.T) {
	m := newModel(testAgents())

	// Select all compatible agents
	m = updateModel(m, "a")

	ids := m.selectedIDs()
	assert.NotContains(t, ids, "_done", "Done sentinel should not appear in selected IDs")
}

func TestModel_ToggleAllSkipsDone(t *testing.T) {
	m := newModel(testAgents())

	// toggleAll with 'a'
	m = updateModel(m, "a")

	// All compatible agents should be checked
	for i, a := range m.agents {
		if a.IsSelectAll || a.IsDone || a.Blocked {
			continue
		}
		assert.True(t, m.checked[i], "%s should be checked after toggleAll", a.ID)
	}

	// Done should not be in checked map (access returns zero value)
	doneIdx := len(m.agents) - 1
	assert.False(t, m.checked[doneIdx], "Done sentinel should remain unchecked after toggleAll")
}

func TestModel_DoneItemCursorWrap(t *testing.T) {
	m := newModel(testAgents())
	doneIdx := len(m.agents) - 1

	// Navigate down to Done
	for range doneIdx {
		m = updateModel(m, "j")
	}
	assert.Equal(t, doneIdx, m.cursor, "cursor should be on Done sentinel")

	// One more down wraps to select-all (index 0)
	m = updateModel(m, "j")
	assert.Equal(t, 0, m.cursor, "cursor should wrap from Done to first item")

	// Up from first wraps to Done
	m = updateModel(m, "k")
	assert.Equal(t, doneIdx, m.cursor, "cursor should wrap from first to Done sentinel")
}
