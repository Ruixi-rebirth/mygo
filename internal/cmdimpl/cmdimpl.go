// Package cmdimpl implements business logic for mygo commands.
// Each function takes a gocmd.Runner interface for testability.
package cmdimpl

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/config"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
)

// Version prints version information.
func Version(ctx context.Context, g gocmd.Runner) error {
	goVer, err := g.Version(ctx)
	if err != nil {
		return err
	}

	myVer := "(devel)"

	output.Banner("mygo")
	output.Printf(output.Info, "%s %s", output.Bold("mygo"), output.Cyan(myVer))
	output.Printf(output.Info, "%s", output.Dim(goVer))

	prj, err := project.Find()
	if err == nil {
		output.Printf(output.Info, "%s: %s (go %s)",
			output.Dim("project"), output.Bold(prj.Module), output.Dim(prj.GoVer))
		if prj.IsWork {
			output.Printf(output.Info, "%s", output.Yellow("workspace mode"))
		}
	}
	return nil
}

// NewProject scaffolds a new Go project.
func NewProject(ctx context.Context, g gocmd.Runner, name, modulePrefix string, isLib bool) error {
	if name == "" {
		return fmt.Errorf("project name is required; use --name NAME")
	}

	if !isValidProjectName(name) {
		return fmt.Errorf("invalid project name %q; must match [a-zA-Z0-9_.\\-/]+", name)
	}

	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory %q already exists", name)
	}

	modulePath := name
	if modulePrefix != "" {
		modulePath = modulePrefix + "/" + name
	}

	kind := "binary"
	if isLib {
		kind = "library"
	}

	output.Header("Creating %s project: %s", kind, output.Bold(name))
	fmt.Println()

	// 1. Directory
	output.Step("Initialize directory")
	if err := os.Mkdir(name, 0755); err != nil {
		output.StepFail("Initialize directory")
		return fmt.Errorf("create directory: %w", err)
	}

	absDir, err := filepath.Abs(name)
	if err != nil {
		return err
	}

	// 2. Module
	output.Step("Initialize go module: %s", output.Dim(modulePath))
	if err := g.ModInit(ctx, absDir, modulePath); err != nil {
		output.StepFail("Initialize go module")
		return fmt.Errorf("initialize module: %w", err)
	}

	// 3. Source files
	output.Step("Create source files")
	data := templateData{
		Name:   sanitizePkgName(name),
		Module: modulePath,
	}
	if isLib {
		if err := scaffoldLibrary(absDir, data); err != nil {
			output.StepFail("Create source files")
			return err
		}
	} else {
		if err := scaffoldBinary(absDir, data); err != nil {
			output.StepFail("Create source files")
			return err
		}
	}

	// 4. Config
	output.Step("Create mygo.toml")
	cfg := config.Default()
	cfg.Project.Name = name
	cfg.Project.Version = "0.1.0"
	cfg.Build.OutDir = "bin"
	if err := config.Save(cfg, filepath.Join(absDir, "mygo.toml")); err != nil {
		output.Warnf("Could not create mygo.toml: %v", err)
	}

	// 5. .gitignore
	gitignore := "/bin/\n*.exe\n*.test\n*.out\n"
	_ = os.WriteFile(filepath.Join(absDir, ".gitignore"), []byte(gitignore), 0644)

	// 6. Git
	output.Step("Initialize git repository")
	if err := initGit(ctx, absDir); err != nil {
		output.Warnf("Could not initialize git: %v", err)
	}

	fmt.Println()
	output.Successf("Project %s created", output.Bold(name))
	fmt.Println()
	output.Printf(output.Info, "%s", output.Bold("Next steps:"))
	output.Printf(output.Info, "  cd %s", name)
	if isLib {
		output.Printf(output.Info, "  # Write your code in %s.go", data.Name)
	} else {
		output.Printf(output.Info, "  mygo run")
		output.Printf(output.Info, "  mygo build --release")
	}
	output.Printf(output.Info, "  mygo test")

	return nil
}

