# 02 — CLI framework + `version`

## Description
Wire in Cobra so we have a real subcommand tree. Prove the wiring with a `version` subcommand — the smallest possible smoke test that both the framework and `main` are hooked up.

## Status
done

## Depends on
[01 — Project scaffold](01-project-scaffold.md)

## What / where to change

Add the dependency:
```
go get github.com/spf13/cobra@latest
```

Files to create/modify:

- `cmd/terminal-helper/root.go` — `NewRootCmd() *cobra.Command` returning a root command with `Use: "terminal-helper"`, short description matching the spec §1 one-liner, and no default action (print help on bare invocation).
- `cmd/terminal-helper/version.go` — `NewVersionCmd()` returning a `version` subcommand that prints a package-level `Version = "0.0.0-dev"` constant.
- `cmd/terminal-helper/main.go` — replace body with: build root, register `version`, `Execute()`. On error, exit 1.

The `Version` const should live at the top of `root.go` (or a new `version.go` constant file) so step 10+ can bump it from `goreleaser` metadata later.

## How to verify

```
go build ./cmd/terminal-helper
./terminal-helper                # prints help
./terminal-helper --help         # prints help
./terminal-helper version        # prints: 0.0.0-dev
./terminal-helper version; echo $?   # exit 0
./terminal-helper bogus          # exits non-zero with Cobra "unknown command" error
```

`go vet ./...` still clean.
