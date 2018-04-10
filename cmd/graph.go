package cmd

import (
	"errors"
	"os/exec"
	"runtime"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco"
	"github.com/ReconfigureIO/reco/logger"
)

var errorGraphNotFound = errors.New("No graph with that ID could be found. Run 'reco graph list' to view available graphs")

var graphCmd = &cobra.Command{
	Use:     "graph",
	Aliases: []string{"g", "graphs"},
	Short:   "Manage your graphs",
	Long: `Manage your graphs.
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

var errInvalidGraphSourceDirectory = errors.New("Invalid source directory. Directory must have a main.go file")

func init() {
	graphCmd.AddCommand(
		graphCmdGenerate,
		graphCmdOpen,
		genListSubcommand("graphs", "Graph"),
	)
	graphCmd.PersistentFlags().StringVar(&project, "project", project, "Project to use. If unset, the active project is used")

	RootCmd.AddCommand(graphCmd)
}

func generateGraph(cmd *cobra.Command, args []string) {
	if !validGraphDir(srcDir) {
		exitWithError(errInvalidGraphSourceDirectory)
	}
	id, err := tool.Graph().Generate(reco.Args{srcDir})
	if err != nil {
		exitWithError(interpretErrorGraph(err))
	}
	logger.Std.Println("Graph submitted. Run 'reco graph list' to track the status of your graph")
	logger.Std.Println("Once the graph has been completed run 'reco graph open " + id + "' to view it")
}

func openGraph(_ *cobra.Command, args []string) {
	if len(args) == 0 {
		exitWithError("ID required")
	}
	file, err := tool.Graph().Open(args[0])
	if err != nil {
		exitWithError(interpretErrorGraph(err))
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

func interpretErrorGraph(err error) error {
	switch err {
	case reco.ErrNotFound:
		return errorGraphNotFound
	default:
		return err
	}
}

func validGraphDir(srcDir string) bool {
	// Do we have a main.go to graph?
	if !hasMain(srcDir) {
		return false
	}
	return true
}
