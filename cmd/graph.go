package cmd

import (
	"os/exec"
	"runtime"

	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/ReconfigureIO/cobra"
)

var graphCmd = &cobra.Command{
	Use:     "graph",
	Aliases: []string{"g", "graphs"},
	Short:   "Manage graphs",
	Long: `Manage graphs.
You can generate a dataflow graph, list your graphs or open a previously generated graph.`,
	PersistentPreRun: initializeCmd,
	Annotations: map[string]string{
		"type": "dev",
	},
}

var graphCmdGenerate = &cobra.Command{
	Use:     "gen",
	Aliases: []string{"g", "generate"},
	Short:   "Generate a graph",
	Long: `Generate a dataflow graph for your source code.
This usually takes few minutes.
`,
	Run: generateGraph,
}

var graphCmdOpen = &cobra.Command{
	Use:     "open",
	Aliases: []string{"o"},
	Short:   "Open a generated graph",
	Long: `Open a generated graph.
This attempts to use your default pdf viewer to open the graph.
`,
	Run: openGraph,
}

func init() {
	graphCmd.AddCommand(
		graphCmdGenerate,
		graphCmdOpen,
		genListSubcommand("graphs", "Graph"),
	)

	RootCmd.AddCommand(graphCmd)
}

func generateGraph(cmd *cobra.Command, args []string) {
	if !validBuildDir(srcDir) {
		exitWithError(errInvalidSourceDirectory)
	}
	id, err := tool.Graph().Generate(reco.Args{srcDir})
	if err != nil {
		exitWithError(err)
	}
	logger.Std.Println(id)
}

func openGraph(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("ID required")
	}
	file, err := tool.Graph().Open(args[0])
	if err != nil {
		exitWithError(err)
	}
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", file)
	case "linux":
		if _, err := exec.LookPath("xdg-open"); err != nil {
			break
		}
		cmd = exec.Command("xdg-open", file)
	case "windows":
		cmd = exec.Command("start", file)
	}
	// could not open with default pdf handler.
	if cmd == nil || cmd.Run() != nil {
		logger.Std.Printf("Your graph is available here: %s", file)
		return
	}
}
