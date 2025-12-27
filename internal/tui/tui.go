package tui

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/emaland/ccs/internal/claude"
	"github.com/emaland/ccs/internal/config"
	"github.com/emaland/ccs/internal/git"
	"github.com/emaland/ccs/internal/session"
	"github.com/emaland/ccs/internal/state"
)

// viewState represents the current view
type viewState int

const (
	viewDashboard viewState = iota
	viewNewSession
	viewDiff
	viewLog
	viewFinish
	viewFinishConfirm
)

// SessionData extends session state with display data
type SessionData struct {
	Name       string
	RepoName   string
	Path       string
	Branch     string
	BaseBranch string
	Status     *session.Status
}

// Model is the main TUI model
type Model struct {
	// View state
	view     viewState
	showHelp bool

	// Data
	sessions []SessionData
	selected int
	repoPath string

	// Managers
	sessMgr  *session.Manager
	stateMgr *state.Manager
	cfg      *config.Config

	// Sub-components
	viewport  viewport.Model
	textinput textinput.Model
	spinner   spinner.Model

	// New session state
	newSessionName string
	newSessionBase string

	// Finish state
	finishAction  int
	finishSession *SessionData

	// UI state
	loading    bool
	err        error
	filterText string
	width      int
	height     int

	// Keys
	keys KeyMap

	// For switching session on exit
	switchToSession string
}

// Messages
type tickMsg time.Time
type sessionsLoadedMsg []SessionData
type errorMsg error
type sessionCreatedMsg struct{ name string }
type sessionFinishedMsg struct{ name string }
type diffLoadedMsg string
type logLoadedMsg string

// New creates a new TUI model
func New(cfg *config.Config, sessMgr *session.Manager, stateMgr *state.Manager, repoPath string) Model {
	ti := textinput.New()
	ti.Placeholder = "session-name"
	ti.CharLimit = 50

	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		view:     viewDashboard,
		sessMgr:  sessMgr,
		stateMgr: stateMgr,
		cfg:      cfg,
		repoPath: repoPath,
		keys:     DefaultKeyMap(),

		textinput: ti,
		spinner:   s,
		viewport:  viewport.New(80, 20),

		newSessionBase: cfg.DefaultBase,
		loading:        true,
	}
}

// Init implements tea.Model
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.loadSessions,
		m.spinner.Tick,
		tickCmd(),
	)
}

// SwitchToSession returns the session to switch to after TUI exits
func (m Model) SwitchToSession() string {
	return m.switchToSession
}

// tickCmd returns a command that ticks every 2 seconds
func tickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

// loadSessions loads session data from global state
func (m Model) loadSessions() tea.Msg {
	if m.stateMgr == nil {
		return errorMsg(fmt.Errorf("state manager not available"))
	}

	globalSessions := m.stateMgr.GetAllSessions()
	var data []SessionData

	for _, sess := range globalSessions {
		// Get status for each session
		status := &session.Status{
			ClaudeState: claude.GetState(sess.WorkTree),
		}

		// Try to get diff stats if the worktree exists
		if _, err := os.Stat(sess.WorkTree); err == nil {
			g, err := git.NewExecGit(sess.WorkTree)
			if err == nil {
				mergeBase, _ := g.MergeBase(sess.BaseBranch, "HEAD")
				if mergeBase == "" {
					mergeBase = sess.BaseBranch
				}
				if diffStat, err := g.DiffStat(mergeBase, "HEAD"); err == nil {
					status.FilesChanged = diffStat.FilesChanged
				}
				if commitCount, err := g.CommitCount(mergeBase, "HEAD"); err == nil {
					status.CommitsAhead = commitCount
				}
			}
		}

		data = append(data, SessionData{
			Name:       sess.Name,
			RepoName:   sess.RepoName,
			Path:       sess.WorkTree,
			Branch:     sess.Branch,
			BaseBranch: sess.BaseBranch,
			Status:     status,
		})
	}

	return sessionsLoadedMsg(data)
}

// Update implements tea.Model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	// Handle Ctrl-Z suspend globally
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if key.Matches(keyMsg, m.keys.Suspend) {
			return m, tea.Suspend
		}
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6 // Leave room for header/footer

	case tickMsg:
		// Refresh sessions periodically
		cmds = append(cmds, m.loadSessions, tickCmd())

	case sessionsLoadedMsg:
		m.sessions = msg
		m.loading = false
		// Keep selected in bounds
		if m.selected >= len(m.sessions) {
			m.selected = max(0, len(m.sessions)-1)
		}

	case errorMsg:
		m.err = msg
		m.loading = false

	case sessionCreatedMsg:
		m.view = viewDashboard
		m.textinput.Reset()
		cmds = append(cmds, m.loadSessions)

	case sessionFinishedMsg:
		m.view = viewDashboard
		cmds = append(cmds, m.loadSessions)

	case diffLoadedMsg:
		m.viewport.SetContent(string(msg))
		m.viewport.GotoTop()

	case logLoadedMsg:
		m.viewport.SetContent(string(msg))
		m.viewport.GotoTop()

	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	// Route to view-specific update
	switch m.view {
	case viewDashboard:
		return m.updateDashboard(msg, cmds)
	case viewNewSession:
		return m.updateNewSession(msg, cmds)
	case viewDiff, viewLog:
		return m.updateViewport(msg, cmds)
	case viewFinish:
		return m.updateFinish(msg, cmds)
	case viewFinishConfirm:
		return m.updateFinishConfirm(msg, cmds)
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m Model) View() string {
	if m.showHelp {
		return m.viewHelp()
	}

	switch m.view {
	case viewDashboard:
		return m.viewDashboard()
	case viewNewSession:
		return m.viewNewSession()
	case viewDiff:
		return m.viewDiff()
	case viewLog:
		return m.viewLog()
	case viewFinish:
		return m.viewFinish()
	case viewFinishConfirm:
		return m.viewFinishConfirm()
	default:
		return m.viewDashboard()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
