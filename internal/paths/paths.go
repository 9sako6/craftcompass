package paths

import "path/filepath"

type Paths struct {
	ConfigPath   string
	StateDir     string
	EventsPath   string
	SummariesDir string
	SessionsDir  string
}

func Resolve(env map[string]string, home string) Paths {
	configHome := env["XDG_CONFIG_HOME"]
	if configHome == "" {
		configHome = filepath.Join(home, ".config")
	}

	stateHome := env["XDG_STATE_HOME"]
	if stateHome == "" {
		stateHome = filepath.Join(home, ".local", "state")
	}

	stateDir := filepath.Join(stateHome, "craftcompass")
	return Paths{
		ConfigPath:   filepath.Join(configHome, "craftcompass", "config.toml"),
		StateDir:     stateDir,
		EventsPath:   filepath.Join(stateDir, "events.jsonl"),
		SummariesDir: filepath.Join(stateDir, "summaries"),
		SessionsDir:  filepath.Join(stateDir, "sessions"),
	}
}

func FindRepoRoot(start string) (string, bool) {
	current := filepath.Clean(start)
	for {
		gitPath := filepath.Join(current, ".git")
		if _, err := stat(gitPath); err == nil {
			return current, true
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", false
		}
		current = parent
	}
}

func MatchesIgnoredPath(path string, ignorePaths []string) bool {
	cleanPath := filepath.Clean(path)
	for _, ignorePath := range ignorePaths {
		cleanIgnorePath := filepath.Clean(ignorePath)
		rel, err := filepath.Rel(cleanIgnorePath, cleanPath)
		if err != nil {
			continue
		}
		if rel == "." || rel == "" {
			return true
		}
		if rel != ".." && rel != "." && rel != "" && rel[0] != '.' {
			return true
		}
	}
	return false
}
