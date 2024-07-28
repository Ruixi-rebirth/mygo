package gocmd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExecRunner runs go commands via os/exec.
type ExecRunner struct {
	GoBin string
}

func NewExecRunner() *ExecRunner {
	goBin := "go"
	if g, err := exec.LookPath("go"); err == nil {
		goBin = g
	}
	return &ExecRunner{GoBin: goBin}
}

// --- Core ---

func (e *ExecRunner) Build(ctx context.Context, dir string, opts BuildOpts) (*BuildResult, error) {
	start := time.Now()
	args := []string{"build"}

	if opts.Output != "" {
		args = append(args, "-o", opts.Output)
	}
	if opts.Ldflags != "" {
		args = append(args, "-ldflags", opts.Ldflags)
	}
	if len(opts.Tags) > 0 {
		args = append(args, "-tags", strings.Join(opts.Tags, ","))
	}
	if opts.Gcflags != "" {
		args = append(args, "-gcflags", opts.Gcflags)
	}
	if opts.Race {
		args = append(args, "-race")
	}
	if opts.Verbose {
		args = append(args, "-v")
	}
	if opts.Trimpath {
		args = append(args, "-trimpath")
	}
	if opts.Buildmode != "" {
		args = append(args, "-buildmode", opts.Buildmode)
	}
	if opts.Mod != "" {
		args = append(args, "-mod", opts.Mod)
	}
	args = append(args, ".")

	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Dir = dir
	cmd.Env = e.buildEnv(opts.GOOS, opts.GOARCH)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go build: %w", err)
	}

	result := &BuildResult{
		Binary:   opts.Output,
		Duration: time.Since(start),
	}
	if result.Binary == "" {
		result.Binary = filepath.Base(dir)
	}
	if fi, err := os.Stat(filepath.Join(dir, result.Binary)); err == nil {
		result.Size = fi.Size()
	}
	return result, nil
}

func (e *ExecRunner) Run(ctx context.Context, dir string, pkg string, args []string, opts RunOpts) error {
	cmdArgs := []string{"run"}
	if opts.Race {
		cmdArgs = append(cmdArgs, "-race")
	}
	cmdArgs = append(cmdArgs, pkg)
	cmdArgs = append(cmdArgs, args...)

	cmd := exec.CommandContext(ctx, e.GoBin, cmdArgs...)
	cmd.Dir = dir
	cmd.Env = e.buildEnv(opts.GOOS, opts.GOARCH)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go run: %w", err)
	}
	return nil
}

func (e *ExecRunner) Test(ctx context.Context, dir string, opts TestOpts) (*TestResult, error) {
	start := time.Now()
	args := []string{"test"}

	if opts.Verbose {
		args = append(args, "-v")
	}
	if opts.Cover {
		args = append(args, "-cover")
	}
	if opts.CoverProfile != "" {
		args = append(args, "-coverprofile", opts.CoverProfile)
	}
	if opts.CoverMode != "" {
		args = append(args, "-covermode", opts.CoverMode)
	}
	if opts.Race {
		args = append(args, "-race")
	}
	if opts.Timeout != "" {
		args = append(args, "-timeout", opts.Timeout)
	}
	if opts.Bench != "" {
		args = append(args, "-bench", opts.Bench)
	}
	if opts.BenchMem {
		args = append(args, "-benchmem")
	}
	if opts.Run != "" {
		args = append(args, "-run", opts.Run)
	}
	if opts.Count > 0 {
		args = append(args, "-count", fmt.Sprintf("%d", opts.Count))
	}
	if opts.FailFast {
		args = append(args, "-failfast")
	}
	if opts.Shuffle != "" {
		args = append(args, "-shuffle", opts.Shuffle)
	}
	if opts.Short {
		args = append(args, "-short")
	}
	if opts.Parallel > 0 {
		args = append(args, "-parallel", fmt.Sprintf("%d", opts.Parallel))
	}
	if opts.JSON {
		args = append(args, "-json")
	}
	if len(opts.Tags) > 0 {
		args = append(args, "-tags", strings.Join(opts.Tags, ","))
	}
	args = append(args, "./...")

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = &stdout

	err := cmd.Run()
	output := stdout.String()
	duration := time.Since(start)

	result := &TestResult{
		Output:   output,
		Duration: duration,
	}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "ok ") {
			result.Pass++
		} else if strings.HasPrefix(line, "FAIL") {
			result.Fail++
		}
	}

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			result.Fail = 1
			return result, nil
		}
		return result, fmt.Errorf("go test: %w", err)
	}
	return result, nil
}

// --- Module operations ---

func (e *ExecRunner) ModInit(ctx context.Context, dir, modulePath string) error {
	cmd := exec.CommandContext(ctx, e.GoBin, "mod", "init", modulePath)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod init: %w", err)
	}
	return nil
}

func (e *ExecRunner) ModTidy(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, e.GoBin, "mod", "tidy")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod tidy: %w", err)
	}
	return nil
}

