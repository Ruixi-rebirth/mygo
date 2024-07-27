package cmd

import (
	"fmt"
	"github.com/Ruixi-rebirth/mygo/pkg/constants"
	"github.com/Ruixi-rebirth/mygo/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project",
	Run: func(cmd *cobra.Command, args []string) {
		runProject()
	},
}

func runProject() {
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

	// Run 'go mod tidy' to tidy up dependencies
	if output, err := exec.Command("go", "mod", "tidy").CombinedOutput(); err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to tidy up dependencies: %s\n%s\x1b[0m\n", constants.Error, err, string(output))
		return
	}

	// Run the project
	runOutput, err := exec.Command("go", "run", ".").CombinedOutput()
	if err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to run project: %s\x1b[0m\n", constants.Error, err)
		return
	}

	fmt.Println(string(runOutput))
}
