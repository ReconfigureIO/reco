package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ReconfigureIO/reco"
	"github.com/spf13/cobra"
)

var logPreRun = func(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("id required")
	}
}

func genLogSubcommand(name string, job string) *cobra.Command {
	return &cobra.Command{
		Use:     "log id",
		Aliases: []string{"logs"},
		Short:   fmt.Sprintf("Stream logs for a %s", name),
		Long:    fmt.Sprintf("Stream logs for a %s previously started with 'reco %s run'", name, name),
		PreRun:  logPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			l := reflect.ValueOf(tool).MethodByName(job).Call(nil)[0].Interface()
			if err := l.(reco.Job).Log(args[0], os.Stdout); err != nil {
				exitWithError(err)
			}
		},
	}
}
