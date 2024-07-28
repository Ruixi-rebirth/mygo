// Package gocmd abstracts the Go toolchain for testability.
package gocmd

import (
	"context"
	"time"
)

// Runner executes Go toolchain commands. Mock this interface in tests.
type Runner interface {
	// Core
	Build(ctx context.Context, dir string, opts BuildOpts) (*BuildResult, error)
	Run(ctx context.Context, dir string, pkg string, args []string, opts RunOpts) error
	Test(ctx context.Context, dir string, opts TestOpts) (*TestResult, error)

	// Modules
	ModInit(ctx context.Context, dir, modulePath string) error
	ModTidy(ctx context.Context, dir string) error
	ModGraph(ctx context.Context, dir string) ([]Edge, error)
	ModEdit(ctx context.Context, dir string, args ...string) error

	// Dependencies
	Get(ctx context.Context, dir, pkg string) error

	// Workspace
	WorkInit(ctx context.Context, dir string, moduleDirs ...string) error
	WorkUse(ctx context.Context, dir string, moduleDirs ...string) error
	WorkEdit(ctx context.Context, dir string, args ...string) error
	WorkSync(ctx context.Context, dir string) error

	// Security
	Govulncheck(ctx context.Context, dir string) (*VulnResult, error)

	// Info
	Clean(ctx context.Context, dir string, cacheOnly bool) error
	Env(ctx context.Context, keys ...string) (map[string]string, error)
	Version(ctx context.Context) (string, error)
	List(ctx context.Context, dir string, opts ListOpts) (*ListResult, error)
}

// ---- Option types ----

type BuildOpts struct {
	Output    string
	Ldflags   string
	Tags      []string
	GOOS      string
	GOARCH    string
	Gcflags   string
	Race      bool
	Verbose   bool
	Trimpath  bool
	Buildmode string
	Mod       string
}

type RunOpts struct {
	GOOS   string
	GOARCH string
	Race   bool
}

type TestOpts struct {
	Verbose      bool
	Cover        bool
	CoverProfile string
	CoverMode    string
	Race         bool
	Timeout      string
	Bench        string
	BenchMem     bool
	Run          string
	Count        int
	FailFast     bool
	Shuffle      string
	Tags         []string
	Short        bool
	Parallel     int
	JSON         bool
}

type ListOpts struct {
	JSON bool
	U    bool
	M    bool
}

// ---- Result types ----

type BuildResult struct {
	Binary   string
	Size     int64
	Duration time.Duration
}

type TestResult struct {
	Pass     int
	Fail     int
	Skip     int
	Duration time.Duration
	Output   string
}

type Edge struct {
	From string
	To   string
}

type VulnResult struct {
	Vulnerabilities []Vuln
	Output          string
}

type Vuln struct {
	ID      string
	Package string
	Details string
}

type ListResult struct {
	Packages []PackageInfo
}

type PackageInfo struct {
	ImportPath string
	Name       string
	Module     *ModuleInfo
	Deps       []string
}

type ModuleInfo struct {
	Path    string
	Version string
	Update  *ModuleUpdate
}

type ModuleUpdate struct {
	Path    string
	Version string
}
