package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func genDevCommand(name string, aliases ...string) *cobra.Command {
	print := name
	if len(aliases) > 0 {
		print += ", " + aliases[0]
	}
	cmd := &cobra.Command{
		Use:     name,
		Aliases: aliases,
		Short:   fmt.Sprintf("Manage %ss", name),
		Long: fmt.Sprintf(`Manage %ss.
You can start, stop, list %ss and stream logs.`, name, name),
		PersistentPreRun: initializeCmd,
		Annotations: map[string]string{
			"type": "dev",
		},
	}

	cmd.AddCommand(
		genListSubcommand(name),
		genLogSubcommand(name),
		genStopSubcommand(name),
	)

	return cmd
}
