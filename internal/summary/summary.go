package summary

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/9sako6/craftcompass/internal/store"
)

type Period string

const (
	PeriodDay  Period = "day"
	PeriodWeek Period = "week"
)

type FailureLoop struct {
	Category string
	Count    int
}

type Entry struct {
	Name       string
	Count      int
	Duration   time.Duration
	Percentage float64
}

type Report struct {
	Period            Period
	WindowStart       time.Time
	WindowEnd         time.Time
	TotalActive       time.Duration
	RepoTotals        []Entry
	CategoryTotals    []Entry
	RepoSwitches      int
	IdleGaps          int
	LongestFocusBlock time.Duration
	HourlyTotals      map[int]time.Duration
	FailureLoops      []FailureLoop
	TestRuns          int
}

func Build(events []store.Event, now time.Time, period Period, idleThreshold time.Duration) Report {
	start, end := bounds(now, period)
	filtered := filter(events, start, end)
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].StartedAt.Before(filtered[j].StartedAt)
	})

	report := Report{
		Period:       period,
		WindowStart:  start,
		WindowEnd:    end,
		HourlyTotals: map[int]time.Duration{},
	}

	repoTotals := map[string]time.Duration{}
	categoryTotals := map[string]time.Duration{}

	var currentFocusRepo string
	var currentFocus time.Duration
	var previous *store.Event
	var failureCategory string
	var failureCount int

	flushFailureLoop := func() {
		if failureCount >= 2 {
			report.FailureLoops = append(report.FailureLoops, FailureLoop{
				Category: failureCategory,
				Count:    failureCount,
			})
		}
		failureCategory = ""
		failureCount = 0
	}

	for i := range filtered {
		event := filtered[i]
		duration := time.Duration(event.DurationMS) * time.Millisecond
		report.TotalActive += duration
		report.HourlyTotals[event.StartedAt.In(now.Location()).Hour()] += duration
		if event.Repo != "" {
			repoTotals[event.Repo] += duration
		}
		categoryTotals[event.CommandCategory] += duration
		if event.CommandCategory == "test" {
			report.TestRuns++
		}

		if event.ExitCode != 0 {
			if failureCategory == event.CommandCategory {
				failureCount++
			} else {
				flushFailureLoop()
				failureCategory = event.CommandCategory
				failureCount = 1
			}
		} else {
			flushFailureLoop()
		}

		if previous == nil {
			currentFocusRepo = event.Repo
			currentFocus = duration
			previous = &filtered[i]
			continue
		}

		gap := event.StartedAt.Sub(previous.EndedAt)
		if gap > idleThreshold {
			report.IdleGaps++
		}
		if previous.Repo != "" && event.Repo != "" && previous.Repo != event.Repo {
			report.RepoSwitches++
		}

		if event.Repo != "" && event.Repo == currentFocusRepo && gap <= idleThreshold {
			currentFocus += duration
		} else {
			if currentFocus > report.LongestFocusBlock {
				report.LongestFocusBlock = currentFocus
			}
			currentFocusRepo = event.Repo
			currentFocus = duration
		}

		previous = &filtered[i]
	}

	flushFailureLoop()
	if currentFocus > report.LongestFocusBlock {
		report.LongestFocusBlock = currentFocus
	}

	report.RepoTotals = rankedEntries(repoTotals, report.TotalActive)
	report.CategoryTotals = rankedEntries(categoryTotals, report.TotalActive)
	return report
}

func TopRepos(events []store.Event, limit int) []Entry {
	repoTotals := map[string]time.Duration{}
	repoCounts := map[string]int{}
	for _, event := range events {
		if event.Repo == "" {
			continue
		}
		repoTotals[event.Repo] += time.Duration(event.DurationMS) * time.Millisecond
		repoCounts[event.Repo]++
	}
	entries := rankedEntries(repoTotals, 0)
	for index := range entries {
		entries[index].Count = repoCounts[entries[index].Name]
	}
	return limitEntries(entries, limit)
}

