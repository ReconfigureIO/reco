package cmd

import (
	"github.com/ReconfigureIO/cobra"
)

// authCmd represents the auth command
var authCmd = &cobra.Command{
	Use:     "auth",
	Aliases: []string{"login"},
	Short:   "Authenticate your account",
	Long: `Authenticate your account.

You will be directed to your reconfigure.io dashboard to copy your API key.
An oauth login flow may be required to access reconfigure.io.
`,
	Run:    auth,
	PreRun: initializeCmd,
}

func init() {
	RootCmd.AddCommand(authCmd)
}

func auth(cmd *cobra.Command, args []string) {
	var token = ""
	if len(args) > 0 {
		token = args[0]
	}
	if err := tool.Auth(token); err != nil {
		exitWithError(err)
	}
}
