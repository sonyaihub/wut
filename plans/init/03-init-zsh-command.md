# 03 — `init zsh` command

## Description
Implement `terminal-helper init zsh`, which prints the zsh shell snippet to stdout. Users source it from `~/.zshrc`. For M0 the snippet is deliberately minimal — just enough to route the unresolved line through `terminal-helper detect`.

## Status
done

## Depends on
[02 — CLI framework](02-cli-framework.md)

## What / where to change

- `internal/shell/zsh.go` — export the snippet as a `go:embed` string (embed from `internal/shell/snippets/zsh.sh`) plus a `ZshSnippet() string` accessor. Embedding keeps the snippet editable as a real shell file with syntax highlighting.
- `internal/shell/snippets/zsh.sh` — the snippet. See spec §10 for the shape:
  ```zsh
  command_not_found_handler() {
    if terminal-helper detect --line "$*"; then
      return 0
    fi
    print -u2 "zsh: command not found: $1"
    return 127
  }
  ```
  Keep it to this. Do not add conditional installs, quote escaping gymnastics, or version pinning yet — step 05 will validate what's actually needed.
- `cmd/terminal-helper/init.go` — add:
  - `NewInitCmd()` — parent `init` command, no action of its own.
  - `NewInitZshCmd()` — `zsh` subcommand. `Run` writes `shell.ZshSnippet()` to stdout.
- Register `init` on the root in `cmd/terminal-helper/main.go` (or wherever root assembly lives).

## How to verify

```
./terminal-helper init                     # prints help listing the `zsh` subcommand
./terminal-helper init zsh                 # prints the snippet verbatim
./terminal-helper init zsh | zsh -n        # zsh parses the snippet with no syntax errors
diff <(./terminal-helper init zsh) internal/shell/snippets/zsh.sh   # empty diff
```

`zsh -n` (no-exec syntax check) is the key check — it proves the snippet is valid zsh without actually installing anything.
