package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove stale sessions from global state",
	Long: `Remove sessions from global state whose worktrees no longer exist.

This cleans up the state file after worktrees have been manually deleted
or lost (e.g., after a system restore).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if stateMgr == nil {
			return fmt.Errorf("state manager not initialized")
		}

		removed, err := stateMgr.Cleanup()
		if err != nil {
			return fmt.Errorf("cleanup failed: %w", err)
		}

		if len(removed) == 0 {
			fmt.Println("No stale sessions found.")
			return nil
		}

		fmt.Printf("Removed %d stale session(s):\n", len(removed))
		for _, s := range removed {
			fmt.Printf("  - %s/%s (%s)\n", s.RepoName, s.Name, s.WorkTree)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
}
