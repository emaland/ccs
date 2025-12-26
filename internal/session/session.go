package session

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/emaland/ccs/internal/claude"
	"github.com/emaland/ccs/internal/config"
	"github.com/emaland/ccs/internal/git"
	"github.com/emaland/ccs/internal/state"
	"github.com/emaland/ccs/internal/terminal"
)

// Session represents a CCS session
type Session struct {
	Name       string
	Path       string
	Branch     string
	BaseBranch string
	BaseCommit string
	RepoRoot   string
}

// Status contains runtime status information for a session
type Status struct {
	FilesChanged int
	CommitsAhead int
	ClaudeState  claude.State
	TerminalInfo string // e.g., "[tmux:2]"
}

// Manager handles session operations
type Manager struct {
	cfg      *config.Config
	git      git.Git
	terminal terminal.Terminal
	state    *state.Manager
}

// NewManager creates a new session manager
func NewManager(cfg *config.Config, g git.Git, term terminal.Terminal, stateMgr *state.Manager) *Manager {
	return &Manager{
		cfg:      cfg,
		git:      g,
		terminal: term,
		state:    stateMgr,
	}
}

// ValidateName validates a session name
func ValidateName(name string) error {
	if name == "" {
		return fmt.Errorf("session name cannot be empty")
	}
	if name == "-" {
		return nil // Special case for previous session
	}
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, name)
	if !matched {
		return fmt.Errorf("session name must be alphanumeric with hyphens and underscores only")
	}
	return nil
}

// Create creates a new session
func (m *Manager) Create(name string, opts CreateOptions) (*Session, error) {
	if err := ValidateName(name); err != nil {
		return nil, err
	}

	// Determine base - always use configured default (main) unless --from specified
	baseBranch := opts.From
	if baseBranch == "" {
		baseBranch = m.cfg.DefaultBase
	}

	baseCommit, err := m.git.ResolveRef(baseBranch)
	if err != nil {
		return nil, fmt.Errorf("could not resolve base ref %q: %w", baseBranch, err)
	}

	// Determine worktree path
	var worktreePath string
	if opts.Here {
		worktreePath = m.cfg.GetLocalWorktreePath(name)
	} else {
		worktreePath = m.cfg.GetWorktreePath(m.git.RepoName(), name)
	}

	// Create parent directory if needed
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		return nil, fmt.Errorf("could not create parent directory: %w", err)
	}

	// Create branch name
	branchName := m.cfg.GetBranchName(name)

	// Check if session already exists
	if m.git.BranchExists(branchName) {
		return nil, fmt.Errorf("session %q already exists\n\nUse 'ccs switch %s' to switch to it, or 'ccs finish %s --delete' to remove it", name, name, name)
	}

	// Create worktree with new branch
	if err := m.git.WorktreeAdd(worktreePath, branchName, baseBranch); err != nil {
		return nil, fmt.Errorf("could not create worktree: %w", err)
	}

	session := &Session{
		Name:       name,
		Path:       worktreePath,
		Branch:     branchName,
		BaseBranch: baseBranch,
		BaseCommit: baseCommit,
		RepoRoot:   m.git.RepoRoot(),
	}

	// Save to global state
	if m.state != nil {
		m.state.AddSession(state.SessionState{
			Name:       name,
			RepoPath:   m.git.RepoRoot(),
			RepoName:   m.git.RepoName(),
			WorkTree:   worktreePath,
			Branch:     branchName,
			BaseBranch: baseBranch,
			CreatedAt:  time.Now(),
			LastAccess: time.Now(),
		})
	}

	// Run post-create hook
	if m.cfg.Hooks.PostCreate != "" {
		if err := m.runHook(m.cfg.Hooks.PostCreate, worktreePath); err != nil {
			// Clean up on hook failure
			m.git.WorktreeRemove(worktreePath, true)
			m.git.BranchDelete(branchName, true)
			return nil, fmt.Errorf("post_create hook failed: %w", err)
		}
	}

	// Create terminal window
	if !opts.NoTerminal && m.terminal.Name() != "none" {
		// Pass claude command - terminal handles shell wrapping
		startCmd := ""
		if !opts.NoClaude && m.cfg.AutoStartClaude {
			startCmd = "claude"
			if len(opts.ClaudeArgs) > 0 {
				startCmd += " " + strings.Join(opts.ClaudeArgs, " ")
			}
		}
		if err := m.terminal.CreateWindow(name, worktreePath, startCmd); err != nil {
			// Non-fatal, just warn
			fmt.Fprintf(os.Stderr, "Warning: could not create terminal window: %v\n", err)
		}
	} else if !opts.NoClaude && m.cfg.AutoStartClaude {
		// Start claude in current terminal
		cmd := exec.Command("claude", opts.ClaudeArgs...)
		cmd.Dir = worktreePath
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// Don't wait - let claude take over
		if err := cmd.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: could not start claude: %v\n", err)
		}
	}

	return session, nil
}

