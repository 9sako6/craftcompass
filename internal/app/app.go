package app

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/9sako6/craftcompass/internal/config"
	"github.com/9sako6/craftcompass/internal/doctor"
	"github.com/9sako6/craftcompass/internal/paths"
	"github.com/9sako6/craftcompass/internal/record"
	"github.com/9sako6/craftcompass/internal/store"
	"github.com/9sako6/craftcompass/internal/summary"
)

const helpText = `craftcompass

Commands:
  record
  summary
  top
  doctor
  export
`

type Dependencies struct {
	Getenv   func(string) string
	Getwd    func() (string, error)
	HomeDir  func() (string, error)
	Hostname func() (string, error)
	Now      func() time.Time
}

func Run(args []string, stdout, stderr io.Writer) int {
	return RunWithDependencies(args, stdout, stderr, defaultDependencies())
}

func RunWithDependencies(args []string, stdout, stderr io.Writer, deps Dependencies) int {
	deps = withDefaults(deps)
	if len(args) == 0 || args[0] == "help" || args[0] == "--help" || args[0] == "-h" {
		_, _ = io.WriteString(stdout, helpText)
		return 0
	}

	homeDir, err := deps.HomeDir()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "resolve home dir: %v\n", err)
		return 1
	}

	pathSet := paths.Resolve(envMap(deps), homeDir)
	cfg, cfgErr := config.Load(pathSet.ConfigPath)
	if cfgErr != nil {
		_, _ = fmt.Fprintf(stderr, "warning: using defaults because config load failed: %v\n", cfgErr)
		cfg = config.Default()
	}

	eventStore := store.New(pathSet)
	recorder := record.New(eventStore, deps.Now, deps.Hostname)

	switch args[0] {
	case "record":
		return runRecord(args[1:], stdout, stderr, deps, cfg, recorder)
	case "summary":
		return runSummary(args[1:], stdout, stderr, deps, cfg, eventStore)
	case "top":
		return runTop(args[1:], stdout, stderr, eventStore)
	case "doctor":
		return runDoctor(stdout, stderr, deps, pathSet, eventStore)
	case "export":
		return runExport(args[1:], stdout, stderr, eventStore)
	default:
		_, _ = fmt.Fprintf(stderr, "unknown command: %s\n\n%s", strings.Join(args, " "), helpText)
		return 1
	}
}

func runRecord(args []string, _ io.Writer, stderr io.Writer, deps Dependencies, cfg config.Config, recorder record.Recorder) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, "usage: craftcompass record <start|end>\n")
		return 1
	}

	switch args[0] {
	case "start":
		flags := flag.NewFlagSet("record start", flag.ContinueOnError)
		flags.SetOutput(stderr)
		sessionID := flags.String("session", "", "session identifier")
		shellName := flags.String("shell", "zsh", "shell name")
		cwd := flags.String("cwd", "", "current working directory")
		command := flags.String("command", "", "shell command line")
		if err := flags.Parse(args[1:]); err != nil {
			return 1
		}
		if *sessionID == "" || *command == "" {
			_, _ = io.WriteString(stderr, "record start requires --session and --command\n")
			return 1
		}
		value := *cwd
		if value == "" {
			current, err := deps.Getwd()
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "resolve cwd: %v\n", err)
				return 1
			}
			value = current
		}
		if err := recorder.Start(cfg, record.StartOptions{
			SessionID: *sessionID,
			Shell:     *shellName,
			CWD:       value,
			Command:   *command,
		}); err != nil {
			_, _ = fmt.Fprintf(stderr, "record start: %v\n", err)
			return 1
		}
		return 0
	case "end":
		flags := flag.NewFlagSet("record end", flag.ContinueOnError)
		flags.SetOutput(stderr)
		sessionID := flags.String("session", "", "session identifier")
		exitCode := flags.Int("exit-code", 0, "command exit code")
		if err := flags.Parse(args[1:]); err != nil {
			return 1
		}
		if *sessionID == "" {
			_, _ = io.WriteString(stderr, "record end requires --session\n")
			return 1
		}
		if err := recorder.End(record.EndOptions{
			SessionID: *sessionID,
			ExitCode:  *exitCode,
		}); err != nil {
			_, _ = fmt.Fprintf(stderr, "record end: %v\n", err)
			return 1
		}
		return 0
	default:
		_, _ = io.WriteString(stderr, "usage: craftcompass record <start|end>\n")
		return 1
	}
}

