package tui

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/emaland/ccs/internal/claude"
)

// updateDashboard handles updates for the dashboard view
func (m Model) updateDashboard(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Clear error on any keypress
		m.err = nil

		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, m.keys.Help):
			m.showHelp = !m.showHelp

		case key.Matches(msg, m.keys.Up):
			if m.selected > 0 {
				m.selected--
			}

		case key.Matches(msg, m.keys.Down):
			if m.selected < len(m.sessions)-1 {
				m.selected++
			}

		case key.Matches(msg, m.keys.Select):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				m.switchToSession = sess.Name
				return m, tea.Quit
			}

		case key.Matches(msg, m.keys.New):
			m.view = viewNewSession
			m.textinput.Focus()
			m.textinput.Reset()
			return m, nil

		case key.Matches(msg, m.keys.Resume):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				cmds = append(cmds, m.resumeSession(sess.Name))
			}

		case key.Matches(msg, m.keys.Pause):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				cmds = append(cmds, m.pauseSession(sess.Name))
			}

		case key.Matches(msg, m.keys.PauseAll):
			cmds = append(cmds, m.pauseAllSessions())

		case key.Matches(msg, m.keys.Diff):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				m.view = viewDiff
				cmds = append(cmds, m.loadDiff(sess.Session.Path))
			}

		case key.Matches(msg, m.keys.Log):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				m.view = viewLog
				cmds = append(cmds, m.loadLog(sess.Session.Path))
			}

		case key.Matches(msg, m.keys.Finish):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				m.finishSession = &sess
				m.finishAction = 0
				m.view = viewFinish
			}
		}
	}

	return m, tea.Batch(cmds...)
}

// viewDashboard renders the dashboard view
func (m Model) viewDashboard() string {
	var b strings.Builder

	// Title bar
	title := titleStyle.Render("CCS - Claude Code Sessions")
	helpHint := helpHintStyle.Render("[?] help  [q] quit")
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(helpHint))),
		helpHint,
	)
	b.WriteString(titleBar + "\n\n")

	// Loading state
	if m.loading && len(m.sessions) == 0 {
		b.WriteString(m.spinner.View() + " Loading sessions...\n")
		return b.String()
	}

	// No sessions
	if len(m.sessions) == 0 {
		b.WriteString("No sessions found.\n\n")
		b.WriteString("Press " + keyHint("n", "ew") + " to create a session.\n")
		return b.String()
	}

	// Table header
	header := fmt.Sprintf("  %-18s %-12s %-8s %-10s %s",
		"NAME", "STATUS", "FILES", "CLAUDE", "PATH")
	b.WriteString(tableHeaderStyle.Render(header) + "\n")

	// Table rows
	for i, sess := range m.sessions {
		cursor := "  "
		rowStyle := tableRowStyle
		if i == m.selected {
			cursor = cursorStyle.Render("► ")
			rowStyle = tableSelectedStyle
		}

		// Status (commits ahead or uncommitted)
		status := ""
		if sess.Status != nil {
			if sess.Status.CommitsAhead > 0 {
				status = fmt.Sprintf("%d ahead", sess.Status.CommitsAhead)
			}
			if sess.Status.FilesChanged > 0 && sess.Status.CommitsAhead == 0 {
				status = "uncommitted"
			}
		}

		// Files changed
		files := ""
		if sess.Status != nil && sess.Status.FilesChanged > 0 {
			files = fmt.Sprintf("%d", sess.Status.FilesChanged)
		}

		// Claude state
		claudeState := "unknown"
		if sess.Status != nil {
			claudeState = string(sess.Status.ClaudeState)
		}

		// Path (shortened)
		path := sess.Session.Path
		if len(path) > 30 {
			path = "..." + path[len(path)-27:]
		}

		row := fmt.Sprintf("%-18s %-12s %-8s %-10s %s",
			truncate(sess.Name, 18),
			truncate(status, 12),
			files,
			styledClaudeState(claudeState),
			path,
		)

		b.WriteString(cursor + rowStyle.Render(row) + "\n")
	}

	// Error display
	if m.err != nil {
		b.WriteString("\n" + errorStyle.Render("Error: "+m.err.Error()) + "\n")
	}

	// Status bar
	b.WriteString("\n")
	statusBar := fmt.Sprintf("%s  %s  %s  %s  %s  %s  %s  %s",
		keyHint("n", "ew"),
		keyHint("enter", " switch"),
		keyHint("r", "esume"),
		keyHint("p", "ause"),
		keyHint("d", "iff"),
		keyHint("l", "og"),
		keyHint("f", "inish"),
		keyHint("/", "filter"),
	)
	b.WriteString(statusBarStyle.Render(statusBar))

	return b.String()
}

// Helper commands
func (m Model) resumeSession(name string) tea.Cmd {
	return func() tea.Msg {
		sess, err := m.sessMgr.Get(name)
		if err != nil {
			return errorMsg(err)
		}
		if err := claude.StartProcess(sess.Path, []string{"--continue"}); err != nil {
			return errorMsg(err)
		}
		return sessionsLoadedMsg(nil) // Trigger refresh
	}
}

func (m Model) pauseSession(name string) tea.Cmd {
	return func() tea.Msg {
		sess, err := m.sessMgr.Get(name)
		if err != nil {
			return errorMsg(err)
		}
		if err := claude.StopProcess(sess.Path); err != nil {
			return errorMsg(err)
		}
		return sessionsLoadedMsg(nil) // Trigger refresh
	}
}

func (m Model) pauseAllSessions() tea.Cmd {
	return func() tea.Msg {
		sessions, err := m.sessMgr.List()
		if err != nil {
			return errorMsg(err)
		}
		for _, sess := range sessions {
			claude.StopProcess(sess.Path)
		}
		return sessionsLoadedMsg(nil) // Trigger refresh
	}
}

func (m Model) loadDiff(path string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "diff", m.cfg.DefaultBase+"..HEAD")
		cmd.Dir = path
		out, err := cmd.Output()
		if err != nil {
			return errorMsg(err)
		}
		if len(out) == 0 {
			return diffLoadedMsg("No changes")
		}
		return diffLoadedMsg(string(out))
	}
}

func (m Model) loadLog(path string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("git", "log", "--oneline", m.cfg.DefaultBase+"..HEAD")
		cmd.Dir = path
		out, err := cmd.Output()
		if err != nil {
			return errorMsg(err)
		}
		if len(out) == 0 {
			return logLoadedMsg("No commits")
		}
		return logLoadedMsg(string(out))
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
