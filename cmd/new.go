package cmd

import (
	"bufio"
	"fmt"
	"github.com/Ruixi-rebirth/mygo/pkg/constants"
	"github.com/Ruixi-rebirth/mygo/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

var NewCmd = &cobra.Command{
	Use:   "new",
	Short: "Create and initialize a new Go project",
	Run: func(cmd *cobra.Command, args []string) {
		newProject()
	},
}

func newProject() {
	reader := bufio.NewReader(os.Stdin)

	// Get project name
	fmt.Print(constants.Prompt + "  Enter project name: ")
	projectName, _ := reader.ReadString('\n')
	projectName = strings.TrimSpace(projectName)

	if projectName == "" || !utils.ValidateInput(projectName) {
		fmt.Printf("%s  \x1b[1;31mInvalid input.\x1b[0m\n", constants.Error)
		return
	}

	// Prompt user for module name format
	fmt.Print(constants.Prompt + "  Do you want to add a custom prefix to the module name (e.g., github.com/username)? (y/n): ")
	addPrefix, _ := reader.ReadString('\n')
	addPrefix = strings.ToLower(strings.TrimSpace(addPrefix))

	var moduleName string

	if addPrefix == "y" || addPrefix == "yes" {
		// Get custom prefix interactively
		fmt.Print(constants.Prompt + " Enter your custom prefix: ")
		customPrefix, _ := reader.ReadString('\n')
		customPrefix = strings.TrimSpace(customPrefix)

		if customPrefix == "" || !utils.ValidateInput(customPrefix) {
			fmt.Printf("%s  \x1b[1;31mInvalid input.\x1b[0m\n", constants.Error)
			return
		}

		moduleName = fmt.Sprintf("%s/%s", customPrefix, projectName)
	} else {
		moduleName = projectName
	}

	// After all inputs are gathered, execute the project creation
	if err := createProject(projectName, moduleName); err != nil {
		fmt.Printf("%s  \x1b[1;31m%s\x1b[0m\n", constants.Error, err)
	}
}

func createProject(projectName, moduleName string) error {
	// Check if the directory already exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("project directory already exists: %s", projectName)
	} else {
		// Attempt to create the directory
		if err := os.Mkdir(projectName, 0755); err != nil {
			return fmt.Errorf("failed to create project directory: %s", err)
		}
	}

	// Change into the directory
	if err := os.Chdir(projectName); err != nil {
		return fmt.Errorf("failed to change directory: %s", err)
	}

	// Initialize Go module
	if err := exec.Command("go", "mod", "init", moduleName).Run(); err != nil {
		return fmt.Errorf("failed to initialize project with go mod: %s", err)
	}

	// Create main.go file
	mainFileContent := `package main

import (
	"fmt"
)

func main() {
    fmt.Println("Hello, World!")
}
`
	if err := os.WriteFile("main.go", []byte(mainFileContent), 0644); err != nil {
		return fmt.Errorf("failed to write main.go file: %s", err)
	}

	// Initialize Git repository
	if err := exec.Command("git", "init").Run(); err != nil {
		return fmt.Errorf("failed to initialize Git repository: %s", err)
	}

	// Print success message
	fmt.Printf("%s  Created project: %s\n", constants.Success, projectName)
	return nil
}
