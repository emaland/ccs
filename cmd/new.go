package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/session"
)

var (
	newFrom       string
	newHere       bool
	newNoClaude   bool
	newNoTerminal bool
)

var newCmd = &cobra.Command{
	Use:   "new <name> [-- claude-args...]",
	Short: "Create a new session",
	Long: `Create a new session with its own git worktree and branch.

The session will be created from the current branch (or --from if specified).
A new terminal window/tab will be opened, and Claude will be started.

Any arguments after -- are passed to Claude:
  ccs new my-feature -- --dangerously-skip-permissions
  ccs new bugfix -- --continue --model sonnet`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Arguments after the session name are passed to Claude
		var claudeArgs []string
		if len(args) > 1 {
			claudeArgs = args[1:]
		}

		opts := session.CreateOptions{
			From:       newFrom,
			Here:       newHere,
			NoClaude:   newNoClaude,
			NoTerminal: newNoTerminal,
			ClaudeArgs: claudeArgs,
		}

		sess, err := sessMgr.Create(name, opts)
		if err != nil {
			return err
		}

		fmt.Printf("Created session %s\n", sess.Name)
		fmt.Printf("  Branch: %s (from %s)\n", sess.Branch, sess.BaseBranch)
		fmt.Printf("  Path:   %s\n", sess.Path)

		if !newNoClaude && cfg.AutoStartClaude {
			if len(claudeArgs) > 0 {
				fmt.Printf("Starting claude with args: %v\n", claudeArgs)
			} else {
				fmt.Println("Starting claude...")
			}
		}

		return nil
	},
}

func init() {
	newCmd.Flags().StringVar(&newFrom, "from", "", "Base branch/commit (default: current branch)")
	newCmd.Flags().BoolVar(&newHere, "here", false, "Create worktree in ./.worktrees/<name>")
	newCmd.Flags().BoolVar(&newNoClaude, "no-claude", false, "Don't start Claude after creation")
	newCmd.Flags().BoolVar(&newNoTerminal, "no-terminal", false, "Don't create terminal window/tab")
}
