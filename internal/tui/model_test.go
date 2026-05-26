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
func updateModel(m model, key string) model {
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)})
	return updated.(model)
}

// updateModelKey sends a tea.KeyMsg with the given type and returns the updated model.
func updateModelKey(m model, keyType tea.KeyType) model {
	updated, _ := m.Update(tea.KeyMsg{Type: keyType})
	return updated.(model)
}

// agentsForWizard returns agents where opencode is installed+compatible and
// aider is installed+compatible, so deselecting them triggers the wizard.
func agentsForWizard() []AgentItem {
	return []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "claude-code", Name: "Claude Code", Blocked: false, PreChecked: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false, PreChecked: true},
		{ID: "aider", Name: "Aider", Blocked: false},
	}
}

func TestModel_InitialState(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

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
	assert.Empty(t, m.showDialog)
	assert.Nil(t, m.wizard)
	assert.Nil(t, m.wizardOut)
}

func TestModel_CursorDown(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	assert.Equal(t, 0, m.cursor, "starts at select-all row")
	m = updateModel(m, "j")
	assert.Equal(t, 1, m.cursor, "moves to first agent")

	m = updateModel(m, "j")
	assert.Equal(t, 2, m.cursor)
}

func TestModel_CursorDownWraps(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	for range m.agents {
		m = updateModel(m, "j")
	}
	assert.Equal(t, 0, m.cursor, "wraps to select-all row")
}

func TestModel_CursorUp(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// Up from select-all row wraps to last
	m = updateModel(m, "k")
	assert.Equal(t, len(m.agents)-1, m.cursor)

	m = updateModel(m, "k")
	assert.Equal(t, len(m.agents)-2, m.cursor)
}

func TestModel_CursorArrows(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// Up arrow wraps to last
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyUp})
	assert.Equal(t, len(m.agents)-1, updated.(model).cursor)

	// Down arrow works from last → wraps to 0
	updated, _ = updated.Update(tea.KeyMsg{Type: tea.KeyDown})
	assert.Equal(t, 0, updated.(model).cursor)
}

func TestModel_ToggleCheck(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	assert.True(t, m.checked[4], "aider starts checked (PreChecked)")

	// Navigate to aider (index 4)
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
	m := 	newModel(testAgents(), "0.1.0")

	// Navigate to codex (index 3, blocked)
	for range 3 {
		m = updateModel(m, "j")
	}

	assert.True(t, m.agents[3].Blocked)
	assert.False(t, m.checked[3])

	// Try to toggle — should be no-op
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[3], "blocked agent should remain unchecked")
}

func TestModel_EnterOnApplyConfirms(t *testing.T) {
	// Use agents where only claude-code (not aider) is PreChecked
	// to avoid triggering the wizard when toggling.
	agents := []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "claude-code", Name: "Claude Code", Blocked: false, PreChecked: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
		{ID: "codex", Name: "Codex CLI", Blocked: true, BlockReason: "requires Node.js 22+"},
		{ID: "aider", Name: "Aider", Blocked: false},
	}
	m := newModel(agents, "0.1.0")

	// Check opencode (index 2, not PreChecked)
	m = updateModel(m, "j") // index 1 (claude-code, PreChecked)
	m = updateModel(m, "j") // index 2 (opencode)
	m = updateModelKey(m, tea.KeySpace)

	// Navigate past aider to Apply (last item) and press Enter
	m = updateModel(m, "j") // index 3 (codex, blocked)
	m = updateModel(m, "j") // index 4 (aider, not PreChecked)
	m = updateModel(m, "j") // index 5 (Apply)
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mFinal := mResult.(model)

	assert.True(t, mFinal.isSubmitted)
	assert.ElementsMatch(t, []string{"claude-code", "opencode"}, mFinal.selectedIDs())
}

func TestModel_EnterOnApplyReturnsSelection(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

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

	// Navigate to Apply (last item) and press Enter
	m = updateModel(m, "j") // index 5 (Apply)
	final, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	ids := final.(model).selectedIDs()
	assert.ElementsMatch(t, []string{"opencode"}, ids)
}

func TestModel_CtrlCQuits(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	assert.False(t, mResult.(model).isSubmitted)
}

