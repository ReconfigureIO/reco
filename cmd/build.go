package cmd

import (
	"os"
	"path/filepath"

	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/spf13/cobra"
)

var buildVars = struct {
	wait  bool
	force bool
}{
	wait: true,
}

// buildCmd represents the upload command
var buildCmdStart = &cobra.Command{
	Use:     "run",
	Aliases: []string{"r", "start", "starts", "create"},
	Short:   "Start a new build",
	Long: `Start a new build. The build (if successful) can be deployed afterwards.

Your project should have a top-level directory "cmd". On build, each
subdirectory in "cmd" with a main.go will be built and put into your
$PATH automatically.
`,
	Run: startBuild,
}

func init() {
	buildCmdStart.PersistentFlags().BoolVarP(&buildVars.wait, "wait", "w", buildVars.wait, "wait for the build to complete. If false, it only kicks off the build without waiting.")
	buildCmdStart.PersistentFlags().BoolVarP(&buildVars.force, "force", "f", buildVars.force, "force build. Ignore source code validation.")

	buildCmd := genDevCommand("build", "b", "builds")
	buildCmd.AddCommand(genListSubcommand("builds", "Build"))
	buildCmd.AddCommand(genLogSubcommand("builds", "Build"))
	buildCmd.AddCommand(genStopSubcommand("builds", "Build"))
	buildCmd.AddCommand(buildCmdStart)

	RootCmd.AddCommand(buildCmd)
}

func startBuild(cmd *cobra.Command, args []string) {
	if !buildVars.force && !validBuildDir(srcDir) {
		exitWithError(errInvalidSourceDirectory)
	}

	id, err := tool.Build().Start(reco.Args{srcDir, buildVars.wait})
	if err != nil {
		exitWithError(err)
	}

	logger.Std.Println(id)
}

func validBuildDir(srcDir string) bool {
	// src directory
	if !hasMain(srcDir) {
		return false
	}

	// cmd subdirectory
	f, err := os.Open(filepath.Join(srcDir, "cmd"))
	if err != nil {
		return false
	}
	children, err := f.Readdirnames(-1)
	if err != nil {
		return false
	}

	for i := range children {
		if !hasMain(filepath.Join(f.Name(), children[i])) {
			return false
		}
	}
	return true
}

func hasMain(dir string) bool {
	mainGo := "main.go"
	f, err := os.Open(srcDir)
	if err != nil {
		return false
	}
	children, err := f.Readdirnames(-1)
	if err != nil {
		return false
	}

	foundMain := false
	for i := range children {
		if children[i] == mainGo {
			foundMain = true
			break
		}
	}
	return foundMain
}
