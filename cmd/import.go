package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/state"
)

var importCmd = &cobra.Command{
	Use:   "import [path...]",
	Short: "Import existing worktrees into CCS state",
	Long: `Scan directories for git worktrees and import them into CCS global state.

By default, scans ~/scratch/git/*/ for repositories and imports any
worktrees with ccs/ prefix branches.

You can specify paths to scan:
  ccs import ~/projects/*/`,
	RunE: func(cmd *cobra.Command, args []string) error {
		stateMgr, err := state.NewManager()
		if err != nil {
			return fmt.Errorf("could not initialize state manager: %w", err)
		}

		// Default to ~/scratch/git/*/
		paths := args
		if len(paths) == 0 {
			home, _ := os.UserHomeDir()
			defaultPath := filepath.Join(home, "scratch", "git", "*")
			matches, _ := filepath.Glob(defaultPath)
			paths = matches
		}

		var imported int
		for _, path := range paths {
			info, err := os.Stat(path)
			if err != nil || !info.IsDir() {
				continue
			}

			// Check if it's a git repo
			gitDir := filepath.Join(path, ".git")
			if _, err := os.Stat(gitDir); os.IsNotExist(err) {
				continue
			}

			// Get worktrees for this repo
			out, err := exec.Command("git", "-C", path, "worktree", "list", "--porcelain").Output()
			if err != nil {
				continue
			}

			worktrees := parseWorktreeList(string(out))
			repoName := filepath.Base(path)

			for _, wt := range worktrees {
				// Only import worktrees with ccs/ prefix branches
				if !strings.HasPrefix(wt.branch, "ccs/") {
					continue
				}

				// Skip if already in state
				if stateMgr.GetSession(wt.path) != nil {
					continue
				}

				name := strings.TrimPrefix(wt.branch, "ccs/")
				stateMgr.AddSession(state.SessionState{
					Name:       name,
					RepoPath:   path,
					RepoName:   repoName,
					WorkTree:   wt.path,
					Branch:     wt.branch,
					BaseBranch: "main", // Assume main, we can't easily determine original base
					CreatedAt:  time.Now(),
					LastAccess: time.Now(),
				})
				imported++
				fmt.Printf("Imported: %s/%s (%s)\n", repoName, name, wt.path)
			}
		}

		if imported == 0 {
			fmt.Println("No new sessions found to import.")
		} else {
			fmt.Printf("\nImported %d session(s).\n", imported)
		}

		return nil
	},
}

type worktreeInfo struct {
	path   string
	branch string
}

func parseWorktreeList(output string) []worktreeInfo {
	var worktrees []worktreeInfo
	var current worktreeInfo

	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "worktree ") {
			if current.path != "" {
				worktrees = append(worktrees, current)
			}
			current = worktreeInfo{path: strings.TrimPrefix(line, "worktree ")}
		} else if strings.HasPrefix(line, "branch ") {
			branch := strings.TrimPrefix(line, "branch refs/heads/")
			current.branch = branch
		}
	}
	if current.path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// Not exposed as a command - use internally when needed
// func init() {
// 	rootCmd.AddCommand(importCmd)
// }
