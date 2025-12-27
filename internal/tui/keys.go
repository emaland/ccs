package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines all keybindings
type KeyMap struct {
	// Navigation
	Up     key.Binding
	Down   key.Binding
	Select key.Binding

	// Actions
	New       key.Binding
	Resume    key.Binding
	Pause     key.Binding
	PauseAll  key.Binding
	Finish    key.Binding
	Diff      key.Binding
	Log       key.Binding
	Status    key.Binding
	Filter    key.Binding

	// General
	Help    key.Binding
	Quit    key.Binding
	Suspend key.Binding
	Back    key.Binding
	Cancel  key.Binding
	Confirm key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "switch to session"),
		),
		New: key.NewBinding(
			key.WithKeys("n"),
			key.WithHelp("n", "new session"),
		),
		Resume: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "resume session"),
		),
		Pause: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "pause session"),
		),
		PauseAll: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "pause all"),
		),
		Finish: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "finish session"),
		),
		Diff: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "show diff"),
		),
		Log: key.NewBinding(
			key.WithKeys("l"),
			key.WithHelp("l", "show log"),
		),
		Status: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "show status"),
		),
		Filter: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "filter"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		Suspend: key.NewBinding(
			key.WithKeys("ctrl+z"),
			key.WithHelp("ctrl+z", "suspend"),
		),
		Back: key.NewBinding(
			key.WithKeys("q", "esc"),
			key.WithHelp("q/esc", "back"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Confirm: key.NewBinding(
			key.WithKeys("y", "enter"),
			key.WithHelp("y/enter", "confirm"),
		),
	}
}
