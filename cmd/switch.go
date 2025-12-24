package cmd

import (
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
		if current, err := sessMgr.GetCurrent(); err == nil {
			previousSession = current.Name
		}

		return sessMgr.Switch(name)
	},
}
