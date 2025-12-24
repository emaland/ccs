package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var shellInitCmd = &cobra.Command{
	Use:   "shell-init [shell]",
	Short: "Output shell integration code",
	Long: `Output shell integration code for bash/zsh/fish.

Add to your shell config:
  eval "$(ccs shell-init)"

Provides:
  - Tab completion
  - Prompt integration (show current session)
  - cd hooks for auto-context`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := detectShell()
		if len(args) > 0 {
			shell = args[0]
		}

		switch shell {
		case "bash":
			fmt.Print(bashInit)
		case "zsh":
			fmt.Print(zshInit)
		case "fish":
			fmt.Print(fishInit)
		default:
			return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish)", shell)
		}

		return nil
	},
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	return filepath.Base(shell)
}

const bashInit = `# CCS shell integration for bash

# Completion
_ccs_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local cmd="${COMP_WORDS[1]}"

    if [[ ${COMP_CWORD} -eq 1 ]]; then
        COMPREPLY=($(compgen -W "new ls switch status finish diff log pause resume hooks shell-init" -- "$cur"))
        return
    fi

    case "$cmd" in
        switch|status|finish|diff|log|pause|resume)
            local sessions=$(ccs ls --json 2>/dev/null | grep -o '"name":"[^"]*"' | cut -d'"' -f4)
            COMPREPLY=($(compgen -W "$sessions" -- "$cur"))
            ;;
    esac
}
complete -F _ccs_completions ccs

# Prompt integration
_ccs_prompt() {
    local session=$(ccs _current-session 2>/dev/null)
    if [[ -n "$session" ]]; then
        echo "[ccs:$session]"
    fi
}

# Add to PS1 if not already there
if [[ "$PS1" != *'$(_ccs_prompt)'* ]]; then
    PS1='$(_ccs_prompt)'"$PS1"
fi
`

const zshInit = `# CCS shell integration for zsh

# Completion
_ccs() {
    local -a commands sessions
    commands=(
        'new:Create a new session'
        'ls:List sessions'
        'switch:Switch to a session'
        'status:Show session status'
        'finish:Finish a session'
        'diff:Show diff for a session'
        'log:Show log for a session'
        'pause:Pause a session'
        'resume:Resume a session'
        'hooks:Manage Claude Code hooks'
        'shell-init:Output shell integration'
    )

    if (( CURRENT == 2 )); then
        _describe 'command' commands
        return
    fi

    case "$words[2]" in
        switch|status|finish|diff|log|pause|resume)
            sessions=(${(f)"$(ccs ls --json 2>/dev/null | grep -o '"name":"[^"]*"' | cut -d'"' -f4)"})
            _describe 'session' sessions
            ;;
    esac
}
compdef _ccs ccs

# Prompt integration
_ccs_prompt() {
    local session=$(ccs _current-session 2>/dev/null)
    if [[ -n "$session" ]]; then
        echo "[ccs:$session]"
    fi
}

# Add to precmd if not already there
if [[ ! "$precmd_functions" =~ "_ccs_precmd" ]]; then
    _ccs_precmd() {
        # Could update prompt here
        true
    }
    precmd_functions+=(_ccs_precmd)
fi
`

const fishInit = `# CCS shell integration for fish

# Completion
function __ccs_sessions
    ccs ls --json 2>/dev/null | string match -r '"name":"[^"]*"' | string replace -r '"name":"([^"]*)"' '$1'
end

complete -c ccs -n "__fish_use_subcommand" -a "new" -d "Create a new session"
complete -c ccs -n "__fish_use_subcommand" -a "ls" -d "List sessions"
complete -c ccs -n "__fish_use_subcommand" -a "switch" -d "Switch to a session"
complete -c ccs -n "__fish_use_subcommand" -a "status" -d "Show session status"
complete -c ccs -n "__fish_use_subcommand" -a "finish" -d "Finish a session"
complete -c ccs -n "__fish_use_subcommand" -a "diff" -d "Show diff for a session"
complete -c ccs -n "__fish_use_subcommand" -a "log" -d "Show log for a session"
complete -c ccs -n "__fish_use_subcommand" -a "pause" -d "Pause a session"
complete -c ccs -n "__fish_use_subcommand" -a "resume" -d "Resume a session"
complete -c ccs -n "__fish_use_subcommand" -a "hooks" -d "Manage Claude Code hooks"
complete -c ccs -n "__fish_use_subcommand" -a "shell-init" -d "Output shell integration"

complete -c ccs -n "__fish_seen_subcommand_from switch status finish diff log pause resume" -a "(__ccs_sessions)"

# Prompt integration
function _ccs_prompt
    set -l session (ccs _current-session 2>/dev/null)
    if test -n "$session"
        echo "[ccs:$session]"
    end
end
`

// Internal commands for shell integration

var currentSessionCmd = &cobra.Command{
	Use:    "_current-session",
	Hidden: true,
	Short:  "Print current session name",
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := sessMgr.GetCurrent()
		if err != nil {
			return nil // Silent fail
		}
		fmt.Println(sess.Name)
		return nil
	},
}

var previousSessionCmd = &cobra.Command{
	Use:    "_previous-session",
	Hidden: true,
	Short:  "Print previous session name",
	RunE: func(cmd *cobra.Command, args []string) error {
		if previousSession != "" {
			fmt.Println(previousSession)
		}
		return nil
	},
}

var sessionPathCmd = &cobra.Command{
	Use:    "_session-path <name>",
	Hidden: true,
	Short:  "Print session path",
	Args:   cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		sess, err := sessMgr.Get(args[0])
		if err != nil {
			return err
		}
		fmt.Println(sess.Path)
		return nil
	},
}

// Import for filepath.Base
func init() {
	_ = strings.TrimSpace // Avoid unused import
}
