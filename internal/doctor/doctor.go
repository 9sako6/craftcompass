package doctor

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/9sako6/craftcompass/internal/paths"
	"github.com/9sako6/craftcompass/internal/store"
)

type Report struct {
	ConfigPath       string
	ConfigFound      bool
	StateDir         string
	StateWritable    bool
	HooksActive      bool
	RecentEventCount int
	LastEventAt      *time.Time
}

func Check(paths paths.Paths, env map[string]string, events []store.Event) Report {
	report := Report{
		ConfigPath:    paths.ConfigPath,
		StateDir:      paths.StateDir,
		HooksActive:   env["CRAFTCOMPASS_HOOKS_ACTIVE"] == "1",
		StateWritable: canWrite(paths.StateDir),
	}

	if _, err := os.Stat(paths.ConfigPath); err == nil {
		report.ConfigFound = true
	}
	if len(events) > 0 {
		report.RecentEventCount = len(events)
		last := events[len(events)-1].EndedAt
		report.LastEventAt = &last
	}
	return report
}

func Format(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "config path: %s\n", report.ConfigPath)
	fmt.Fprintf(&builder, "config found: %t\n", report.ConfigFound)
	fmt.Fprintf(&builder, "state dir: %s\n", report.StateDir)
	fmt.Fprintf(&builder, "state writable: %t\n", report.StateWritable)
	fmt.Fprintf(&builder, "shell hooks active: %t\n", report.HooksActive)
	fmt.Fprintf(&builder, "recent event count: %d\n", report.RecentEventCount)
	if report.LastEventAt != nil {
		fmt.Fprintf(&builder, "last event at: %s\n", report.LastEventAt.Format(time.RFC3339))
	}
	return builder.String()
}

func canWrite(path string) bool {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return false
	}
	file, err := os.CreateTemp(path, "write-check-*.tmp")
	if err != nil {
		return false
	}
	file.Close()
	return os.Remove(file.Name()) == nil
}
