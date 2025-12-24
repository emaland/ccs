package terminal

import (
	"os"

	"github.com/emaland/ccs/internal/config"
)

// Terminal is the interface for terminal operations
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

// Detect detects and returns the appropriate terminal implementation
func Detect(cfg *config.Config) Terminal {
	termType := cfg.Terminal

	if termType == "none" {
		return &NoopTerminal{}
	}

	if termType == "auto" || termType == "" {
		// Auto-detect
		if os.Getenv("TMUX") != "" {
			return NewTmuxTerminal(cfg)
		}
		if kitty := NewKittyTerminal(cfg); kitty != nil {
			return kitty
		}
		return &NoopTerminal{}
	}

	switch termType {
	case "tmux":
		return NewTmuxTerminal(cfg)
	case "kitty":
		if kitty := NewKittyTerminal(cfg); kitty != nil {
			return kitty
		}
		return &NoopTerminal{}
	default:
		return &NoopTerminal{}
	}
}
