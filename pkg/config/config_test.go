package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg.Test.Timeout != "30s" {
		t.Errorf("Test.Timeout = %q, want %q", cfg.Test.Timeout, "30s")
	}
	if cfg.Lint.Tool != "golangci-lint" {
		t.Errorf("Lint.Tool = %q, want %q", cfg.Lint.Tool, "golangci-lint")
	}
	if cfg.Profiles.Release.Ldflags != "-s -w" {
		t.Errorf("Profiles.Release.Ldflags = %q, want %q", cfg.Profiles.Release.Ldflags, "-s -w")
	}
	if cfg.Profiles.Release.Env["CGO_ENABLED"] != "0" {
		t.Errorf("Profiles.Release.Env[CGO_ENABLED] = %q, want %q", cfg.Profiles.Release.Env["CGO_ENABLED"], "0")
	}
}

func TestLoadSave(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "mygo.toml")

	// Loading non-existent file returns defaults
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load of non-existent file should not error: %v", err)
	}
	if cfg.Test.Timeout != "30s" {
		t.Errorf("default Test.Timeout = %q, want %q", cfg.Test.Timeout, "30s")
	}

	// Save and reload
	cfg.Project.Name = "testproj"
	cfg.Project.Version = "1.0.0"
	cfg.Build.Ldflags = "-X main.version=1.0.0"
	cfg.Test.Timeout = "60s"

	if err := Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Reload
	loaded, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}
	if loaded.Project.Name != "testproj" {
		t.Errorf("Project.Name = %q, want %q", loaded.Project.Name, "testproj")
	}
	if loaded.Project.Version != "1.0.0" {
		t.Errorf("Project.Version = %q, want %q", loaded.Project.Version, "1.0.0")
	}
	if loaded.Build.Ldflags != "-X main.version=1.0.0" {
		t.Errorf("Build.Ldflags = %q", loaded.Build.Ldflags)
	}
	if loaded.Test.Timeout != "60s" {
		t.Errorf("Test.Timeout = %q, want %q", loaded.Test.Timeout, "60s")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "mygo.toml")

	cfg := Default()
	cfg.Project.Name = "hello"
	cfg.Project.Description = "A test project"
	cfg.Build.OutDir = "bin"

	if err := Save(cfg, cfgPath); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	content := string(data)
	if content == "" {
		t.Error("Saved config is empty")
	}
}
