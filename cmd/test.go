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
	Simulation test builds and deployment and runs much faster.
	It is a good way to check for errors before a build.
	`,
	Run: startTest,
}

func init() {
	testCmd := genDevCommand("test", "t", "sim", "simulation", "simulations")
	testCmd.AddCommand(genListSubcommand("tests", "Test"))
	testCmd.AddCommand(genLogSubcommand("tests", "Tests"))
	testCmd.AddCommand(genStopSubcommand("tests", "Tests"))
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