func TestModel_EscapeQuits(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	assert.False(t, mResult.(model).isSubmitted)
}

func TestModel_QQuits(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	m = updateModel(m, "q")
	assert.False(t, m.isSubmitted)
}

func TestModel_EmptyAgents(t *testing.T) {
	ids, choices, err := RunSelection(nil, "0.1.0")
	assert.Nil(t, err)
	assert.Nil(t, ids)
	assert.Nil(t, choices)

	ids, choices, err = RunSelection([]AgentItem{}, "0.1.0")
	assert.Nil(t, err)
	assert.Nil(t, ids)
	assert.Nil(t, choices)
}

func TestModel_ViewRenders(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true
	m.width = 60

	view := m.View()
	assert.Contains(t, view, "Squad AI")
	assert.Contains(t, view, "version 0.1.0")
	assert.Contains(t, view, "Select Your AI Coding Agents")
	assert.Contains(t, view, "select all")
	assert.Contains(t, view, "Claude Code")
	assert.Contains(t, view, "OpenCode")
	assert.Contains(t, view, "Codex CLI (requires Node.js 22+)")
	assert.Contains(t, view, "Aider")
	assert.Contains(t, view, "Apply")
	assert.NotContains(t, view, "✓ Done", "Done renamed to Apply")
	assert.NotContains(t, view, "✅", "no installed emoji")
	assert.NotContains(t, view, "⛔", "no blocked emoji")
	assert.Contains(t, view, "a apply", "help bar mentions a apply")
	assert.Contains(t, view, "navigate")
	assert.Contains(t, view, "toggle")
	assert.Contains(t, view, "○", "select-all uses ○ for unchecked style")
	assert.NotContains(t, view, "[ ]", "no bracket-style checkboxes")
	assert.NotContains(t, view, "[x]", "no bracket-style checkboxes")
}

func TestModel_NotReady(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = false

	view := m.View()
	assert.Contains(t, view, "Loading...")
}

func TestModel_InstalledAgentToggleable(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

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
	m := 	newModel(testAgents(), "0.1.0")

	// toggleAll via Space on select-all row (index 0)
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[1], "claude-code should be checked")
	assert.True(t, m.checked[2], "opencode should be checked")
	assert.True(t, m.checked[4], "aider should be checked")
	assert.False(t, m.checked[3], "codex (blocked) should remain unchecked")
	assert.False(t, m.checked[0], "select-all sentinel should stay unchecked")
}

func TestModel_ToggleAllDeselect(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// First check all via Space on select-all
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[1], "claude-code should be checked")
	assert.True(t, m.checked[2], "opencode should be checked")

	// Space on select-all again should uncheck all compatible
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[1], "claude-code should be unchecked")
	assert.False(t, m.checked[2], "opencode should be unchecked")
	assert.False(t, m.checked[4], "aider should be unchecked")
	assert.False(t, m.checked[3], "codex (blocked) should remain unchecked")
}

func TestModel_SelectAllRowToggle(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

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
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true

	// Initially: claude-code and aider checked, opencode unchecked
	view := m.View()
	assert.Contains(t, view, "○ select all")

	// Uncheck claude-code — still not all checked → "select all"
	m = updateModelKey(m, tea.KeyDown) // move to claude-code (index 1)
	m = updateModelKey(m, tea.KeySpace)
	m.isReady = true

	view2 := m.View()
	assert.Contains(t, view2, "○ select all", "still select all — not all checked")

	// Check all compatible agents via Space on select-all (cursor is at index 1, need to go to 0)
	m = updateModelKey(m, tea.KeyUp) // move to select-all
	m = updateModelKey(m, tea.KeySpace)
	m.isReady = true

	view3 := m.View()
	assert.Contains(t, view3, "◉ unselect all", "all checked → unselect all")
}

func TestModel_BlankLineAfterSelectAll(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true
	m.width = 60

	view := m.View()
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
	assert.Equal(t, selectAllIdx+2, claudeIdx,
		"should be exactly one blank line between select-all and first agent")
}

func TestModel_PreCheckedSetsInitialState(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

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
	m := newModel(agents, "0.1.0")

	assert.False(t, m.checked[1], "blocked agent should not be checked")
}

