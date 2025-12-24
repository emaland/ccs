package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/session"
)

var statusCmd = &cobra.Command{
	Use:   "status [name]",
	Short: "Show detailed status of a session",
	Long:  `Show detailed status of a session. Defaults to current session.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var sess *session.Session
		var err error

		if len(args) > 0 {
			sess, err = sessMgr.Get(args[0])
		} else {
			sess, err = sessMgr.GetCurrent()
		}
		if err != nil {
			return err
		}

		status, err := sessMgr.GetStatus(sess)
		if err != nil {
			return err
		}

		// Get more details
		wtGit := gitRepo.InWorktree(sess.Path)
		mergeBase, _ := wtGit.MergeBase(cfg.DefaultBase, "HEAD")
		commitCount, _ := wtGit.CommitCount(mergeBase, "HEAD")

		fmt.Printf("Session: %s\n", sess.Name)
		fmt.Printf("Branch:  %s (based on %s, %d commits ahead)\n",
			sess.Branch, cfg.DefaultBase, commitCount)
		fmt.Printf("Path:    %s\n", sess.Path)
		fmt.Println()

		// Show files changed
		files, err := wtGit.DiffFiles(mergeBase, "HEAD")
		if err == nil && len(files) > 0 {
			fmt.Println("Files changed:")
			for _, f := range files {
				fmt.Printf("  %s %s\n", f.Status, f.Path)
			}
			fmt.Println()
		}

		// Show Claude status
		fmt.Printf("Claude: %s\n", status.ClaudeState)

		return nil
	},
}
