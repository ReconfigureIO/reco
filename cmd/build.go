package cmd

import (
	"os"
	"path/filepath"

	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/cobra"
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
	Long: `Start a new build. A successful build creates an image that can be deployed to an F1 instance. Your FPGA code will be compiled, optimized and assigned a unique ID. Each subdirectory within "cmd" containing a main.go file will become a runnable command available for use when you deploy your build - reco deploy run <build_ID> <my_cmd>.
`,
	Run: startBuild,
}

func init() {
	buildCmdStart.PersistentFlags().BoolVarP(&buildVars.wait, "wait", "w", buildVars.wait, "Wait for the build to complete. If wait=false, logs will only be displayed up to where the build is started and assigned its unique ID. Use 'reco build list' to check the status of your builds.")
	buildCmdStart.PersistentFlags().BoolVarP(&buildVars.force, "force", "f", buildVars.force, "Force a build to start. Ignore source code validation.")

	buildCmd := genDevCommand("build", "b", "builds")
	buildCmd.AddCommand(genListSubcommand("builds", "Build"))
	buildCmd.AddCommand(genLogSubcommand("build", "Build"))
	buildCmd.AddCommand(genStopSubcommand("build", "Build"))
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
