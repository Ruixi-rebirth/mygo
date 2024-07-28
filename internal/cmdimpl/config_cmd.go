package cmdimpl

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/config"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
)

// configEntry describes a single config key with getter and setter.
type configEntry struct {
	get func(*config.Config) string
	set func(*config.Config, string)
}

var configKeys = map[string]configEntry{
	"build.ldflags": {
		get: func(c *config.Config) string { return c.Build.Ldflags },
		set: func(c *config.Config, v string) { c.Build.Ldflags = v },
	},
	"build.out-dir": {
		get: func(c *config.Config) string { return c.Build.OutDir },
		set: func(c *config.Config, v string) { c.Build.OutDir = v },
	},
	"test.timeout": {
		get: func(c *config.Config) string { return c.Test.Timeout },
		set: func(c *config.Config, v string) { c.Test.Timeout = v },
	},
	"test.race": {
		get: func(c *config.Config) string { return fmt.Sprintf("%v", c.Test.Race) },
		set: func(c *config.Config, v string) { c.Test.Race = v == "true" },
	},
	"project.name": {
		get: func(c *config.Config) string { return c.Project.Name },
		set: func(c *config.Config, v string) { c.Project.Name = v },
	},
	"project.version": {
		get: func(c *config.Config) string { return c.Project.Version },
		set: func(c *config.Config, v string) { c.Project.Version = v },
	},
}

var configListKeys = []struct {
	key string
	get func(*config.Config) string
}{
	{"build.ldflags", func(c *config.Config) string { return c.Build.Ldflags }},
	{"build.out-dir", func(c *config.Config) string { return c.Build.OutDir }},
	{"build.tags", func(c *config.Config) string { return fmt.Sprintf("%v", c.Build.Tags) }},
	{"test.timeout", func(c *config.Config) string { return c.Test.Timeout }},
	{"test.race", func(c *config.Config) string { return fmt.Sprintf("%v", c.Test.Race) }},
	{"lint.tool", func(c *config.Config) string { return c.Lint.Tool }},
	{"project.name", func(c *config.Config) string { return c.Project.Name }},
	{"project.version", func(c *config.Config) string { return c.Project.Version }},
	{"project.description", func(c *config.Config) string { return c.Project.Description }},
	{"bench.count", func(c *config.Config) string { return fmt.Sprintf("%d", c.Bench.Count) }},
	{"bench.benchmem", func(c *config.Config) string { return fmt.Sprintf("%v", c.Bench.BenchMem) }},
}

func ConfigGet(ctx context.Context, prj *project.Project, key string) error {
	cfg, err := loadProjectConfig(prj)
	if err != nil {
		return err
	}

	e, ok := configKeys[key]
	if !ok {
		return fmt.Errorf("unknown config key: %s (use 'mygo config list' to see all keys)", key)
	}
	fmt.Println(e.get(cfg))
	return nil
}

func ConfigSet(ctx context.Context, prj *project.Project, key, value string) error {
	e, ok := configKeys[key]
	if !ok {
		return fmt.Errorf("unknown config key: %s (use 'mygo config list' to see all keys)", key)
	}

	cfg, err := loadProjectConfig(prj)
	if err != nil {
		return err
	}

	e.set(cfg, value)

	cfgPath := filepath.Join(prj.Root, "mygo.toml")
	if err := config.Save(cfg, cfgPath); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	output.Successf("Set %s = %s", output.Bold(key), output.Cyan(value))
	return nil
}

func ConfigList(ctx context.Context, prj *project.Project) error {
	cfg, err := loadProjectConfig(prj)
	if err != nil {
		return err
	}

	fmt.Println()
	t := output.NewTable("Key", "Value")
	for _, k := range configListKeys {
		t.Row(k.key, k.get(cfg))
	}
	t.Print()
	return nil
}

func loadProjectConfig(prj *project.Project) (*config.Config, error) {
	cfgPath := filepath.Join(prj.Root, "mygo.toml")
	cfg, err := config.Load(cfgPath)
	if err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}
	return cfg, nil
}
