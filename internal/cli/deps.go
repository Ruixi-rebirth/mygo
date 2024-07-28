package cli

import (
	"fmt"
	"path/filepath"

	"github.com/Ruixi-rebirth/mygo/internal/cmdimpl"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/spf13/cobra"
)

// ---- Dependency Commands ----

func AddCmd(r gocmd.Runner) *cobra.Command {
	var version string
	var noTidy bool

	cmd := &cobra.Command{
		Use:   "add <package>",
		Short: "Add a dependency (go get + tidy)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.AddDependency(cobraCmd.Context(), r, prj, args[0], version, !noTidy)
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "Version (e.g., v1.2.3 or latest)")
	cmd.Flags().BoolVar(&noTidy, "no-tidy", false, "Skip go mod tidy")

	return cmd
}

func RemoveCmd(r gocmd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "remove <package>",
		Short: "Remove a dependency",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.RemoveDependency(cobraCmd.Context(), r, prj, args[0])
		},
	}
}

func UpdateCmd(r gocmd.Runner) *cobra.Command {
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "update [packages...]",
		Short: "Update dependencies (no args = all)",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			all := len(args) == 0
			return cmdimpl.UpdateDependencies(cobraCmd.Context(), r, prj, args, dryRun, all)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would change without applying")

	return cmd
}

func TreeCmd(r gocmd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "tree",
		Short: "Print a visual dependency tree",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.DepTree(cobraCmd.Context(), r, prj)
		},
	}
}

func OutdatedCmd(r gocmd.Runner) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "outdated",
		Short: "List dependencies with newer versions available",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.OutdatedDeps(cobraCmd.Context(), r, prj, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "JSON output")

	return cmd
}

func SearchCmd(r gocmd.Runner) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search GitHub for Go repositories",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmdimpl.SearchPackages(cobraCmd.Context(), args[0], limit)
		},
	}

	cmd.Flags().IntVarP(&limit, "limit", "n", 10, "Max results")

	return cmd
}

// ---- Advanced Commands ----

func UninstallCmd(r gocmd.Runner) *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall <binary-name>",
		Short: "Remove a go-installed binary from GOBIN",
		Args:  cobra.ExactArgs(1),
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			return cmdimpl.UninstallBinary(cobraCmd.Context(), r, args[0])
		},
	}
}

func BenchCmd(r gocmd.Runner) *cobra.Command {
	var count int
	var benchMem bool
	var timeout string

	cmd := &cobra.Command{
		Use:   "bench [pattern]",
		Short: "Run benchmarks",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			pattern := "."
			if len(args) > 0 {
				pattern = args[0]
			}
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.Bench(cobraCmd.Context(), r, prj, pattern, count, benchMem, timeout)
		},
	}

	cmd.Flags().IntVar(&count, "count", 1, "Iterations")
	cmd.Flags().BoolVar(&benchMem, "benchmem", false, "Show memory stats")
	cmd.Flags().StringVar(&timeout, "timeout", "", "Timeout")

	return cmd
}

func AuditCmd(r gocmd.Runner) *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "audit",
		Short: "Check dependencies for known vulnerabilities (govulncheck)",
		RunE: func(cobraCmd *cobra.Command, args []string) error {
			prj, _, err := loadProjectAndConfig()
			if err != nil {
				return err
			}
			return cmdimpl.Audit(cobraCmd.Context(), r, prj, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "JSON output")

	return cmd
}

func WorkspaceCmd(r gocmd.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workspace",
		Short: "Manage Go workspace (go.work)",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "init [dirs...]",
			Short: "Create a new workspace",
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				return cmdimpl.WorkspaceInit(cobraCmd.Context(), r, args)
			},
		},
		&cobra.Command{
			Use:   "add <dirs...>",
			Short: "Add modules to the workspace",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				return cmdimpl.WorkspaceAdd(cobraCmd.Context(), r, args)
			},
		},
		&cobra.Command{
			Use:   "remove <dirs...>",
			Short: "Remove modules from the workspace",
			Args:  cobra.MinimumNArgs(1),
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				return cmdimpl.WorkspaceRemove(cobraCmd.Context(), r, args)
			},
		},
		&cobra.Command{
			Use:   "sync",
			Short: "Sync workspace dependencies",
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				return cmdimpl.WorkspaceSync(cobraCmd.Context(), r)
			},
		},
	)

	return cmd
}

func ConfigCmd(r gocmd.Runner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage mygo.toml",
	}

	cmd.AddCommand(
		&cobra.Command{
			Use:   "path",
			Short: "Show config file path",
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				prj, _, err := loadProjectAndConfig()
				if err != nil {
					return err
				}
				fmt.Println(filepath.Join(prj.Root, "mygo.toml"))
				return nil
			},
		},
		&cobra.Command{
			Use:   "get <key>",
			Short: "Get a config value",
			Args:  cobra.ExactArgs(1),
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				prj, _, err := loadProjectAndConfig()
				if err != nil {
					return err
				}
				return cmdimpl.ConfigGet(cobraCmd.Context(), prj, args[0])
			},
		},
		&cobra.Command{
			Use:                "set <key> <value>",
			Short:              "Set a config value",
			Args:               cobra.ExactArgs(2),
			DisableFlagParsing: true,
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				prj, _, err := loadProjectAndConfig()
				if err != nil {
					return err
				}
				return cmdimpl.ConfigSet(cobraCmd.Context(), prj, args[0], args[1])
			},
		},
		&cobra.Command{
			Use:   "list",
			Short: "List all config values",
			RunE: func(cobraCmd *cobra.Command, args []string) error {
				prj, _, err := loadProjectAndConfig()
				if err != nil {
					return err
				}
				return cmdimpl.ConfigList(cobraCmd.Context(), prj)
			},
		},
	)

	return cmd
}
