package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/claude"
	"github.com/emaland/ccs/internal/session"
)

var pauseAll bool

var pauseCmd = &cobra.Command{
	Use:   "pause [name]",
	Short: "Pause a session (stop Claude)",
	Long:  `Stop the Claude process for a session while keeping the worktree.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if pauseAll {
			sessions, err := sessMgr.List()
			if err != nil {
				return err
			}
			for _, sess := range sessions {
				if err := claude.StopProcess(sess.Path); err != nil {
					fmt.Printf("Warning: could not stop Claude for %s: %v\n", sess.Name, err)
				} else {
					fmt.Printf("Paused %s\n", sess.Name)
				}
			}
			return nil
		}

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

		if err := claude.StopProcess(sess.Path); err != nil {
			return fmt.Errorf("could not stop Claude: %w", err)
		}

		fmt.Printf("Paused %s\n", sess.Name)
		return nil
	},
}

func init() {
	pauseCmd.Flags().BoolVar(&pauseAll, "all", false, "Pause all sessions")
}
