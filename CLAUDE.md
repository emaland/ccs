# CCS - Claude Code Sessions

A CLI tool for managing multiple concurrent Claude Code sessions using git worktrees.

## Project Overview

CCS solves the problem of running multiple Claude Code sessions against the same repository. Each session gets its own git worktree and branch, providing real isolation while sharing the git object store. The tool integrates with terminal multiplexers/emulators (tmux, Kitty, etc.) for window management and understands Claude Code's session state.

### Design Philosophy

1. **Git worktrees are the right primitive** — Real isolation with shared object store. No copying, no weird state.
2. **Sessions should be cheap** — Spinning one up is one command, not a workflow.
3. **Stay out of the way** — Plumbing, not a platform. No daemons, no databases, no web UIs.
4. **Integrate with existing tools** — tmux, Kitty, git, shell. Don't reinvent.
5. **Claude Code awareness** — Understand Claude's session files, hook into its lifecycle.
6. **Terminal agnostic** — Abstract terminal integration behind an interface.

### Tech Stack

- **Language**: Go
- **Dependencies**: Minimal. Only git required. Optional: tmux, kitty, or other supported terminals.
- **State storage**: Filesystem only. No databases. Session list derived from worktree directories.

## Architecture

```
ccs (binary)
├── cmd/                    # Command implementations
│   ├── root.go            # Root command, global flags
│   ├── new.go             # ccs new
│   ├── ls.go              # ccs ls
│   ├── switch.go          # ccs switch
│   ├── status.go          # ccs status
│   ├── finish.go          # ccs finish
│   ├── pause.go           # ccs pause
│   ├── resume.go          # ccs resume
│   ├── diff.go            # ccs diff
│   ├── log.go             # ccs log
│   ├── hooks.go           # ccs hooks
│   └── shell_init.go      # ccs shell-init
├── internal/
│   ├── config/            # Configuration loading
│   ├── git/               # Git abstraction
│   │   ├── git.go         # Interface definition
│   │   ├── exec.go        # Exec-based implementation
│   │   └── native.go      # Optional go-git/git2go implementation
│   ├── terminal/          # Terminal abstraction
│   │   ├── terminal.go    # Interface definition
│   │   ├── tmux.go        # Tmux implementation
│   │   ├── kitty.go       # Kitty implementation
│   │   └── noop.go        # No-op fallback
│   ├── claude/            # Claude Code state detection
│   ├── session/           # Session management logic
│   └── hooks/             # Claude Code hooks management
├── go.mod
├── go.sum
└── main.go
```

## Data Model

### Session

A session is defined by its worktree directory. No separate metadata storage.

```go
type Session struct {
    Name       string    // e.g., "auth-refactor"
    Path       string    // e.g., "~/.ccs/myrepo/auth-refactor"
    Branch     string    // e.g., "ccs/auth-refactor"
    BaseBranch string    // e.g., "main"
    BaseCommit string    // SHA of commit session branched from
    RepoRoot   string    // Original repo path
}
```

Session state is derived at runtime:
- **Files changed**: `git diff --stat` against base
- **Commits ahead**: `git rev-list --count base..HEAD`
- **Claude state**: Process inspection + `.claude/` directory parsing

### Configuration

Global config at `~/.config/ccs/config.toml`:

```toml
worktree_location = "centralized"  # or "local"
worktree_root = "~/.ccs"
branch_prefix = "ccs/"
auto_start_claude = true
terminal = "auto"  # "auto", "tmux", "kitty", "wezterm", "none"
default_base = "main"

[hooks]
post_create = ""      # Run after creating worktree
pre_finish = ""       # Must exit 0 to proceed with finish

[terminal.tmux]
use_sessions = false
window_prefix = ""

[terminal.kitty]
use_os_windows = false
tab_prefix = ""
```

Per-repo config at `<repo>/.ccs.toml` overrides global.

### Directory Structure

```
~/.ccs/
├── myrepo/
│   ├── auth-refactor/     # Worktree directory
│   │   ├── .git           # File pointing to main repo's .git
│   │   ├── .claude/       # Claude Code session data
│   │   └── ...
│   └── db-migration/
└── other-repo/

~/.config/ccs/
└── config.toml
```

## Command Specifications

### `ccs new <n> [flags]`

Create a new session.

**Flags:**
- `--from <ref>`: Base branch/commit (default: current branch)
- `--here`: Create worktree in `./.worktrees/<n>` instead of centralized
- `--no-claude`: Don't start Claude Code after creation
- `--no-terminal`: Don't create terminal window/tab