func (e *ExecRunner) ModGraph(ctx context.Context, dir string) ([]Edge, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, e.GoBin, "mod", "graph")
	cmd.Dir = dir
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go mod graph: %w", err)
	}

	var edges []Edge
	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) >= 2 {
			edges = append(edges, Edge{From: parts[0], To: parts[1]})
		}
	}
	return edges, scanner.Err()
}

func (e *ExecRunner) ModEdit(ctx context.Context, dir string, args ...string) error {
	cmdArgs := append([]string{"mod", "edit"}, args...)
	cmd := exec.CommandContext(ctx, e.GoBin, cmdArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go mod edit: %w", err)
	}
	return nil
}

// --- Dependencies ---

func (e *ExecRunner) Get(ctx context.Context, dir, pkg string) error {
	cmd := exec.CommandContext(ctx, e.GoBin, "get", pkg)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go get: %w", err)
	}
	return nil
}

// --- Workspace ---

func (e *ExecRunner) WorkInit(ctx context.Context, dir string, moduleDirs ...string) error {
	args := append([]string{"work", "init"}, moduleDirs...)
	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go work init: %w", err)
	}
	return nil
}

func (e *ExecRunner) WorkUse(ctx context.Context, dir string, moduleDirs ...string) error {
	args := append([]string{"work", "use"}, moduleDirs...)
	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go work use: %w", err)
	}
	return nil
}

func (e *ExecRunner) WorkEdit(ctx context.Context, dir string, args ...string) error {
	cmdArgs := append([]string{"work", "edit"}, args...)
	cmd := exec.CommandContext(ctx, e.GoBin, cmdArgs...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go work edit: %w", err)
	}
	return nil
}

func (e *ExecRunner) WorkSync(ctx context.Context, dir string) error {
	cmd := exec.CommandContext(ctx, e.GoBin, "work", "sync")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go work sync: %w", err)
	}
	return nil
}

// --- Security ---

func (e *ExecRunner) Govulncheck(ctx context.Context, dir string) (*VulnResult, error) {
	// govulncheck is not part of the standard Go distribution.
	// Check if it's installed first.
	if _, err := exec.LookPath("govulncheck"); err != nil {
		return nil, fmt.Errorf("govulncheck not found; install it with: go install golang.org/x/vuln/cmd/govulncheck@latest")
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, "govulncheck", "-json", "./...")
	cmd.Dir = dir
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	output := stdout.String()

	result := &VulnResult{Output: output}
	if output != "" {
		var raw struct {
			Vulnerabilities []struct {
				ID      string `json:"id"`
				Details string `json:"details"`
				Modules []struct {
					Path string `json:"path"`
				} `json:"modules"`
			} `json:"vulnerabilities"`
		}
		if json.Unmarshal([]byte(output), &raw) == nil {
			for _, v := range raw.Vulnerabilities {
				pkg := ""
				if len(v.Modules) > 0 {
					pkg = v.Modules[0].Path
				}
				result.Vulnerabilities = append(result.Vulnerabilities, Vuln{
					ID: v.ID, Package: pkg, Details: v.Details,
				})
			}
		}
	}

	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			return result, nil
		}
		return result, fmt.Errorf("govulncheck: %w", err)
	}
	return result, nil
}

// --- Info ---

func (e *ExecRunner) Clean(ctx context.Context, dir string, cacheOnly bool) error {
	args := []string{"clean"}
	if cacheOnly {
		args = append(args, "-cache", "-testcache")
	}
	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go clean: %w", err)
	}
	return nil
}

func (e *ExecRunner) Env(ctx context.Context, keys ...string) (map[string]string, error) {
	args := append([]string{"env"}, keys...)

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go env: %w", err)
	}

	result := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(stdout.String()), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx >= 0 {
			key := line[:idx]
			val := strings.Trim(line[idx+1:], "'")
			result[key] = val
		}
	}
	return result, nil
}

func (e *ExecRunner) Version(ctx context.Context) (string, error) {
	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, e.GoBin, "version")
	cmd.Stdout = &stdout
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("go version: %w", err)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func (e *ExecRunner) List(ctx context.Context, dir string, opts ListOpts) (*ListResult, error) {
	args := []string{"list"}
	if opts.JSON {
		args = append(args, "-json")
	}
	if opts.U {
		args = append(args, "-u")
	}
	if opts.M {
		args = append(args, "-m")
	}
	if opts.M {
		args = append(args, "all")
	} else {
		args = append(args, "./...")
	}

	var stdout bytes.Buffer
	cmd := exec.CommandContext(ctx, e.GoBin, args...)
	cmd.Dir = dir
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("go list: %w", err)
	}

	result := &ListResult{}
	if opts.JSON {
		decoder := json.NewDecoder(&stdout)
		for decoder.More() {
			var info PackageInfo
			if err := decoder.Decode(&info); err == nil {
				result.Packages = append(result.Packages, info)
			}
		}
	}
	return result, nil
}

// --- Helpers ---

func (e *ExecRunner) buildEnv(goos, goarch string) []string {
	env := os.Environ()
	if goos != "" {
		env = append(env, "GOOS="+goos)
	}
	if goarch != "" {
		env = append(env, "GOARCH="+goarch)
	}
	return env
}
