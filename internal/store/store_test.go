package store

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/9sako6/craftcompass/internal/paths"
)

func TestStorePersistsSessionsAndEvents(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	p := paths.Paths{
		StateDir:     root,
		EventsPath:   filepath.Join(root, "events.jsonl"),
		SummariesDir: filepath.Join(root, "summaries"),
		SessionsDir:  filepath.Join(root, "sessions"),
	}
	s := New(p)

	session := Session{
		ID:              "shell-123",
		StartedAt:       time.Date(2026, 3, 26, 9, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
		Host:            "mbp",
		Shell:           "zsh",
		Repo:            "/repo/demo",
		CWD:             "/repo/demo/api",
		CommandCategory: "git",
		CommandName:     "git",
	}
	if err := s.SaveSession(session); err != nil {
		t.Fatalf("save session: %v", err)
	}

	loaded, err := s.LoadSession(session.ID)
	if err != nil {
		t.Fatalf("load session: %v", err)
	}
	if loaded.CommandName != "git" {
		t.Fatalf("expected session command name git, got %q", loaded.CommandName)
	}

	event := Event{
		StartedAt:       session.StartedAt,
		EndedAt:         session.StartedAt.Add(3 * time.Second),
		Host:            session.Host,
		Shell:           session.Shell,
		Repo:            session.Repo,
		CWD:             session.CWD,
		CommandCategory: session.CommandCategory,
		CommandName:     session.CommandName,
		ExitCode:        0,
		DurationMS:      3000,
	}
	if err := s.AppendEvent(event); err != nil {
		t.Fatalf("append event: %v", err)
	}

	events, err := s.ReadEvents()
	if err != nil {
		t.Fatalf("read events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}
	if err := s.DeleteSession(session.ID); err != nil {
		t.Fatalf("delete session: %v", err)
	}
}
