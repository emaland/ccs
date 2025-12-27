package cmd

import (
	"fmt"
	"os"
	"os/exec"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/tui"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Open interactive TUI dashboard",
	Long: `Launch an interactive terminal UI for managing Claude Code sessions.

The TUI provides a dashboard view of all sessions with keyboard navigation
and actions for creating, switching, pausing, resuming, and finishing sessions.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if stateMgr == nil {
			return fmt.Errorf("state manager not available")
		}

		// Rename current tab to "CCS UI"
		if os.Getenv("KITTY_WINDOW_ID") != "" {
			exec.Command("kitty", "@", "set-tab-title", "CCS UI").Run()
		} else if os.Getenv("TMUX") != "" {
			exec.Command("tmux", "rename-window", "CCS UI").Run()
		}

		// Get repo path if available (for creating new sessions)
		repoPath := ""
		if gitRepo != nil {
			repoPath = gitRepo.RepoRoot()
		}

		model := tui.New(cfg, sessMgr, stateMgr, repoPath)

		p := tea.NewProgram(model, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		// Check if we should switch to a session
		if m, ok := finalModel.(tui.Model); ok {
			if switchTo := m.SwitchToSession(); switchTo != "" {
				// Look up session path in global state
				sessions := stateMgr.GetAllSessions()
				for _, sess := range sessions {
					if sess.Name == switchTo {
						// Try to switch terminal window
						if term != nil && term.Name() != "none" {
							if err := term.SwitchWindow(switchTo); err == nil {
								return nil
							}
						}
						// Otherwise print cd command
						fmt.Printf("cd %s\n", sess.WorkTree)
						return nil
					}
				}
				return fmt.Errorf("session %q not found", switchTo)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
