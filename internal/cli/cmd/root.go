package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "dropper",
	Short: "Tool for fast file drop between machines in local net",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

// Execute executes the root command
func Execute() {
	_ = rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(dropCmd)
	rootCmd.AddCommand(getCmd)
}
