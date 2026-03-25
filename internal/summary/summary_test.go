package summary

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/9sako6/craftcompass/internal/store"
)

func TestBuildDayReportComputesCoreMetrics(t *testing.T) {
	t.Parallel()

	jst := time.FixedZone("JST", 9*60*60)
	now := time.Date(2026, 3, 26, 23, 0, 0, 0, jst)
	events := []store.Event{
		{
			StartedAt:       time.Date(2026, 3, 26, 9, 0, 0, 0, jst),
			EndedAt:         time.Date(2026, 3, 26, 9, 30, 0, 0, jst),
			Repo:            "/repo/a",
			CommandCategory: "git",
			CommandName:     "git",
			DurationMS:      30 * 60 * 1000,
			ExitCode:        0,
		},
		{
			StartedAt:       time.Date(2026, 3, 26, 9, 35, 0, 0, jst),
			EndedAt:         time.Date(2026, 3, 26, 9, 50, 0, 0, jst),
			Repo:            "/repo/a",
			CommandCategory: "test",
			CommandName:     "go",
			DurationMS:      15 * 60 * 1000,
			ExitCode:        1,
		},
		{
			StartedAt:       time.Date(2026, 3, 26, 10, 20, 0, 0, jst),
			EndedAt:         time.Date(2026, 3, 26, 10, 50, 0, 0, jst),
			Repo:            "/repo/b",
			CommandCategory: "search",
			CommandName:     "rg",
			DurationMS:      30 * 60 * 1000,
			ExitCode:        0,
		},
		{
			StartedAt:       time.Date(2026, 3, 26, 11, 0, 0, 0, jst),
			EndedAt:         time.Date(2026, 3, 26, 11, 10, 0, 0, jst),
			Repo:            "/repo/b",
			CommandCategory: "test",
			CommandName:     "pytest",
			DurationMS:      10 * 60 * 1000,
			ExitCode:        1,
		},
		{
			StartedAt:       time.Date(2026, 3, 26, 11, 12, 0, 0, jst),
			EndedAt:         time.Date(2026, 3, 26, 11, 18, 0, 0, jst),
			Repo:            "/repo/b",
			CommandCategory: "test",
			CommandName:     "pytest",
			DurationMS:      6 * 60 * 1000,
			ExitCode:        1,
		},
	}

	report := Build(events, now, PeriodDay, 15*time.Minute)
	if report.TotalActive != 91*time.Minute {
		t.Fatalf("expected total active 91m, got %s", report.TotalActive)
	}
	if report.RepoSwitches != 1 {
		t.Fatalf("expected 1 repo switch, got %d", report.RepoSwitches)
	}
	if report.IdleGaps != 1 {
		t.Fatalf("expected 1 idle gap, got %d", report.IdleGaps)
	}
	if report.LongestFocusBlock != 46*time.Minute {
		t.Fatalf("expected longest focus 46m, got %s", report.LongestFocusBlock)
	}
	if report.TestRuns != 3 {
		t.Fatalf("expected 3 test runs, got %d", report.TestRuns)
	}
	if len(report.FailureLoops) != 1 || report.FailureLoops[0].Count != 2 {
		t.Fatalf("expected one failure loop of length 2, got %#v", report.FailureLoops)
	}
}

func TestTopHelpersAndExportJSON(t *testing.T) {
	t.Parallel()

	events := []store.Event{
		{Repo: "/repo/a", CommandName: "git", DurationMS: 1000},
		{Repo: "/repo/a", CommandName: "git", DurationMS: 2000},
		{Repo: "/repo/b", CommandName: "rg", DurationMS: 500},
	}

	repos := TopRepos(events, 2)
	if len(repos) != 2 || repos[0].Name != "/repo/a" {
		t.Fatalf("unexpected repo ranking: %#v", repos)
	}

	commands := TopCommands(events, 2)
	if len(commands) != 2 || commands[0].Name != "git" || commands[0].Count != 2 {
		t.Fatalf("unexpected command ranking: %#v", commands)
	}

	data, err := ExportJSON(events)
	if err != nil {
		t.Fatalf("export json: %v", err)
	}
	var decoded []store.Event
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal export: %v", err)
	}
	if len(decoded) != 3 {
		t.Fatalf("expected 3 exported events, got %d", len(decoded))
	}
}
