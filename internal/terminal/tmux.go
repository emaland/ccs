package terminal

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/emaland/ccs/internal/config"
)

// TmuxTerminal implements Terminal for tmux
type TmuxTerminal struct {
	windowPrefix string
}

// NewTmuxTerminal creates a new TmuxTerminal
func NewTmuxTerminal(cfg *config.Config) *TmuxTerminal {
	return &TmuxTerminal{
		windowPrefix: cfg.Tmux.WindowPrefix,
	}
}

func (t *TmuxTerminal) Name() string {
	return "tmux"
}

func (t *TmuxTerminal) windowName(name string) string {
	return t.windowPrefix + name
}

func (t *TmuxTerminal) CreateWindow(name, path, startCmd string) error {
	windowName := t.windowName(name)
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/bash"
	}
	args := []string{"new-window", "-n", windowName, "-c", path}
	if startCmd != "" {
		// Start interactive login shell with job control that runs command
		// Source profile files for PATH, then run command
		initCmd := fmt.Sprintf(`[[ -r ~/.bash_profile ]] && . ~/.bash_profile || { [[ -r ~/.profile ]] && . ~/.profile; }; [[ -r ~/.bashrc ]] && . ~/.bashrc; %s`, startCmd)
		bashCmd := fmt.Sprintf(`exec %s --rcfile <(echo %q) -i`, shell, initCmd)
		args = append(args, shell, "-c", bashCmd)
	}
	return exec.Command("tmux", args...).Run()
}

func (t *TmuxTerminal) SwitchWindow(name string) error {
	windowName := t.windowName(name)
	return exec.Command("tmux", "select-window", "-t", windowName).Run()
}

func (t *TmuxTerminal) CloseWindow(name string) error {
	windowName := t.windowName(name)
	return exec.Command("tmux", "kill-window", "-t", windowName).Run()
}

func (t *TmuxTerminal) WindowExists(name string) bool {
	windowName := t.windowName(name)
	err := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Run()
	if err != nil {
		return false
	}
	out, _ := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Output()
	for _, w := range strings.Split(string(out), "\n") {
		if strings.TrimSpace(w) == windowName {
			return true
		}
	}
	return false
}

func (t *TmuxTerminal) RenameWindow(oldName, newName string) error {
	oldWindowName := t.windowName(oldName)
	newWindowName := t.windowName(newName)
	return exec.Command("tmux", "rename-window", "-t", oldWindowName, newWindowName).Run()
}

func (t *TmuxTerminal) ListWindows() ([]string, error) {
	out, err := exec.Command("tmux", "list-windows", "-F", "#{window_name}").Output()
	if err != nil {
		return nil, err
	}
	var windows []string
	prefix := t.windowPrefix
	for _, w := range strings.Split(string(out), "\n") {
		w = strings.TrimSpace(w)
		if w != "" && strings.HasPrefix(w, prefix) {
			windows = append(windows, strings.TrimPrefix(w, prefix))
		}
	}
	return windows, nil
}

func (t *TmuxTerminal) CurrentWindow() (string, error) {
	out, err := exec.Command("tmux", "display-message", "-p", "#{window_name}").Output()
	if err != nil {
		return "", err
	}
	name := strings.TrimSpace(string(out))
	if strings.HasPrefix(name, t.windowPrefix) {
		return strings.TrimPrefix(name, t.windowPrefix), nil
	}
	return name, nil
}