func scaffoldBinary(dir string, data templateData) error {
	return copyTemplate(
		filepath.Join(dir, "main.go"),
		"templates/binary/main.go.tmpl",
	)
}

func scaffoldLibrary(dir string, data templateData) error {
	if err := writeTemplate(
		filepath.Join(dir, data.Name+".go"),
		"templates/library/lib.go.tmpl",
		data,
	); err != nil {
		return err
	}
	return writeTemplate(
		filepath.Join(dir, data.Name+"_test.go"),
		"templates/library/lib_test.go.tmpl",
		data,
	)
}

// copyTemplate copies a file from the embedded filesystem.
func copyTemplate(dst, src string) error {
	data, err := templateFS.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read template %s: %w", src, err)
	}
	return os.WriteFile(dst, data, 0644)
}

// writeTemplate writes a template file with data substitution.
func writeTemplate(dst, src string, data templateData) error {
	tmplData, err := templateFS.ReadFile(src)
	if err != nil {
		return fmt.Errorf("read template %s: %w", src, err)
	}

	content := string(tmplData)
	content = strings.ReplaceAll(content, "{{.Name}}", data.Name)
	content = strings.ReplaceAll(content, "{{.Module}}", data.Module)

	return os.WriteFile(dst, []byte(content), 0644)
}

// InitProject initializes mygo config in an existing directory.
func InitProject(ctx context.Context, g gocmd.Runner) error {
	prj, err := project.Find()
	if err != nil {
		return fmt.Errorf("not in a Go project: %w", err)
	}

	if prj.HasMygoTOML() {
		output.Warnf("mygo.toml already exists in %s", prj.Root)
		return nil
	}

	cfg := config.Default()
	cfg.Project.Name = filepath.Base(prj.Root)
	cfg.Project.GoVersion = prj.GoVer

	cfgPath := filepath.Join(prj.Root, "mygo.toml")
	if err := config.Save(cfg, cfgPath); err != nil {
		return fmt.Errorf("create mygo.toml: %w", err)
	}

	output.Successf("Created mygo.toml in %s", prj.Root)
	return nil
}

// Build compiles the project.
func Build(ctx context.Context, g gocmd.Runner, prj *project.Project, cfg *config.Config, opts gocmd.BuildOpts, release bool, target string) error {
	// Apply profile if --release
	if release {
		opts.Ldflags, opts.Tags = cfg.MergeProfile("release", opts.Ldflags, opts.Tags)
	}

	// Parse --target (os/arch)
	if target != "" {
		parts := strings.Split(target, "/")
		if len(parts) >= 2 {
			opts.GOOS = parts[0]
			opts.GOARCH = parts[1]
		} else {
			opts.GOOS = target
		}
	}

	// Set default output name
	if opts.Output == "" {
		binName := filepath.Base(prj.Root)
		if cfg.Build.OutDir != "" {
			opts.Output = filepath.Join(cfg.Build.OutDir, binName)
		} else {
			opts.Output = binName
		}
	}
	if opts.GOOS == "windows" {
		opts.Output += ".exe"
	}

	// Apply config defaults
	if opts.Ldflags == "" {
		opts.Ldflags = cfg.Build.Ldflags
	}

	spinner := output.NewSpinner("Building...")
	spinner.Start()

	result, err := g.Build(ctx, prj.Root, opts)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("build failed: %w", err)
	}

	output.Successf("Built %s (%s) in %s",
		output.Bold(result.Binary),
		output.Dim(formatSize(result.Size)),
		output.Dim(result.Duration.Round(time.Millisecond).String()),
	)
	return nil
}

