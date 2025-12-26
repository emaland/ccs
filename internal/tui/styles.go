package tui

import "github.com/charmbracelet/lipgloss"

// Colors - subtle, works on light and dark terminals
var (
	colorPrimary   = lipgloss.Color("12")  // Blue
	colorSecondary = lipgloss.Color("244") // Gray
	colorSuccess   = lipgloss.Color("10")  // Green
	colorWarning   = lipgloss.Color("11")  // Yellow
	colorError     = lipgloss.Color("9")   // Red
	colorMuted     = lipgloss.Color("240") // Dim gray
)

// Styles
var (
	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)

	// Help text in title bar
	helpHintStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	// Table styles
	tableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(colorSecondary).
				BorderBottom(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(colorMuted)

	tableSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("255"))

	tableRowStyle = lipgloss.NewStyle()

	// Claude state styles
	claudeRunningStyle = lipgloss.NewStyle().Foreground(colorSuccess)
	claudeWaitingStyle = lipgloss.NewStyle().Foreground(colorWarning)
	claudeIdleStyle    = lipgloss.NewStyle().Foreground(colorMuted)
	claudePausedStyle  = lipgloss.NewStyle().Foreground(colorMuted)

	// Status bar at bottom
	statusBarStyle = lipgloss.NewStyle().
			Foreground(colorSecondary).
			BorderTop(true).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(colorMuted)

	// Key hint style (e.g., "[n]ew")
	keyStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)

	keyDescStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(colorError)

	// Input styles
	inputLabelStyle = lipgloss.NewStyle().
			Foreground(colorSecondary)

	inputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(0, 1)

	// Modal/overlay styles
	modalStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(colorPrimary).
			Padding(1, 2)

	modalTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			MarginBottom(1)

	// Diff view styles
	diffAddStyle    = lipgloss.NewStyle().Foreground(colorSuccess)
	diffRemoveStyle = lipgloss.NewStyle().Foreground(colorError)
	diffHeaderStyle = lipgloss.NewStyle().Foreground(colorPrimary).Bold(true)

	// Selection cursor
	cursorStyle = lipgloss.NewStyle().
			Foreground(colorPrimary).
			Bold(true)
)

// Helper to style Claude state
func styledClaudeState(state string) string {
	switch state {
	case "running":
		return claudeRunningStyle.Render(state)
	case "waiting":
		return claudeWaitingStyle.Render(state)
	case "idle":
		return claudeIdleStyle.Render(state)
	case "paused":
		return claudePausedStyle.Render(state)
	default:
		return claudeIdleStyle.Render(state)
	}
}

// Helper to render a key hint like "[n]ew"
func keyHint(key, desc string) string {
	return keyStyle.Render("["+key+"]") + keyDescStyle.Render(desc)
}
