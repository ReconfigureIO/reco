package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
)

var (
	testCmdStart = &cobra.Command{
		Use:     "run [flags] command -- [args]",
		Aliases: []string{"r", "start", "starts", "create"},
		Short:   "Start a simulation",
		Long: `Start a simulation.
		Running a hardware simulation checks how your program will build and deploy on hardware.
		It's much faster than running a build and is a good way to check for errors in your code.
		`,
		Run: startTest,
	}

	testCmdLog = &cobra.Command{
		Use:     fmt.Sprintf("log [simulation_ID]"),
		Aliases: []string{"logs"},
		Short:   fmt.Sprintf("Stream logs for a simulation"),
		Long:    fmt.Sprintf("Stream logs for a simulation previously started with 'reco sim run'."),
		PreRun:  testLogPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tool.Test().Log(args[0], os.Stdout); err != nil {
				exitWithError(err)
			}
		},
	}

	testCmdReport = &cobra.Command{
		Use:     fmt.Sprintf("report [simulation_ID]"),
		Aliases: []string{"reports"},
		Short:   fmt.Sprintf("View reports for a completed simulation"),
		Long:    fmt.Sprintf("View reports for a completed simulation, containing area and resource utilisation figures."),
		Run:     openTestReport,
	}

	testLogPreRun = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exitWithError("ID required")
		}
	}

	testCmdStop = &cobra.Command{
		Use:     "stop [simulation_ID]",
		Aliases: []string{"s", "stp", "stops"},
		Short:   fmt.Sprintf("Stop a simulation"),
		Long:    fmt.Sprintf("Stop a simulation previously started with 'reco sim run'"),
		PreRun:  testStopPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tool.Test().Stop(args[0]); err != nil {
				exitWithError(err)
			}
			logger.Std.Printf("Simulation stopped successfully")
		},
	}

	testStopPreRun = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exitWithError("ID required")
		}
	}
)

func init() {
	testCmd := genDevCommand("sim", "simulation", "simulation", "simulations", "test", "tests", "t")
	testCmd.AddCommand(genListSubcommand("simulations", tool.Test()))
	testCmd.AddCommand(testCmdLog)
	testCmd.AddCommand(testCmdStop)
	testCmd.AddCommand(testCmdStart)
	testCmd.AddCommand(testCmdReport)
	testCmd.PersistentFlags().StringVar(&project, "project", project, "Project to use. If unset, the active project is used")

	RootCmd.AddCommand(testCmd)
}

func startTest(cmd *cobra.Command, args []string) {
	if !validBuildDir(srcDir) {
		exitWithError(errInvalidSourceDirectory)
	}
	if len(args) < 1 {
		exitWithUsage(cmd, "command is required")
	}
	command := args[0]
	commandArgs := []string{}
	if dash := cmd.ArgsLenAtDash(); dash > 0 {
		commandArgs = args[dash:]
	} else if len(args) > 1 {
		commandArgs = args[1:]
	}
	id, err := tool.Test().Start(reco.Args{srcDir, command, commandArgs})
	if err != nil {
		exitWithError(err)
	}

	status := tool.Test().Status(id)
	logger.Std.Println("Simulation ID: " + id + " Status: " + strings.Title(status))
}

func openTestReport(_ *cobra.Command, args []string) {
	if len(args) != 1 {
		exitWithError("ID required")
	}

	report, err := tool.Test().(reco.SimulationReporter).Report(args[0]) // TODO campgareth: Anything but this weird form.
	if err != nil {
		exitWithError(err)
	}

	logger.Std.Println(report)
}
