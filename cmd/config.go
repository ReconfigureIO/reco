package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/pkg/logger"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:    "config",
	Short:  "Show config file path",
	Run:    config,
	Hidden: true,
}

func init() {
	RootCmd.AddCommand(configCmd)
}

func config(cmd *cobra.Command, args []string) {
	logger.Std.Println(filepath.Join(getConfigDir(), "reco.yml"))
}

func exitWithError(err interface{}) {
	if err != nil {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
	os.Exit(1)
}

func exitWithUsage(cmd *cobra.Command, err interface{}) {
	if err != nil {
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Error:", err)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, cmd.UsageString())
	}
	os.Exit(1)
}
