# CCS - Claude Code Sessions

A CLI tool for managing multiple concurrent Claude Code sessions using git worktrees.

## The Problem

Running multiple Claude Code sessions against the same repository is challenging. Each session needs its own isolated working directory to avoid conflicts, but manually managing worktrees, branches, and terminal windows is tedious.

## The Solution

CCS uses git worktrees to provide real isolation with a shared object store. Each session gets its own directory and branch, with automatic terminal window management and Claude Code integration.

## Features

- **Git Worktree Management**: Automatically creates and manages worktrees for each session
- **Terminal Integration**: Creates terminal windows/tabs for each session (supports tmux and Kitty)
- **Claude Code Awareness**: Detects Claude state (running/waiting/idle), manages processes
- **Job Control**: Ctrl-Z suspends Claude and returns to shell, `fg` resumes
- **Auto-cleanup**: Terminal tabs close automatically when Claude exits
- **Minimal Dependencies**: Only requires git. Terminal integration is optional.

## Installation

```bash
go install github.com/emaland/ccs@latest
```

Or build from source:

```bash
git clone https://github.com/emaland/ccs.git
cd ccs
go build -o ccs .
```

## Quick Start

```bash
# Create a new session
ccs new my-feature

# List all sessions
ccs ls

# Switch to a session
ccs switch my-feature

# Resume a paused session (restarts Claude with --continue)
ccs resume my-feature

# Pause Claude in a session
ccs pause my-feature

# View session status
ccs status my-feature

# View changes in a session
ccs diff my-feature
ccs log my-feature

# Finish a session
ccs finish my-feature --squash  # Squash merge to main
ccs finish my-feature --pr      # Push for PR
ccs finish my-feature --delete  # Delete without merging

# Global commands (work from anywhere)
ccs sessions                    # List all sessions across repos
ccs cleanup                     # Remove stale sessions

# Interactive TUI
ccs ui                          # Open dashboard for all sessions
```

## Commands

### `ccs new <name> [-- claude-args...]`

Create a new session with its own worktree and branch.

```bash
ccs new auth-refactor                           # Create session branching from main
ccs new auth-refactor --from develop            # Branch from specific ref
ccs new auth-refactor -- --dangerously-skip-permissions  # Pass flags to Claude
ccs new auth-refactor --no-claude               # Don't start Claude
ccs new auth-refactor --no-terminal             # Don't create terminal window
ccs new auth-refactor --here                    # Create worktree in ./.worktrees/
```

### `ccs ls`

List all sessions for the current repository.

```bash
ccs ls           # List sessions
ccs ls -v        # Verbose output
ccs ls --json    # JSON output
```

Output shows session name, branch, files changed, Claude state, and terminal info.

### `ccs switch <name>`

Switch to a session. In tmux/Kitty, switches to that window/tab.

```bash
ccs switch my-feature
ccs switch -           # Switch to previous session
```

### `ccs status [name]`

Show detailed status of a session.

```bash
ccs status              # Current session
ccs status my-feature   # Specific session
```

### `ccs resume [name] [-- claude-args...]`

Resume a paused session. Restarts Claude with `--continue` automatically.

```bash
ccs resume my-feature
ccs resume my-feature -- --dangerously-skip-permissions
```

### `ccs pause [name]`

Pause a session by stopping the Claude process.

```bash
ccs pause              # Current session
ccs pause my-feature   # Specific session
ccs pause --all        # All sessions
```

### `ccs diff [name]` / `ccs log [name]`

View changes in a session compared to the base branch.

```bash
ccs diff my-feature
ccs log my-feature
```

### `ccs finish <name>`

Finish a session with various options.

```bash
ccs finish my-feature --squash   # Squash merge to main
ccs finish my-feature --merge    # Merge to main (keep commits)
ccs finish my-feature --pr       # Push branch for PR
ccs finish my-feature --delete   # Delete without merging
ccs finish my-feature --force    # Skip confirmation/hooks
```

### `ccs sessions`

List all sessions globally, across all repositories.

```bash
ccs sessions        # List all sessions
ccs sessions --json # JSON output
```

### `ccs cleanup`

Remove stale sessions from global state (worktrees that no longer exist).

```bash
ccs cleanup
```

### `ccs ui`

Open an interactive TUI dashboard for managing all sessions across all repositories.

```bash
ccs ui
```

The dashboard shows:
- Session name
- Repository name
- Status (commits ahead, uncommitted changes)
- Files changed
- Claude state (running/waiting/idle)

**Keyboard shortcuts:**

| Key | Action |
|-----|--------|
| `j`/`k` or `↑`/`↓` | Navigate sessions |
| `Enter` | Switch to session |
| `n` | Create new session |
| `r` | Resume session (restart Claude) |
| `p` | Pause session (stop Claude) |
| `P` | Pause all sessions |
| `d` | View diff |
| `l` | View log |
| `f` | Finish session |
| `?` | Toggle help |
| `q` | Quit |

The TUI updates automatically every 2 seconds to reflect Claude state changes.

## Configuration

Global config at `~/.config/ccs/config.toml`:

```toml
# Where to store worktrees
worktree_location = "centralized"  # or "local" for ./.worktrees/
worktree_root = "~/.ccs"

# Branch naming
branch_prefix = "ccs/"
default_base = "main"

# Claude integration
auto_start_claude = true

# Terminal (auto-detected: tmux, kitty, or none)
terminal = "auto"

[hooks]
post_create = ""      # Run after creating worktree
pre_finish = ""       # Must exit 0 to proceed with finish

[terminal.tmux]
window_prefix = ""

[terminal.kitty]
tab_prefix = ""
```

Per-repo config at `<repo>/.ccs.toml` overrides global settings.

## Terminal Integration

### Tmux

CCS automatically detects tmux via `$TMUX` and creates windows for each session. Windows are named with the session name.

### Kitty

CCS detects Kitty via `$KITTY_WINDOW_ID`. Requires remote control enabled:

```
# In kitty.conf
allow_remote_control yes
```

### Job Control

When a session is created or resumed:
- Claude starts in an interactive shell
- **Ctrl-Z** suspends Claude and returns to a shell prompt
- **`fg`** resumes Claude
- When Claude exits, the terminal tab/window closes automatically

## Directory Structure

```
~/.ccs/
├── myrepo/
│   ├── auth-refactor/     # Worktree directory
│   │   ├── .git           # Points to main repo's .git
│   │   ├── .claude/       # Claude Code session data
│   │   └── ...
│   └── db-migration/
└── other-repo/

~/.config/ccs/
├── config.toml
└── state.json         # Tracks all sessions globally
```

## Session Lifecycle

1. **Create**: `ccs new my-feature` creates a worktree, branch, and terminal window
2. **Work**: Claude runs in the session, you can suspend/resume as needed
3. **Switch**: `ccs switch other-feature` to work on something else
4. **Finish**: `ccs finish my-feature --pr` pushes for review and cleans up

## Requirements

- **Git** (required)
- **Go 1.21+** (for building)
- **tmux** or **Kitty** (optional, for terminal integration)
- **Claude Code** (for Claude integration features)

## Design Philosophy

1. **Git worktrees are the right primitive** - Real isolation with shared object store
2. **Sessions should be cheap** - One command to create, minimal overhead
3. **Stay out of the way** - Plumbing, not a platform
4. **Integrate with existing tools** - tmux, Kitty, git, shell
5. **Terminal agnostic** - Abstract terminal integration behind an interface

## License

MIT License - see [LICENSE](LICENSE)
