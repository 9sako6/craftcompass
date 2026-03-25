package config

import (
	"errors"
	"os"

	toml "github.com/pelletier/go-toml/v2"
)

type PrivacyMode string

const (
	PrivacyStrict   PrivacyMode = "strict"
	PrivacyBalanced PrivacyMode = "balanced"
	PrivacyDebug    PrivacyMode = "debug"
)

type TrackingConfig struct {
	OnlyGitRepos bool `toml:"only_git_repos"`
}

type IgnoreConfig struct {
	Paths []string `toml:"paths"`
}

type Config struct {
	PrivacyMode          PrivacyMode         `toml:"privacy_mode"`
	IdleThresholdMinutes int                 `toml:"idle_threshold_minutes"`
	RetentionDays        int                 `toml:"retention_days"`
	Tracking             TrackingConfig      `toml:"tracking"`
	Ignore               IgnoreConfig        `toml:"ignore"`
	Categories           map[string][]string `toml:"categories"`
}

func Default() Config {
	return Config{
		PrivacyMode:          PrivacyBalanced,
		IdleThresholdMinutes: 15,
		RetentionDays:        90,
		Categories: map[string][]string{
			"git":    {"git"},
			"test":   {"bun", "pytest", "cargo", "go", "npm", "pnpm", "yarn"},
			"build":  {"go", "cargo", "npm", "pnpm", "yarn", "bun"},
			"search": {"rg", "fd", "grep", "find", "ast-grep"},
			"edit":   {"nvim", "vim", "code"},
			"nav":    {"cd", "z", "zi", "pwd", "ls", "eza"},
		},
	}
}

func Load(path string) (Config, error) {
	cfg := Default()
	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}
		return cfg, err
	}

	var raw Config
	if err := toml.Unmarshal(data, &raw); err != nil {
		return cfg, err
	}

	if raw.PrivacyMode != "" {
		cfg.PrivacyMode = raw.PrivacyMode
	}
	if raw.IdleThresholdMinutes != 0 {
		cfg.IdleThresholdMinutes = raw.IdleThresholdMinutes
	}
	if raw.RetentionDays != 0 {
		cfg.RetentionDays = raw.RetentionDays
	}
	cfg.Tracking = raw.Tracking
	if raw.Ignore.Paths != nil {
		cfg.Ignore.Paths = raw.Ignore.Paths
	}
	for category, commands := range raw.Categories {
		cfg.Categories[category] = commands
	}

	return cfg, nil
}
