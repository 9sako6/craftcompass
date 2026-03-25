package record

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/9sako6/craftcompass/internal/config"
	"github.com/9sako6/craftcompass/internal/paths"
	"github.com/9sako6/craftcompass/internal/store"
)

func TestRecorderStartAndEndWritesEvent(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	p := paths.Paths{
		StateDir:     root,
		EventsPath:   filepath.Join(root, "events.jsonl"),
		SummariesDir: filepath.Join(root, "summaries"),
		SessionsDir:  filepath.Join(root, "sessions"),
	}
	clock := newFakeClock(
		time.Date(2026, 3, 26, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
		time.Date(2026, 3, 26, 10, 0, 3, 120000000, time.FixedZone("JST", 9*60*60)),
	)
	rec := New(store.New(p), clock.Now, func() (string, error) { return "mbp", nil })
	cfg := config.Default()

	cwd := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(cwd, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}

	if err := rec.Start(cfg, StartOptions{
		SessionID: "shell-1",
		Shell:     "zsh",
		CWD:       filepath.Join(cwd, "api"),
		Command:   "git status",
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := rec.End(EndOptions{SessionID: "shell-1", ExitCode: 0}); err != nil {
		t.Fatalf("end: %v", err)
	}

	events, err := store.New(p).ReadEvents()
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if events[0].CommandCategory != "git" {
		t.Fatalf("expected git category, got %q", events[0].CommandCategory)
	}
	if events[0].DurationMS != 3120 {
		t.Fatalf("expected duration 3120ms, got %d", events[0].DurationMS)
	}
}

func TestRecorderStrictModeOmitsCommandName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	p := paths.Paths{
		StateDir:     root,
		EventsPath:   filepath.Join(root, "events.jsonl"),
		SummariesDir: filepath.Join(root, "summaries"),
		SessionsDir:  filepath.Join(root, "sessions"),
	}
	clock := newFakeClock(
		time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 26, 10, 0, 1, 0, time.UTC),
	)
	rec := New(store.New(p), clock.Now, func() (string, error) { return "mbp", nil })
	cfg := config.Default()
	cfg.PrivacyMode = config.PrivacyStrict

	cwd := filepath.Join(root, "repo")
	if err := os.MkdirAll(filepath.Join(cwd, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir repo: %v", err)
	}
	if err := rec.Start(cfg, StartOptions{
		SessionID: "shell-2",
		Shell:     "zsh",
		CWD:       cwd,
		Command:   "git status",
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := rec.End(EndOptions{SessionID: "shell-2", ExitCode: 0}); err != nil {
		t.Fatalf("end: %v", err)
	}

	events, err := store.New(p).ReadEvents()
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if events[0].CommandName != "" {
		t.Fatalf("expected empty command name in strict mode, got %q", events[0].CommandName)
	}
}

func TestRecorderSkipsIgnoredPaths(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	p := paths.Paths{
		StateDir:     root,
		EventsPath:   filepath.Join(root, "events.jsonl"),
		SummariesDir: filepath.Join(root, "summaries"),
		SessionsDir:  filepath.Join(root, "sessions"),
	}
	clock := newFakeClock(time.Now(), time.Now())
	rec := New(store.New(p), clock.Now, func() (string, error) { return "mbp", nil })
	cfg := config.Default()
	cfg.Ignore.Paths = []string{filepath.Join(root, "ignore-me")}

	if err := rec.Start(cfg, StartOptions{
		SessionID: "shell-3",
		Shell:     "zsh",
		CWD:       filepath.Join(root, "ignore-me", "repo"),
		Command:   "git status",
	}); err != nil {
		t.Fatalf("start: %v", err)
	}
	if err := rec.End(EndOptions{SessionID: "shell-3", ExitCode: 0}); err != nil {
		t.Fatalf("end: %v", err)
	}

	events, err := store.New(p).ReadEvents()
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 0 {
		t.Fatalf("expected no events for ignored path, got %d", len(events))
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
