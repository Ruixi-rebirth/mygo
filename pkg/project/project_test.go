package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFrom(t *testing.T) {
	// Create a temp project
	dir := t.TempDir()

	// No go.mod yet - should fail
	_, err := FindFrom(dir)
	if err == nil {
		t.Error("expected error for directory without go.mod")
	}

	// Create go.mod
	modContent := "module example.com/test\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(modContent), 0644); err != nil {
		t.Fatal(err)
	}

	prj, err := FindFrom(dir)
	if err != nil {
		t.Fatalf("FindFrom failed: %v", err)
	}

	if prj.Root != dir {
		t.Errorf("Root = %q, want %q", prj.Root, dir)
	}
	if prj.Module != "example.com/test" {
		t.Errorf("Module = %q, want %q", prj.Module, "example.com/test")
	}
	if prj.GoVer != "1.22" {
		t.Errorf("GoVer = %q, want %q", prj.GoVer, "1.22")
	}
	if prj.IsWork {
		t.Error("IsWork = true, want false")
	}
}

func TestFindFromNested(t *testing.T) {
	dir := t.TempDir()

	// Create go.mod in subdirectory
	subDir := filepath.Join(dir, "sub", "project")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	modContent := "module example.com/nested\n\ngo 1.21\n"
	if err := os.WriteFile(filepath.Join(subDir, "go.mod"), []byte(modContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Find from the subdirectory
	prj, err := FindFrom(subDir)
	if err != nil {
		t.Fatalf("FindFrom failed: %v", err)
	}
	if prj.Root != subDir {
		t.Errorf("Root = %q, want %q", prj.Root, subDir)
	}

	// Find from a deeper directory
	deeperDir := filepath.Join(subDir, "internal", "foo")
	if err := os.MkdirAll(deeperDir, 0755); err != nil {
		t.Fatal(err)
	}

	prj2, err := FindFrom(deeperDir)
	if err != nil {
		t.Fatalf("FindFrom from deeper dir failed: %v", err)
	}
	if prj2.Root != subDir {
		t.Errorf("Root from deeper = %q, want %q", prj2.Root, subDir)
	}
}

func TestWorkspaceDetection(t *testing.T) {
	dir := t.TempDir()

	// Create go.work
	workContent := "go 1.22\n\nuse ./\n"
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte(workContent), 0644); err != nil {
		t.Fatal(err)
	}

	prj, err := FindFrom(dir)
	if err != nil {
		t.Fatalf("FindFrom failed: %v", err)
	}
	if !prj.IsWork {
		t.Error("IsWork = false, want true")
	}
}

func TestMainPackages(t *testing.T) {
	dir := t.TempDir()

	// Create go.mod
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module test\n\ngo 1.22\n"), 0644)

	// Create root main.go
	os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	prj := &Project{Root: dir, Module: "test", GoVer: "1.22"}

	mains, err := prj.MainPackages()
	if err != nil {
		t.Fatalf("MainPackages failed: %v", err)
	}
	if len(mains) != 1 || mains[0] != "." {
		t.Errorf("MainPackages = %v, want [.]", mains)
	}

	// Create cmd/myapp/main.go
	cmdDir := filepath.Join(dir, "cmd", "myapp")
	os.MkdirAll(cmdDir, 0755)
	os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\nfunc main() {}\n"), 0644)

	mains, err = prj.MainPackages()
	if err != nil {
		t.Fatalf("MainPackages failed: %v", err)
	}
	if len(mains) != 2 {
		t.Errorf("MainPackages count = %d, want 2", len(mains))
	}
}
