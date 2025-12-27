package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// updateViewport handles updates for viewport-based views (diff, log)
func (m Model) updateViewport(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Back):
			m.view = viewDashboard
			return m, nil
		}
	}

	// Update viewport
	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// viewDiff renders the diff view
func (m Model) viewDiff() string {
	var b strings.Builder

	// Title - show both session name and repo
	sessName := ""
	if len(m.sessions) > 0 && m.selected < len(m.sessions) {
		sess := m.sessions[m.selected]
		sessName = fmt.Sprintf("%s (%s)", sess.Name, sess.RepoName)
	}

	title := diffHeaderStyle.Render(fmt.Sprintf("Diff: %s", sessName))
	helpHint := helpHintStyle.Render("[q] back")
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(helpHint))),
		helpHint,
	)
	b.WriteString(titleBar + "\n\n")

	// Colorize diff content
	content := m.viewport.View()
	b.WriteString(colorizeDiff(content))

	// Scroll info
	b.WriteString(fmt.Sprintf("\n%3.f%%", m.viewport.ScrollPercent()*100))

	return b.String()
}

// viewLog renders the log view
func (m Model) viewLog() string {
	var b strings.Builder

	// Title - show both session name and repo
	sessName := ""
	if len(m.sessions) > 0 && m.selected < len(m.sessions) {
		sess := m.sessions[m.selected]
		sessName = fmt.Sprintf("%s (%s)", sess.Name, sess.RepoName)
	}

	title := diffHeaderStyle.Render(fmt.Sprintf("Log: %s", sessName))
	helpHint := helpHintStyle.Render("[q] back")
	titleBar := lipgloss.JoinHorizontal(
		lipgloss.Top,
		title,
		strings.Repeat(" ", max(0, m.width-lipgloss.Width(title)-lipgloss.Width(helpHint))),
		helpHint,
	)
	b.WriteString(titleBar + "\n\n")

	b.WriteString(m.viewport.View())

	// Scroll info
	b.WriteString(fmt.Sprintf("\n%3.f%%", m.viewport.ScrollPercent()*100))

	return b.String()
}

// colorizeDiff adds basic coloring to diff output
func colorizeDiff(content string) string {
	var result strings.Builder
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		switch {
		case strings.HasPrefix(line, "+") && !strings.HasPrefix(line, "+++"):
			result.WriteString(diffAddStyle.Render(line))
		case strings.HasPrefix(line, "-") && !strings.HasPrefix(line, "---"):
			result.WriteString(diffRemoveStyle.Render(line))
		case strings.HasPrefix(line, "@@"):
			result.WriteString(diffHeaderStyle.Render(line))
		case strings.HasPrefix(line, "diff ") || strings.HasPrefix(line, "index "):
			result.WriteString(diffHeaderStyle.Render(line))
		default:
			result.WriteString(line)
		}
		result.WriteString("\n")
	}

	return result.String()
}
