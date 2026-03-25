autoload -Uz add-zsh-hook

if [[ -n "${CRAFTCOMPASS_HOOKS_LOADED:-}" ]]; then
  return 0
fi

export CRAFTCOMPASS_HOOKS_LOADED=1
export CRAFTCOMPASS_HOOKS_ACTIVE=1
export CRAFTCOMPASS_SESSION_ID="${CRAFTCOMPASS_SESSION_ID:-zsh-$$}"

_craftcompass_bin() {
  if [[ -n "${CRAFTCOMPASS_BIN:-}" ]]; then
    printf '%s\n' "$CRAFTCOMPASS_BIN"
    return 0
  fi
  if command -v craftcompass >/dev/null 2>&1; then
    printf '%s\n' "craftcompass"
    return 0
  fi
  return 1
}

_craftcompass_preexec() {
  emulate -L zsh
  setopt local_options no_aliases

  local command_line="$1"
  local bin
  bin="$(_craftcompass_bin 2>/dev/null)" || return 0

  case "$command_line" in
    *craftcompass*) return 0 ;;
  esac

  "$bin" record start \
    --session "$CRAFTCOMPASS_SESSION_ID" \
    --shell "zsh" \
    --cwd "$PWD" \
    --command "$command_line" >/dev/null 2>&1 || true
}

_craftcompass_precmd() {
  emulate -L zsh
  setopt local_options no_aliases

  local exit_code=$?
  local bin
  bin="$(_craftcompass_bin 2>/dev/null)" || return 0

  "$bin" record end \
    --session "$CRAFTCOMPASS_SESSION_ID" \
    --exit-code "$exit_code" >/dev/null 2>&1 || true
}

add-zsh-hook preexec _craftcompass_preexec
add-zsh-hook precmd _craftcompass_precmd