// CreateOptions contains options for session creation
type CreateOptions struct {
	From       string   // Base branch/commit
	Here       bool     // Create in ./.worktrees/
	NoClaude   bool     // Don't start Claude
	NoTerminal bool     // Don't create terminal window
	ClaudeArgs []string // Arguments to pass to Claude
}

// List lists all sessions for the current repository
func (m *Manager) List() ([]*Session, error) {
	worktrees, err := m.git.WorktreeList()
	if err != nil {
		return nil, err
	}

	var sessions []*Session
	prefix := m.cfg.BranchPrefix

	for _, wt := range worktrees {
		if wt.Bare {
			continue
		}
		if !strings.HasPrefix(wt.Branch, prefix) {
			continue
		}

		name := strings.TrimPrefix(wt.Branch, prefix)
		sessions = append(sessions, &Session{
			Name:     name,
			Path:     wt.Path,
			Branch:   wt.Branch,
			RepoRoot: m.git.RepoRoot(),
		})
	}

	return sessions, nil
}

// Get gets a session by name
func (m *Manager) Get(name string) (*Session, error) {
	sessions, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, s := range sessions {
		if s.Name == name {
			return s, nil
		}
	}

	return nil, &ErrSessionNotFound{Name: name}
}

// GetCurrent gets the current session (if in one)
func (m *Manager) GetCurrent() (*Session, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	sessions, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, s := range sessions {
		if strings.HasPrefix(cwd, s.Path) {
			return s, nil
		}
	}

	return nil, &ErrNotInSession{}
}

// GetStatus gets the runtime status of a session
func (m *Manager) GetStatus(session *Session) (*Status, error) {
	status := &Status{}

	// Get files changed and commits ahead
	wtGit := m.git.InWorktree(session.Path)

	// Find merge base with main branch
	mergeBase, err := wtGit.MergeBase(m.cfg.DefaultBase, "HEAD")
	if err != nil {
		mergeBase = m.cfg.DefaultBase
	}

	diffStat, err := wtGit.DiffStat(mergeBase, "HEAD")
	if err == nil {
		status.FilesChanged = diffStat.FilesChanged
	}

	commitCount, err := wtGit.CommitCount(mergeBase, "HEAD")
	if err == nil {
		status.CommitsAhead = commitCount
	}

	// Get Claude state
	status.ClaudeState = claude.GetState(session.Path)

	// Get terminal info
	if m.terminal.Name() != "none" {
		windows, _ := m.terminal.ListWindows()
		for i, w := range windows {
			if w == session.Name {
				status.TerminalInfo = fmt.Sprintf("[%s:%d]", m.terminal.Name(), i+1)
				break
			}
		}
	}

	return status, nil
}

// Switch switches to a session
func (m *Manager) Switch(name string) error {
	session, err := m.Get(name)
	if err != nil {
		return err
	}

	// If in a terminal, switch window
	if m.terminal.Name() != "none" {
		if err := m.terminal.SwitchWindow(name); err == nil {
			return nil
		}
	}

	// Otherwise, print cd command for shell integration
	fmt.Printf("cd %s\n", session.Path)
	return nil
}

