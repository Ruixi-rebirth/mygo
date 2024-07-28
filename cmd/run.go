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

var RunCmd = &cobra.Command{
	Use:   "run [app]",
	Short: "Run the project or the specified app",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runProject("")
		} else {
			runProject(args[0])
		}
	},
}

func runProject(appName string) {
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

	// Run 'go mod tidy' to tidy up dependencies
	if output, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to tidy up dependencies: %s\n%s\x1b[0m\n", constants.Error, err, string(output))
		return
	}

	// Run the project
	var runCmd *exec.Cmd
	if appName == "" {
		runCmd = exec.Command("go", "run", ".")
	} else {
		runCmd = exec.Command("go", "run", filepath.Join("cmd", appName))
	}

	runOutput, err := runCmd.CombinedOutput()
	if err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to run project: %s\x1b[0m\n", constants.Error, err)
		return
	}

	fmt.Print(string(runOutput))
}
