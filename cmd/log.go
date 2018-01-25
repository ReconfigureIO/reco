package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/cobra"
)

var logPreRun = func(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("ID required")
	}
}

func genLogSubcommand(name string, job string) *cobra.Command {
	return &cobra.Command{
		Use:     "log ID",
		Aliases: []string{"logs"},
		Short:   fmt.Sprintf("Stream logs for a %s", name),
		Long:    fmt.Sprintf("Stream logs for a %s previously started with 'reco %s run'", name, job),
		PreRun:  logPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			l := reflect.ValueOf(tool).MethodByName(job).Call(nil)[0].Interface()
			if err := l.(reco.Job).Log(args[0], os.Stdout); err != nil {
				exitWithError(err)
			}
		},
	}
}
