package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultsValidates(t *testing.T) {
	cfg := Defaults()
	if err := cfg.Validate(); err != nil {
		t.Fatalf("Defaults().Validate() = %v", err)
	}
	if _, err := cfg.ActiveInvocation(ModeInteractive); err != nil {
		t.Fatalf("ActiveInvocation(interactive) on defaults: %v", err)
	}
}

func TestLoadMissingReturnsDefaults(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.toml"))
	if err != nil {
		t.Fatalf("Load on missing file: %v", err)
	}
	if cfg.ActiveHarness != "claude" {
		t.Fatalf("want default claude, got %q", cfg.ActiveHarness)
	}
}

func TestLoadOverrides(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
active_harness = "my-agent"
default_mode   = "interactive"

[harness.my-agent]
interactive = { command = "my-agent", args = ["--tui", "{prompt}"] }
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ActiveHarness != "my-agent" {
		t.Fatalf("active_harness = %q", cfg.ActiveHarness)
	}
	inv, err := cfg.ActiveInvocation(ModeInteractive)
	if err != nil {
		t.Fatalf("ActiveInvocation: %v", err)
	}
	if inv.Command != "my-agent" || len(inv.Args) != 2 || inv.Args[1] != "{prompt}" {
		t.Fatalf("unexpected invocation: %+v", inv)
	}
}

func TestValidateRejectsUnknownActive(t *testing.T) {
	cfg := &Config{ActiveHarness: "ghost", Harness: map[string]Harness{}}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for unknown active_harness")
	}
}
