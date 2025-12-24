package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/session"
)

var (
	finishSquash bool
	finishMerge  bool
	finishPR     bool
	finishDelete bool
	finishForce  bool
)

var finishCmd = &cobra.Command{
	Use:   "finish <name>",
	Short: "Finish a session",
	Long: `Finish a session by merging, creating a PR, or deleting.

Without flags, shows an interactive menu.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		// Check if any action was specified
		hasAction := finishSquash || finishMerge || finishPR || finishDelete

		if !hasAction {
			// Interactive mode
			return interactiveFinish(name)
		}

		opts := session.FinishOptions{
			Squash: finishSquash,
			Merge:  finishMerge,
			PR:     finishPR,
			Delete: finishDelete,
			Force:  finishForce,
		}

		return sessMgr.Finish(name, opts)
	},
}

func interactiveFinish(name string) error {
	sess, err := sessMgr.Get(name)
	if err != nil {
		return err
	}

	status, _ := sessMgr.GetStatus(sess)

	fmt.Printf("Session %s has %d files changed.\n\n", name, status.FilesChanged)
	fmt.Printf("[s] Squash and merge to %s\n", cfg.DefaultBase)
	fmt.Printf("[m] Merge to %s (keep commits)\n", cfg.DefaultBase)
	fmt.Println("[p] Push branch for PR")
	fmt.Println("[d] Delete without merging")
	fmt.Println("[c] Cancel")
	fmt.Println()
	fmt.Print("Choice: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(strings.ToLower(choice))

	var opts session.FinishOptions
	switch choice {
	case "s":
		opts.Squash = true
	case "m":
		opts.Merge = true
	case "p":
		opts.PR = true
	case "d":
		opts.Delete = true
	case "c":
		fmt.Println("Cancelled.")
		return nil
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	return sessMgr.Finish(name, opts)
}

func init() {
	finishCmd.Flags().BoolVar(&finishSquash, "squash", false, "Squash all commits and merge to base")
	finishCmd.Flags().BoolVar(&finishMerge, "merge", false, "Merge to base (keep commits)")
	finishCmd.Flags().BoolVar(&finishPR, "pr", false, "Push branch for PR, don't merge locally")
	finishCmd.Flags().BoolVar(&finishDelete, "delete", false, "Delete without merging")
	finishCmd.Flags().BoolVar(&finishForce, "force", false, "Skip confirmation and hooks")
}
