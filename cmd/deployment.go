package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/pkg/logger"
	"github.com/ReconfigureIO/reco/pkg/reco"
)

var (
	deploymentVars = struct {
		wait string
	}{
		wait: "true",
	}

	errorDeploymentNotFound = errors.New("No deployment with that ID could be found. Run 'reco deploy list' to view available deployments")

	deploymentCmdStart = &cobra.Command{
		Use:     "run [flags] <build_ID> <your_cmd> -- [args]",
		Aliases: []string{"r", "start", "starts"},
		Short:   "Deploy a build image and command to an F1 instance",
		Long: `Deploy a build image and run a command from that build on an F1 instance.

More about commands:

Reconfigure.io programs have a main.go at the top level and a top level
directory "cmd". On build, the top level main.go will be compiled and optimized
into a deployable image, and each subdirectory in "cmd" containing a main.go
file will be available as a runnable command for the host CPU.

For example, if your program has a file at "cmd/my-cmd/main.go", you will have
a runnable command named "my-cmd" available to you.

Passing arguments:

Arguments that are not captured by this tool are passed to the command.

For example, "reco deploy run my-cmd 1" would pass the argument "1" to
"my-cmd". It's equivalent to calling "my-cmd 1". However, some
of your arguments may conflict with this command. If this is the case,
use "--" to specify that all further arguments should be provided to
your command. The two forms are equivalent:
"reco run my-image my-cmd -- 1" and "reco run my-image my-cmd 1"
	`,
		Run: startDeployment,
	}

	deploymentCmdConnect = &cobra.Command{
		Use:     "connect <deploy_ID>",
		Aliases: []string{"c", "connects"},
		Short:   "Connect to a running deployment",
		Long:    "Connect to a running deployment.",
		Run:     connectDeployment,
	}

	deploymentCmdLog = &cobra.Command{
		Use:     fmt.Sprintf("log [deployment_ID]"),
		Aliases: []string{"logs"},
		Short:   fmt.Sprintf("Stream logs for a deployment"),
		Long:    fmt.Sprintf("Stream logs for a deployment previously started with 'reco deploy run'."),
		PreRun:  deploymentLogPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tool.Deployment().Log(args[0], os.Stdout); err != nil {
				exitWithError(interpretErrorDeployment(err))
			}
		},
	}

	deploymentLogPreRun = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exitWithError("ID required")
		}
	}

	deploymentCmdStop = &cobra.Command{
		Use:     "stop [deployment_ID]",
		Aliases: []string{"s", "stp", "stops"},
		Short:   fmt.Sprintf("Stop a deployment"),
		Long:    fmt.Sprintf("Stop a deployment previously started with 'reco deploy run'"),
		PreRun:  deploymentStopPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			if err := tool.Deployment().Stop(args[0]); err != nil {
				exitWithError(interpretErrorDeployment(err))
			}
			logger.Std.Printf("deployment stopped successfully")
		},
	}

	deploymentStopPreRun = func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			exitWithError("ID required")
		}
	}
)

func init() {
	deploymentCmdStart.PersistentFlags().StringVarP(&deploymentVars.wait, "wait", "w", deploymentVars.wait, "Wait for the run to complete. If false, it only starts the command without waiting for it to complete")

	deploymentCmd := genDevCommand("deploy", "deployment", "d", "dep", "deps", "deployments", "deployment")
	deploymentCmd.AddCommand(genListSubcommand("deployments", tool.Deployment()))
	deploymentCmd.AddCommand(deploymentCmdLog)
	deploymentCmd.AddCommand(deploymentCmdStop)
	deploymentCmd.AddCommand(deploymentCmdStart)
	deploymentCmd.AddCommand(deploymentCmdConnect)
	deploymentCmd.PersistentFlags().StringVar(&project, "project", project, "Project to use. If unset, the active project is used")

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
		exitWithError(interpretErrorDeployment(err))
	}
	logger.Std.Println(out)
}

func connectDeployment(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("deployment ID required")
	}
	if err := tool.Deployment().(reco.DeploymentProxy).Connect(args[0], true); err != nil {
		exitWithError(interpretErrorDeployment(err))
	}
}

func interpretErrorDeployment(err error) error {
	switch err {
	case reco.ErrNotFound:
		return errorDeploymentNotFound
	default:
		return err
	}
}
