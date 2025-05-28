package main

import (
	"github.com/dmitysh/dropper/cli/command"
	"github.com/dmitysh/dropper/configs/envconfig"
	"github.com/spf13/cobra"
	"os"
)

func main() {
	if runDropperErr := runDropper(); runDropperErr != nil {
		os.Exit(1)
	}
}

func runDropper() error {
	envconfig.LoadEmbeddedEnvConfig()
	topCmd := newDropperCommand()
	return topCmd.Execute()
}

func newDropperCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dropper",
		Short: "Tool for fast file drop between machines in local net",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	command.AddAllCommands(cmd)

	return cmd
}