**Behavior:**
1. Validate name (alphanumeric, hyphens, underscores)
2. Create branch `<prefix><n>` from base
3. Create worktree at appropriate path
4. Run post_create hook if configured
5. Create terminal window/tab (if enabled)
6. Start `claude` (if enabled)

**Output:**
```
Created session auth-refactor
  Branch: ccs/auth-refactor (from main)
  Path:   ~/.ccs/myrepo/auth-refactor
Starting claude...
```

---

### `ccs ls [flags]`

List sessions for current repository.

**Flags:**
- `-v, --verbose`: Show additional details
- `--running`: Only sessions with active Claude process
- `--json`: Output as JSON

**Output:**
```
  auth-refactor     ccs/auth-refactor   3 files   claude: waiting   [tmux:2]
  db-migration      ccs/db-migration    12 files  claude: running   [tmux:3]
* api-docs          ccs/api-docs        1 file    claude: idle      [tmux:1]
```

**Claude States:** `running`, `waiting`, `idle`, `unknown`

---

### `ccs switch <n>`

Switch to a session. Use `-` for previous session.

**Behavior:**
- If in tmux/kitty: switches to that window/tab
- Otherwise: prints `cd <path>` (use with shell integration)

---

### `ccs status [name]`

Show detailed status of a session (defaults to current).

**Output:**
```
Session: auth-refactor
Branch:  ccs/auth-refactor (based on main, 4 commits ahead)
Path:    ~/.ccs/myrepo/auth-refactor

Files changed:
  M src/auth/middleware.ts
  A src/auth/oauth.ts

Claude: waiting (last active 12m ago)
Tokens: ~14k in / ~8k out
```

---

### `ccs diff [name]` / `ccs log [name]`

Pass-through to `git diff` / `git log` for session's changes against base. Supports standard git flags.

---

### `ccs pause [name]` / `ccs resume [name]`

**pause**: Stop Claude process, keep worktree. Use `--all` for all sessions.

**resume**: Restart Claude. Use `-c` to pass `--continue` to Claude.

---

### `ccs finish <n> [flags]`

Finish a session.

**Flags:**
- `--squash`: Squash all commits and merge to base
- `--merge`: Merge to base (keep commits)
- `--pr`: Push branch for PR, don't merge locally
- `--delete`: Delete without merging
- `--force`: Skip confirmation

**Interactive mode (no flags):**
```
Session auth-refactor has 4 commits.

[s] Squash and merge to main
[m] Merge to main (keep commits)
[p] Push branch for PR
[d] Delete without merging
[c] Cancel
```

After completion: removes worktree, deletes branch (unless --pr), closes terminal window.

---

### `ccs hooks <subcommand>`

- `install`: Install ccs hooks into Claude Code's hook system
- `uninstall`: Remove hooks
- `status`: Show current hook configuration

---

### `ccs shell-init [shell]`

Output shell integration code for bash/zsh/fish. Provides:
- Tab completion
- Prompt integration (show current session)
- `cd` hooks for auto-context

```bash
eval "$(ccs shell-init)"
```

---

### Internal Commands

- `ccs _current-session`: Print current session name (for shell integration)
- `ccs _previous-session`: Print previous session name
- `ccs _session-path <n>`: Print filesystem path for session

---

## Git Operations

Git operations are abstracted behind an interface, allowing exec-based (default) or native implementations.

### Git Interface

```go
type Git interface {
    RepoRoot() string
    
    // Worktree operations
    WorktreeAdd(path, branch, base string) error
    WorktreeList() ([]WorktreeInfo, error)
    WorktreeRemove(path string, force bool) error
    
    // Branch operations
    BranchCreate(name, ref string) error
    BranchDelete(name string, force bool) error
    BranchCurrent() (string, error)
    BranchExists(name string) bool
    
    // Ref operations
    ResolveRef(ref string) (string, error)
    MergeBase(ref1, ref2 string) (string, error)
    
    // Diff and status
    DiffStat(base, head string) (*DiffStat, error)
    DiffFiles(base, head string) ([]FileChange, error)
    CommitCount(base, head string) (int, error)
    IsClean() (bool, error)
    
    // Commit operations
    MergeSquash(branch string) error
    Merge(branch string, ff bool) error
    Commit(message string) error
    
    // Remote
    Push(branch string, force bool) error
    RemoteURL(name string) (string, error)
}

type WorktreeInfo struct {
    Path, Branch, HEAD string
    Bare               bool
}

type DiffStat struct {
    FilesChanged, Insertions, Deletions int
}

type FileChange struct {
    Path   string
    Status FileStatus // "A", "M", "D", "R", "C"
}
```

