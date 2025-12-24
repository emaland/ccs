package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var hooksCmd = &cobra.Command{
	Use:   "hooks",
	Short: "Manage Claude Code hooks",
	Long:  `Manage CCS integration with Claude Code hooks.`,
}

var hooksInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install CCS hooks into Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		hookDir := filepath.Join(gitRepo.RepoRoot(), ".claude", "hooks")
		if err := os.MkdirAll(hookDir, 0755); err != nil {
			return fmt.Errorf("could not create hooks directory: %w", err)
		}

		// Create stop hook
		stopHook := `#!/bin/bash
# CCS hook: stop
CCS_SESSION=$(ccs _current-session 2>/dev/null)
if [[ -n "$CCS_SESSION" ]]; then
    # Session stopped notification could be added here
    true
fi
`
		stopPath := filepath.Join(hookDir, "stop")
		if err := os.WriteFile(stopPath, []byte(stopHook), 0755); err != nil {
			return fmt.Errorf("could not write stop hook: %w", err)
		}

		fmt.Println("Installed CCS hooks:")
		fmt.Printf("  %s\n", stopPath)
		return nil
	},
}

var hooksUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove CCS hooks from Claude Code",
	RunE: func(cmd *cobra.Command, args []string) error {
		hookDir := filepath.Join(gitRepo.RepoRoot(), ".claude", "hooks")

		// Remove our hooks
		hooks := []string{"stop"}
		for _, hook := range hooks {
			path := filepath.Join(hookDir, hook)
			if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Warning: could not remove %s: %v\n", path, err)
			}
		}

		fmt.Println("Uninstalled CCS hooks")
		return nil
	},
}

var hooksStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show CCS hooks status",
	RunE: func(cmd *cobra.Command, args []string) error {
		hookDir := filepath.Join(gitRepo.RepoRoot(), ".claude", "hooks")

		hooks := []string{"stop"}
		installed := false

		for _, hook := range hooks {
			path := filepath.Join(hookDir, hook)
			if _, err := os.Stat(path); err == nil {
				if !installed {
					fmt.Println("Installed hooks:")
					installed = true
				}
				fmt.Printf("  %s\n", path)
			}
		}

		if !installed {
			fmt.Println("No CCS hooks installed.")
			fmt.Println("Run 'ccs hooks install' to install them.")
		}

		return nil
	},
}

func init() {
	hooksCmd.AddCommand(hooksInstallCmd)
	hooksCmd.AddCommand(hooksUninstallCmd)
	hooksCmd.AddCommand(hooksStatusCmd)
}
