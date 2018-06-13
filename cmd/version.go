package cmd

import (
	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/pkg/logger"
)

// BuildInfo is the build information of reco binary. This is
// set at build time by ldflags.
var BuildInfo struct {
	Version, BuildTime, Builder, GoVersion, Target string
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show which version of reco you are currently running.",
	Run:   version,
}

func init() {
	RootCmd.AddCommand(versionCmd)
}

func version(cmd *cobra.Command, args []string) {
	if BuildInfo.Version == "" {
		logger.Std.Println("reco version: untracked dev build")
		return
	}
	logger.Std.Println("reco version: ", BuildInfo.Version)
	if BuildInfo.GoVersion != "" {
		logger.Std.Println("Go version: ", BuildInfo.GoVersion)
	}
	if BuildInfo.BuildTime != "" {
		logger.Std.Println("Build Time: ", BuildInfo.BuildTime)
	}
	if BuildInfo.Builder != "" {
		logger.Std.Println("Built By: ", BuildInfo.Builder)
	}
}
