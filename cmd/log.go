package cmd

import (
	"fmt"
	"os"
	"reflect"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco"
)

var logPreRun = func(cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("ID required")
	}
}

func genLogSubcommand(commandName string, jobType string) *cobra.Command {
	return &cobra.Command{
		Use:     fmt.Sprintf("log [%s_ID]", jobType),
		Aliases: []string{"logs"},
		Short:   fmt.Sprintf("Stream logs for a %s", jobType),
		Long:    fmt.Sprintf("Stream logs for a %s previously started with 'reco %s run'", jobType, commandName),
		PreRun:  logPreRun,
		Run: func(cmd *cobra.Command, args []string) {
			l := reflect.ValueOf(tool).MethodByName(commandName).Call(nil)[0].Interface()
			if err := l.(reco.Job).Log(args[0], os.Stdout); err != nil {
				exitWithError(err)
			}
		},
	}
}