func runSummary(args []string, stdout, stderr io.Writer, deps Dependencies, cfg config.Config, eventStore store.Store) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, "usage: craftcompass summary <day|week>\n")
		return 1
	}

	events, err := eventStore.ReadEvents()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "read events: %v\n", err)
		return 1
	}

	period := summary.Period(args[0])
	if period != summary.PeriodDay && period != summary.PeriodWeek {
		_, _ = io.WriteString(stderr, "summary period must be day or week\n")
		return 1
	}

	report := summary.Build(events, deps.Now(), period, time.Duration(cfg.IdleThresholdMinutes)*time.Minute)
	if err := eventStore.WriteSummaryCache(summary.CacheKey(period, deps.Now()), report); err != nil {
		_, _ = fmt.Fprintf(stderr, "warning: summary cache write failed: %v\n", err)
	}
	_, _ = io.WriteString(stdout, summary.Format(report))
	return 0
}

func runTop(args []string, stdout, stderr io.Writer, eventStore store.Store) int {
	if len(args) == 0 {
		_, _ = io.WriteString(stderr, "usage: craftcompass top <repos|commands>\n")
		return 1
	}

	events, err := eventStore.ReadEvents()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "read events: %v\n", err)
		return 1
	}

	var entries []summary.Entry
	switch args[0] {
	case "repos":
		entries = summary.TopRepos(events, 5)
	case "commands":
		entries = summary.TopCommands(events, 5)
	default:
		_, _ = io.WriteString(stderr, "top target must be repos or commands\n")
		return 1
	}

	for _, entry := range entries {
		_, _ = fmt.Fprintf(stdout, "%s\t%d\t%s\n", entry.Name, entry.Count, entry.Duration)
	}
	return 0
}

func runDoctor(stdout, stderr io.Writer, deps Dependencies, pathSet paths.Paths, eventStore store.Store) int {
	events, err := eventStore.ReadEvents()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "read events: %v\n", err)
		return 1
	}
	_, _ = io.WriteString(stdout, doctor.Format(doctor.Check(pathSet, envMap(deps), events)))
	return 0
}

func runExport(args []string, stdout, stderr io.Writer, eventStore store.Store) int {
	flags := flag.NewFlagSet("export", flag.ContinueOnError)
	flags.SetOutput(stderr)
	format := flags.String("format", "json", "export format")
	if err := flags.Parse(args); err != nil {
		return 1
	}
	if *format != "json" {
		_, _ = io.WriteString(stderr, "export format must be json\n")
		return 1
	}

	events, err := eventStore.ReadEvents()
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "read events: %v\n", err)
		return 1
	}
	data, err := summary.ExportJSON(events)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "export json: %v\n", err)
		return 1
	}
	_, _ = fmt.Fprintf(stdout, "%s\n", data)
	return 0
}

func defaultDependencies() Dependencies {
	return Dependencies{
		Getenv:   os.Getenv,
		Getwd:    os.Getwd,
		HomeDir:  os.UserHomeDir,
		Hostname: os.Hostname,
		Now:      time.Now,
	}
}

func withDefaults(deps Dependencies) Dependencies {
	defaults := defaultDependencies()
	if deps.Getenv == nil {
		deps.Getenv = defaults.Getenv
	}
	if deps.Getwd == nil {
		deps.Getwd = defaults.Getwd
	}
	if deps.HomeDir == nil {
		deps.HomeDir = defaults.HomeDir
	}
	if deps.Hostname == nil {
		deps.Hostname = defaults.Hostname
	}
	if deps.Now == nil {
		deps.Now = defaults.Now
	}
	return deps
}

func envMap(deps Dependencies) map[string]string {
	return map[string]string{
		"XDG_CONFIG_HOME":           deps.Getenv("XDG_CONFIG_HOME"),
		"XDG_STATE_HOME":            deps.Getenv("XDG_STATE_HOME"),
		"CRAFTCOMPASS_HOOKS_ACTIVE": deps.Getenv("CRAFTCOMPASS_HOOKS_ACTIVE"),
	}
}