func TestModel_PreCheckedToggleWorks(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// claude-code (index 1) starts PreChecked — toggle should uncheck
	m = updateModel(m, "j") // cursor 1
	m = updateModelKey(m, tea.KeySpace)
	assert.False(t, m.checked[1], "pre-checked agent should uncheck on toggle")

	// Toggle again — should re-check
	m = updateModelKey(m, tea.KeySpace)
	assert.True(t, m.checked[1], "pre-checked agent should re-check on second toggle")
}

func TestModel_BlockedAgentNoEmoji(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true

	view := m.View()
	assert.NotContains(t, view, "⛔", "no blocked emoji in view")
	assert.Contains(t, view, "Codex CLI (requires Node.js 22+)", "block reason as parenthetical")
}

func TestModel_NoEmojiInView(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true

	view := m.View()
	assert.NotContains(t, view, "🤖", "no title emoji")
	assert.NotContains(t, view, "✅", "no installed emoji")
	assert.NotContains(t, view, "⛔", "no blocked emoji")
}

func TestModel_EnterTogglesAgent(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

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

func TestModel_ApplyItemRenders(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true
	m.width = 60

	view := m.View()
	assert.Contains(t, view, "Apply", "Apply item should appear in the rendered view")
	assert.NotContains(t, view, "✓ Done", "Done should be renamed to Apply")
	assert.Contains(t, view, "───────────────────────────────────────", "separator should appear above Apply")
	assert.Contains(t, view, "a apply", "help bar should mention a apply key")
}

func TestModel_AKeySubmits(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// Press 'a' — should submit (not toggle all)
	m = updateModel(m, "a")
	assert.True(t, m.isSubmitted, "a key should submit the selection")
}

func TestModel_NoChangesDialog(t *testing.T) {
	// Use agents where nothing is PreChecked — empty selection
	agents := []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
	}
	m := newModel(agents, "0.1.0")

	// Navigate to Apply and press Enter — nothing selected
	m = updateModel(m, "j") // move to opencode
	m = updateModel(m, "j") // move to Apply
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.Equal(t, "no-changes", m.showDialog, "dialog should appear when nothing selected")
	assert.False(t, m.isSubmitted, "should not submit")
}

func TestModel_NoChangesDismiss(t *testing.T) {
	agents := []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
	}
	m := newModel(agents, "0.1.0")

	// Trigger dialog
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, "no-changes", m.showDialog)

	// Dismiss with Enter
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Empty(t, m.showDialog, "dialog should be dismissed after Enter")
}

func TestModel_NoChangesDialogViaAKey(t *testing.T) {
	agents := []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
	}
	m := newModel(agents, "0.1.0")

	// Press 'a' with nothing selected — should show dialog
	m = updateModel(m, "a")
	assert.Equal(t, "no-changes", m.showDialog, "dialog should appear via a key")
	assert.False(t, m.isSubmitted)
}

func TestModel_WizardInit(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code (index 1, PreChecked) and opencode (index 2, PreChecked)
	m = updateModel(m, "j") // index 1
	m = updateModelKey(m, tea.KeySpace) // uncheck claude-code
	m = updateModel(m, "j") // index 2
	m = updateModelKey(m, tea.KeySpace) // uncheck opencode

	// Navigate to Apply and press Enter
	m = updateModel(m, "j") // index 3 (aider, not PreChecked)
	m = updateModel(m, "j") // index 4 (Apply)
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.NotNil(t, m.wizard, "wizard should be initialized when installed agents deselected")
	assert.Equal(t, 2, m.wizard.total, "wizard should have 2 steps")
	assert.Equal(t, 0, m.wizard.step, "wizard should start at step 0")
	assert.False(t, m.isSubmitted, "should not submit yet")
}