// Run executes the project.
func Run(ctx context.Context, g gocmd.Runner, prj *project.Project, cfg *config.Config, appName string, args []string, release bool) error {
	// Determine which package to run
	pkg := "."
	if appName != "" {
		pkg = "./cmd/" + appName
		// Verify it exists
		mainFile := filepath.Join(prj.Root, "cmd", appName, "main.go")
		if _, err := os.Stat(mainFile); os.IsNotExist(err) {
			return fmt.Errorf("no main package at cmd/%s/main.go", appName)
		}
	} else {
		// Check root main.go
		if _, err := os.Stat(filepath.Join(prj.Root, "main.go")); os.IsNotExist(err) {
			// Try to find a main package in cmd/
			mains, err := prj.MainPackages()
			if err != nil || len(mains) == 0 {
				return fmt.Errorf("no main package found; create main.go or specify an app")
			}
			if len(mains) == 1 {
				pkg = mains[0]
			} else {
				return fmt.Errorf("multiple main packages found: %v; specify an app name", mains)
			}
		}
	}

	runOpts := gocmd.RunOpts{}
	if release {
		ldflags, tags := cfg.MergeProfile("release", "", nil)
		_ = ldflags // run doesn't use ldflags
		_ = tags
	}

	output.Infof("Running %s", output.Bold(pkg))
	return g.Run(ctx, prj.Root, pkg, args, runOpts)
}

// Test runs tests for the project.
func Test(ctx context.Context, g gocmd.Runner, prj *project.Project, cfg *config.Config, opts gocmd.TestOpts) error {
	// Apply config defaults
	if opts.Timeout == "" {
		opts.Timeout = cfg.Test.Timeout
	}
	if !opts.Race {
		opts.Race = cfg.Test.Race
	}

	spinner := output.NewSpinner("Running tests...")
	spinner.Start()

	result, err := g.Test(ctx, prj.Root, opts)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("test failed: %w", err)
	}

	// Print results
	fmt.Println()
	if result.Fail > 0 {
		output.Errorf("Tests: %d passed, %d failed in %s",
			result.Pass, result.Fail,
			result.Duration.Round(time.Millisecond))
	} else {
		output.Successf("Tests: %d passed in %s",
			result.Pass,
			result.Duration.Round(time.Millisecond))
	}
	return nil
}

// Check performs a fast compilation check without producing a binary.
func Check(ctx context.Context, g gocmd.Runner, prj *project.Project) error {
	spinner := output.NewSpinner("Checking...")
	spinner.Start()

	// go build -o /dev/null ./...
	opts := gocmd.BuildOpts{
		Output: os.DevNull,
	}
	_, err := g.Build(ctx, prj.Root, opts)
	spinner.Stop()

	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	output.Successf("Check passed")
	return nil
}

// Clean removes build artifacts.
func Clean(ctx context.Context, g gocmd.Runner, prj *project.Project, cacheOnly bool) error {
	spinner := output.NewSpinner("Cleaning...")
	spinner.Start()

	if err := g.Clean(ctx, prj.Root, cacheOnly); err != nil {
		spinner.Stop()
		return err
	}

	// Also clean the project's binary
	binName := filepath.Base(prj.Root)
	_ = os.Remove(filepath.Join(prj.Root, binName))
	_ = os.Remove(filepath.Join(prj.Root, binName+".exe"))

	spinner.Stop()
	output.Successf("Cleaned")
	return nil
}

func isValidProjectName(name string) bool {
	if name == "" {
		return false
	}
	for _, c := range name {
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.' || c == '-' || c == '/' {
			continue
		}
		return false
	}
	return true
}

func sanitizePkgName(name string) string {
	// Go package names must be valid identifiers
	name = filepath.Base(name) // take last path component
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	if len(name) > 0 && name[0] >= '0' && name[0] <= '9' {
		name = "_" + name
	}
	return name
}

func initGit(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, "git", "init")
	cmd.Dir = dir
	return cmd.Run()
}

func formatSize(n int64) string {
	const unit = 1024.0
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	v, i := float64(n), 0
	for v >= unit && i < len(units)-1 {
		v /= unit
		i++
	}
	return fmt.Sprintf("%.1f %s", v, units[i])
}
