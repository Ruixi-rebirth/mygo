package cmd

import (
	"fmt"
	"github.com/Ruixi-rebirth/mygo/pkg/constants"
	"github.com/Ruixi-rebirth/mygo/pkg/utils"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

var TestCmd = &cobra.Command{
	Use:   "test",
	Short: "Run tests",
	Run: func(cmd *cobra.Command, args []string) {
		testProject()
	},
}

func testProject() {
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

	// Run the tests
	output, err := exec.Command("go", "test", "./...").CombinedOutput()
	if err != nil {
		fmt.Printf("%s  \x1b[1;31mFailed to test project: %s\n%s\x1b[0m\n", constants.Error, err, string(output))
		return
	}

	fmt.Print(string(output))
}
