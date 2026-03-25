package record

import (
	"errors"
	"os"
	"time"

	"github.com/9sako6/craftcompass/internal/classify"
	"github.com/9sako6/craftcompass/internal/config"
	"github.com/9sako6/craftcompass/internal/paths"
	"github.com/9sako6/craftcompass/internal/store"
)

type StartOptions struct {
	SessionID string
	Shell     string
	CWD       string
	Command   string
}

type EndOptions struct {
	SessionID string
	ExitCode  int
}

type Recorder struct {
	store    store.Store
	now      func() time.Time
	hostname func() (string, error)
}

func New(store store.Store, now func() time.Time, hostname func() (string, error)) Recorder {
	return Recorder{
		store:    store,
		now:      now,
		hostname: hostname,
	}
}

func (r Recorder) Start(cfg config.Config, options StartOptions) error {
	if paths.MatchesIgnoredPath(options.CWD, cfg.Ignore.Paths) {
		return nil
	}

	repo, ok := paths.FindRepoRoot(options.CWD)
	if cfg.Tracking.OnlyGitRepos && !ok {
		return nil
	}

	host, err := r.hostname()
	if err != nil {
		return err
	}

	commandName := classify.CommandName(options.Command)
	if cfg.PrivacyMode == config.PrivacyStrict {
		commandName = ""
	}

	return r.store.SaveSession(store.Session{
		ID:              options.SessionID,
		StartedAt:       r.now(),
		Host:            host,
		Shell:           options.Shell,
		Repo:            repo,
		CWD:             options.CWD,
		CommandCategory: classify.Category(options.Command, cfg.Categories),
		CommandName:     commandName,
	})
}

func (r Recorder) End(options EndOptions) error {
	session, err := r.store.LoadSession(options.SessionID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	endedAt := r.now()
	event := store.Event{
		StartedAt:       session.StartedAt,
		EndedAt:         endedAt,
		Host:            session.Host,
		Shell:           session.Shell,
		Repo:            session.Repo,
		CWD:             session.CWD,
		CommandCategory: session.CommandCategory,
		CommandName:     session.CommandName,
		ExitCode:        options.ExitCode,
		DurationMS:      endedAt.Sub(session.StartedAt).Milliseconds(),
	}
	if err := r.store.AppendEvent(event); err != nil {
		return err
	}
	return r.store.DeleteSession(options.SessionID)
}
