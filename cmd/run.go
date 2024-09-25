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
	Use:   "run [project-path] [app]",
	Short: "Run the project in the specified path or the specified app",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			runProject("", "")
		} else if len(args) == 1 {
			runProject(args[0], "")
		} else {
			runProject(args[0], args[1])
		}
	},
}

func runProject(projectPath string, appName string) {
	var projectRoot string
	var err error

	if projectPath == "" {
		projectRoot, err = utils.GetProjectRoot()
		if err != nil {
			fmt.Printf("%s  \x1b[1;31m%s\x1b[0m\n", constants.Error, err)
			return
		}
	} else {
		projectRoot, err = filepath.Abs(projectPath)
		if err != nil {
			fmt.Printf("%s  \x1b[1;31mFailed to resolve project path: %s\x1b[0m\n", constants.Error, err)
			return
		}
	}

	if err := os.Chdir(projectRoot); err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to change to project root directory: %s\x1b[0m\n", constants.Error, err)
		return
	}

	var mainGoPath string

	if appName == "" {
		mainGoPath = filepath.Join(projectRoot, "main.go")
		if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
			fmt.Printf("%s  \x1b[1;31mmain.go not found in the project root directory\x1b[0m\n", constants.Error)
			return
		} else if err != nil {
			fmt.Printf("%s  \x1b[1;31mError checking for main.go: %s\x1b[0m\n", constants.Error, err)
			return
		}
	} else {
		mainGoPath = filepath.Join(projectRoot, "cmd", appName, "main.go")
		if _, err := os.Stat(mainGoPath); os.IsNotExist(err) {
			fmt.Printf("%s  \x1b[1;31mmain.go not found in the cmd/%s directory\x1b[0m\n", constants.Error, appName)
			return
		} else if err != nil {
			fmt.Printf("%s  \x1b[1;31mError checking for main.go: %s\x1b[0m\n", constants.Error, err)
			return
		}
	}

	runCmd := exec.Command("go", "run", ".")
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr

	if err := runCmd.Run(); err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to run project: %s\x1b[0m\n", constants.Error, err)
		return
	}
}
