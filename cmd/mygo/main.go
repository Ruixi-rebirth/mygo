package main

import (
	"os"

	"github.com/Ruixi-rebirth/mygo/internal/cli"
	"github.com/Ruixi-rebirth/mygo/internal/output"
	"github.com/Ruixi-rebirth/mygo/pkg/gocmd"
)

func main() {
	runner := gocmd.NewExecRunner()
	rootCmd := cli.RootCmd(runner)

	if err := rootCmd.Execute(); err != nil {
		output.Errorf("%v", err)
		os.Exit(1)
	}
}
