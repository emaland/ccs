package cmd

import (
	"fmt"

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
		if sessMgr == nil {
			return fmt.Errorf("not in a git repository")
		}

		model := tui.New(cfg, sessMgr, stateMgr, gitRepo.RepoRoot())

		p := tea.NewProgram(model, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			return err
		}

		// Check if we should switch to a session
		if m, ok := finalModel.(tui.Model); ok {
			if switchTo := m.SwitchToSession(); switchTo != "" {
				// Switch to the session
				return sessMgr.Switch(switchTo)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(uiCmd)
}
