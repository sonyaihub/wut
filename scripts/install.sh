#!/bin/sh
# terminal-helper installer.
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/<owner>/terminal-helper/main/scripts/install.sh | sh
#
# Requires: Go 1.22+.
#
# For release-binary installs once we publish to a tap or GitHub releases,
# this script will switch to a download-and-verify flow. For now it uses
# `go install` so users only need a Go toolchain.

set -eu

REPO="github.com/sonyatalona/terminal-helper"
PKG="$REPO/cmd/terminal-helper@latest"

log()  { printf "==> %s\n" "$*" >&2; }
warn() { printf "!!  %s\n" "$*" >&2; }
die()  { printf "xx  %s\n" "$*" >&2; exit 1; }

command -v go >/dev/null 2>&1 || die "Go toolchain not found. Install Go 1.22+ from https://go.dev/dl/"

log "installing terminal-helper via go install…"
GOBIN_DIR="$(go env GOBIN)"
if [ -z "$GOBIN_DIR" ]; then
  GOBIN_DIR="$(go env GOPATH)/bin"
fi

go install "$PKG"

BIN="$GOBIN_DIR/terminal-helper"
if [ ! -x "$BIN" ]; then
  die "expected binary at $BIN after go install, but it is missing"
fi

log "installed: $BIN"

# Warn loudly if the install dir isn't on PATH — the shell hook will recurse
# if terminal-helper isn't resolvable.
case ":$PATH:" in
  *":$GOBIN_DIR:"*) ;;
  *)
    warn "$GOBIN_DIR is not on your \$PATH."
    warn "add this to your shell rc file:"
    warn "  export PATH=\"$GOBIN_DIR:\$PATH\""
    ;;
esac

cat >&2 <<NEXT

Next steps:

  1. Make sure terminal-helper is on your \$PATH (see above if warned).
  2. Pick a harness and default mode:
       terminal-helper setup
  3. Wire the shell hook. Pick your shell:
       # zsh:   add to ~/.zshrc
       eval "\$(terminal-helper init zsh)"
       # bash:  add to ~/.bashrc
       eval "\$(terminal-helper init bash)"
       # fish:  save to ~/.config/fish/conf.d/terminal-helper.fish
       terminal-helper init fish > ~/.config/fish/conf.d/terminal-helper.fish
  4. Open a new shell, then verify:
       terminal-helper doctor

NEXT
