package cmd

import (
	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/cobra"
)

var projectCmd = &cobra.Command{
	Use:     "project",
	Aliases: []string{"p", "prj", "projects"},
	Short:   "Manage projects",
	Long: `Manage projects.
You can create a new project, set and get the active project.`,
	PersistentPreRun: initializeCmd,
}

var projectCmdCreate = &cobra.Command{
	Use:     "create",
	Aliases: []string{"c", "new"},
	Short:   "Create a new project",
	Long: `Create a new project. All future builds with be associated with this
project if you do not have any other projects.

To set the active project, use 'reco project set'.
`,
	Run: createProject,
}

var projectCmdSet = &cobra.Command{
	Use:   "set name",
	Short: "Set project to use for builds",
	Long: `Set project to use for builds.
Builds created afterwards will be associated with this project.
You can verify active project with "reco project list".

This is a directory level config.
`,
	Run: setProject,
}

var projectCmdGet = &cobra.Command{
	Use:   "get",
	Short: "Get the active project",
	Long: `Get the active project previously set with 'reco project set'.

This is a directory level config.
`,
	Run: getProject,
}

func init() {
	projectCmd.AddCommand(
		projectCmdSet,
		projectCmdGet,
		projectCmdCreate,
		genListSubcommand("projects", "Project"),
	)

	RootCmd.AddCommand(projectCmd)
}

func createProject(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("name required")
	}
	if err := tool.Project().Create(args[0]); err != nil {
		exitWithError(err)
	}
}

func setProject(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("name required")
	}
	if err := tool.Project().Set(args[0]); err != nil {
		exitWithError(err)
	}
}

func getProject(cmd *cobra.Command, args []string) {
	name, err := tool.Project().Get()
	if err != nil {
		exitWithError(err)
	}
	logger.Std.Println(name)
}
