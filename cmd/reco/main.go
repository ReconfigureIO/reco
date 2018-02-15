package main

import "github.com/ReconfigureIO/reco/cmd"

func main() {
	// set build info
	cmd.BuildInfo.Version = version
	cmd.BuildInfo.BuildTime = buildTime
	cmd.BuildInfo.Builder = builder
	cmd.BuildInfo.GoVersion = goversion
	cmd.BuildInfo.Target = target

	// execute
	cmd.Execute()
}

var (
	version   string
	buildTime string
	builder   string
	goversion string
	target    string
)
