// Package cli defines cobra commands for mygo.
package cli

import (
	"github.com/Ruixi-rebirth/mygo/internal/cmdimpl"
	"github.com/Ruixi-rebirth/mygo/pkg/config"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
	"github.com/spf13/cobra"
)

// splitAppArgs splits args into [appName, binArgs].
// "myapp" → app="myapp", extra=[]
// "myapp arg1 arg2" → app="myapp", extra=["arg1","arg2"]
// "" → app="", extra=[]
func splitAppArgs(args []string) (app string, extra []string) {
	for i, a := range args {
		if a == "--" {
			return "", args[i+1:]
		}
	}
	if len(args) > 0 {
		return args[0], args[1:]
	}
	return "", nil
}

func loadProjectAndConfig() (*project.Project, *config.Config, error) {
	prj, err := project.Find()
	if err != nil {
		return nil, nil, err
	}
	cfg, err := config.Find()
	if err != nil {
		return nil, nil, err
	}
	return prj, cfg, nil
}

// RootCmd returns the root command with dependencies injected.
func RootCmd(runner gocmd.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mygo",
		Short: "A Cargo-like build tool for Go projects",
		Long: `mygo augments the Go toolchain with project scaffolding, dependency
management, and quality-of-life features that go doesn't provide.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		VersionCmd(runner),
		NewCmd(runner),
		InitCmd(runner),
		BuildCmd(runner),
		RunCmd(runner),
		TestCmd(runner),
		CheckCmd(runner),
		CleanCmd(runner),
		AddCmd(runner),
		RemoveCmd(runner),
		UpdateCmd(runner),
		TreeCmd(runner),
		OutdatedCmd(runner),
		SearchCmd(runner),
		UninstallCmd(runner),
		BenchCmd(runner),
		AuditCmd(runner),
		WorkspaceCmd(runner),
		ConfigCmd(runner),
	)

	return cmd
}

// ---- Project Commands ----

func VersionCmd(r gocmd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdimpl.Version(cmd.Context(), r)
		},
	}
}

func NewCmd(r gocmd.Runner) *cobra.Command {
	var name, modulePrefix string
	var isLib bool

	cmd := &cobra.Command{
		Use:   "new [path]",
		Short: "Create a new Go project",
		Long: `Scaffold a new Go project with go.mod, source files, git, and mygo.toml.

  mygo new hello          # binary (default)
  mygo new hello --lib    # library`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				name = args[0]
			}
			return cmdimpl.NewProject(cobraCmd.Context(), r, name, modulePrefix, isLib)
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Project name")
	cmd.Flags().StringVar(&modulePrefix, "prefix", "", "Module prefix (e.g., github.com/username)")
	cmd.Flags().BoolVar(&isLib, "lib", false, "Create a library instead of a binary")

	return cmd
}

func InitCmd(r gocmd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a mygo.toml in the current project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmdimpl.InitProject(cmd.Context(), r)
		},
	}
}

func BuildCmd(r gocmd.Runner) *cobra.Command {
	var release bool
	var target, output string
	var verbose, trimpath bool
	var buildmode, modFlag string

	cmd := &cobra.Command{
		Use:   "build",
		Short: "Compile the project",
		Long: `Build the project with profile-driven config and cross-compilation.

Targets: linux/amd64, darwin/arm64, windows/amd64, etc.
Build modes: pie, plugin, c-shared, c-archive`,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, cfg, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			opts := gocmd.BuildOpts{
				Output:    output,
				Verbose:   verbose,
				Trimpath:  trimpath,
				Buildmode: buildmode,
				Mod:       modFlag,
			}
			return cmdimpl.Build(cobraCmd.Context(), r, prj, cfg, opts, release, target)
		},
	}

	cmd.Flags().BoolVar(&release, "release", false, "Build with release profile")
	cmd.Flags().StringVar(&target, "target", "", "Cross-compile target (e.g., linux/amd64)")
	cmd.Flags().StringVarP(&output, "output", "o", "", "Output binary path")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().BoolVar(&trimpath, "trimpath", false, "Strip file paths from binary")
	cmd.Flags().StringVar(&buildmode, "buildmode", "", "Build mode: pie, plugin, c-shared")
	cmd.Flags().StringVar(&modFlag, "mod", "", "Module mode: readonly, vendor")

	return cmd
}

func RunCmd(r gocmd.Runner) *cobra.Command {
	var release bool

	cmd := &cobra.Command{
		Use:   "run [app] [-- args...]",
		Short: "Build and run the project",
		Long: `Run the project's main package.

  mygo run                  # run root main.go
  mygo run --release        # run with release profile
  mygo run myapp            # run ./cmd/myapp
  mygo run myapp arg1 arg2  # run myapp, pass args to binary
  mygo run -- arg1 arg2     # pass args directly, no app`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, cfg, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			appName, extra := splitAppArgs(args)
			return cmdimpl.Run(cobraCmd.Context(), r, prj, cfg, appName, extra, release)
		},
	}

	cmd.Flags().BoolVar(&release, "release", false, "Run with release profile")
	cmd.Flags().SetInterspersed(false)

	return cmd
}

func TestCmd(r gocmd.Runner) *cobra.Command {
	opts := gocmd.TestOpts{}

	cmd := &cobra.Command{
		Use:   "test",
		Short: "Run tests with coverage and race detection",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, cfg, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.Test(cobraCmd.Context(), r, prj, cfg, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.Verbose, "verbose", "v", false, "Verbose output")
	cmd.Flags().BoolVar(&opts.Cover, "cover", false, "Enable coverage")
	cmd.Flags().StringVar(&opts.CoverProfile, "coverprofile", "", "Write coverage profile to file")
	cmd.Flags().StringVar(&opts.CoverMode, "covermode", "", "Coverage mode: set, count, atomic")
	cmd.Flags().BoolVar(&opts.Race, "race", false, "Enable race detector")
	cmd.Flags().StringVar(&opts.Timeout, "timeout", "", "Test timeout")
	cmd.Flags().StringVar(&opts.Bench, "bench", "", "Run benchmarks matching pattern")
	cmd.Flags().BoolVar(&opts.BenchMem, "benchmem", false, "Show memory allocs")
	cmd.Flags().StringVar(&opts.Run, "run", "", "Run specific test pattern")
	cmd.Flags().IntVar(&opts.Count, "count", 0, "Repeat tests N times")
	cmd.Flags().BoolVar(&opts.FailFast, "fail-fast", false, "Stop on first failure")
	cmd.Flags().BoolVar(&opts.Short, "short", false, "Skip long-running tests")
	cmd.Flags().IntVar(&opts.Parallel, "parallel", 0, "Parallel test count")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "JSON output")

	return cmd
}

func CheckCmd(r gocmd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check that the project compiles (no binary output)",
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.Check(cmd.Context(), r, prj)
		},
	}
}

func CleanCmd(r gocmd.Runner) *cobra.Command {
	var cacheOnly bool

	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Remove build artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.Clean(cmd.Context(), r, prj, cacheOnly)
		},
	}

	cmd.Flags().BoolVar(&cacheOnly, "cache", false, "Also clean module and test caches")

	return cmd
}
