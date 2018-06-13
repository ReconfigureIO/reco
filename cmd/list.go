package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/pkg/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/reco/printer"
)

var listVars struct {
	resourceType string
	table        printer.Table
	err          error

	noScroll    bool
	limit       int
	status      string
	allProjects bool
	public      bool
}

type lister interface {
	List(filter reco.M) (printer.Table, error)
}

func genListSubcommand(name string, job lister) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "lst", "lists"},
		Short:   fmt.Sprintf("List all %s for your current project", name),
		Long:    fmt.Sprintf("List all %s for your current project - status information, start times and unique IDs will be displayed.", name),
		Run: func(cmd *cobra.Command, args []string) {
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

			listVars.resourceType = name
			listVars.table, listVars.err = job.List(filters)
		},
		PostRun: listPostRun,
	}
	listCmdAddFlags(cmd)
	return cmd
}

var listPostRun = func(cmd *cobra.Command, args []string) {
	if listVars.err != nil {
		exitWithError(listVars.err)
	}

	if listVars.table.Empty() {
		logger.Std.Printf("You have no %s.", listVars.resourceType)
		return
	}

	var err error

	if listVars.noScroll {
		err = printer.Fprint(os.Stdout, listVars.table)
	} else {
		err = printer.Print(listVars.table)
	}

	if err != nil {
		exitWithError(err)
	}
}

func listCmdAddFlags(listCmd *cobra.Command) {
	listCmd.PersistentFlags().BoolVar(&listVars.noScroll, "no-scroll", listVars.noScroll, "Disable scrollable paged output even if output is longer than the terminal height")
	listCmd.PersistentFlags().IntVarP(&listVars.limit, "limit", "l", listVars.limit, "Limit the number of results displayed")
	listCmd.PersistentFlags().StringVar(&listVars.status, "status", listVars.status, "Filter result by status: completed, errored, timed-out etc")
	listCmd.PersistentFlags().BoolVar(&listVars.allProjects, "all-projects", listVars.allProjects, "List items for all projects, not just the active project")
	listCmd.PersistentFlags().BoolVar(&listVars.public, "public", listVars.public, "Only list publically available items")
}
