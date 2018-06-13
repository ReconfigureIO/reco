package cmd

import (
	"strconv"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/pkg/logger"
	"github.com/ReconfigureIO/reco/pkg/reco"
)

var (
	projectCmd = &cobra.Command{
		Use:     "project",
		Aliases: []string{"p", "prj", "projects"},
		Short:   "Manage your projects",
		Long: `Manage your projects.
You can create a new project, set a project to work within and get the name of the currently active project.`,
		PersistentPreRun: initializeCmd,
	}

	projectCmdCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"c", "new"},
		Short:   "Create a new project",
		Long: `Create a new project.
		Once you have created a project it will be available to set as active for any location. To set an active project, use 'reco project set <my_project>'.
`,
		Run: createProject,
	}

	projectCmdSet = &cobra.Command{
		Use:   "set name",
		Short: "Set the active project for your current location",
		Long: `Set the active project for your current location.
Simulations, builds, graphs and deployments created after setting an active project will be associated with that project.
You can verify the active project with "reco project list" or "reco project get".
This is a directory level configuration so you need to set a project for each new location you work in.
`,
		Run: setProject,
	}

	projectCmdGet = &cobra.Command{
		Use:   "get",
		Short: "Get the name of the active project",
		Long: `Get the name of the active project for your current location.
This is a directory level configuration so you need to check in each new location you work in.
`,
		Run: getProject,
	}

	projectCmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "lst", "lists"},
		Short:   "List all projects for your account",
		Long: `List all projects for your account.
If you have an active project set for your current location this will be highlighted in the list.`,
		Run:     listProject,
		PostRun: listPostRun,
	}
)

func init() {
	projectCmd.AddCommand(
		projectCmdSet,
		projectCmdGet,
		projectCmdCreate,
		projectCmdList,
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

func listProject(cmd *cobra.Command, args []string) {
	filters := reco.M{}
	if listVars.status != "" {
		filters["status"] = listVars.status
	}
	if listVars.limit != 0 {
		filters["limit"] = strconv.Itoa(listVars.limit)
	}
	if listVars.allProjects {
		filters["all"] = "1"
	}
	if listVars.public {
		filters["public"] = "1"
	}

	listVars.resourceType = "project"
	listVars.table, listVars.err = tool.Project().List(filters)
	listCmdAddFlags(cmd)
}
