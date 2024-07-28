package cmdimpl

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
)

func UninstallBinary(ctx context.Context, g gocmd.Runner, name string) error {
	env, err := g.Env(ctx, "GOBIN", "GOPATH")
	if err != nil {
		return fmt.Errorf("get env: %w", err)
	}

	gobin := env["GOBIN"]
	if gobin == "" {
		gopath := env["GOPATH"]
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		gobin = filepath.Join(gopath, "bin")
	}

	binPath := filepath.Join(gobin, name)
	if _, err := os.Stat(binPath); os.IsNotExist(err) {
		return fmt.Errorf("binary %s not found at %s", name, binPath)
	}

	if err := os.Remove(binPath); err != nil {
		return fmt.Errorf("remove binary: %w", err)
	}

	output.Successf("Uninstalled %s from %s", output.Bold(name), output.Dim(gobin))
	return nil
}

func Bench(ctx context.Context, g gocmd.Runner, prj *project.Project, pattern string, count int, benchMem bool, timeout string) error {
	opts := gocmd.TestOpts{
		Bench:    pattern,
		Count:    count,
		BenchMem: benchMem,
		Timeout:  timeout,
	}

	output.Infof("Running benchmarks...")

	result, err := g.Test(ctx, prj.Root, opts)
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println(result.Output)
	output.Successf("Benchmarks completed in %s", result.Duration)
	return nil
}

func Audit(ctx context.Context, g gocmd.Runner, prj *project.Project, jsonOutput bool) error {
	output.Infof("Checking for vulnerabilities...")

	result, err := g.Govulncheck(ctx, prj.Root)
	if err != nil {
		return fmt.Errorf("audit: %w", err)
	}

	if len(result.Vulnerabilities) == 0 {
		output.Successf("No vulnerabilities found")
		return nil
	}

	output.Errorf("Found %d vulnerability(s):", len(result.Vulnerabilities))
	for _, v := range result.Vulnerabilities {
		fmt.Printf("  %s %s - %s\n", output.Red("✗"), output.Bold(v.ID), output.Dim(v.Package))
		if v.Details != "" {
			fmt.Printf("    %s\n", v.Details)
		}
	}
	return nil
}
