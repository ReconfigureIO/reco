package cmd

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/ReconfigureIO/reco"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var provider string
var project string
var srcDir string
var tool reco.Client

var errInvalidSourceDirectory = errors.New("invalid source directory. Directory and all cmd/<directory> subdirectories must have a main.go file")

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "reco",
	Short: "reco is the Reconfigure.io command line tool",
	Long:  `reco is the Reconfigure.io command line tool`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	// template
	RootCmd.SetUsageTemplate(usageTemplate)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", `Config file (default "`+filepath.Join(getConfigDir(), "reco.yml")+`")`)
	RootCmd.PersistentFlags().StringVar(&provider, "provider", "", "Service provider")
	RootCmd.PersistentFlags().StringVarP(&srcDir, "source", "s", "", `Source directory (default is current directory "`+getCurrentDir()+`")`)
	RootCmd.PersistentFlags().StringVar(&project, "project", project, "Project to use. If unset, the active project is used")

	// hide provider and config. It is for internal use
	RootCmd.PersistentFlags().MarkHidden("provider")
	RootCmd.PersistentFlags().MarkHidden("config")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var recoDir string
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
		recoDir = filepath.Dir(cfgFile)
	} else {
		recoDir = getConfigDir()
		if err := os.MkdirAll(recoDir, 0755); err != nil {
			exitWithError(err)
		}
	}

	viper.AddConfigPath(recoDir)
	viper.SetConfigName("reco") // name of config file (without extension)
	viper.AutomaticEnv()        // read in environment variables that match

	// If a config file is found, read it in.
	viper.ReadInConfig()

	// save config for other uses.
	viper.Set(reco.GlobalConfigDirKey, recoDir)

	// source directory
	if srcDir == "" {
		srcDir = getCurrentDir()
	}

	// local config dir
	configDir := filepath.Join(srcDir, ".reco")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		exitWithError(err)
	}
	viper.Set(reco.ConfigDirKey, configDir)
}

func initTool() {
	viper.Set("project", project)
	tool = reco.NewClient()
	if err := tool.Init(); err != nil {
		exitWithError(err)
	}
}

var initializeCmd = func(*cobra.Command, []string) {
	initConfig()
	initTool()
}

const usageTemplate = `Usage:{{if .Runnable}}
  {{if .HasAvailableFlags}}{{appendIfNotPresent .UseLine "[flags]"}}{{else}}{{.UseLine}}{{end}}{{end}}{{if .HasAvailableSubCommands}}
  {{ .CommandPath}} [command]{{end}}{{if gt .Aliases 0}}

Aliases:
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

Examples:
{{ .Example }}{{end}}{{if .HasAvailableSubCommands}}
{{if eq .Name "reco"}}
Development Commands: {{range .Commands}}{{if eq .Annotations.type "dev"}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{end}}

Other Commands: {{range .Commands}}{{if .Annotations.type}}{{else}}{{if .IsAvailableCommand}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{end}}{{else}}
Available Commands:{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}{{end}}

Flags:
{{.LocalFlags.FlagUsages | trimRightSpace}}{{if .HasAvailableInheritedFlags}}

Global Flags:
{{.InheritedFlags.FlagUsages | trimRightSpace}}{{end}}{{if .HasHelpSubCommands}}

Additional help topics:{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

Use "{{.CommandPath}} [command] --help" for more information about a command.{{end}}
`
