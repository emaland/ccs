package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/emaland/ccs/internal/config"
)

// KittyTerminal implements Terminal for Kitty
type KittyTerminal struct {
	tabPrefix string
}

// NewKittyTerminal creates a new KittyTerminal if running in Kitty
func NewKittyTerminal(cfg *config.Config) *KittyTerminal {
	if os.Getenv("KITTY_WINDOW_ID") == "" {
		return nil
	}
	// Test remote control
	if exec.Command("kitty", "@", "ls").Run() != nil {
		fmt.Fprintln(os.Stderr, "Warning: Kitty remote control disabled. Add 'allow_remote_control yes' to kitty.conf")
		return nil
	}
	return &KittyTerminal{
		tabPrefix: cfg.Kitty.TabPrefix,
	}
}

func (k *KittyTerminal) Name() string {
	return "kitty"
}

func (k *KittyTerminal) tabName(name string) string {
	return k.tabPrefix + name
}

func (k *KittyTerminal) CreateWindow(name, path, startCmd string) error {
	tabName := k.tabName(name)

	// Create tab with default shell - capture the window ID from output
	args := []string{"@", "launch", "--type=tab", "--tab-title", tabName, "--cwd", path}
	out, err := exec.Command("kitty", args...).Output()
	if err != nil {
		return err
	}

	if startCmd != "" {
		// Send command as keystrokes to the shell - this gives proper job control
		// Ctrl-Z suspends and returns to shell prompt, fg resumes
		// Use the window ID returned by launch
		windowID := strings.TrimSpace(string(out))
		// Use PROMPT_COMMAND to exit when no jobs remain (handles Ctrl-Z + fg case)
		cmd := `PROMPT_COMMAND='[[ -z "$(jobs)" ]] && exit'; ` + startCmd + "\n"
		return exec.Command("kitty", "@", "send-text", "--match", "id:"+windowID, cmd).Run()
	}
	return nil
}

func (k *KittyTerminal) SwitchWindow(name string) error {
	tabName := k.tabName(name)
	return exec.Command("kitty", "@", "focus-tab", "--match", "title:^"+tabName+"$").Run()
}

func (k *KittyTerminal) CloseWindow(name string) error {
	tabName := k.tabName(name)
	return exec.Command("kitty", "@", "close-tab", "--match", "title:^"+tabName+"$").Run()
}

func (k *KittyTerminal) WindowExists(name string) bool {
	windows, err := k.ListWindows()
	if err != nil {
		return false
	}
	for _, w := range windows {
		if w == name {
			return true
		}
	}
	return false
}

func (k *KittyTerminal) RenameWindow(oldName, newName string) error {
	oldTabName := k.tabName(oldName)
	newTabName := k.tabName(newName)
	return exec.Command("kitty", "@", "set-tab-title", "--match", "title:^"+oldTabName+"$", newTabName).Run()
}

func (k *KittyTerminal) ListWindows() ([]string, error) {
	out, err := exec.Command("kitty", "@", "ls").Output()
	if err != nil {
		return nil, err
	}

	var osWindows []struct {
		Tabs []struct {
			Title string `json:"title"`
		} `json:"tabs"`
	}
	if err := json.Unmarshal(out, &osWindows); err != nil {
		return nil, err
	}

	var titles []string
	for _, w := range osWindows {
		for _, t := range w.Tabs {
			if strings.HasPrefix(t.Title, k.tabPrefix) {
				titles = append(titles, strings.TrimPrefix(t.Title, k.tabPrefix))
			}
		}
	}
	return titles, nil
}

func (k *KittyTerminal) CurrentWindow() (string, error) {
	out, err := exec.Command("kitty", "@", "ls").Output()
	if err != nil {
		return "", err
	}

	var osWindows []struct {
		IsFocused bool `json:"is_focused"`
		Tabs      []struct {
			Title     string `json:"title"`
			IsFocused bool   `json:"is_focused"`
		} `json:"tabs"`
	}
	if err := json.Unmarshal(out, &osWindows); err != nil {
		return "", err
	}

	for _, w := range osWindows {
		if w.IsFocused {
			for _, t := range w.Tabs {
				if t.IsFocused {
					name := t.Title
					if strings.HasPrefix(name, k.tabPrefix) {
						return strings.TrimPrefix(name, k.tabPrefix), nil
					}
					return name, nil
				}
			}
		}
	}
	return "", nil
}
