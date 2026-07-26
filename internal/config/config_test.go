package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadCreatesGlobalConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Codex.Version != "latest" {
		t.Fatalf("codex version = %q, want latest", cfg.Codex.Version)
	}

	path := filepath.Join(homeDir, DirName, FileName)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat config: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("config permissions = %o, want 600", got)
	}
}

func TestSaveAndLoadGlobalConfig(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	want := Default()
	want.Codex.Version = "1.2.3"
	if err := Save(want); err != nil {
		t.Fatalf("save config: %v", err)
	}

	got, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if got.Codex.Version != want.Codex.Version {
		t.Fatalf("codex version = %q, want %q", got.Codex.Version, want.Codex.Version)
	}
}

func TestSaveRejectsControlCharactersInMount(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	cfg := Default()
	cfg.Mounts = []Mount{{
		Source: "\x1b[B",
		Target: "/mnt/bad",
		Mode:   "rw",
	}}

	if err := Save(cfg); err == nil {
		t.Fatal("save config with control character succeeded")
	}
}

func TestLoadRemovesCorruptedMount(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	path := filepath.Join(homeDir, DirName, FileName)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("create config directory: %v", err)
	}

	data := []byte("mounts:\n  - source: \"\\e[B\"\n    target: /mnt/bad\n    mode: rw\n")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if len(cfg.Mounts) != 0 {
		t.Fatalf("mount count = %d, want 0", len(cfg.Mounts))
	}

	saved, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read cleaned config: %v", err)
	}
	if strings.Contains(string(saved), "\x1b") {
		t.Fatal("cleaned config still contains an escape character")
	}
}
