# terminal-helper

You open a terminal, start typing `how do I rebase onto main without losing my stash`, and the shell yells `command not found`. terminal-helper catches that moment and hands the typed text off to your configured AI harness (Claude Code, aider, codex, or a custom CLI) as a prompt.

Zero latency on normal commands — we only run when the shell has already decided your first token isn't a real command.

## Example

```sh
$ how do I rebase onto main without losing my stash
╭─ claude ──────────────────────────────────────────────────────╮
│ Stash your work first with `git stash`, then run              │
│ `git rebase origin/main`. After the rebase, `git stash pop`   │
│ to re-apply…                                                  │
╰───────────────────────────────────────────────────────────────╯
$
$ gti status
zsh: command not found: gti
$ ls -la
(runs ls normally)
```

## Install

> Release artifacts aren't published yet. The URLs below are placeholders — links will be filled in once the homebrew tap and GitHub releases go live.

**Homebrew** (once the tap is published):

```sh
brew install <owner>/tap/terminal-helper
```

**curl | sh** (Go toolchain required for now):

```sh
curl -fsSL <install-script-url> | sh
```

**From source**:

```sh
go install <module-path>/cmd/terminal-helper@latest
```

## Setup

```sh
terminal-helper setup            # pick your harness and default mode
terminal-helper install-hook     # wire the hook into ~/.zshrc (or bash/fish)
```

Open a new shell, then:

```sh
terminal-helper doctor           # sanity-check the install
```

`install-hook` auto-detects your shell from `$SHELL`. Override with `--shell zsh|bash|fish`, add `-y` to skip the confirmation, or do it by hand if you'd rather:

```sh
# zsh
echo 'eval "$(terminal-helper init zsh)"' >> ~/.zshrc

# bash (requires 4.0+; macOS stock bash is 3.2)
echo 'eval "$(terminal-helper init bash)"' >> ~/.bashrc

# fish
terminal-helper init fish > ~/.config/fish/conf.d/terminal-helper.fish
```

## What it does

- **Detects** natural language at the prompt with near-zero false positives on real commands (heuristic: token count + stopword + interrogative + punctuation signals, with explicit metacharacter and path gates).
- **Routes** the line to your configured harness in one of three modes:
  - `interactive` — hand the terminal over (`exec` into the harness).
  - `headless` — one-shot answer streamed back inline with a box or markdown renderer; shell stays in control.
  - `ask` — arrow-key picker chooses per invocation.
- **Escape hatches** — prefix a line with `?` or `\` to force-route, `!` to force-passthrough. `??` forces headless, `?!` forces interactive.
- **Passthrough allowlist** — list first-tokens in `behavior.passthrough` that should never route even if NL-shaped.

## Commands

| Command | Purpose |
|---|---|
| `terminal-helper install-hook` | Wire the hook into your shell's rc file. Idempotent; `--shell` to override detection, `-y` to skip prompt. |
| `terminal-helper init zsh\|bash\|fish` | Print the shell-hook snippet (lower-level than `install-hook`). |
| `terminal-helper setup` | Guided config wizard. Supports `--harness` / `--mode` for non-interactive use. |
| `terminal-helper harness list\|use\|test\|add` | Manage harnesses. `use` takes `--command <bin>` to swap the binary of a preset (e.g. point claude at `claude-yolo`). `test` invokes directly without running detection. |
| `terminal-helper detect --line "<text>"` | Classify + act. Used by the shell hook; exit 127 = pass through. |
| `terminal-helper run --line "<text>" [--mode ...]` | Force a launch, skipping detection. |
| `terminal-helper mode set <mode>` | Shortcut for `config set default_mode <mode>`. |
| `terminal-helper config path\|edit\|get\|set` | Inspect or modify the config file. |
| `terminal-helper doctor` | Verify config, harness binary, and hook install. |
| `terminal-helper version` / `-v` / `--version` | Print version. |

### Using a wrapper binary

If you already have a wrapper like `claude-yolo` that calls
`claude --dangerously-skip-permissions` under the hood, swap just the binary
and keep the preset's args:

```sh
terminal-helper harness use claude --command claude-yolo
```

That flips active to `claude` and rewrites `command = "claude-yolo"` in both
the interactive and headless blocks (leaving `args` intact).

## Configuration

Lives at `~/.config/terminal-helper/config.toml` (or `$XDG_CONFIG_HOME/terminal-helper/config.toml`). Missing file is fine — presets for claude / aider / codex ship with the binary.

```toml
active_harness = "claude"
default_mode   = "headless"

[behavior]
confirm           = false
spinner           = true
passthrough       = ["gti", "sl"]
headless_fallback = "interactive"

[harness.claude]
interactive = { command = "claude", args = ["{prompt}"] }

[harness.claude.headless]
command = "claude"
args    = ["-p", "{prompt}"]
render  = "box"
```

Full schema is in [`spec.md`](./spec.md) §7.

## Design

- Detection runs at the shell's `command_not_found_handler` / `_handle` / `fish_command_not_found` hook. We never wrap normal command execution.
- The exit-code contract between the hook and the binary: **127 = pass through** (shell prints its usual "command not found"); **anything else = handled** (propagated as the child harness's exit code).
- Interactive mode uses `syscall.Exec` so the harness fully owns the tty.
- Headless mode spawns without a pty, streams stdout through a renderer, forwards stderr, and supports SIGINT forwarding (double-Ctrl-C escalates to SIGKILL) and `timeout_sec`.

## Development

```sh
go build ./cmd/terminal-helper
go test ./...
```

Design docs and milestone plans live in [`plans/`](./plans/). Start with [`spec.md`](./spec.md).

## Security

terminal-helper only passes the detected text as a string argument to a user-configured local binary. It never executes the detected text as a shell command, never makes network calls on its own, and has no telemetry.
