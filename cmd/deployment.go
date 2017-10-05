package cmd

import (
	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/spf13/cobra"
)

var deploymentVars = struct {
	wait bool
}{
	wait: true,
}

var deploymentCmdStart = &cobra.Command{
	Use:     "run [flags] image executable -- [args]",
	Aliases: []string{"r", "start", "starts"},
	Short:   "Run a command from a build",
	Long: `Run a command from a build on a machine equipped with an FPGA.

Defining commands:

Your project should have a top-level directory "cmd". On build, each
subdirectory in "cmd" with a main.go will be built and put into your
$PATH automatically.

For example, if you have a file at "cmd/my-cmd/main.go", you will have
a binary named "my-cmd" available to you.

Passing arguments:

Arguments that are not captured by this tool are passed to the command.

For example, "reco run my-cmd 1" would pass the argument "1" to your
"my-cmd" binary. It's equivalent to calling "my-cmd 1". However, some
of your arguments may conflict with this binary. If this is the case,
use "--" to specify that all further arguments should be provided to
your command instead. The two forms are equivalent:
"reco run my-image my-cmd -- 1" and "reco run my-image my-cmd 1"
`,
	Run: startDeployment,
}

func init() {
	deploymentCmdStart.PersistentFlags().BoolVarP(&deploymentVars.wait, "wait", "w", deploymentVars.wait, "wait for the run to complete. If false, it only starts the command without waiting for it to complete.")

	deploymentCmd := genDevCommand("deployment", "d", "dep", "deps", "deployments")
	deploymentCmd.AddCommand(deploymentCmdStart)

	RootCmd.AddCommand(deploymentCmd)
}

func startDeployment(cmd *cobra.Command, args []string) {
	if len(args) < 2 {
		exitWithUsage(cmd, "image and executable are required")
	}
	image := args[0]
	command := args[1]
	commandArgs := []string{}
	if dash := cmd.ArgsLenAtDash(); dash > 0 {
		commandArgs = args[dash:]
	} else if len(args) > 2 {
		commandArgs = args[2:]
	}
	out, err := tool.Deployment().Start(reco.Args{image, command, commandArgs, deploymentVars.wait})
	if err != nil {
		exitWithError(err)
	}
	logger.Std.Println(out)
}
