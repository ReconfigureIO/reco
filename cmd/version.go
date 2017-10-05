package cmd

import (
	"strings"

	"github.com/ReconfigureIO/reco/logger"
	"github.com/spf13/cobra"
)

// BuildInfo is the build information of reco binary. This is
// set at build time by ldflags.
var BuildInfo struct {
	Version, BuildTime, Builder, GoVersion string
}

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show app version",
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
	if goVersion := strings.Fields(BuildInfo.GoVersion); len(goVersion) > 3 {
		logger.Std.Println("Go version: ", goVersion[2])
	}
	if BuildInfo.BuildTime != "" {
		logger.Std.Println("Build Time: ", BuildInfo.BuildTime)
	}
	if BuildInfo.Builder != "" {
		logger.Std.Println("Built By: ", BuildInfo.Builder)
	}
}
