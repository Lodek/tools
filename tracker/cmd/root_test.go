package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultRootUsesTrackerRootEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TRACKER_ROOT", root)

	if got := defaultRoot(); got != root {
		t.Fatalf("defaultRoot() = %q, want %q", got, root)
	}
}

func TestCommandUsesTrackerRootEnv(t *testing.T) {
	root := t.TempDir()
	t.Setenv("TRACKER_ROOT", root)

	a := &app{root: defaultRoot()}
	cmd := a.rootCommand()
	cmd.SetArgs([]string{"init"})
	cmd.SetOut(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(init) error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(root, "habits", "active")); err != nil {
		t.Fatalf("env root was not initialized: %v", err)
	}
}

func TestRootFlagOverridesTrackerRootEnv(t *testing.T) {
	envRoot := t.TempDir()
	flagRoot := t.TempDir()
	t.Setenv("TRACKER_ROOT", envRoot)

	a := &app{root: defaultRoot()}
	cmd := a.rootCommand()
	cmd.SetArgs([]string{"--root", flagRoot, "init"})
	cmd.SetOut(&bytes.Buffer{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute(init --root) error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(flagRoot, "habits", "active")); err != nil {
		t.Fatalf("flag root was not initialized: %v", err)
	}
	if _, err := os.Stat(filepath.Join(envRoot, "habits", "active")); !os.IsNotExist(err) {
		t.Fatalf("env root was initialized despite --root override; stat err = %v", err)
	}
}
