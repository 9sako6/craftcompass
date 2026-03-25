# craftcompass

English README: [README.md](README.md)

`craftcompass` は、開発習慣をローカルで観測するための local-first observability tool です。shell command のメタデータを手元で記録し、外部送信なしで terminal 上に summary を表示します。

## Principles

- 保存先はローカルのみ
- privacy-first の既定値
- command arguments、stdin、env、file contents、secrets は保存しない
- zsh hook で低負荷にイベント収集
- 日次・週次の summary を terminal で確認できる

## Stack

- `mise` で管理する Go 1.26
- JSONL event store
- zsh `preexec` / `precmd` hook

## Setup

toolchain を入れて binary を build します。

```bash
mise trust .mise.toml
mise install
mise exec -- go build -o ./bin/craftcompass ./cmd/craftcompass
```

`./bin` を `PATH` に通すか、shell alias を設定してください。hook script は既定で `PATH` 上の `craftcompass` を探します。必要なら `CRAFTCOMPASS_BIN` で binary path を明示できます。

## zsh Hook

`.zshrc` に次を追加します。

```zsh
export CRAFTCOMPASS_BIN="$HOME/path/to/craftcompass/bin/craftcompass"
source "$HOME/path/to/craftcompass/shell/craftcompass.zsh"
```

hook 側は失敗しても握りつぶすので、record に失敗しても shell 操作は壊しません。

## Commands

hook からは次のように記録します。

```bash
craftcompass record start --session zsh-123 --shell zsh --cwd "$PWD" --command "git status"
craftcompass record end --session zsh-123 --exit-code 0
```

summary や診断は次です。

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

設定例:

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

- `strict`: category、time、repo、cwd、exit code、duration だけ保存する
- `balanced`: `strict` に加えて先頭 command token を保存する
- `debug`: ローカル開発用の予約モード。現状の永続化内容は `balanced` と同じ

## Development

test は `mise` 経由で実行します。

```bash
mise exec -- go test ./...
```
