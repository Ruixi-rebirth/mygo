package cmd

import (
	"fmt"
	"github.com/Ruixi-rebirth/mygo/pkg/constants"
	"github.com/Ruixi-rebirth/mygo/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build the project",
	Run: func(cmd *cobra.Command, args []string) {
		buildProject()
	},
}

func buildProject() {
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

	// Build the project
	output, err := exec.Command("go", "build").CombinedOutput()
	if err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to build project: %s\n%s\x1b[0m\n", constants.Error, err, string(output))
		return
	}

	fmt.Println(string(output))
	fmt.Println(constants.Success + "  \x1b[1;32mProject built successfully\x1b[0m")
}
