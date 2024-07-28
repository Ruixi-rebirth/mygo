package main

import (
	"github.com/Ruixi-rebirth/mygo/cmd"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	rootCmd := &cobra.Command{Use: "mygo"}

	rootCmd.AddCommand(cmd.NewCmd)
	rootCmd.AddCommand(cmd.BuildCmd)
	rootCmd.AddCommand(cmd.RunCmd)
	rootCmd.AddCommand(cmd.TestCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
