package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/pkg/logger"
	"github.com/ReconfigureIO/reco/pkg/reco"
)

var (
	buildVars = struct {
		wait  bool
		force bool
	}{
		wait: true,
	}

	// buildCmd represents the upload command
	buildCmdStart = &cobra.Command{
		Use:     "run",
		Aliases: []string{"r", "start", "starts", "create"},
		Short:   "Start a new build",
		Long: `Start a new build.
	A successful build creates an image that can be deployed to an F1 instance. Your FPGA code will be compiled, optimized and assigned a unique ID.
	Each subdirectory within "cmd" containing a main.go file will become a runnable command available for use when you deploy your build - reco deploy run <build_ID> <my_cmd>.
	`,
		Run: startBuild,
	}

	buildCmdLog = &cobra.Command{
		Use:     fmt.Sprintf("log [build_ID]"),
		Aliases: []string{"logs"},
		Short:   fmt.Sprintf("Stream logs for a build"),
		Long:    fmt.Sprintf("Stream logs for a build previously started with 'reco build run'."),
		PreRun:  buildLogPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tool.Build().Log(args[0], os.Stdout); err != nil {
				exitWithError(err)
			}
		},
	}

	buildLogPreRun = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exitWithError("ID required")
		}
	}

	buildCmdStop = &cobra.Command{
		Use:     "stop [build_ID]",
		Aliases: []string{"s", "stp", "stops"},
		Short:   fmt.Sprintf("Stop a build"),
		Long:    fmt.Sprintf("Stop a build previously started with 'reco build run'"),
		PreRun:  buildStopPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tool.Build().Stop(args[0]); err != nil {
				exitWithError(err)
			}
			logger.Std.Printf("build stopped successfully")
		},
	}

	buildStopPreRun = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exitWithError("ID required")
		}
	}
)

func init() {
	buildCmdStart.PersistentFlags().BoolVarP(&buildVars.wait, "wait", "w", buildVars.wait, "Wait for the build to complete. If wait=false, logs will only be displayed up to where the build is started and assigned its unique ID. Use 'reco build list' to check the status of your builds")
	buildCmdStart.PersistentFlags().BoolVarP(&buildVars.force, "force", "f", buildVars.force, "Force a build to start. Ignore source code validation")

	buildCmd := genDevCommand("build", "build", "b", "builds")
	buildCmd.AddCommand(genListSubcommand("builds", tool.Build()))
	buildCmd.AddCommand(buildCmdLog)
	buildCmd.AddCommand(buildCmdStop)
	buildCmd.AddCommand(buildCmdStart)
	buildCmd.PersistentFlags().StringVar(&project, "project", project, "Project to use. If unset, the active project is used")

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

	status := tool.Build().Status(id)
	logger.Std.Println("Build ID: " + id + " Status: " + strings.Title(status))
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
