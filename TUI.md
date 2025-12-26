# Add Terminal UI to CCS

## Overview

Add an interactive terminal UI (TUI) to CCS using Bubble Tea (charmbracelet/bubbletea). The TUI should provide a dashboard for managing Claude Code sessions without memorizing commands.

## Technical Stack

- **bubbletea** - Core TUI framework (Elm architecture)
- **bubbles** - Pre-built components (table, viewport, textinput, spinner)
- **lipgloss** - Styling and layout

## Entry Point

Add a new command `ccs ui` (or just `ccs` with no arguments) that launches the interactive interface.

## Main Dashboard View

The primary view should show all sessions in a table:

```
CCS - Claude Code Sessions                                    [?] help  [q] quit

  NAME              BRANCH              STATUS      FILES   CLAUDE
► auth-refactor     ccs/auth-refactor   3 ahead     +5/-2   running
  db-migration      ccs/db-migration    1 ahead     +12/-0  waiting
  fix-tests         ccs/fix-tests       uncommitted +1/-1   idle

[n]ew  [enter]switch  [r]esume  [p]ause  [d]iff  [l]og  [f]inish  [/]filter
```

### Table Columns
- **Name**: Session name
- **Branch**: Git branch name
- **Status**: Commits ahead of base, or "uncommitted" if dirty
- **Files**: Lines added/removed vs base branch
- **Claude**: Current state (running/waiting/idle/paused/unknown)

### Keybindings
- `j`/`k` or `↑`/`↓`: Navigate sessions
- `enter`: Switch to selected session (exits TUI, switches terminal)
- `n`: New session (opens input prompt)
- `r`: Resume selected session
- `p`: Pause selected session
- `P`: Pause all sessions
- `d`: Show diff view for selected session
- `l`: Show log view for selected session
- `s`: Show detailed status
- `f`: Finish session (opens finish workflow)
- `/`: Filter/search sessions
- `?`: Toggle help overlay
- `q` or `ctrl+c`: Quit

## Secondary Views

### New Session Input
Modal/overlay for creating a new session:
```
New Session
━━━━━━━━━━━
Name: █
Base: main

[enter] create  [esc] cancel
```

Fields:
- Name (required): Text input with validation (no spaces, valid git branch chars)
- Base (optional): Defaults to configured default_base

### Diff View
Full-screen view showing `git diff` output:
```
Diff: auth-refactor (ccs/auth-refactor vs main)                    [q] back

diff --git a/auth/handler.go b/auth/handler.go
index abc123..def456 100644
--- a/auth/handler.go
+++ b/auth/handler.go
@@ -10,6 +10,15 @@ func HandleAuth(w http.ResponseWriter, r *http.Request) {
...
```
- Use viewport component for scrolling
- Syntax highlighting if practical (or just raw diff)
- `q` or `esc` returns to dashboard

### Log View
Similar to diff view but showing `git log --oneline` for the session branch:
```
Log: auth-refactor                                                 [q] back

a1b2c3d Add JWT validation
e4f5g6h Refactor auth middleware  
h7i8j9k Initial auth handler
```

### Finish Workflow
Step-through workflow for finishing a session:
```
Finish: auth-refactor
━━━━━━━━━━━━━━━━━━━━

Choose action:
  ► Squash merge to main
    Merge to main (keep commits)
    Push for PR
    Delete without merging

[enter] select  [esc] cancel
```

Then confirm:
```
Squash merge auth-refactor → main?

This will:
  • Squash 3 commits into one
  • Merge to main
  • Delete worktree and branch

[y] confirm  [n] cancel
```

### Help Overlay
Toggle-able overlay showing all keybindings:
```
┌─────────────────────────────────────┐
│ Keyboard Shortcuts                  │
├─────────────────────────────────────┤
│ Navigation                          │
│   j/k, ↑/↓    Move selection        │
│   enter       Switch to session     │
│   /           Filter sessions       │
│                                     │
│ Actions                             │
│   n           New session           │
│   r           Resume session        │
│   p           Pause session         │
│   P           Pause all             │
│   f           Finish session        │
│                                     │
│ Views                               │
│   d           Show diff             │
│   l           Show log              │
│   s           Show status           │
│                                     │
│ General                             │
│   ?           Toggle help           │
│   q           Quit                  │
└─────────────────────────────────────┘
```

## Architecture

### Package Structure
```
internal/
  tui/
    tui.go          # Main model, entry point
    dashboard.go    # Dashboard view & update logic
    newsession.go   # New session input component
    diffview.go     # Diff viewport
    logview.go      # Log viewport  
    finish.go       # Finish workflow
    help.go         # Help overlay
    styles.go       # Lipgloss styles
    keys.go         # Keybinding definitions
```

### Model Structure
```go
type model struct {
    // Current view state
    view        viewState  // dashboard, newSession, diff, log, finish
    
    // Data
    sessions    []Session
    selected    int
    
    // Sub-components
    table       table.Model
    viewport    viewport.Model
    textinput   textinput.Model
    spinner     spinner.Model
    
    // State
    loading     bool
    err         error
    showHelp    bool
    filterText  string
    
    // Dimensions
    width       int
    height      int
}
```

### Session Data Integration
Reuse existing session discovery logic. Create an interface or extract functions that both CLI commands and TUI can use:

```go
// internal/session/session.go
type Session struct {
    Name        string
    Branch      string
    WorktreePath string
    BaseBranch  string
    ClaudeState string  // running, waiting, idle, paused, unknown
    Ahead       int
    Behind      int
    FilesChanged int
    LinesAdded  int
    LinesRemoved int
    IsDirty     bool
}

func List(repoPath string) ([]Session, error)
func Get(repoPath, name string) (*Session, error)
func Create(repoPath, name string, opts CreateOpts) (*Session, error)
// etc.
```

### Auto-refresh
Use Bubble Tea's `tick` command to refresh session state periodically (every 2-3 seconds) to catch Claude state changes.

```go
type tickMsg time.Time

func tickCmd() tea.Cmd {
    return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
        return tickMsg(t)
    })
}
```

## Styling Guidelines

- Use subtle colors that work on both light and dark terminals
- Highlight selected row clearly
- Use dim text for less important info
- Status indicators:
  - `running` → green or cyan
  - `waiting` → yellow
  - `idle` → dim/gray
  - `paused` → dim/gray
- Keep it minimal and readable

## Error Handling

- Show errors inline (e.g., "Failed to create session: branch already exists")
- Non-fatal errors should display briefly then clear
- Fatal errors should show message and offer to quit

## Testing Considerations

- The TUI model should be testable by sending messages and checking resulting model state
- Extract business logic from view code so it can be unit tested independently
