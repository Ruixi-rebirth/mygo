package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// GetProjectRoot tries to get the project root directory using git command,
// and falls back to searching for go.mod file if git is not available.
func GetProjectRoot() (string, error) {
	// Try to get the project root directory using git
	projectRoot, err := getProjectRootUsingGit()
	if err != nil {
		// If git command fails, try to find the project root by searching for go.mod
		projectRoot, err = findProjectRoot()
		if err != nil {
			return "", fmt.Errorf("failed to find project root directory: %w", err)
		}
	}
	return projectRoot, nil
}

// getProjectRootUsingGit tries to get the project root directory using git command
func getProjectRootUsingGit() (string, error) {
	output, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// findProjectRoot recursively searches for the project root directory by looking for go.mod
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		if parentDir := filepath.Dir(dir); parentDir == dir {
			break
		} else {
			dir = parentDir
		}
	}

	return "", fmt.Errorf("project root not found")
}

// ValidateInput checks if the input meets naming conventions
func ValidateInput(input string) bool {
	// Check if input matches naming conventions (allow alphanumeric, underscores, and slashes)
	isValid := regexp.MustCompile(`^[a-zA-Z0-9_./]+$`).MatchString
	return isValid(input)
}
