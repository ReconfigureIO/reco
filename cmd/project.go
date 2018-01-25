package cmd

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
)

var (
	projectCmd = &cobra.Command{
		Use:     "project",
		Aliases: []string{"p", "prj", "projects"},
		Short:   "Manage your projects.",
		Long: `Manage your projects.
You can create a new project, set and get the active project.`,
		PersistentPreRun: initializeCmd,
	}

	projectCmdCreate = &cobra.Command{
		Use:     "create",
		Aliases: []string{"c", "new"},
		Short:   "Create a new project",
		Long: `Create a new project. All future builds with be associated with this
project if you do not have any other projects.

To set the active project, use 'reco project set'.
`,
		Run: createProject,
	}

	projectCmdSet = &cobra.Command{
		Use:   "set name",
		Short: "Set project to use for builds",
		Long: `Set project to use for builds.
Builds created afterwards will be associated with this project.
You can verify active project with "reco project list".

This is a directory level config.
`,
		Run: setProject,
	}

	projectCmdGet = &cobra.Command{
		Use:   "get",
		Short: "Get the active project",
		Long: `Get the active project previously set with 'reco project set'.

This is a directory level config.
`,
		Run: getProject,
	}

	projectCmdList = &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "lst", "lists"},
		Short:   fmt.Sprintf("List all %s for your current project.", "project"),
		Long:    fmt.Sprintf("List all %s for your current project - status information, start times and unique IDs will be displayed.", "project"),
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
	l := reflect.ValueOf(tool).MethodByName("Project").Call(nil)[0].Interface()
	listVars.table, listVars.err = l.(lister).List(filters)
	listCmdAddFlags(cmd)
}