// Delete deletes a session
func (m *Manager) Delete(name string, force bool) error {
	session, err := m.Get(name)
	if err != nil {
		return err
	}

	// Stop claude process if running
	claude.StopProcess(session.Path)

	// Close terminal window
	if m.terminal.Name() != "none" {
		m.terminal.CloseWindow(name)
	}

	// Remove worktree
	if err := m.git.WorktreeRemove(session.Path, force); err != nil {
		return fmt.Errorf("could not remove worktree: %w", err)
	}

	// Delete branch
	if err := m.git.BranchDelete(session.Branch, force); err != nil {
		// Non-fatal
		fmt.Fprintf(os.Stderr, "Warning: could not delete branch %s: %v\n", session.Branch, err)
	}

	// Remove from global state
	if m.state != nil {
		m.state.RemoveSession(session.Path)
	}

	return nil
}

// Finish finishes a session with the given options
func (m *Manager) Finish(name string, opts FinishOptions) error {
	session, err := m.Get(name)
	if err != nil {
		return err
	}

	// Run pre-finish hook
	if m.cfg.Hooks.PreFinish != "" && !opts.Force {
		if err := m.runHook(m.cfg.Hooks.PreFinish, session.Path); err != nil {
			return fmt.Errorf("pre_finish hook failed: %w\nUse --force to skip", err)
		}
	}

	switch {
	case opts.Delete:
		return m.Delete(name, opts.Force)

	case opts.PR:
		// Push branch for PR
		wtGit := m.git.InWorktree(session.Path)
		if err := wtGit.Push(session.Branch, false); err != nil {
			return fmt.Errorf("could not push branch: %w", err)
		}
		// Don't delete branch or worktree when creating PR
		fmt.Printf("Branch %s pushed. Create a PR at your repository.\n", session.Branch)

		// Stop claude process
		claude.StopProcess(session.Path)

		// Close terminal window
		if m.terminal.Name() != "none" {
			m.terminal.CloseWindow(name)
		}

		// Remove worktree but keep branch
		return m.git.WorktreeRemove(session.Path, opts.Force)

	case opts.Squash:
		// Checkout base, squash merge, cleanup
		return m.mergeSession(session, true, opts.Force)

	case opts.Merge:
		return m.mergeSession(session, false, opts.Force)

	default:
		return fmt.Errorf("no finish action specified")
	}
}

// FinishOptions contains options for finishing a session
type FinishOptions struct {
	Squash bool
	Merge  bool
	PR     bool
	Delete bool
	Force  bool
}

func (m *Manager) mergeSession(session *Session, squash, force bool) error {
	// Get commit count for message
	mergeBase, _ := m.git.MergeBase(m.cfg.DefaultBase, session.Branch)
	commitCount, _ := m.git.CommitCount(mergeBase, session.Branch)

	// Checkout base branch
	baseGit := m.git // Use main repo git

	// First ensure base branch is checked out in main worktree
	currentBranch, _ := baseGit.BranchCurrent()
	if currentBranch != m.cfg.DefaultBase {
		// We're in a worktree, need to use main repo
		fmt.Printf("Please checkout %s in the main worktree before merging.\n", m.cfg.DefaultBase)
		return fmt.Errorf("not on base branch")
	}

	if squash {
		if err := baseGit.MergeSquash(session.Branch); err != nil {
			return fmt.Errorf("squash merge failed: %w", err)
		}
		msg := fmt.Sprintf("Squashed %d commits from %s", commitCount, session.Name)
		if err := baseGit.Commit(msg); err != nil {
			return fmt.Errorf("commit failed: %w", err)
		}
	} else {
		if err := baseGit.Merge(session.Branch, false); err != nil {
			return fmt.Errorf("merge failed: %w", err)
		}
	}

	// Cleanup
	return m.Delete(session.Name, force)
}

func (m *Manager) runHook(command, dir string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Error types

type ErrSessionNotFound struct {
	Name string
}

func (e *ErrSessionNotFound) Error() string {
	return fmt.Sprintf("session %q not found", e.Name)
}

type ErrSessionExists struct {
	Name string
}

func (e *ErrSessionExists) Error() string {
	return fmt.Sprintf("session %q already exists", e.Name)
}

type ErrNotInSession struct{}

func (e *ErrNotInSession) Error() string {
	return "not in a session"
}
