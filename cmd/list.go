package cmd

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/reco/printer"
	"github.com/spf13/cobra"
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

func genListSubcommand(name string) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls", "lst", "lists"},
		Short:   fmt.Sprintf("List %ss for your account", name),
		Long:    fmt.Sprintf("List %ss created by your account.", name),
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
			l := reflect.ValueOf(tool).MethodByName(strings.Title(name)).Call(nil)[0].Interface()
			listVars.table, listVars.err = l.(lister).List(filters)
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
		logger.Std.Printf("You have no %ss.", listVars.resourceType)
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
	listCmd.PersistentFlags().BoolVar(&listVars.noScroll, "no-scroll", listVars.noScroll, "disable scrollable paged output even if output is longer than terminal height")
	listCmd.PersistentFlags().IntVarP(&listVars.limit, "limit", "l", listVars.limit, "limit the number of results")
	listCmd.PersistentFlags().StringVar(&listVars.status, "status", listVars.status, "filter result by status")
	listCmd.PersistentFlags().BoolVar(&listVars.allProjects, "all-projects", listVars.allProjects, "list for all projects")
	listCmd.PersistentFlags().BoolVar(&listVars.public, "public", listVars.public, "list for public items")
}

type lister interface {
	List(filter reco.M) (printer.Table, error)
}
