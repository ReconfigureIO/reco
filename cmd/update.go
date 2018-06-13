package cmd

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/go-update"
	"github.com/ReconfigureIO/reco/pkg/logger"
	"github.com/google/go-github/github"
)

const (
	recoDownloadAddress = "https://s3.amazonaws.com/reconfigure.io/reco/releases/"
)

// updateCmd represents the update command
var (
	updateCmd = &cobra.Command{
		Use:   "update",
		Short: "Update reco to the latest version",
		Run:   updateHandler,
	}

	justDoIt bool
)

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().BoolVar(&justDoIt, "just-do-it", justDoIt, "Download and apply update without user interaction")
}

func updateHandler(cmd *cobra.Command, args []string) {
	if BuildInfo.Version == "" {
		logger.Std.Println("reco version: untracked dev build")
		logger.Std.Println("Cannot automatically update from this version")
		return
	}
	logger.Std.Println("You are using reco ", BuildInfo.Version)
	latest, err := latestRelease(github.NewClient(nil))
	if err != nil {
		logger.Std.Println("Could not retrieve latest verion info from Github: ", err)
		return
	} else {
		logger.Std.Println("The latest release is reco ", latest)
	}

	if justDoIt {
		err = UpgradeTo(latest, BuildInfo.Target)
		if err != nil {
			exitWithError(err)
		} else {
			return
		}
	}

	if latest != BuildInfo.Version {
		logger.Std.Println("Would you like to upgrade? (Y/N)")
		upgrade := askForConfirmation()
		if upgrade == true {
			if err := UpgradeTo(latest, BuildInfo.Target); err != nil {
				exitWithError(err)
			}
		}

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

func askForConfirmation() bool {
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		log.Fatal(err)
	}
	okayResponses := []string{"y", "Y", "yes", "Yes", "YES"}
	nokayResponses := []string{"n", "N", "no", "No", "NO"}
	if containsString(okayResponses, response) {
		return true
	} else if containsString(nokayResponses, response) {
		return false
	} else {
		fmt.Println("Please type yes or no and then press enter:")
		return askForConfirmation()
	}
}

func posString(slice []string, element string) int {
	for index, elem := range slice {
		if elem == element {
			return index
		}
	}
	return -1
}

// containsString returns true iff slice contains element
func containsString(slice []string, element string) bool {
	return !(posString(slice, element) == -1)
}

func UpgradeTo(version string, platform string) error {
	downloadURL := recoDownloadAddress + "reco-" + version + "-" + platform + ".zip"
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	zipFile, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	unzip, err := zip.NewReader(bytes.NewReader(zipFile), int64(len(zipFile)))
	if err != nil {
		return err
	}
	newReco, err := unzip.File[0].Open()
	err = update.Apply(newReco, update.Options{})
	if err != nil {
		logger.Std.Println("Error occured during self-update")
		return err
	}
	logger.Std.Println("Self-update successful")
	return err
}
