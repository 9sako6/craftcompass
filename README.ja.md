# craftcompass

English README: [README.md](README.md)

**開発時間の使い方を、感覚ではなく見える形にする。**

`craftcompass` は、毎日の shell activity を local・privacy-first に集計して、focus できている時間、repo の切替、test の失敗ループを terminal で振り返れるようにします。

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

## うれしさ

- 今週どの repo に時間を使っているか、すぐわかる
- repo switch や idle gap が多い日を見つけやすい
- test-fail-repeat の詰まり方が見える
- データは手元だけに残る。SaaS も過剰追跡もない

## すぐ始める

まず binary を build します。

```bash
mise trust .mise.toml
mise install
mise exec -- go build -o ./bin/craftcompass ./cmd/craftcompass
```

つぎに `zsh` に hook を入れます。

```zsh
export CRAFTCOMPASS_BIN="$HOME/path/to/craftcompass/bin/craftcompass"
source "$HOME/path/to/craftcompass/shell/craftcompass.zsh"
```

あとは使うだけです。

```bash
craftcompass summary day
craftcompass summary week
craftcompass top repos
craftcompass doctor
```

## 見えるもの

### Daily / Weekly Summary

- active time
- top repos
- command category breakdown
- repo switches
- idle gaps
- longest focus block
- test-run pressure
- failure loop hints

### 壊れにくい Hook

- zsh `preexec` / `precmd`
- 記録に失敗しても shell は壊さない
- 保存は local JSONL のみ

## Privacy

`craftcompass` が保存するのはメタデータ中心です。

- `strict`: command name を保存しない
- `balanced`: 先頭 command token だけ保存する
- `debug`: ローカル診断用の予約モード

arguments、stdin、env、file contents、secrets は保存しません。

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

設定ファイルは `~/.config/craftcompass/config.toml` に置きます。

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
