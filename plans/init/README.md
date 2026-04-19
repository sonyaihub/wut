# Plan: `init` — walking skeleton

Goal: stand up the minimum end-to-end path so that typing natural language at a zsh prompt routes to a stub harness, while real commands and typos behave normally. This plan covers M0 from `spec.md`; everything else (real harness launch, config, headless mode) builds on what lands here.

Work the steps in order. Each step has a `Status` field — update it from `pending` → `in progress` → `done` as you go.

| # | Step | Status |
|---|---|---|
| 01 | [Project scaffold](01-project-scaffold.md) | done |
| 02 | [CLI framework + `version`](02-cli-framework.md) | done |
| 03 | [`init zsh` command](03-init-zsh-command.md) | done |
| 04 | [`detect` stub](04-detect-stub.md) | done |
| 05 | [Shell hook install](05-shell-hook-install.md) | done |
| 06 | [Detection heuristic](06-heuristic.md) | done |
| 07 | [Echo harness](07-echo-harness.md) | done |
| 08 | [End-to-end verification](08-end-to-end.md) | done |

Exit criteria for the whole plan: at a fresh zsh prompt, `how do I rebase onto main` invokes the echo harness; `ls` runs normally; `gti` shows `command not found`.
