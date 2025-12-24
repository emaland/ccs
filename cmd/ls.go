package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	lsVerbose bool
	lsRunning bool
	lsJSON    bool
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List sessions",
	Long:  `List all sessions for the current repository.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := sessMgr.List()
		if err != nil {
			return err
		}

		if len(sessions) == 0 {
			fmt.Println("No sessions found.")
			fmt.Println("Run 'ccs new <name>' to create one.")
			return nil
		}

		// Get current session for marking
		currentSession, _ := sessMgr.GetCurrent()

		type sessionOutput struct {
			Name         string `json:"name"`
			Branch       string `json:"branch"`
			Path         string `json:"path"`
			FilesChanged int    `json:"files_changed"`
			CommitsAhead int    `json:"commits_ahead"`
			ClaudeState  string `json:"claude_state"`
			TerminalInfo string `json:"terminal_info,omitempty"`
			IsCurrent    bool   `json:"is_current"`
		}

		var outputs []sessionOutput

		for _, sess := range sessions {
			status, _ := sessMgr.GetStatus(sess)

			// Filter for running only
			if lsRunning && status.ClaudeState != "running" {
				continue
			}

			isCurrent := currentSession != nil && currentSession.Name == sess.Name

			out := sessionOutput{
				Name:         sess.Name,
				Branch:       sess.Branch,
				Path:         sess.Path,
				FilesChanged: status.FilesChanged,
				CommitsAhead: status.CommitsAhead,
				ClaudeState:  string(status.ClaudeState),
				TerminalInfo: status.TerminalInfo,
				IsCurrent:    isCurrent,
			}
			outputs = append(outputs, out)
		}

		if lsJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(outputs)
		}

		// Table output
		for _, out := range outputs {
			marker := " "
			if out.IsCurrent {
				marker = "*"
			}

			filesStr := fmt.Sprintf("%d file", out.FilesChanged)
			if out.FilesChanged != 1 {
				filesStr += "s"
			}

			claudeStr := fmt.Sprintf("claude: %s", out.ClaudeState)

			line := fmt.Sprintf("%s %-18s %-22s %-10s %s",
				marker,
				out.Name,
				out.Branch,
				filesStr,
				claudeStr,
			)

			if out.TerminalInfo != "" {
				line += "   " + out.TerminalInfo
			}

			if lsVerbose {
				line += "\n    " + out.Path
			}

			fmt.Println(strings.TrimRight(line, " "))
		}

		return nil
	},
}

func init() {
	lsCmd.Flags().BoolVarP(&lsVerbose, "verbose", "v", false, "Show additional details")
	lsCmd.Flags().BoolVar(&lsRunning, "running", false, "Only show sessions with active Claude process")
	lsCmd.Flags().BoolVar(&lsJSON, "json", false, "Output as JSON")
}
