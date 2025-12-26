package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/emaland/ccs/internal/claude"
)

var (
	sessionsJSON bool
)

var sessionsCmd = &cobra.Command{
	Use:   "sessions",
	Short: "List all sessions across all repositories",
	Long: `List all CCS sessions globally, across all repositories.

This shows sessions tracked in the global state, regardless of which
repository you're currently in.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if stateMgr == nil {
			return fmt.Errorf("state manager not initialized")
		}

		sessions := stateMgr.GetAllSessions()

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			fmt.Println("\nRun 'ccs new <name>' to create a session.")
			return nil
		}

		if sessionsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(sessions)
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintf(w, "REPO\tSESSION\tBRANCH\tPATH\tSTATUS\n")

		for _, s := range sessions {
			// Check if worktree still exists
			status := "ok"
			if _, err := os.Stat(s.WorkTree); os.IsNotExist(err) {
				status = "missing"
			} else {
				// Check claude state
				claudeState := claude.GetState(s.WorkTree)
				if claudeState != claude.StateUnknown {
					status = string(claudeState)
				}
			}

			// Shorten path for display
			path := s.WorkTree
			if home, err := os.UserHomeDir(); err == nil {
				if rel, err := filepath.Rel(home, path); err == nil && len(rel) < len(path) {
					path = "~/" + rel
				}
			}

			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				s.RepoName,
				s.Name,
				s.Branch,
				path,
				status,
			)
		}

		return w.Flush()
	},
}

func init() {
	sessionsCmd.Flags().BoolVar(&sessionsJSON, "json", false, "Output as JSON")
	rootCmd.AddCommand(sessionsCmd)
}
