# typed: false
# frozen_string_literal: true
#
# Homebrew formula for terminal-helper.
#
# This is a TEMPLATE — the placeholders marked TODO must be filled in from
# release tooling (goreleaser / release workflow) before this ships. Do not
# publish as-is.

class TerminalHelper < Formula
  desc "Route accidental natural-language shell input to your AI harness"
  homepage "https://github.com/sonyatalona/terminal-helper"
  license "MIT" # TODO: confirm or change

  # Source build — works today via `brew install --build-from-source`.
  url "https://github.com/sonyatalona/terminal-helper/archive/refs/tags/v0.0.0.tar.gz" # TODO: update per release
  sha256 "0000000000000000000000000000000000000000000000000000000000000000"          # TODO: update per release
  version "0.0.0"

  depends_on "go" => :build

  def install
    system "go", "build",
           "-ldflags", "-s -w -X main.Version=#{version}",
           "-o", bin/"terminal-helper",
           "./cmd/terminal-helper"

    # Shell-completion scripts once we add them (see `terminal-helper completion`).
    # generate_completions_from_executable(bin/"terminal-helper", "completion")
  end

  def caveats
    <<~EOS
      To finish setup:
        terminal-helper setup

      Install the shell hook (pick one):
        # zsh
        echo 'eval "$(terminal-helper init zsh)"' >> ~/.zshrc
        # bash
        echo 'eval "$(terminal-helper init bash)"' >> ~/.bashrc
        # fish
        terminal-helper init fish > ~/.config/fish/conf.d/terminal-helper.fish

      Then open a new shell and run:
        terminal-helper doctor
    EOS
  end

  test do
    assert_match "terminal-helper", shell_output("#{bin}/terminal-helper --help")
    assert_match version.to_s, shell_output("#{bin}/terminal-helper version")
  end
end
