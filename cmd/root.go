package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/config"
	"github.com/emaland/ccs/internal/git"
	"github.com/emaland/ccs/internal/session"
	"github.com/emaland/ccs/internal/state"
	"github.com/emaland/ccs/internal/terminal"
)

var (
	cfg      *config.Config
	gitRepo  git.Git
	term     terminal.Terminal
	stateMgr *state.Manager
	sessMgr  *session.Manager
	rootCmd  = &cobra.Command{
		Use:   "ccs",
		Short: "Claude Code Sessions - manage multiple concurrent Claude sessions",
		Long: `CCS manages multiple concurrent Claude Code sessions using git worktrees.

Each session gets its own git worktree and branch, providing real isolation
while sharing the git object store.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip initialization for commands that don't need it
			if cmd.Name() == "help" || cmd.Name() == "version" {
				return nil
			}

			var err error
			cfg, err = config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Initialize state manager (global, not repo-specific)
			stateMgr, err = state.NewManager()
			if err != nil {
				return fmt.Errorf("failed to initialize state: %w", err)
			}

			// Try to find git repo root
			repoRoot, err := git.FindRepoRoot(".")
			if err != nil {
				// Some commands work without a repo (using global state)
				if cmd.Name() == "shell-init" || cmd.Name() == "sessions" || cmd.Name() == "cleanup" || cmd.Name() == "switch" || cmd.Name() == "ui" {
					term = terminal.Detect(cfg)
					return nil
				}
				return fmt.Errorf("not in a git repository")
			}

			gitRepo, err = git.NewExecGit(repoRoot)
			if err != nil {
				return fmt.Errorf("failed to initialize git: %w", err)
			}

			term = terminal.Detect(cfg)
			sessMgr = session.NewManager(cfg, gitRepo, term, stateMgr)

			return nil
		},
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(switchCmd)
	rootCmd.AddCommand(statusCmd)
	rootCmd.AddCommand(finishCmd)
	rootCmd.AddCommand(diffCmd)
	rootCmd.AddCommand(logCmd)
	rootCmd.AddCommand(pauseCmd)
	rootCmd.AddCommand(resumeCmd)
	rootCmd.AddCommand(hooksCmd)
	rootCmd.AddCommand(shellInitCmd)
	rootCmd.AddCommand(currentSessionCmd)
	rootCmd.AddCommand(previousSessionCmd)
	rootCmd.AddCommand(sessionPathCmd)
}

func printError(err error) {
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
}
