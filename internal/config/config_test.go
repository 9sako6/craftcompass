package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultReturnsBalancedPrivacyAndBuiltInCategories(t *testing.T) {
	t.Parallel()

	cfg := Default()
	if cfg.PrivacyMode != PrivacyBalanced {
		t.Fatalf("expected balanced privacy, got %q", cfg.PrivacyMode)
	}
	if cfg.IdleThresholdMinutes != 15 {
		t.Fatalf("expected idle threshold 15, got %d", cfg.IdleThresholdMinutes)
	}
	if len(cfg.Categories["git"]) == 0 {
		t.Fatalf("expected built-in git category")
	}
}

func TestLoadMergesOverridesIntoDefaults(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
privacy_mode = "strict"
idle_threshold_minutes = 20

[tracking]
only_git_repos = true

[ignore]
paths = ["/tmp/demo"]

[categories]
search = ["rg", "grep", "sg"]
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.PrivacyMode != PrivacyStrict {
		t.Fatalf("expected strict privacy, got %q", cfg.PrivacyMode)
	}
	if !cfg.Tracking.OnlyGitRepos {
		t.Fatalf("expected only_git_repos to be true")
	}
	if got := cfg.Categories["search"]; len(got) != 3 || got[2] != "sg" {
		t.Fatalf("expected category override to win, got %#v", got)
	}
	if len(cfg.Categories["git"]) == 0 {
		t.Fatalf("expected other defaults to remain")
	}
}
