# craftcompass

Japanese README: [README.ja.md](README.ja.md)

**See where your dev time actually goes.**

`craftcompass` turns everyday shell activity into local, privacy-first summaries so you can spot focus windows, repo churn, and failure loops before they become habits.

```text
$ craftcompass summary day

Day summary
active time: 5h42m
repo switches: 9
idle gaps: 4
longest focus: 1h18m
test runs: 23

top repos:
  ~/src/api 2h31m
  ~/src/web 1h44m
  ~/dotfiles 28m

categories:
  test 41%
  edit 24%
  git 18%
  search 11%
  misc 6%
```

## Why It Feels Useful

- See which repos are eating your week.
- Catch fragmented days with too many switches and idle gaps.
- Notice when you are stuck in test-fail-repeat loops.
- Keep everything local. No SaaS. No surveillance creep.

## Get Started Fast

Build the binary:

```bash
mise trust .mise.toml
mise install
mise exec -- go build -o ./bin/craftcompass ./cmd/craftcompass
```

Wire it into `zsh`:

```zsh
export CRAFTCOMPASS_BIN="$HOME/path/to/craftcompass/bin/craftcompass"
source "$HOME/path/to/craftcompass/shell/craftcompass.zsh"
```

Then use it:

```bash
craftcompass summary day
craftcompass summary week
craftcompass top repos
craftcompass doctor
```

## What You Get

### Daily and Weekly Readouts

- Active time
- Top repos
- Command category breakdown
- Repo switches
- Idle gaps
- Longest focus block
- Test-run pressure
- Failure loop hints

### Zero-Fragility Shell Hooks

- zsh `preexec` / `precmd`
- Recording failures do not break your shell
- Local JSONL storage

## Privacy By Default

`craftcompass` stores metadata, not the dangerous parts.

- `strict`: no command name
- `balanced`: first command token only
- `debug`: reserved for local diagnostics

It does **not** store arguments, stdin, env, file contents, or secrets.

## Main Commands

```bash
craftcompass summary day
craftcompass summary week
craftcompass top repos
craftcompass top commands
craftcompass doctor
craftcompass export --format json
```

## Config

Config lives at `~/.config/craftcompass/config.toml`.

```toml
privacy_mode = "balanced"
idle_threshold_minutes = 15
retention_days = 90

[tracking]
only_git_repos = false

[ignore]
paths = [
  "/private/tmp",
  "/tmp",
]
```

## Storage

- Config: `~/.config/craftcompass/config.toml`
- Events: `~/.local/state/craftcompass/events.jsonl`
- Summary cache: `~/.local/state/craftcompass/summaries/`

## Development

```bash
mise exec -- go test ./...
```
