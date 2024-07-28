package cmd

import (
	"fmt"
	"github.com/Ruixi-rebirth/mygo/pkg/constants"
	"github.com/Ruixi-rebirth/mygo/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
)

var BuildCmd = &cobra.Command{
	Use:   "build [app]",
	Short: "Build the project or the specified app",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			buildProject("")
		} else {
			buildProject(args[0])
		}
	},
}

func buildProject(appName string) {
	// Get the project root directory
	projectRoot, err := utils.GetProjectRoot()
	if err != nil {
		fmt.Printf("%s  \x1b[1;31m%s\x1b[0m\n", constants.Error, err)
		return
	}

	// Change to the project root directory
	if err := os.Chdir(projectRoot); err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to change to project root directory: %s\x1b[0m\n", constants.Error, err)
		return
	}

	var mainGoPath string

	if appName == "" {
		// Check if main.go exists in the project root directory for small projects
		mainGoPath = filepath.Join(projectRoot, "main.go")
		if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
			fmt.Printf("%s  \x1b[1;31mmain.go not found in the project root directory\x1b[0m\n", constants.Error)
			return
		} else if err != nil {
			fmt.Printf("%s  \x1b[1;31mError checking for main.go: %s\x1b[0m\n", constants.Error, err)
			return
		}
	} else {
		// Check if main.go exists in the specified app directory under cmd
		mainGoPath = filepath.Join(projectRoot, "cmd", appName, "main.go")
		if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
			fmt.Printf("%s  \x1b[1;31mmain.go not found in the cmd/%s directory\x1b[0m\n", constants.Error, appName)
			return
		} else if err != nil {
			fmt.Printf("%s  \x1b[1;31mError checking for main.go: %s\x1b[0m\n", constants.Error, err)
			return
		}
	}

	// Determine the output binary name and build path
	outputName := appName
	buildPath := "."

	if appName != "" {
		buildPath = filepath.Join("./cmd", appName)
		outputName = appName
	} else {
		outputName = filepath.Base(projectRoot)
		buildPath = "."
	}

	// Build the project
	output, err := exec.Command("go", "build", "-o", outputName, buildPath).CombinedOutput()
	if err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to build project: %s\n%s\x1b[0m\n", constants.Error, err, string(output))
		return
	}

	fmt.Print(string(output))
}
