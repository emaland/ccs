package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionState represents the persistent state of a session
type SessionState struct {
	Name       string    `json:"name"`
	RepoPath   string    `json:"repo_path"`
	RepoName   string    `json:"repo_name"`
	WorkTree   string    `json:"worktree"`
	Branch     string    `json:"branch"`
	BaseBranch string    `json:"base_branch"`
	CreatedAt  time.Time `json:"created_at"`
	LastAccess time.Time `json:"last_access"`
}

// GlobalState represents all tracked sessions across repos
type GlobalState struct {
	Sessions []SessionState `json:"sessions"`
	Version  int            `json:"version"`
}

// Manager handles global state persistence
type Manager struct {
	path  string
	state GlobalState
	mu    sync.RWMutex
}

// NewManager creates a new state manager
func NewManager() (*Manager, error) {
	home := os.Getenv("HOME")
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return nil, err
		}
	}

	stateDir := filepath.Join(home, ".config", "ccs")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		return nil, err
	}

	m := &Manager{
		path: filepath.Join(stateDir, "state.json"),
		state: GlobalState{
			Sessions: []SessionState{},
			Version:  1,
		},
	}

	// Load existing state if it exists
	if err := m.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return m, nil
}

// Load reads state from disk
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &m.state)
}

// Save writes state to disk
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}

// AddSession adds or updates a session in the global state
func (m *Manager) AddSession(sess SessionState) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Update if exists, otherwise add
	found := false
	for i, s := range m.state.Sessions {
		if s.WorkTree == sess.WorkTree {
			m.state.Sessions[i] = sess
			found = true
			break
		}
	}

	if !found {
		m.state.Sessions = append(m.state.Sessions, sess)
	}

	return m.saveUnlocked()
}

// RemoveSession removes a session from the global state
func (m *Manager) RemoveSession(worktreePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.state.Sessions {
		if s.WorkTree == worktreePath {
			m.state.Sessions = append(m.state.Sessions[:i], m.state.Sessions[i+1:]...)
			break
		}
	}

	return m.saveUnlocked()
}

// GetSession returns a session by worktree path
func (m *Manager) GetSession(worktreePath string) *SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.state.Sessions {
		if s.WorkTree == worktreePath {
			return &s
		}
	}
	return nil
}

// GetSessionByName returns a session by name and repo
func (m *Manager) GetSessionByName(name, repoPath string) *SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, s := range m.state.Sessions {
		if s.Name == name && s.RepoPath == repoPath {
			return &s
		}
	}
	return nil
}

// GetAllSessions returns all tracked sessions
func (m *Manager) GetAllSessions() []SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SessionState, len(m.state.Sessions))
	copy(result, m.state.Sessions)
	return result
}

// GetSessionsForRepo returns sessions for a specific repo
func (m *Manager) GetSessionsForRepo(repoPath string) []SessionState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []SessionState
	for _, s := range m.state.Sessions {
		if s.RepoPath == repoPath {
			result = append(result, s)
		}
	}
	return result
}

// UpdateLastAccess updates the last access time for a session
func (m *Manager) UpdateLastAccess(worktreePath string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, s := range m.state.Sessions {
		if s.WorkTree == worktreePath {
			m.state.Sessions[i].LastAccess = time.Now()
			return m.saveUnlocked()
		}
	}
	return nil
}

// Cleanup removes sessions whose worktrees no longer exist
func (m *Manager) Cleanup() ([]SessionState, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var removed []SessionState
	var remaining []SessionState

	for _, s := range m.state.Sessions {
		if _, err := os.Stat(s.WorkTree); os.IsNotExist(err) {
			removed = append(removed, s)
		} else {
			remaining = append(remaining, s)
		}
	}

	m.state.Sessions = remaining
	if err := m.saveUnlocked(); err != nil {
		return nil, err
	}

	return removed, nil
}

func (m *Manager) saveUnlocked() error {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.path, data, 0644)
}
