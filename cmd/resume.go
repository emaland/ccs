package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/claude"
	"github.com/emaland/ccs/internal/session"
)

var resumeCmd = &cobra.Command{
	Use:   "resume [name] [-- claude-args...]",
	Short: "Resume a session (restart Claude)",
	Long: `Restart Claude for a paused session.

Any arguments after -- are passed to Claude:
  ccs resume my-feature -- --continue
  ccs resume -- --dangerously-skip-permissions`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var sessionName string
		var claudeArgs []string

		// Parse args: [session-name] [claude-args...]
		for i, arg := range args {
			if i == 0 && !startsWithDash(arg) {
				sessionName = arg
			} else {
				claudeArgs = append(claudeArgs, arg)
			}
		}

		var sess *session.Session
		var err error

		if sessionName != "" {
			sess, err = sessMgr.Get(sessionName)
		} else {
			sess, err = sessMgr.GetCurrent()
		}
		if err != nil {
			return err
		}

		// Check if already running
		state := claude.GetState(sess.Path)
		if state == claude.StateRunning || state == claude.StateWaiting {
			return fmt.Errorf("Claude is already running for %s", sess.Name)
		}

		if err := claude.StartProcess(sess.Path, claudeArgs); err != nil {
			return fmt.Errorf("could not start Claude: %w", err)
		}

		if len(claudeArgs) > 0 {
			fmt.Printf("Resumed %s with args: %v\n", sess.Name, claudeArgs)
		} else {
			fmt.Printf("Resumed %s\n", sess.Name)
		}
		return nil
	},
}

func startsWithDash(s string) bool {
	return len(s) > 0 && s[0] == '-'
}
