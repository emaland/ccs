package terminal

// NoopTerminal is a no-op implementation of Terminal
type NoopTerminal struct{}

func (n *NoopTerminal) Name() string {
	return "none"
}

func (n *NoopTerminal) CreateWindow(name, path, startCmd string) error {
	return nil
}

func (n *NoopTerminal) SwitchWindow(name string) error {
	return nil
}

func (n *NoopTerminal) CloseWindow(name string) error {
	return nil
}

func (n *NoopTerminal) WindowExists(name string) bool {
	return false
}

func (n *NoopTerminal) RenameWindow(oldName, newName string) error {
	return nil
}

func (n *NoopTerminal) ListWindows() ([]string, error) {
	return nil, nil
}

func (n *NoopTerminal) CurrentWindow() (string, error) {
	return "", nil
}
