package cmd

import (
	"fmt"

	"github.com/ReconfigureIO/cobra"
)

func genDevCommand(name string, jobType string, aliases ...string) *cobra.Command {
	print := name
	if len(aliases) > 0 {
		print += ", " + aliases[0]
	}
	cmd := &cobra.Command{
		Use:     name,
		Aliases: aliases,
		Short:   fmt.Sprintf("Manage your %ss", jobType),
		Long: fmt.Sprintf(`Manage your %ss.
You can start, stop, list %ss and stream logs.`, jobType, jobType),
		PersistentPreRun: initializeCmd,
		Annotations: map[string]string{
			"type": "dev",
		},
	}

	cmd.AddCommand()

	return cmd
}
