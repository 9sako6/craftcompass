package app

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRunHelpListsPrimaryCommands(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	exitCode := Run([]string{"help"}, &stdout, &stderr)
	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d, stderr=%q", exitCode, stderr.String())
	}

	output := stdout.String()
	for _, snippet := range []string{
		"record",
		"summary",
		"top",
		"doctor",
		"export",
	} {
		if !bytes.Contains(stdout.Bytes(), []byte(snippet)) {
			t.Fatalf("expected help output to include %q, got %q", snippet, output)
		}
	}
}

func TestRunWorkflowCommands(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	repo := filepath.Join(root, "workspace", "demo")
	cwd := filepath.Join(repo, "api")
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo git dir: %v", err)
	}
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	clock := newFakeClock(
		time.Date(2026, 3, 26, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
		time.Date(2026, 3, 26, 10, 0, 3, 0, time.FixedZone("JST", 9*60*60)),
		time.Date(2026, 3, 26, 10, 0, 3, 0, time.FixedZone("JST", 9*60*60)),
	)
	deps := Dependencies{
		Getenv: func(key string) string {
			switch key {
			case "XDG_CONFIG_HOME":
				return filepath.Join(root, "config")
			case "XDG_STATE_HOME":
				return filepath.Join(root, "state")
			case "CRAFTCOMPASS_HOOKS_ACTIVE":
				return "1"
			default:
				return ""
			}
		},
		Getwd:    func() (string, error) { return cwd, nil },
		HomeDir:  func() (string, error) { return root, nil },
		Hostname: func() (string, error) { return "mbp", nil },
		Now:      clock.Now,
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	if exitCode := RunWithDependencies([]string{"record", "start", "--session", "shell-1", "--shell", "zsh", "--command", "git status"}, &stdout, &stderr, deps); exitCode != 0 {
		t.Fatalf("record start failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if exitCode := RunWithDependencies([]string{"record", "end", "--session", "shell-1", "--exit-code", "0"}, &stdout, &stderr, deps); exitCode != 0 {
		t.Fatalf("record end failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	stdout.Reset()
	stderr.Reset()
	if exitCode := RunWithDependencies([]string{"summary", "day"}, &stdout, &stderr, deps); exitCode != 0 {
		t.Fatalf("summary failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("active time: 0m")) && !bytes.Contains(stdout.Bytes(), []byte("active time:")) {
		t.Fatalf("expected summary output, got %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := RunWithDependencies([]string{"export", "--format", "json"}, &stdout, &stderr, deps); exitCode != 0 {
		t.Fatalf("export failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("command_category")) {
		t.Fatalf("expected exported json, got %q", stdout.String())
	}

	stdout.Reset()
	stderr.Reset()
	if exitCode := RunWithDependencies([]string{"doctor"}, &stdout, &stderr, deps); exitCode != 0 {
		t.Fatalf("doctor failed: exit=%d stderr=%q", exitCode, stderr.String())
	}
	if !bytes.Contains(stdout.Bytes(), []byte("shell hooks active: true")) {
		t.Fatalf("expected doctor output, got %q", stdout.String())
	}
}

type fakeClock struct {
	values []time.Time
	index  int
}

func newFakeClock(values ...time.Time) *fakeClock {
	return &fakeClock{values: values}
}

func (f *fakeClock) Now() time.Time {
	value := f.values[f.index]
	if f.index < len(f.values)-1 {
		f.index++
	}
	return value
}
