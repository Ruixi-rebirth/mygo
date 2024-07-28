// Package config manages mygo.toml configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config represents the full mygo.toml configuration.
type Config struct {
	Project  ProjectConfig  `toml:"project"`
	Build    BuildConfig    `toml:"build"`
	Run      RunConfig      `toml:"run"`
	Test     TestConfig     `toml:"test"`
	Lint     LintConfig     `toml:"lint"`
	Bench    BenchConfig    `toml:"bench"`
	Profiles ProfileSection `toml:"profile"`
}

// ProjectConfig contains project metadata.
type ProjectConfig struct {
	Name        string   `toml:"name"`
	Version     string   `toml:"version"`
	Description string   `toml:"description"`
	Authors     []string `toml:"authors"`
	License     string   `toml:"license"`
	GoVersion   string   `toml:"go-version"`
	Repository  string   `toml:"repository"`
}

// BuildConfig holds default build settings.
type BuildConfig struct {
	Ldflags string   `toml:"ldflags"`
	Tags    []string `toml:"tags"`
	Gcflags string   `toml:"gcflags"`
	OutDir  string   `toml:"out-dir"`
}

// RunConfig holds default run settings.
type RunConfig struct {
	Env map[string]string `toml:"env"`
}

// TestConfig holds default test settings.
type TestConfig struct {
	Timeout string `toml:"timeout"`
	Race    bool   `toml:"race"`
}

// LintConfig holds linting tool settings.
type LintConfig struct {
	Tool string `toml:"tool"`
}

// BenchConfig holds benchmark settings.
type BenchConfig struct {
	Timeout  string `toml:"timeout"`
	BenchMem bool   `toml:"benchmem"`
	Count    int    `toml:"count"`
}

// ProfileSection groups build profiles.
type ProfileSection struct {
	Release ProfileOpts `toml:"release"`
	Dev     ProfileOpts `toml:"dev"`
}

// ProfileOpts holds build options for a specific profile.
type ProfileOpts struct {
	Ldflags string            `toml:"ldflags"`
	Tags    []string          `toml:"tags"`
	Gcflags string            `toml:"gcflags"`
	Env     map[string]string `toml:"env"`
}

// Default returns a Config with sensible defaults.
func Default() *Config {
	return &Config{
		Test: TestConfig{
			Timeout: "30s",
		},
		Lint: LintConfig{
			Tool: "golangci-lint",
		},
		Bench: BenchConfig{
			Count: 1,
		},
		Profiles: ProfileSection{
			Release: ProfileOpts{
				Ldflags: "-s -w",
				Env:     map[string]string{"CGO_ENABLED": "0"},
			},
		},
	}
}

// Load reads a mygo.toml file from the given path.
func Load(path string) (*Config, error) {
	cfg := Default()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // no config = defaults
		}
		return nil, fmt.Errorf("read config: %w", err)
	}

	if err := toml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Find walks up from the current directory to find and load a mygo.toml.
func Find() (*Config, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		cfgPath := filepath.Join(dir, "mygo.toml")
		if _, err := os.Stat(cfgPath); err == nil {
			return Load(cfgPath)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return Default(), nil
		}
		dir = parent
	}
}

// Save writes a Config to a mygo.toml file, skipping zero-value sections.
func Save(cfg *Config, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	// Write a clean config by hand to avoid zero-value noise
	fmt.Fprintln(f, "# mygo.toml — configuration for mygo build tool")
	fmt.Fprintln(f)

	if cfg.Project.Name != "" || cfg.Project.Version != "" {
		fmt.Fprintln(f, "[project]")
		if cfg.Project.Name != "" {
			fmt.Fprintf(f, "name = %q\n", cfg.Project.Name)
		}
		if cfg.Project.Version != "" {
			fmt.Fprintf(f, "version = %q\n", cfg.Project.Version)
		}
		if cfg.Project.Description != "" {
			fmt.Fprintf(f, "description = %q\n", cfg.Project.Description)
		}
		if cfg.Project.License != "" {
			fmt.Fprintf(f, "license = %q\n", cfg.Project.License)
		}
		fmt.Fprintln(f)
	}

	if cfg.Build.Ldflags != "" || cfg.Build.OutDir != "" || len(cfg.Build.Tags) > 0 {
		fmt.Fprintln(f, "[build]")
		if cfg.Build.Ldflags != "" {
			fmt.Fprintf(f, "ldflags = %q\n", cfg.Build.Ldflags)
		}
		if cfg.Build.OutDir != "" {
			fmt.Fprintf(f, "out-dir = %q\n", cfg.Build.OutDir)
		}
		if len(cfg.Build.Tags) > 0 {
			fmt.Fprintf(f, "tags = %v\n", cfg.Build.Tags)
		}
		fmt.Fprintln(f)
	}

	if cfg.Test.Timeout != "" || cfg.Test.Race {
		fmt.Fprintln(f, "[test]")
		if cfg.Test.Timeout != "" {
			fmt.Fprintf(f, "timeout = %q\n", cfg.Test.Timeout)
		}
		if cfg.Test.Race {
			fmt.Fprintln(f, "race = true")
		}
		fmt.Fprintln(f)
	}

	fmt.Fprintln(f, "[profile.release]")
	fmt.Fprintf(f, "ldflags = %q\n", cfg.Profiles.Release.Ldflags)
	if len(cfg.Profiles.Release.Env) > 0 {
		fmt.Fprintln(f, "[profile.release.env]")
		for k, v := range cfg.Profiles.Release.Env {
			fmt.Fprintf(f, "%s = %q\n", k, v)
		}
	}

	return nil
}

// MergeProfile merges profile settings into build options.
func (c *Config) MergeProfile(profile string, ldflags string, tags []string) (string, []string) {
	var p ProfileOpts
	switch profile {
	case "release":
		p = c.Profiles.Release
	case "dev":
		p = c.Profiles.Dev
	default:
		return ldflags, tags
	}

	if ldflags == "" && p.Ldflags != "" {
		ldflags = p.Ldflags
	}
	if len(tags) == 0 && len(p.Tags) > 0 {
		tags = p.Tags
	}
	return ldflags, tags
}
