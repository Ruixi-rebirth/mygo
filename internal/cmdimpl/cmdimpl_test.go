package cmdimpl

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Ruixi-rebirth/mygo/pkg/config"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
)

func TestNewProject_Binary(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "myapp")

	var modInitCalled bool
	var modInitPath string
	mock := &gocmd.MockRunner{
		ModInitFn: func(ctx context.Context, dir, modulePath string) error {
			modInitCalled = true
			modInitPath = modulePath
			// Actually write go.mod so file checks pass
			return os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module myapp\n"), 0644)
		},
	}
	ctx := context.Background()

	err := NewProject(ctx, mock, name, "", false)
	if err != nil {
		t.Fatalf("NewProject failed: %v", err)
	}

	if !modInitCalled {
		t.Error("ModInit was not called")
	}
	if modInitPath != name {
		t.Errorf("ModInit module = %q, want %q", modInitPath, name)
	}

	// Verify scaffolding files
	for _, f := range []string{"go.mod", "main.go", "mygo.toml"} {
		if _, err := os.Stat(filepath.Join(name, f)); os.IsNotExist(err) {
			t.Errorf("expected file %s not found", f)
		}
	}

	mainData, _ := os.ReadFile(filepath.Join(name, "main.go"))
	if !strings.Contains(string(mainData), "package main") {
		t.Errorf("main.go should be a main package")
	}
}

func TestNewProject_Library(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "mylib")

	mock := &gocmd.MockRunner{}
	ctx := context.Background()

	err := NewProject(ctx, mock, name, "", true)
	if err != nil {
		t.Fatalf("NewProject(lib) failed: %v", err)
	}

	// Library should have <name>.go not main.go
	if _, err := os.Stat(filepath.Join(name, "mylib.go")); os.IsNotExist(err) {
		t.Errorf("expected mylib.go not found")
	}
	if _, err := os.Stat(filepath.Join(name, "main.go")); err == nil {
		t.Errorf("library should not have main.go")
	}
}

func TestNewProject_Invalid(t *testing.T) {
	mock := &gocmd.MockRunner{}
	ctx := context.Background()

	// Empty name
	err := NewProject(ctx, mock, "", "", false)
	if err == nil {
		t.Error("expected error for empty name")
	}

	// Name with spaces
	err = NewProject(ctx, mock, "bad name", "", false)
	if err == nil {
		t.Error("expected error for name with spaces")
	}

	// Existing directory
	dir := t.TempDir()
	existing := filepath.Join(dir, "exists")
	os.Mkdir(existing, 0755)
	err = NewProject(ctx, mock, existing, "", false)
	if err == nil {
		t.Error("expected error for existing directory")
	}
}

func TestAddDependency(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "testmod"}
	mock := &gocmd.MockRunner{}
	ctx := context.Background()

	err := AddDependency(ctx, mock, prj, "github.com/spf13/cobra", "v1.8.1", true)
	if err != nil {
		t.Fatalf("AddDependency failed: %v", err)
	}
}

func TestAddDependency_Error(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "testmod"}
	mock := &gocmd.MockRunner{
		GetFn: func(ctx context.Context, dir, pkg string) error {
			return os.ErrNotExist
		},
	}
	ctx := context.Background()

	err := AddDependency(ctx, mock, prj, "bad/pkg", "", false)
	if err == nil {
		t.Error("expected error for failed get")
	}
}

func TestRemoveDependency(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "testmod"}
	mock := &gocmd.MockRunner{}
	ctx := context.Background()

	err := RemoveDependency(ctx, mock, prj, "github.com/old/dep")
	if err != nil {
		t.Fatalf("RemoveDependency failed: %v", err)
	}
}

func TestBuild(t *testing.T) {
	dir := t.TempDir()
	prj := &project.Project{Root: dir, Module: "test"}
	cfg := config.Default()
	mock := &gocmd.MockRunner{}
	ctx := context.Background()

	opts := gocmd.BuildOpts{Output: "bin/test"}
	err := Build(ctx, mock, prj, cfg, opts, false, "")
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Test with release profile
	err = Build(ctx, mock, prj, cfg, opts, true, "")
	if err != nil {
		t.Fatalf("Build with release failed: %v", err)
	}

	// Test with target
	err = Build(ctx, mock, prj, cfg, opts, false, "linux/amd64")
	if err != nil {
		t.Fatalf("Build with target failed: %v", err)
	}
}

func TestBuild_Error(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "test"}
	cfg := config.Default()
	mock := &gocmd.MockRunner{
		BuildFn: func(ctx context.Context, dir string, opts gocmd.BuildOpts) (*gocmd.BuildResult, error) {
			return nil, os.ErrNotExist
		},
	}
	err := Build(context.Background(), mock, prj, cfg, gocmd.BuildOpts{}, false, "")
	if err == nil {
		t.Error("expected build error")
	}
}