func TestModel_WizardNavigationJK(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Navigate to Apply and press Enter to start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.NotNil(t, m.wizard)
	assert.Equal(t, 0, m.wizard.cursor, "cursor starts at option 0")

	// j moves through 5 positions (0→1→2→3→4→0)
	m = updateModel(m, "j")
	assert.Equal(t, 1, m.wizard.cursor)

	m = updateModel(m, "j")
	assert.Equal(t, 2, m.wizard.cursor)

	m = updateModel(m, "j")
	assert.Equal(t, wizardBackIdx, m.wizard.cursor, "cursor at Back button after 3 j presses")

	m = updateModel(m, "j")
	assert.Equal(t, wizardNextIdx, m.wizard.cursor, "cursor at Next button after 4 j presses")

	// j wraps to 0
	m = updateModel(m, "j")
	assert.Equal(t, 0, m.wizard.cursor, "cursor wraps to radio 0")

	// k wraps back to Next button
	m = updateModel(m, "k")
	assert.Equal(t, wizardNextIdx, m.wizard.cursor, "k from radio 0 wraps to Next")
}

func TestModel_WizardEnterSelectsAndAutoAdvances(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode — need 2 agents to test multi-step
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard (2 steps)
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.NotNil(t, m.wizard)
	assert.Equal(t, 2, m.wizard.total)
	assert.Equal(t, -1, m.wizard.choices[0], "unselected before enter")
	assert.Equal(t, -1, m.wizard.choices[1], "step 1 unselected")

	// Enter on option 0 at step 0 — should select AND auto-advance to step 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 0, m.wizard.choices[0], "enter should select option 0")
	assert.Equal(t, 1, m.wizard.step, "enter should auto-advance to step 1")
	assert.False(t, m.wizard.showingSummary, "should not show summary yet")

	// At step 1, move to option 2 and select it — should auto-advance to summary
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 2, m.wizard.choices[1], "should select option 2 for second agent")
	assert.True(t, m.wizard.showingSummary, "should show summary after last step")
}

func TestModel_WizardNextBack(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 for step 0 — auto-advances to step 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 1, m.wizard.step, "Enter on radio auto-advances to step 1")
	assert.Equal(t, 0, m.wizard.choices[0], "step 0 choice is stored")

	// n without a confirmed choice on step 1 should NOT advance
	m = updateModel(m, "n")
	assert.Equal(t, 1, m.wizard.step, "n without choice should not advance")

	// b goes back to step 0
	m = updateModel(m, "b")
	assert.Equal(t, 0, m.wizard.step, "b should go back to step 0")

	// At step 0, b should stay at step 0 (can't go below 0)
	m = updateModel(m, "b")
	assert.Equal(t, 0, m.wizard.step, "b at step 0 should stay at step 0")

	// n with confirmed choice advances (step 0 already has choice 0)
	m = updateModel(m, "n")
	assert.Equal(t, 1, m.wizard.step, "n with confirmed choice should advance to step 1")
}

func TestModel_WizardCancel(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.NotNil(t, m.wizard)

	// q cancels wizard
	m = updateModel(m, "q")
	assert.Nil(t, m.wizard, "wizard should be nil after cancel")
	assert.False(t, m.isSubmitted, "should not submit on cancel")
}

func TestModel_WizardCompleteThroughSummary(t *testing.T) {
	m := newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code (index 1) — only 1 agent in wizard
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard from Apply
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 (app only) for step 0 — auto-advances to summary
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.True(t, m.wizard.showingSummary, "should show summary after last step")
	assert.NotNil(t, m.wizard, "wizard should still be active")

	// Enter on Apply in summary view — completes wizard and submits immediately
	mResult, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Nil(t, m.wizard, "wizard should be nil after summary Apply")
	assert.True(t, m.isSubmitted, "should be submitted immediately after summary Apply")
	assert.NotNil(t, m.wizardOut, "wizardOut should be populated")
	assert.Equal(t, 0, m.wizardOut["claude-code"], "claude-code should have choice 0")

	// Verify tea.Quit was returned
	assert.NotNil(t, cmd, "should return tea.Quit command")
}

func TestModel_SelectedIDsExcludesApply(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// Select all compatible agents via Space on select-all
	m = updateModelKey(m, tea.KeySpace)

	ids := m.selectedIDs()
	assert.NotContains(t, ids, "_done", "Apply sentinel should not appear in selected IDs")
}