### ExecGit Implementation

Primary implementation shells out to git CLI:

```go
type ExecGit struct {
    repoRoot string
}

func (g *ExecGit) git(args ...string) *exec.Cmd {
    cmd := exec.Command("git", args...)
    cmd.Dir = g.repoRoot
    return cmd
}

func (g *ExecGit) WorktreeAdd(path, branch, base string) error {
    return g.git("worktree", "add", "-b", branch, path, base).Run()
}

func (g *ExecGit) WorktreeList() ([]WorktreeInfo, error) {
    out, _ := g.git("worktree", "list", "--porcelain").Output()
    return parseWorktreeList(string(out))
}

func (g *ExecGit) DiffStat(base, head string) (*DiffStat, error) {
    out, _ := g.git("diff", "--shortstat", base+".."+head).Output()
    return parseDiffStat(string(out))
}

// Other methods follow same pattern: construct args, run git, parse output
```

Key git commands used:
- `git worktree add -b <branch> <path> <base>`
- `git worktree list --porcelain`
- `git worktree remove [--force] <path>`
- `git branch [-d|-D] <n>`
- `git rev-parse --abbrev-ref HEAD`
- `git merge-base <ref1> <ref2>`
- `git diff --shortstat|--name-status <base>..<head>`
- `git rev-list --count <base>..<head>`
- `git merge [--squash] <branch>`

### Native Implementation (Optional)

For CGO environments, could use go-git + git2go:

```go
// +build cgo

type NativeGit struct {
    repo    *git.Repository   // go-git for most operations
    libRepo *git2go.Repository // git2go for worktree support
}
```

---

## Terminal Integration

Terminal operations are abstracted behind an interface. Detection priority: tmux → kitty → none.

### Terminal Interface

```go
type Terminal interface {
    Name() string
    CreateWindow(name, path, startCmd string) error
    SwitchWindow(name string) error
    CloseWindow(name string) error
    WindowExists(name string) bool
    RenameWindow(oldName, newName string) error
    ListWindows() ([]string, error)
    CurrentWindow() (string, error)
}
```

### Detection

```go
func DetectTerminal() Terminal {
    if os.Getenv("TMUX") != "" {
        return NewTmuxTerminal()
    }
    if kitty, ok := NewKittyTerminal(); ok {
        return kitty
    }
    return &NoopTerminal{}
}
```

### Tmux Implementation

```go
func (t *TmuxTerminal) CreateWindow(name, path, startCmd string) error {
    args := []string{"new-window", "-n", name, "-c", path}
    if startCmd != "" {
        args = append(args, startCmd)
    }
    return exec.Command("tmux", args...).Run()
}

func (t *TmuxTerminal) SwitchWindow(name string) error {
    return exec.Command("tmux", "select-window", "-t", name).Run()
}

func (t *TmuxTerminal) CloseWindow(name string) error {
    return exec.Command("tmux", "kill-window", "-t", name).Run()
}
```

### Kitty Implementation

Requires `allow_remote_control yes` in kitty.conf.

```go
func NewKittyTerminal() (*KittyTerminal, bool) {
    if os.Getenv("KITTY_WINDOW_ID") == "" {
        return nil, false
    }
    // Test remote control
    if exec.Command("kitty", "@", "ls").Run() != nil {
        fmt.Fprintln(os.Stderr, "Warning: Kitty remote control disabled")
        return nil, false
    }
    return &KittyTerminal{}, true
}

func (k *KittyTerminal) CreateWindow(name, path, startCmd string) error {
    args := []string{"@", "launch", "--type=tab", "--tab-title", name, "--cwd", path}
    if startCmd != "" {
        args = append(args, startCmd)
    } else {
        args = append(args, os.Getenv("SHELL"))
    }
    return exec.Command("kitty", args...).Run()
}

func (k *KittyTerminal) SwitchWindow(name string) error {
    return exec.Command("kitty", "@", "focus-tab", "--match", "title:^"+name+"$").Run()
}

func (k *KittyTerminal) ListWindows() ([]string, error) {
    out, _ := exec.Command("kitty", "@", "ls").Output()
    // Parse JSON: [{"tabs": [{"title": "..."}]}]
    var osWindows []struct {
        Tabs []struct{ Title string } `json:"tabs"`
    }
    json.Unmarshal(out, &osWindows)
    var titles []string
    for _, w := range osWindows {
        for _, t := range w.Tabs {
            titles = append(titles, t.Title)
        }
    }
    return titles, nil
}
```

