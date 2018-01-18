package cmd

import (
	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/spf13/cobra"
)

var testCmdStart = &cobra.Command{
	Use:     "run [flags] command -- [args]",
	Aliases: []string{"r", "start", "starts", "create"},
	Short:   "Start a simulation",
	Long: `Start a simulation.
	Running a hardware simulation checks how your program will build and deploy on hardware.
	It's much faster than running a build and is a good way to check for errors in your code.
	`,
	Run: startTest,
}

func init() {
	testCmd := genDevCommand("sim", "test", "tests", "t", "simulation", "simulations")
	testCmd.AddCommand(genListSubcommand("simulations", "Simulation"))
	testCmd.AddCommand(genLogSubcommand("simulation", "Simulation"))
	testCmd.AddCommand(genStopSubcommand("simulation", "Simulation"))
	testCmd.AddCommand(testCmdStart)

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
	out, err := tool.Test().Start(reco.Args{srcDir, command, commandArgs})
	if err != nil {
		exitWithError(err)
	}
	logger.Std.Println(out)
}