func TestModel_ToggleAllSkipsApply(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")

	// toggleAll via Space on select-all
	m = updateModelKey(m, tea.KeySpace)

	// All compatible agents should be checked
	for i, a := range m.agents {
		if a.IsSelectAll || a.IsDone || a.Blocked {
			continue
		}
		assert.True(t, m.checked[i], "%s should be checked after toggleAll", a.ID)
	}

	// Apply should not be in checked map
	applyIdx := len(m.agents) - 1
	assert.False(t, m.checked[applyIdx], "Apply sentinel should remain unchecked after toggleAll")
}

func TestModel_ApplyItemCursorWrap(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	applyIdx := len(m.agents) - 1

	// Navigate down to Apply
	for range applyIdx {
		m = updateModel(m, "j")
	}
	assert.Equal(t, applyIdx, m.cursor, "cursor should be on Apply sentinel")

	// One more down wraps to select-all (index 0)
	m = updateModel(m, "j")
	assert.Equal(t, 0, m.cursor, "cursor should wrap from Apply to first item")

	// Up from first wraps to Apply
	m = updateModel(m, "k")
	assert.Equal(t, applyIdx, m.cursor, "cursor should wrap from first to Apply sentinel")
}

func TestModel_SeparatorRendering(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true
	m.width = 60

	view := m.View()
	assert.Contains(t, view, "───────────────────────────────────────", "separator should render")
}

func TestModel_HeaderShowsVersion(t *testing.T) {
	m := 	newModel(testAgents(), "0.1.0")
	m.isReady = true
	m.width = 60

	view := m.View()
	assert.Contains(t, view, "Squad AI")
	assert.Contains(t, view, "version 0.1.0")
}

func TestModel_SelectedIDsAfterWizard(t *testing.T) {
	m := newModel(agentsForWizard(), "0.1.0")

	// Keep claude-code checked. Deselect opencode (index 2, PreChecked).
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 2 (skip) for the only wizard step — auto-advances to summary
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.True(t, m.wizard.showingSummary, "should be on summary")

	// Apply on summary completes wizard and submits immediately
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Nil(t, m.wizard)
	assert.True(t, m.isSubmitted, "should be submitted immediately after summary Apply")
	assert.Equal(t, 2, m.wizardOut["opencode"])
	assert.ElementsMatch(t, []string{"claude-code"}, m.selectedIDs(),
		"should only return claude-code (opencode was deselected)")
}

func TestModel_DialogBlocksOtherKeys(t *testing.T) {
	agents := []AgentItem{
		{Name: "select all", IsSelectAll: true},
		{ID: "opencode", Name: "OpenCode", Blocked: false},
	}
	m := newModel(agents, "0.1.0")

	// Trigger dialog
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, "no-changes", m.showDialog)

	// Other keys in dialog mode should not do anything
	m = updateModel(m, "j")
	assert.Equal(t, "no-changes", m.showDialog, "j should not dismiss dialog")

	m = updateModel(m, "q")
	assert.Equal(t, "no-changes", m.showDialog, "q should not dismiss dialog")

	// Only Enter dismisses
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Empty(t, m.showDialog, "Enter should dismiss dialog")
}

// ──── Wizard Button Tests ─────────────────────────────────────────────────

func TestModel_WizardButtonsRender(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	m.isReady = true

	view := m.View()
	assert.Contains(t, view, "[ ◄ Back ]", "wizard should show Back button")
	assert.Contains(t, view, "[ Next ► ]", "wizard should show Next button")
}

func TestModel_WizardCursorArrowKeys(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.Equal(t, 0, m.wizard.cursor)

	// Down arrow wraps through 5 positions
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mResult.(model)
	assert.Equal(t, 1, m.wizard.cursor)

	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mResult.(model)
	assert.Equal(t, 2, m.wizard.cursor)

	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mResult.(model)
	assert.Equal(t, wizardBackIdx, m.wizard.cursor, "down from radio 2 → Back button")

	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mResult.(model)
	assert.Equal(t, wizardNextIdx, m.wizard.cursor, "down from Back → Next button")

	// One more wraps to 0
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = mResult.(model)
	assert.Equal(t, 0, m.wizard.cursor, "down from Next wraps to radio 0")

	// Up wraps to Next
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = mResult.(model)
	assert.Equal(t, wizardNextIdx, m.wizard.cursor, "up from radio 0 wraps to Next")
}