### Adding New Terminals

1. Create `internal/terminal/<n>.go`
2. Implement `Terminal` interface
3. Add detection to `DetectTerminal()`

Environment variables for detection:
- tmux: `$TMUX`
- Kitty: `$KITTY_WINDOW_ID`
- Wezterm: `$WEZTERM_PANE`

---

## Claude Code Integration

### State Detection

```go
type ClaudeState string

const (
    ClaudeRunning ClaudeState = "running"
    ClaudeWaiting ClaudeState = "waiting"
    ClaudeIdle    ClaudeState = "idle"
    ClaudeUnknown ClaudeState = "unknown"
)

func GetClaudeState(sessionPath string) ClaudeState {
    // 1. Check for claude process with cwd in sessionPath
    // 2. If found, determine if running or waiting (check if blocked on stdin)
    // 3. If no process, return idle
}

func GetClaudeProcess(sessionPath string) (*os.Process, error) {
    // ps aux | grep claude | grep <sessionPath>
}

func GetLastPrompt(sessionPath string) (string, error) {
    // Parse .claude/ directory for session history
}

func GetTokenUsage(sessionPath string) (in, out int, err error) {
    // Parse .claude/ directory for usage stats
}
```

### Hooks

Claude Code supports lifecycle hooks. CCS installs hooks to track session state.

```bash
#!/bin/bash
# CCS hook: stop
CCS_SESSION=$(ccs _current-session 2>/dev/null)
if [[ -n "$CCS_SESSION" ]]; then
    ccs _update-status "$CCS_SESSION" stopped
fi
```

Hook events: `pre_tool_use`, `post_tool_use`, `notification`, `stop`

Commands:
- `ccs hooks install` — Install hooks in current repo
- `ccs hooks uninstall` — Remove hooks  
- `ccs hooks status` — Show installed hooks

---

## Error Handling

```go
type ErrSessionNotFound struct{ Name string }
type ErrSessionExists struct{ Name string }
type ErrNotInRepo struct{}
type ErrNotInSession struct{}
type ErrGitError struct{ Command, Output string; Err error }
type ErrHookFailed struct{ Hook, Output string; Err error }
```

User-facing messages should be clear and actionable:

```
Error: Session 'auth-refactor' not found

Available sessions:
  db-migration
  api-docs

Run 'ccs new auth-refactor' to create it.
```

```
Error: pre_finish hook failed

Command: npm test
Exit code: 1

Fix the failing tests or use --force to skip.
```

---

## Implementation Order

### Phase 1: Core (MVP)
- Config system
- Git operations (worktree create/list/remove, branches)
- Session model
- `ccs new`, `ccs ls`, `ccs switch`, `ccs finish --delete`

### Phase 2: Git Integration
- `ccs status`, `ccs diff`, `ccs log`
- `ccs finish --squash`, `--merge`, `--pr`

### Phase 3: Terminal Integration
- Terminal detection and abstraction
- Tmux implementation
- Kitty implementation

### Phase 4: Claude Integration
- Process detection
- State detection (running/waiting/idle)
- `ccs pause`, `ccs resume`
- Token/prompt extraction from .claude/

### Phase 5: Polish
- `ccs hooks` command
- `ccs shell-init` with completion
- Interactive finish menu
- Prompt integration

---

## Testing Strategy

### Unit Tests
- Config parsing
- Session name validation
- Git output parsing (worktree list, diff stat)
- Terminal command construction

### Integration Tests
Use temporary git repository:
1. `ccs new` creates worktree and branch
2. `ccs ls` shows created session
3. `ccs finish --delete` cleans up

### E2E Tests
With real git and tmux/kitty:
1. Full workflow: new → work → finish
2. Multiple concurrent sessions
3. Hook execution

---

## Open Questions

1. **Claude .claude/ structure**: Need to reverse-engineer for prompt/token extraction. May change between versions.

2. **Multiple repos**: Global view across repos, or stay repo-scoped?

3. **Conflict resolution**: What if base branch moves significantly? Offer rebase?

4. **Session templates**: Predefined configs for common workflows?

---

## Non-Goals

- **GUI**: CLI only. Build GUIs on top.
- **Claude wrapper**: Start Claude, don't intercept it.
- **Git replacement**: Use git, don't replace it.
- **Cross-machine sync**: Sessions are local. Use git for sync.
- **Multi-user**: Single-user tool. Teams use git branches/PRs.
