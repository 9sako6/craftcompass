package store

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/9sako6/craftcompass/internal/paths"
)

type Event struct {
	StartedAt       time.Time `json:"ts_start"`
	EndedAt         time.Time `json:"ts_end"`
	Host            string    `json:"host"`
	Shell           string    `json:"shell"`
	Repo            string    `json:"repo,omitempty"`
	CWD             string    `json:"cwd"`
	CommandCategory string    `json:"command_category"`
	CommandName     string    `json:"command_name,omitempty"`
	ExitCode        int       `json:"exit_code"`
	DurationMS      int64     `json:"duration_ms"`
}

type Session struct {
	ID              string    `json:"id"`
	StartedAt       time.Time `json:"started_at"`
	Host            string    `json:"host"`
	Shell           string    `json:"shell"`
	Repo            string    `json:"repo,omitempty"`
	CWD             string    `json:"cwd"`
	CommandCategory string    `json:"command_category"`
	CommandName     string    `json:"command_name,omitempty"`
}

type Store struct {
	paths paths.Paths
}

func New(paths paths.Paths) Store {
	return Store{paths: paths}
}

func (s Store) SaveSession(session Session) error {
	if err := os.MkdirAll(s.paths.SessionsDir, 0o755); err != nil {
		return err
	}

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	path := s.sessionPath(session.ID)
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

func (s Store) LoadSession(sessionID string) (Session, error) {
	data, err := os.ReadFile(s.sessionPath(sessionID))
	if err != nil {
		return Session{}, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, err
	}
	return session, nil
}

func (s Store) DeleteSession(sessionID string) error {
	err := os.Remove(s.sessionPath(sessionID))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s Store) AppendEvent(event Event) error {
	if err := os.MkdirAll(filepath.Dir(s.paths.EventsPath), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(s.paths.EventsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = file.Write(data)
	return err
}

func (s Store) ReadEvents() ([]Event, error) {
	file, err := os.Open(s.paths.EventsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var events []Event
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var event Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s Store) WriteSummaryCache(name string, payload any) error {
	if err := os.MkdirAll(s.paths.SummariesDir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(s.paths.SummariesDir, name), data, 0o644)
}

func (s Store) sessionPath(sessionID string) string {
	return filepath.Join(s.paths.SessionsDir, sessionID+".json")
}
