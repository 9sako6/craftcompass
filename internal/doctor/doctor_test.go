package doctor

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/9sako6/craftcompass/internal/paths"
	"github.com/9sako6/craftcompass/internal/store"
)

func TestCheckReportsHookAndRecentIngestionState(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	p := paths.Paths{
		ConfigPath: filepath.Join(root, "config.toml"),
		StateDir:   filepath.Join(root, "state"),
		EventsPath: filepath.Join(root, "state", "events.jsonl"),
	}
	if err := os.MkdirAll(p.StateDir, 0o755); err != nil {
		t.Fatalf("mkdir state dir: %v", err)
	}

	events := []store.Event{
		{
			StartedAt: time.Date(2026, 3, 26, 10, 0, 0, 0, time.UTC),
			EndedAt:   time.Date(2026, 3, 26, 10, 0, 3, 0, time.UTC),
		},
	}

	report := Check(p, map[string]string{"CRAFTCOMPASS_HOOKS_ACTIVE": "1"}, events)
	if !report.HooksActive {
		t.Fatalf("expected hooks to be active")
	}
	if report.RecentEventCount != 1 {
		t.Fatalf("expected one recent event, got %d", report.RecentEventCount)
	}
	if report.ConfigFound {
		t.Fatalf("expected config to be missing")
	}
	if !report.StateWritable {
		t.Fatalf("expected state dir to be writable")
	}
}
