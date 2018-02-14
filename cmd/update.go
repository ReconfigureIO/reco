package cmd

import (
	"context"

	"github.com/ReconfigureIO/cobra"
	//"github.com/ReconfigureIO/go-update"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/google/go-github/github"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update reco to the latest version",
	Run:   update,
}

func init() {
	RootCmd.AddCommand(updateCmd)
}

func update(cmd *cobra.Command, args []string) {
	if BuildInfo.Version == "" {
		logger.Std.Println("reco version: untracked dev build")
		logger.Std.Println("Cannot automatically update from this version")
		return
	}
	logger.Std.Println("You are using reco ", BuildInfo.Version)
	if BuildInfo.BuildTime != "" {
		logger.Std.Println("Built at: ", BuildInfo.BuildTime)
	}
	latest, err := latestRelease(github.NewClient(nil))
	if err != nil {
		logger.Std.Println("Could not retrieve latest verion info from Github: ", err)
		return
	} else {
		logger.Std.Println("The latest release is reco ", latest)
	}

	if latest != BuildInfo.Version {
		logger.Std.Println("Run reco update --apply to upgrade to ", latest)
	} else {
		logger.Std.Println("You are using the latest version")
		return
	}

	return
}

// latestRelease gets the version number of the latest reco release
func latestRelease(client *github.Client) (string, error) {
	release, _, err := client.Repositories.GetLatestRelease(context.Background(), "ReconfigureIO", "reco")
	if err != nil {
		return "", err
	}
	return *release.TagName, nil
}
