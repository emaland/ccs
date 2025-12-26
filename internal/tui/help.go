package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// viewHelp renders the help overlay
func (m Model) viewHelp() string {
	var b strings.Builder

	helpStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(colorPrimary).
		Padding(1, 2)

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(colorPrimary).
		MarginTop(1)

	keyColStyle := lipgloss.NewStyle().
		Width(14).
		Foreground(colorSecondary)

	descStyle := lipgloss.NewStyle()

	b.WriteString(modalTitleStyle.Render("Keyboard Shortcuts") + "\n")
	b.WriteString(strings.Repeat("─", 35) + "\n")

	// Navigation
	b.WriteString(sectionStyle.Render("Navigation") + "\n")
	b.WriteString(keyColStyle.Render("  j/k, ↑/↓") + descStyle.Render("Move selection") + "\n")
	b.WriteString(keyColStyle.Render("  enter") + descStyle.Render("Switch to session") + "\n")
	b.WriteString(keyColStyle.Render("  /") + descStyle.Render("Filter sessions") + "\n")

	// Actions
	b.WriteString(sectionStyle.Render("Actions") + "\n")
	b.WriteString(keyColStyle.Render("  n") + descStyle.Render("New session") + "\n")
	b.WriteString(keyColStyle.Render("  r") + descStyle.Render("Resume session") + "\n")
	b.WriteString(keyColStyle.Render("  p") + descStyle.Render("Pause session") + "\n")
	b.WriteString(keyColStyle.Render("  P") + descStyle.Render("Pause all") + "\n")
	b.WriteString(keyColStyle.Render("  f") + descStyle.Render("Finish session") + "\n")

	// Views
	b.WriteString(sectionStyle.Render("Views") + "\n")
	b.WriteString(keyColStyle.Render("  d") + descStyle.Render("Show diff") + "\n")
	b.WriteString(keyColStyle.Render("  l") + descStyle.Render("Show log") + "\n")
	b.WriteString(keyColStyle.Render("  s") + descStyle.Render("Show status") + "\n")

	// General
	b.WriteString(sectionStyle.Render("General") + "\n")
	b.WriteString(keyColStyle.Render("  ?") + descStyle.Render("Toggle help") + "\n")
	b.WriteString(keyColStyle.Render("  q") + descStyle.Render("Quit") + "\n")

	b.WriteString("\n" + helpHintStyle.Render("Press ? to close"))

	return helpStyle.Render(b.String())
}
