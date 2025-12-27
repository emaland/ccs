package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var previousSession string // Stores last session for "-" support

var switchCmd = &cobra.Command{
	Use:   "switch <name>",
	Short: "Switch to a session",
	Long: `Switch to a session. Use "-" for the previous session.

If running in tmux/kitty, switches to that window/tab.
Otherwise, prints "cd <path>" for shell integration.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Handle "-" for previous session
		if name == "-" {
			if previousSession == "" {
				return nil
			}
			name = previousSession
		}

		// Store current as previous before switching
		if sessMgr != nil {
			if current, err := sessMgr.GetCurrent(); err == nil {
				previousSession = current.Name
			}
		}

		// Look up session in global state
		if stateMgr == nil {
			return fmt.Errorf("state manager not available")
		}

		sessions := stateMgr.GetAllSessions()
		var sessionPath string
		for _, sess := range sessions {
			if sess.Name == name {
				sessionPath = sess.WorkTree
				break
			}
		}

		if sessionPath == "" {
			return fmt.Errorf("session %q not found", name)
		}

		// If in a terminal, switch window
		if term != nil && term.Name() != "none" {
			if err := term.SwitchWindow(name); err != nil {
				// Window switch failed - show error and fall back to cd
				fmt.Fprintf(os.Stderr, "ccs: could not switch %s window: %v\n", term.Name(), err)
			} else {
				return nil
			}
		}

		// Fall back to print cd command for shell integration
		fmt.Printf("cd %s\n", sessionPath)
		return nil
	},
}
