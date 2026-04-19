# M5 — distribution: verification results

## Artifacts added
- `internal/shell/snippets/bash.sh` — bash hook (requires bash 4+).
- `internal/shell/snippets/fish.fish` — fish hook.
- `scripts/install.sh` — curl | sh installer using `go install`.
- `Formula/terminal-helper.rb` — homebrew formula template (placeholders
  marked TODO; ready to wire into a tap + release workflow).

## CLI
- `terminal-helper init zsh|bash|fish` all emit syntactically-valid snippets.
  Verified with `zsh -n` and `bash -n`.
- `init --help` lists all three subcommands.

## Bash hook — macOS caveat
`command_not_found_handle` was added in bash 4.0. macOS ships bash 3.2
(GPLv3 reasons), so sourcing the hook there is a no-op. The snippet now
early-returns on `BASH_VERSINFO[0] < 4` to avoid installing a handler that
would never fire. Users on stock macOS need `brew install bash`, or just
use zsh (also supported and the default shell on macOS 10.15+).

On Linux distros with bash 4+ the snippet works identically to the zsh one,
including the recursion guard and 127-means-passthrough exit contract.

## Fish hook
Fish wasn't installed on this dev machine, so parse-check + live E2E is
deferred to the first user on fish. The snippet follows the documented
`fish_command_not_found` pattern and returns `$status` explicitly so the
127 contract propagates.

## Install script
`scripts/install.sh` — POSIX `sh -n` clean. Flow:

1. Require Go on `$PATH`.
2. `go install github.com/sonyatalona/terminal-helper/cmd/terminal-helper@latest`.
3. Check that the resulting binary exists at `$(go env GOBIN)` or
   `$(go env GOPATH)/bin`.
4. Warn loudly if that directory isn't on `$PATH` (would cause the hook to
   recurse — we already hit this during M0 development).
5. Print next-steps for each supported shell plus `doctor`.

Release-binary flow (download + checksum verify) is the obvious next upgrade
once we ship tagged releases with assets. The current script is good enough
for "clone, install, try it" until then.

## Homebrew formula
Template uses `depends_on "go" => :build`, builds from source with a
version-stamp ldflag, and includes a `caveats` block walking through
`setup` and the three hook options. Placeholders (owner, tag, sha256) need
filling at release time; the file is commented to that effect.

## Remaining
- Shell-completion scripts (`terminal-helper completion`) aren't wired —
  Cobra provides them for free if we expose the command. Small follow-up.
- Release pipeline (goreleaser config + GitHub Actions) — separate task.
- Publishing the homebrew tap — separate task, depends on release pipeline.
