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
	Use:   "build [project-path] [app]",
	Short: "Build the project in the specified path or the specified app",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			buildProject("", "")
		} else if len(args) == 1 {
			buildProject(args[0], "")
		} else {
			buildProject(args[0], args[1])
		}
	},
}

func buildProject(projectPath string, appName string) {
	var projectRoot string
	var err error

	if projectPath == "" {
		projectRoot, err = utils.GetProjectRoot()
		if err != nil {
			fmt.Printf("%s  \x1b[1;31m%s\x1b[0m\n", constants.Error, err)
			return
		}
	} else {
		// 处理相对路径，将其转换为绝对路径
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

	outputName := appName
	buildPath := "."

	if appName != "" {
		buildPath = filepath.Join("./cmd", appName)
		outputName = appName
	} else {
		outputName = filepath.Base(projectRoot)
		buildPath = "."
	}

	buildCmd := exec.Command("go", "build", "-o", outputName, buildPath)
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr

	if err := buildCmd.Run(); err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to build project: %s\x1b[0m\n", constants.Error, err)
		return
	}

	fmt.Printf("Build successful: %s\n", outputName)
}
