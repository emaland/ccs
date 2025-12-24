package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/session"
)

var logCmd = &cobra.Command{
	Use:   "log [name] [-- git-log-args]",
	Short: "Show log for a session",
	Long: `Show git log for a session since its base.

Defaults to current session. Supports standard git log flags after --.`,
	Args:               cobra.ArbitraryArgs,
	DisableFlagParsing: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var sessionName string
		var gitArgs []string

		// Parse args: [session-name] [-- git-args...]
		for i, arg := range args {
			if arg == "--" {
				gitArgs = args[i+1:]
				break
			}
			if sessionName == "" && !strings.HasPrefix(arg, "-") {
				sessionName = arg
			} else {
				gitArgs = append(gitArgs, arg)
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

		wtGit := gitRepo.InWorktree(sess.Path)
		mergeBase, err := wtGit.MergeBase(cfg.DefaultBase, "HEAD")
		if err != nil {
			mergeBase = cfg.DefaultBase
		}

		output, err := wtGit.Log(mergeBase, "HEAD", gitArgs...)
		if err != nil {
			return err
		}

		fmt.Print(output)
		return nil
	},
}
