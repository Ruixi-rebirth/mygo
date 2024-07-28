package gocmd

import (
	"context"
	"time"
)

// MockRunner is a programmable mock for testing.
type MockRunner struct {
	BuildFn       func(ctx context.Context, dir string, opts BuildOpts) (*BuildResult, error)
	RunFn         func(ctx context.Context, dir string, pkg string, args []string, opts RunOpts) error
	TestFn        func(ctx context.Context, dir string, opts TestOpts) (*TestResult, error)
	ModInitFn     func(ctx context.Context, dir, modulePath string) error
	ModTidyFn     func(ctx context.Context, dir string) error
	ModGraphFn    func(ctx context.Context, dir string) ([]Edge, error)
	ModEditFn     func(ctx context.Context, dir string, args ...string) error
	GetFn         func(ctx context.Context, dir, pkg string) error
	WorkInitFn    func(ctx context.Context, dir string, moduleDirs ...string) error
	WorkUseFn     func(ctx context.Context, dir string, moduleDirs ...string) error
	WorkEditFn    func(ctx context.Context, dir string, args ...string) error
	WorkSyncFn    func(ctx context.Context, dir string) error
	GovulncheckFn func(ctx context.Context, dir string) (*VulnResult, error)
	CleanFn       func(ctx context.Context, dir string, cacheOnly bool) error
	EnvFn         func(ctx context.Context, keys ...string) (map[string]string, error)
	VersionFn     func(ctx context.Context) (string, error)
	ListFn        func(ctx context.Context, dir string, opts ListOpts) (*ListResult, error)
}

func (m *MockRunner) Build(ctx context.Context, dir string, opts BuildOpts) (*BuildResult, error) {
	if m.BuildFn != nil {
		return m.BuildFn(ctx, dir, opts)
	}
	return &BuildResult{Binary: "mock", Size: 1024, Duration: time.Millisecond}, nil
}

func (m *MockRunner) Run(ctx context.Context, dir string, pkg string, args []string, opts RunOpts) error {
	if m.RunFn != nil {
		return m.RunFn(ctx, dir, pkg, args, opts)
	}
	return nil
}

func (m *MockRunner) Test(ctx context.Context, dir string, opts TestOpts) (*TestResult, error) {
	if m.TestFn != nil {
		return m.TestFn(ctx, dir, opts)
	}
	return &TestResult{Pass: 1, Duration: time.Millisecond}, nil
}

func (m *MockRunner) ModInit(ctx context.Context, dir, modulePath string) error {
	if m.ModInitFn != nil {
		return m.ModInitFn(ctx, dir, modulePath)
	}
	return nil
}

func (m *MockRunner) ModTidy(ctx context.Context, dir string) error {
	if m.ModTidyFn != nil {
		return m.ModTidyFn(ctx, dir)
	}
	return nil
}

func (m *MockRunner) ModGraph(ctx context.Context, dir string) ([]Edge, error) {
	if m.ModGraphFn != nil {
		return m.ModGraphFn(ctx, dir)
	}
	return nil, nil
}

func (m *MockRunner) ModEdit(ctx context.Context, dir string, args ...string) error {
	if m.ModEditFn != nil {
		return m.ModEditFn(ctx, dir, args...)
	}
	return nil
}

func (m *MockRunner) Get(ctx context.Context, dir, pkg string) error {
	if m.GetFn != nil {
		return m.GetFn(ctx, dir, pkg)
	}
	return nil
}

func (m *MockRunner) WorkInit(ctx context.Context, dir string, moduleDirs ...string) error {
	if m.WorkInitFn != nil {
		return m.WorkInitFn(ctx, dir, moduleDirs...)
	}
	return nil
}

func (m *MockRunner) WorkUse(ctx context.Context, dir string, moduleDirs ...string) error {
	if m.WorkUseFn != nil {
		return m.WorkUseFn(ctx, dir, moduleDirs...)
	}
	return nil
}

func (m *MockRunner) WorkEdit(ctx context.Context, dir string, args ...string) error {
	if m.WorkEditFn != nil {
		return m.WorkEditFn(ctx, dir, args...)
	}
	return nil
}

func (m *MockRunner) WorkSync(ctx context.Context, dir string) error {
	if m.WorkSyncFn != nil {
		return m.WorkSyncFn(ctx, dir)
	}
	return nil
}

func (m *MockRunner) Govulncheck(ctx context.Context, dir string) (*VulnResult, error) {
	if m.GovulncheckFn != nil {
		return m.GovulncheckFn(ctx, dir)
	}
	return &VulnResult{}, nil
}

func (m *MockRunner) Clean(ctx context.Context, dir string, cacheOnly bool) error {
	if m.CleanFn != nil {
		return m.CleanFn(ctx, dir, cacheOnly)
	}
	return nil
}

func (m *MockRunner) Env(ctx context.Context, keys ...string) (map[string]string, error) {
	if m.EnvFn != nil {
		return m.EnvFn(ctx, keys...)
	}
	return map[string]string{"GOBIN": "/go/bin", "GOPATH": "/go"}, nil
}

func (m *MockRunner) Version(ctx context.Context) (string, error) {
	if m.VersionFn != nil {
		return m.VersionFn(ctx)
	}
	return "go version go1.22.0 linux/amd64", nil
}

func (m *MockRunner) List(ctx context.Context, dir string, opts ListOpts) (*ListResult, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, dir, opts)
	}
	return &ListResult{}, nil
}

// ErrorRunner returns a MockRunner where key methods return the given error.
func ErrorRunner(err error) *MockRunner {
	return &MockRunner{
		BuildFn:   func(ctx context.Context, dir string, opts BuildOpts) (*BuildResult, error) { return nil, err },
		RunFn:     func(ctx context.Context, dir string, pkg string, args []string, opts RunOpts) error { return err },
		TestFn:    func(ctx context.Context, dir string, opts TestOpts) (*TestResult, error) { return nil, err },
		ModInitFn: func(ctx context.Context, dir, modulePath string) error { return err },
		ModTidyFn: func(ctx context.Context, dir string) error { return err },
		GetFn:     func(ctx context.Context, dir, pkg string) error { return err },
		ModEditFn: func(ctx context.Context, dir string, args ...string) error { return err },
	}
}

var _ Runner = (*MockRunner)(nil)
