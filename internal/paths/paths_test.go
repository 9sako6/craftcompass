package paths

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveUsesXDGOverrides(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"XDG_CONFIG_HOME": "/tmp/config-home",
		"XDG_STATE_HOME":  "/tmp/state-home",
	}

	got := Resolve(env, "/Users/tester")
	if got.ConfigPath != "/tmp/config-home/craftcompass/config.toml" {
		t.Fatalf("unexpected config path: %q", got.ConfigPath)
	}
	if got.EventsPath != "/tmp/state-home/craftcompass/events.jsonl" {
		t.Fatalf("unexpected events path: %q", got.EventsPath)
	}
}

func TestFindRepoRootDetectsClosestGitBoundary(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	repo := filepath.Join(root, "demo")
	cwd := filepath.Join(repo, "api", "handlers")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir git dir: %v", err)
	}
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir cwd: %v", err)
	}

	got, ok := FindRepoRoot(cwd)
	if !ok {
		t.Fatalf("expected repo root to be detected")
	}
	if got != repo {
		t.Fatalf("expected repo root %q, got %q", repo, got)
	}
}

func TestMatchesIgnoredPathUsesPrefixMatching(t *testing.T) {
	t.Parallel()

	if !MatchesIgnoredPath("/tmp/demo/project", []string{"/tmp/demo"}) {
		t.Fatalf("expected path to be ignored")
	}
	if MatchesIgnoredPath("/workspace/demo", []string{"/tmp/demo"}) {
		t.Fatalf("expected path not to be ignored")
	}
}
