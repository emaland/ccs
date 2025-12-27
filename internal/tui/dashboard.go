package tui

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

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
				// Switch terminal window without exiting TUI
				cmds = append(cmds, m.switchSession(sess.Name))
			}

		case key.Matches(msg, m.keys.New):
			m.view = viewNewSession
			m.textinput.Focus()
			m.textinput.Reset()
			return m, nil

		case key.Matches(msg, m.keys.Resume):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				cmds = append(cmds, m.resumeSession(sess.Name, sess.Path))
			}

		case key.Matches(msg, m.keys.Pause):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				cmds = append(cmds, m.pauseSession(sess.Path))
			}

		case key.Matches(msg, m.keys.PauseAll):
			cmds = append(cmds, m.pauseAllSessions())

		case key.Matches(msg, m.keys.Diff):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				m.view = viewDiff
				cmds = append(cmds, m.loadDiff(sess.Path, sess.BaseBranch))
			}

		case key.Matches(msg, m.keys.Log):
			if len(m.sessions) > 0 {
				sess := m.sessions[m.selected]
				m.view = viewLog
				cmds = append(cmds, m.loadLog(sess.Path, sess.BaseBranch))
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
	header := fmt.Sprintf("  %-18s %-14s %-12s %-8s %-10s",
		"NAME", "REPO", "STATUS", "FILES", "CLAUDE")
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

		row := fmt.Sprintf("%-18s %-14s %-12s %-8s %-10s",
			truncate(sess.Name, 18),
			truncate(sess.RepoName, 14),
			truncate(status, 12),
			files,
			styledClaudeState(claudeState),
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
func (m Model) switchSession(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		if os.Getenv("KITTY_WINDOW_ID") != "" {
			if err := exec.CommandContext(ctx, "kitty", "@", "focus-tab", "--match", "title:^"+name+"$").Run(); err != nil {
				return errorMsg(fmt.Errorf("kitty focus-tab failed: %w", err))
			}
		} else if os.Getenv("TMUX") != "" {
			if err := exec.CommandContext(ctx, "tmux", "select-window", "-t", name).Run(); err != nil {
				return errorMsg(fmt.Errorf("tmux select-window failed: %w", err))
			}
		}
		return nil
	}
}

func (m Model) resumeSession(name, path string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Try to send command to existing tab, or create new one
		if os.Getenv("KITTY_WINDOW_ID") != "" {
			// Check if tab exists by trying to focus it
			if exec.CommandContext(ctx, "kitty", "@", "focus-tab", "--match", "title:^"+name+"$").Run() == nil {
				// Tab exists - send the command
				cmd := `claude --continue` + "\n"
				if err := exec.CommandContext(ctx, "kitty", "@", "send-text", "--match", "title:^"+name+"$", cmd).Run(); err != nil {
					return errorMsg(fmt.Errorf("send-text failed: %w", err))
				}
			} else {
				// Create new tab and start Claude
				out, err := exec.CommandContext(ctx, "kitty", "@", "launch", "--type=tab", "--tab-title", name, "--cwd", path).CombinedOutput()
				if err != nil {
					return errorMsg(fmt.Errorf("launch failed: %w (output: %s)", err, string(out)))
				}
				// Wait for shell to initialize
				time.Sleep(200 * time.Millisecond)
				cmd := `PROMPT_COMMAND='[[ -z "$(jobs)" ]] && exit'; claude --continue` + "\n"
				// Use window ID if we got one, otherwise match by recent
				windowID := strings.TrimSpace(string(out))
				var sendErr error
				if windowID != "" {
					sendErr = exec.CommandContext(ctx, "kitty", "@", "send-text", "--match", "id:"+windowID, cmd).Run()
				} else {
					sendErr = exec.CommandContext(ctx, "kitty", "@", "send-text", "--match", "recent:0", cmd).Run()
				}
				if sendErr != nil {
					return errorMsg(fmt.Errorf("send-text failed (windowID=%q): %w", windowID, sendErr))
				}
			}
		} else if os.Getenv("TMUX") != "" {
			// Check if window exists
			if exec.CommandContext(ctx, "tmux", "select-window", "-t", name).Run() == nil {
				// Window exists - send the command
				if err := exec.CommandContext(ctx, "tmux", "send-keys", "-t", name, "claude --continue", "Enter").Run(); err != nil {
					return errorMsg(fmt.Errorf("tmux send-keys failed: %w", err))
				}
			} else {
				// Create new window and start Claude
				if err := exec.CommandContext(ctx, "tmux", "new-window", "-n", name, "-c", path, "claude --continue").Run(); err != nil {
					return errorMsg(fmt.Errorf("tmux new-window failed: %w", err))
				}
			}
		}
		return sessionsLoadedMsg(nil) // Trigger refresh
	}
}

func (m Model) pauseSession(path string) tea.Cmd {
	return func() tea.Msg {
		if err := claude.StopProcess(path); err != nil {
			return errorMsg(err)
		}
		return sessionsLoadedMsg(nil) // Trigger refresh
	}
}

func (m Model) pauseAllSessions() tea.Cmd {
	return func() tea.Msg {
		for _, sess := range m.sessions {
			claude.StopProcess(sess.Path)
		}
		return sessionsLoadedMsg(nil) // Trigger refresh
	}
}

func (m Model) loadDiff(path, baseBranch string) tea.Cmd {
	return func() tea.Msg {
		if baseBranch == "" {
			baseBranch = m.cfg.DefaultBase
		}
		cmd := exec.Command("git", "diff", baseBranch+"..HEAD")
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

func (m Model) loadLog(path, baseBranch string) tea.Cmd {
	return func() tea.Msg {
		if baseBranch == "" {
			baseBranch = m.cfg.DefaultBase
		}
		cmd := exec.Command("git", "log", "--oneline", baseBranch+"..HEAD")
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
