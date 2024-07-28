package cmdimpl

import (
	"context"
	"fmt"
	"os"

	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
	"github.com/Ruixi-rebirth/mygo/pkg/project"
)

func WorkspaceInit(ctx context.Context, g gocmd.Runner, dirs []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get working directory: %w", err)
	}

	if len(dirs) == 0 {
		dirs = []string{"."}
	}

	if err := g.WorkInit(ctx, cwd, dirs...); err != nil {
		return fmt.Errorf("workspace init: %w", err)
	}

	output.Successf("Workspace initialized at %s with: %v", output.Bold(cwd), dirs)
	return nil
}

func WorkspaceAdd(ctx context.Context, g gocmd.Runner, dirs []string) error {
	if len(dirs) == 0 {
		return fmt.Errorf("at least one directory required")
	}

	prj, err := project.Find()
	if err != nil {
		return err
	}

	if prj.IsWork {
		if err := g.WorkUse(ctx, prj.Root, dirs...); err != nil {
			return fmt.Errorf("workspace use: %w", err)
		}
	} else {
		return fmt.Errorf("not in a workspace; use 'mygo workspace init' first")
	}

	output.Successf("Added %v to workspace", dirs)
	return nil
}

func WorkspaceRemove(ctx context.Context, g gocmd.Runner, dirs []string) error {
	if len(dirs) == 0 {
		return fmt.Errorf("at least one directory required")
	}

	prj, err := project.Find()
	if err != nil {
		return err
	}

	if !prj.IsWork {
		return fmt.Errorf("not in a workspace")
	}

	for _, d := range dirs {
		if err := g.WorkEdit(ctx, prj.Root, "-dropuse", d); err != nil {
			return fmt.Errorf("workspace remove %s: %w", d, err)
		}
	}

	output.Successf("Removed %v from workspace", dirs)
	return nil
}

func WorkspaceSync(ctx context.Context, g gocmd.Runner) error {
	prj, err := project.Find()
	if err != nil {
		return err
	}

	if !prj.IsWork {
		return fmt.Errorf("not in a workspace")
	}

	if err := g.WorkSync(ctx, prj.Root); err != nil {
		return fmt.Errorf("workspace sync: %w", err)
	}

	output.Successf("Workspace synced")
	return nil
}
