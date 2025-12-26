package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/emaland/ccs/internal/session"
)

var finishActions = []struct {
	name string
	desc string
}{
	{"squash", "Squash merge to main"},
	{"merge", "Merge to main (keep commits)"},
	{"pr", "Push for PR"},
	{"delete", "Delete without merging"},
}

// updateFinish handles updates for the finish action selection
func (m Model) updateFinish(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.view = viewDashboard
			m.finishSession = nil
			return m, nil

		case key.Matches(msg, m.keys.Up):
			if m.finishAction > 0 {
				m.finishAction--
			}

		case key.Matches(msg, m.keys.Down):
			if m.finishAction < len(finishActions)-1 {
				m.finishAction++
			}

		case key.Matches(msg, m.keys.Select):
			m.view = viewFinishConfirm
			return m, nil
		}
	}

	return m, tea.Batch(cmds...)
}

// viewFinish renders the finish action selection
func (m Model) viewFinish() string {
	var b strings.Builder

	sessName := ""
	if m.finishSession != nil {
		sessName = m.finishSession.Name
	}

	b.WriteString(modalTitleStyle.Render(fmt.Sprintf("Finish: %s", sessName)) + "\n")
	b.WriteString(strings.Repeat("━", 30) + "\n\n")

	b.WriteString("Choose action:\n")

	for i, action := range finishActions {
		cursor := "  "
		if i == m.finishAction {
			cursor = cursorStyle.Render("► ")
		}
		b.WriteString(cursor + action.desc + "\n")
	}

	b.WriteString("\n" + keyHint("enter", " select") + "  " + keyHint("esc", " cancel") + "\n")

	return modalStyle.Render(b.String())
}

// updateFinishConfirm handles updates for the finish confirmation
func (m Model) updateFinishConfirm(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.view = viewFinish
			return m, nil

		case msg.String() == "n":
			m.view = viewFinish
			return m, nil

		case key.Matches(msg, m.keys.Confirm):
			if m.finishSession != nil {
				cmds = append(cmds, m.finishSessionCmd())
				m.loading = true
			}
			return m, tea.Batch(cmds...)
		}
	}

	return m, tea.Batch(cmds...)
}

// viewFinishConfirm renders the finish confirmation
func (m Model) viewFinishConfirm() string {
	var b strings.Builder

	if m.finishSession == nil {
		return ""
	}

	action := finishActions[m.finishAction]
	sessName := m.finishSession.Name

	b.WriteString(fmt.Sprintf("%s %s?\n\n", action.desc, sessName))

	b.WriteString("This will:\n")

	switch action.name {
	case "squash":
		commits := 0
		if m.finishSession.Status != nil {
			commits = m.finishSession.Status.CommitsAhead
		}
		b.WriteString(fmt.Sprintf("  • Squash %d commit(s) into one\n", commits))
		b.WriteString("  • Merge to " + m.cfg.DefaultBase + "\n")
		b.WriteString("  • Delete worktree and branch\n")
	case "merge":
		b.WriteString("  • Merge to " + m.cfg.DefaultBase + "\n")
		b.WriteString("  • Delete worktree and branch\n")
	case "pr":
		b.WriteString("  • Push branch to remote\n")
		b.WriteString("  • Delete worktree (keep branch)\n")
	case "delete":
		b.WriteString("  • Delete worktree and branch\n")
		b.WriteString("  • Discard all changes\n")
	}

	b.WriteString("\n")

	if m.loading {
		b.WriteString(m.spinner.View() + " Processing...\n")
	} else {
		b.WriteString(keyHint("y", " confirm") + "  " + keyHint("n", " cancel") + "\n")
	}

	return modalStyle.Render(b.String())
}

// finishSessionCmd executes the finish action
func (m Model) finishSessionCmd() tea.Cmd {
	return func() tea.Msg {
		if m.finishSession == nil {
			return errorMsg(fmt.Errorf("no session selected"))
		}

		action := finishActions[m.finishAction]
		opts := session.FinishOptions{
			Force: true, // Don't prompt again
		}

		switch action.name {
		case "squash":
			opts.Squash = true
		case "merge":
			opts.Merge = true
		case "pr":
			opts.PR = true
		case "delete":
			opts.Delete = true
		}

		if err := m.sessMgr.Finish(m.finishSession.Name, opts); err != nil {
			return errorMsg(err)
		}

		return sessionFinishedMsg{name: m.finishSession.Name}
	}
}
