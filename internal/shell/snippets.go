package shell

import _ "embed"

//go:embed snippets/zsh.sh
var zshSnippet string

//go:embed snippets/bash.sh
var bashSnippet string

//go:embed snippets/fish.fish
var fishSnippet string

// ZshSnippet returns the zsh hook source users should eval from ~/.zshrc.
func ZshSnippet() string { return zshSnippet }

// BashSnippet returns the bash hook source users should eval from ~/.bashrc.
func BashSnippet() string { return bashSnippet }

// FishSnippet returns the fish hook source users should source from fish
// config (e.g. ~/.config/fish/config.fish).
func FishSnippet() string { return fishSnippet }