func TopCommands(events []store.Event, limit int) []Entry {
	commandStats := map[string]Entry{}
	for _, event := range events {
		if event.CommandName == "" {
			continue
		}
		entry := commandStats[event.CommandName]
		entry.Name = event.CommandName
		entry.Count++
		entry.Duration += time.Duration(event.DurationMS) * time.Millisecond
		commandStats[event.CommandName] = entry
	}

	entries := make([]Entry, 0, len(commandStats))
	for _, entry := range commandStats {
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Count == entries[j].Count {
			return entries[i].Duration > entries[j].Duration
		}
		return entries[i].Count > entries[j].Count
	})
	return limitEntries(entries, limit)
}

func ExportJSON(events []store.Event) ([]byte, error) {
	return json.MarshalIndent(events, "", "  ")
}

func Format(report Report) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "%s summary\n", titleCase(string(report.Period)))
	fmt.Fprintf(&builder, "active time: %s\n", humanDuration(report.TotalActive))
	fmt.Fprintf(&builder, "repo switches: %d\n", report.RepoSwitches)
	fmt.Fprintf(&builder, "idle gaps: %d\n", report.IdleGaps)
	fmt.Fprintf(&builder, "longest focus: %s\n", humanDuration(report.LongestFocusBlock))
	fmt.Fprintf(&builder, "test runs: %d\n", report.TestRuns)

	if len(report.RepoTotals) > 0 {
		builder.WriteString("top repos:\n")
		for _, entry := range limitEntries(report.RepoTotals, 3) {
			fmt.Fprintf(&builder, "  %s %s\n", entry.Name, humanDuration(entry.Duration))
		}
	}
	if len(report.CategoryTotals) > 0 {
		builder.WriteString("categories:\n")
		for _, entry := range report.CategoryTotals {
			fmt.Fprintf(&builder, "  %s %.0f%%\n", entry.Name, entry.Percentage)
		}
	}
	return builder.String()
}

func CacheKey(period Period, now time.Time) string {
	switch period {
	case PeriodWeek:
		year, week := now.ISOWeek()
		return fmt.Sprintf("week-%04d-%02d.json", year, week)
	default:
		return fmt.Sprintf("day-%s.json", now.Format("2006-01-02"))
	}
}

func bounds(now time.Time, period Period) (time.Time, time.Time) {
	local := now.In(now.Location())
	switch period {
	case PeriodWeek:
		weekday := int(local.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location()).AddDate(0, 0, -(weekday - 1))
		return start, start.AddDate(0, 0, 7)
	default:
		start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, local.Location())
		return start, start.AddDate(0, 0, 1)
	}
}

func filter(events []store.Event, start, end time.Time) []store.Event {
	filtered := make([]store.Event, 0, len(events))
	for _, event := range events {
		if !event.StartedAt.Before(start) && event.StartedAt.Before(end) {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

func rankedEntries(totals map[string]time.Duration, total time.Duration) []Entry {
	entries := make([]Entry, 0, len(totals))
	for name, duration := range totals {
		entry := Entry{Name: name, Duration: duration}
		if total > 0 {
			entry.Percentage = (float64(duration) / float64(total)) * 100
		}
		entries = append(entries, entry)
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Duration > entries[j].Duration
	})
	return entries
}

func limitEntries(entries []Entry, limit int) []Entry {
	if limit <= 0 || len(entries) <= limit {
		return entries
	}
	return entries[:limit]
}

func humanDuration(duration time.Duration) string {
	if duration == 0 {
		return "0m"
	}
	minutes := int(duration.Round(time.Minute).Minutes())
	hours := minutes / 60
	minutes = minutes % 60
	if hours == 0 {
		return fmt.Sprintf("%dm", minutes)
	}
	if minutes == 0 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dh%dm", hours, minutes)
}

func titleCase(value string) string {
	if value == "" {
		return value
	}
	return strings.ToUpper(value[:1]) + value[1:]
}