func TestModel_WizardEnterOnNextButton(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode (2 steps)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 for step 0 — auto-advances to step 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 1, m.wizard.step)

	// Move cursor to Next button (position 4) on step 1 (no choice yet)
	m.wizard.cursor = wizardNextIdx
	assert.Equal(t, -1, m.wizard.choices[1], "step 1 has no choice")

	// Enter on Next without a choice should NOT advance
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 1, m.wizard.step, "Next without choice should not advance")

	// Now set a choice and try Next
	m.wizard.choices[1] = 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.True(t, m.wizard.showingSummary, "Next with choice should advance to summary")
}

func TestModel_WizardEnterOnBackButton(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode (2 steps)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 for step 0 — auto-advances to step 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 1, m.wizard.step)

	// Navigate to Back button and press Enter
	m.wizard.cursor = wizardBackIdx
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 0, m.wizard.step, "Back button should go to previous step")
}

func TestModel_WizardBackDisabledAtStep0(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.Equal(t, 0, m.wizard.step)

	// Navigate to Back button (step 0 — should be disabled)
	m.wizard.cursor = wizardBackIdx
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Equal(t, 0, m.wizard.step, "Back at step 0 should stay at step 0")
}

// ──── Summary View Tests ──────────────────────────────────────────────────

func TestModel_SummaryRenders(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode (2 steps)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 for step 0 — auto-advances to step 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 2 for step 1 — auto-advances to summary
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	m.isReady = true

	assert.True(t, m.wizard.showingSummary, "should show summary")

	view := m.View()
	assert.Contains(t, view, "Summary", "summary title should appear")
	assert.Contains(t, view, "Claude Code", "agent name should appear")
	assert.Contains(t, view, "OpenCode", "agent name should appear")
	assert.Contains(t, view, "Uninstall app only", "first agent's action should appear")
	assert.Contains(t, view, "Keep installed", "second agent's action should appear")
	assert.Contains(t, view, "Apply", "Apply button should appear")
	assert.Contains(t, view, "[ ◄ Back ]", "Back button should appear")
}

func TestModel_SummaryApplySubmitsAndQuits(t *testing.T) {
	m := newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 — auto-advances to summary
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.True(t, m.wizard.showingSummary)

	// Enter on Apply (cursor=0 in summary) should complete and submit immediately
	mResult, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.Nil(t, m.wizard, "wizard should be nil after Apply on summary")
	assert.True(t, m.isSubmitted, "should be submitted immediately")
	assert.NotNil(t, m.wizardOut, "wizardOut should be set")
	assert.NotNil(t, cmd, "should return tea.Quit command")
}

func TestModel_SummaryBackReturnsToLastStep(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code and opencode (2 steps)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 for step 0 — auto-advances to step 1
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 2 for step 1 — auto-advances to summary
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.True(t, m.wizard.showingSummary)
	assert.Equal(t, 0, m.wizard.cursor, "summary cursor starts at Apply")

	// Navigate to Back and press Enter
	m = updateModel(m, "j") // cursor → Back
	assert.Equal(t, 1, m.wizard.cursor)
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	assert.False(t, m.wizard.showingSummary, "should return to step view")
	assert.Equal(t, 1, m.wizard.step, "should return to last step (index 1)")
	assert.Equal(t, wizardRadio0, m.wizard.cursor, "cursor should reset to first radio")
}

func TestModel_SummaryQQuits(t *testing.T) {
	m := 	newModel(agentsForWizard(), "0.1.0")

	// Deselect claude-code
	m = updateModel(m, "j")
	m = updateModelKey(m, tea.KeySpace)

	// Start wizard
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	m = updateModel(m, "j")
	mResult, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)

	// Select option 0 — auto-advances to summary
	mResult, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mResult.(model)
	assert.True(t, m.wizard.showingSummary)

	// q should cancel from summary
	m = updateModel(m, "q")
	assert.Nil(t, m.wizard, "q should cancel wizard from summary")
	assert.False(t, m.isSubmitted, "q should not submit")
}
