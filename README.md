# craftcompass

`craftcompass` is a local-first observability tool for development habits. It records shell command metadata on your machine, stores everything locally, and renders terminal summaries without sending data anywhere.

## Principles

- Local storage only
- Privacy-first defaults
- No command arguments, stdin, env, file contents, or secrets
- zsh hook integration for low-friction capture
- Daily and weekly summaries in the terminal

## Stack

- Go 1.26 via `mise`
- JSONL event store
- zsh `preexec` and `precmd` hooks

## Setup

Install the toolchain and build the binary:

```bash
mise trust .mise.toml
mise install
mise exec -- go build -o ./bin/craftcompass ./cmd/craftcompass
```

Create a shell alias or put `./bin` on your `PATH`. The hook script looks for `craftcompass` on `PATH` by default, or you can set `CRAFTCOMPASS_BIN` explicitly.

## zsh Hook

Add this to your `.zshrc`:

```zsh
export CRAFTCOMPASS_BIN="$HOME/path/to/craftcompass/bin/craftcompass"
source "$HOME/path/to/craftcompass/shell/craftcompass.zsh"
```

The hook swallows errors so your shell keeps working even if event recording fails.

## Commands

Record commands from the hook:

```bash
craftcompass record start --session zsh-123 --shell zsh --cwd "$PWD" --command "git status"
craftcompass record end --session zsh-123 --exit-code 0
```

Inspect summaries:

```bash
craftcompass summary day
craftcompass summary week
craftcompass top repos
craftcompass top commands
craftcompass doctor
craftcompass export --format json
```

## Storage

- Config: `~/.config/craftcompass/config.toml`
- Events: `~/.local/state/craftcompass/events.jsonl`
- Summary cache: `~/.local/state/craftcompass/summaries/`

## Config

Example config:

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

[categories]
test = ["bun", "pytest", "cargo", "go", "npm", "pnpm", "yarn"]
search = ["rg", "fd", "grep", "find", "ast-grep"]
edit = ["nvim", "vim", "code"]
nav = ["cd", "z", "zi", "pwd", "ls", "eza"]
git = ["git"]
```

## Privacy Modes

- `strict`: store category, time, repo, cwd, exit code, and duration
- `balanced`: `strict` plus the first command token
- `debug`: reserved for local development diagnostics, currently persisted like `balanced`

## Development

Run the tests with `mise`:

```bash
mise exec -- go test ./...
```
