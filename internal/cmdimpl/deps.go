package cmdimpl

import (
	"context"
	"fmt"
	"sort"

	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
)

func AddDependency(ctx context.Context, g gocmd.Runner, prj *project.Project, pkg string, version string, tidy bool) error {
	target := pkg
	if version != "" {
		target = pkg + "@" + version
	}

	output.Infof("Adding %s...", output.Bold(target))

	if err := g.Get(ctx, prj.Root, target); err != nil {
		return fmt.Errorf("add dependency: %w", err)
	}

	if tidy {
		if err := g.ModTidy(ctx, prj.Root); err != nil {
			output.Warnf("go mod tidy failed: %v", err)
		}
	}

	output.Successf("Added %s", output.Bold(pkg))
	return nil
}

func RemoveDependency(ctx context.Context, g gocmd.Runner, prj *project.Project, pkg string) error {
	output.Infof("Removing %s...", output.Bold(pkg))

	if err := g.ModEdit(ctx, prj.Root, "-droprequire", pkg); err != nil {
		return fmt.Errorf("remove dependency: %w", err)
	}
	if err := g.ModTidy(ctx, prj.Root); err != nil {
		return fmt.Errorf("tidy after remove: %w", err)
	}

	output.Successf("Removed %s", output.Bold(pkg))
	return nil
}

func UpdateDependencies(ctx context.Context, g gocmd.Runner, prj *project.Project, packages []string, dryRun, all bool) error {
	if dryRun {
		output.Infof("Checking for updates (dry run)...")
		return OutdatedDeps(ctx, g, prj, false)
	}

	output.Infof("Updating dependencies...")

	if len(packages) > 0 {
		for _, pkg := range packages {
			if err := g.Get(ctx, prj.Root, pkg+"@latest"); err != nil {
				return fmt.Errorf("update %s: %w", pkg, err)
			}
		}
	} else if all {
		if err := g.Get(ctx, prj.Root, "-u"); err != nil {
			return fmt.Errorf("update all: %w", err)
		}
	}

	if err := g.ModTidy(ctx, prj.Root); err != nil {
		return fmt.Errorf("tidy after update: %w", err)
	}
	output.Successf("Dependencies updated")
	return nil
}

func DepTree(ctx context.Context, g gocmd.Runner, prj *project.Project) error {
	edges, err := g.ModGraph(ctx, prj.Root)
	if err != nil {
		return fmt.Errorf("get dependency graph: %w", err)
	}

	// Build adjacency list
	children := make(map[string][]string)
	for _, e := range edges {
		children[e.From] = append(children[e.From], e.To)
	}
	for k := range children {
		sort.Strings(children[k])
	}

	fmt.Println(output.Bold(prj.Module))
	printTree(prj.Module, children, "", make(map[string]bool))
	return nil
}

func printTree(node string, children map[string][]string, prefix string, visited map[string]bool) {
	if visited[node] {
		fmt.Printf("%s%s %s\n", prefix+output.Dim("└──"), node, output.Dim("(cycle)"))
		return
	}
	visited[node] = true

	deps := children[node]
	for i, dep := range deps {
		isLast := i == len(deps)-1
		connector := "├──"
		indent := prefix + "│   "
		if isLast {
			connector = "└──"
			indent = prefix + "    "
		}

		fmt.Printf("%s %s%s\n", prefix+output.Dim(connector), dep, "")
		printTree(dep, children, indent, visited)
	}
}

func OutdatedDeps(ctx context.Context, g gocmd.Runner, prj *project.Project, jsonOutput bool) error {
	result, err := g.List(ctx, prj.Root, gocmd.ListOpts{
		JSON: true, U: true, M: true,
	})
	if err != nil {
		return fmt.Errorf("list deps: %w", err)
	}

	var outdated []struct{ path, current, latest string }
	for _, p := range result.Packages {
		if p.Module != nil && p.Module.Update != nil {
			outdated = append(outdated, struct{ path, current, latest string }{
				p.Module.Path, p.Module.Version, p.Module.Update.Version,
			})
		}
	}

	if len(outdated) == 0 {
		output.Successf("All dependencies up to date")
		return nil
	}

	if jsonOutput {
		fmt.Println("[")
		for i, d := range outdated {
			fmt.Printf("  {\"path\": %q, \"current\": %q, \"latest\": %q}", d.path, d.current, d.latest)
			if i < len(outdated)-1 {
				fmt.Println(",")
			} else {
				fmt.Println()
			}
		}
		fmt.Println("]")
	} else {
		t := output.NewTable("Package", "Current", "Latest")
		for _, d := range outdated {
			t.Row(d.path, output.Yellow(d.current), output.Green(d.latest))
		}
		t.Print()
	}
	return nil
}
