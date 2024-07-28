// Package project handles Go project detection and metadata.
package project

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Project represents a detected Go project.
type Project struct {
	Root   string // absolute path to project root (where go.mod is)
	Module string // module path from go.mod (e.g., "github.com/user/app")
	GoVer  string // go directive version (e.g., "1.22")
	IsWork bool   // true if go.work found instead of go.mod
}

// Find locates the nearest Go project by walking up from the current directory.
// Returns an error if no go.mod or go.work is found.
func Find() (*Project, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}
	return FindFrom(dir)
}

// FindFrom locates the nearest Go project starting from the given path.
func FindFrom(dir string) (*Project, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	root, err := findRoot(dir)
	if err != nil {
		return nil, err
	}

	// Check if it's a workspace
	if _, err := os.Stat(filepath.Join(root, "go.work")); err == nil {
		return &Project{Root: root, IsWork: true}, nil
	}

	// Parse go.mod
	modPath, goVer, err := parseGoMod(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, fmt.Errorf("parse go.mod in %s: %w", root, err)
	}

	return &Project{
		Root:   root,
		Module: modPath,
		GoVer:  goVer,
	}, nil
}

// findRoot walks up from dir looking for go.mod or go.work.
func findRoot(dir string) (string, error) {
	// First, check the starting directory and walk up for go.mod/go.work.
	// This handles the case where we're in a temp dir that's not git-tracked.
	current := dir
	for {
		if _, err := os.Stat(filepath.Join(current, "go.mod")); err == nil {
			return current, nil
		}
		if _, err := os.Stat(filepath.Join(current, "go.work")); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}

	// Fallback: try git to find the project root (from the provided directory)
	if out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output(); err == nil {
		gitRoot := strings.TrimSpace(string(out))
		if _, err := os.Stat(filepath.Join(gitRoot, "go.mod")); err == nil {
			return gitRoot, nil
		}
		if _, err := os.Stat(filepath.Join(gitRoot, "go.work")); err == nil {
			return gitRoot, nil
		}
	}

	return "", fmt.Errorf("no go.mod or go.work found; run 'mygo new' or 'go mod init' first")
}

// parseGoMod extracts the module path and go version from a go.mod file.
func parseGoMod(path string) (module, goVer string, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			module = strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
		if strings.HasPrefix(line, "go ") {
			goVer = strings.TrimSpace(strings.TrimPrefix(line, "go "))
		}
	}
	if module == "" {
		return "", "", fmt.Errorf("no module directive in %s", path)
	}
	return module, goVer, nil
}

// HasMygoTOML checks if a mygo.toml exists in the project root.
func (p *Project) HasMygoTOML() bool {
	_, err := os.Stat(filepath.Join(p.Root, "mygo.toml"))
	return err == nil
}

// MainPackages finds all main packages in the project.
func (p *Project) MainPackages() ([]string, error) {
	var mains []string

	// Check root first
	if hasMain, _ := fileHasMain(filepath.Join(p.Root, "main.go")); hasMain {
		mains = append(mains, ".")
	}

	// Check cmd/ directory
	cmdDir := filepath.Join(p.Root, "cmd")
	entries, err := os.ReadDir(cmdDir)
	if err != nil {
		return mains, nil // cmd/ doesn't exist, just return root main
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		mainFile := filepath.Join(cmdDir, entry.Name(), "main.go")
		if hasMain, _ := fileHasMain(mainFile); hasMain {
			mains = append(mains, "./cmd/"+entry.Name())
		}
	}

	return mains, nil
}

func fileHasMain(path string) (bool, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return false, err
	}
	if f.Name.Name != "main" {
		return false, nil
	}

	// Also verify there's a main function (full parse)
	f, err = parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return false, err
	}
	for _, decl := range f.Decls {
		if fd, ok := decl.(*ast.FuncDecl); ok {
			if fd.Name.Name == "main" && fd.Recv == nil {
				return true, nil
			}
		}
	}
	return false, nil
}
