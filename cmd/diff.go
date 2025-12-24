package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/session"
)

var diffCmd = &cobra.Command{
	Use:   "diff [name] [-- git-diff-args]",
	Short: "Show diff for a session",
	Long: `Show git diff for a session against its base.

Defaults to current session. Supports standard git diff flags after --.`,
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

		output, err := wtGit.DiffRaw(mergeBase, "HEAD", gitArgs...)
		if err != nil {
			return err
		}

		fmt.Print(output)
		return nil
	},
}
