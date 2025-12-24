package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	WorktreeLocation string `toml:"worktree_location"` // "centralized" or "local"
	WorktreeRoot     string `toml:"worktree_root"`     // e.g., "~/.ccs"
	BranchPrefix     string `toml:"branch_prefix"`     // e.g., "ccs/"
	AutoStartClaude  bool   `toml:"auto_start_claude"`
	Terminal         string `toml:"terminal"`     // "auto", "tmux", "kitty", "wezterm", "none"
	DefaultBase      string `toml:"default_base"` // e.g., "main"

	Hooks    HooksConfig    `toml:"hooks"`
	Tmux     TmuxConfig     `toml:"terminal.tmux"`
	Kitty    KittyConfig    `toml:"terminal.kitty"`
}

type HooksConfig struct {
	PostCreate string `toml:"post_create"`
	PreFinish  string `toml:"pre_finish"`
}

type TmuxConfig struct {
	UseSessions  bool   `toml:"use_sessions"`
	WindowPrefix string `toml:"window_prefix"`
}

type KittyConfig struct {
	UseOSWindows bool   `toml:"use_os_windows"`
	TabPrefix    string `toml:"tab_prefix"`
}

func Default() *Config {
	homeDir, _ := os.UserHomeDir()
	return &Config{
		WorktreeLocation: "centralized",
		WorktreeRoot:     filepath.Join(homeDir, ".ccs"),
		BranchPrefix:     "ccs/",
		AutoStartClaude:  true,
		Terminal:         "auto",
		DefaultBase:      "main",
	}
}

func Load() (*Config, error) {
	cfg := Default()

	// Load global config
	globalPath := globalConfigPath()
	if _, err := os.Stat(globalPath); err == nil {
		if _, err := toml.DecodeFile(globalPath, cfg); err != nil {
			return nil, err
		}
	}

	// Load repo-specific config (overrides global)
	repoPath := ".ccs.toml"
	if _, err := os.Stat(repoPath); err == nil {
		if _, err := toml.DecodeFile(repoPath, cfg); err != nil {
			return nil, err
		}
	}

	// Expand ~ in paths
	cfg.WorktreeRoot = expandPath(cfg.WorktreeRoot)

	return cfg, nil
}

func globalConfigPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		homeDir, _ := os.UserHomeDir()
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "ccs", "config.toml")
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, _ := os.UserHomeDir()
		return filepath.Join(homeDir, path[1:])
	}
	return path
}

func (c *Config) GetWorktreePath(repoName, sessionName string) string {
	return filepath.Join(c.WorktreeRoot, repoName, sessionName)
}

func (c *Config) GetLocalWorktreePath(sessionName string) string {
	return filepath.Join(".worktrees", sessionName)
}

func (c *Config) GetBranchName(sessionName string) string {
	return c.BranchPrefix + sessionName
}
