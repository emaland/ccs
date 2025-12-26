package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/emaland/ccs/internal/session"
)

// updateNewSession handles updates for the new session view
func (m Model) updateNewSession(msg tea.Msg, cmds []tea.Cmd) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Cancel):
			m.view = viewDashboard
			m.textinput.Reset()
			m.err = nil
			return m, nil

		case msg.Type == tea.KeyEnter:
			name := strings.TrimSpace(m.textinput.Value())
			if name == "" {
				m.err = fmt.Errorf("session name cannot be empty")
				return m, nil
			}

			// Validate name
			if err := session.ValidateName(name); err != nil {
				m.err = err
				return m, nil
			}

			// Create session
			cmds = append(cmds, m.createSession(name))
			m.loading = true
			return m, tea.Batch(cmds...)
		}
	}

	// Update text input
	var cmd tea.Cmd
	m.textinput, cmd = m.textinput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// viewNewSession renders the new session input view
func (m Model) viewNewSession() string {
	var b strings.Builder

	b.WriteString(modalTitleStyle.Render("New Session") + "\n")
	b.WriteString(strings.Repeat("━", 20) + "\n\n")

	b.WriteString(inputLabelStyle.Render("Name: "))
	b.WriteString(m.textinput.View() + "\n\n")

	b.WriteString(inputLabelStyle.Render("Base: "))
	b.WriteString(m.newSessionBase + "\n\n")

	if m.err != nil {
		b.WriteString(errorStyle.Render("Error: "+m.err.Error()) + "\n\n")
	}

	if m.loading {
		b.WriteString(m.spinner.View() + " Creating session...\n")
	} else {
		b.WriteString(keyHint("enter", " create") + "  " + keyHint("esc", " cancel") + "\n")
	}

	return modalStyle.Render(b.String())
}

// createSession creates a new session
func (m Model) createSession(name string) tea.Cmd {
	return func() tea.Msg {
		opts := session.CreateOptions{
			NoClaude:   true, // Don't start Claude from TUI
			NoTerminal: true, // Don't create terminal from TUI
		}

		_, err := m.sessMgr.Create(name, opts)
		if err != nil {
			return errorMsg(err)
		}

		return sessionCreatedMsg{name: name}
	}
}

// Ensure textinput is properly initialized
func (m *Model) initTextInput() {
	m.textinput = textinput.New()
	m.textinput.Placeholder = "session-name"
	m.textinput.CharLimit = 50
	m.textinput.Width = 30
}
