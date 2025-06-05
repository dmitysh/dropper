package cmd

import (
	"context"
	"runtime/debug"

	"github.com/dmitysh/dropper/internal/pkg/logger"
	"github.com/spf13/cobra"
)

func Version(ctx context.Context) string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		logger.Fatalf(ctx, "no debug info")
	}
	return buildInfo.Main.Version
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:     "version",
	Short:   "Print dropper version",
	Aliases: []string{"v"},
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		logger.Infof(ctx, "dropper version: %s\n", Version(ctx))
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