func TestCheck(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "test"}
	mock := &gocmd.MockRunner{}
	if err := Check(context.Background(), mock, prj); err != nil {
		t.Fatalf("Check failed: %v", err)
	}
}

func TestClean(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "test"}
	mock := &gocmd.MockRunner{}
	for _, cacheOnly := range []bool{false, true} {
		if err := Clean(context.Background(), mock, prj, cacheOnly); err != nil {
			t.Fatalf("Clean(%v) failed: %v", cacheOnly, err)
		}
	}
}

func TestOutdatedDeps(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "test"}
	mock := &gocmd.MockRunner{
		ListFn: func(ctx context.Context, dir string, opts gocmd.ListOpts) (*gocmd.ListResult, error) {
			return &gocmd.ListResult{
				Packages: []gocmd.PackageInfo{
					{Module: &gocmd.ModuleInfo{
						Path: "github.com/old/dep", Version: "v1.0.0",
						Update: &gocmd.ModuleUpdate{Path: "github.com/old/dep", Version: "v2.0.0"},
					}},
				},
			}, nil
		},
	}
	for _, json := range []bool{false, true} {
		if err := OutdatedDeps(context.Background(), mock, prj, json); err != nil {
			t.Fatalf("OutdatedDeps(%v) failed: %v", json, err)
		}
	}
}

func TestAudit(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "test"}
	t.Run("clean", func(t *testing.T) {
		if err := Audit(context.Background(), &gocmd.MockRunner{}, prj, false); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("vulns", func(t *testing.T) {
		mock := &gocmd.MockRunner{
			GovulncheckFn: func(ctx context.Context, dir string) (*gocmd.VulnResult, error) {
				return &gocmd.VulnResult{Vulnerabilities: []gocmd.Vuln{
					{ID: "CVE-2024-0001", Package: "golang.org/x/net", Details: "test"},
				}}, nil
			},
		}
		if err := Audit(context.Background(), mock, prj, false); err != nil {
			t.Fatal(err)
		}
	})
}

func TestBench(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "test"}
	mock := &gocmd.MockRunner{
		TestFn: func(ctx context.Context, dir string, opts gocmd.TestOpts) (*gocmd.TestResult, error) {
			return &gocmd.TestResult{Output: "Benchmark-8 1000 ns/op"}, nil
		},
	}
	if err := Bench(context.Background(), mock, prj, ".", 1, false, ""); err != nil {
		t.Fatal(err)
	}
}

func TestDepTree(t *testing.T) {
	prj := &project.Project{Root: t.TempDir(), Module: "testapp"}
	mock := &gocmd.MockRunner{
		ModGraphFn: func(ctx context.Context, dir string) ([]gocmd.Edge, error) {
			return []gocmd.Edge{
				{From: "testapp", To: "github.com/a/lib@v1.0.0"},
				{From: "github.com/a/lib@v1.0.0", To: "github.com/c/dep@v0.5.0"},
			}, nil
		},
	}
	if err := DepTree(context.Background(), mock, prj); err != nil {
		t.Fatal(err)
	}
}

func TestConfigOperations(t *testing.T) {
	dir := t.TempDir()
	prj := &project.Project{Root: dir, Module: "test"}
	cfg := config.Default()
	cfg.Project.Name = "testproj"
	config.Save(cfg, filepath.Join(dir, "mygo.toml"))

	ctx := context.Background()
	for _, key := range []string{"project.name", "project.version", "build.ldflags"} {
		if err := ConfigGet(ctx, prj, key); err != nil {
			t.Errorf("ConfigGet(%s): %v", key, err)
		}
	}
	if err := ConfigGet(ctx, prj, "bad.key"); err == nil {
		t.Error("expected error for unknown key")
	}
	if err := ConfigSet(ctx, prj, "build.ldflags", "-w"); err != nil {
		t.Fatal(err)
	}
	if err := ConfigSet(ctx, prj, "bad.key", "x"); err == nil {
		t.Error("expected error for unknown key")
	}
	if err := ConfigList(ctx, prj); err != nil {
		t.Fatal(err)
	}
}

func TestInitProject(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22\n"), 0644)
	os.Chdir(dir)
	defer os.Chdir("/")

	if err := InitProject(context.Background(), &gocmd.MockRunner{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "mygo.toml")); os.IsNotExist(err) {
		t.Error("mygo.toml not created")
	}
}

func TestIsValidProjectName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"hello", true}, {"my-app", true}, {"github.com/user/repo", true},
		{"bad name", false}, {"", false},
	}
	for _, tt := range tests {
		if got := isValidProjectName(tt.name); got != tt.valid {
			t.Errorf("isValidProjectName(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		size int64
		want string
	}{
		{0, "0 B"}, {512, "512 B"}, {1024, "1.0 KiB"}, {1048576, "1.0 MiB"},
	}
	for _, tt := range tests {
		if got := formatSize(tt.size); got != tt.want {
			t.Errorf("formatSize(%d) = %q, want %q", tt.size, got, tt.want)
		}
	}
}
