package cmd

import (
	"os"
	"path/filepath"
	"runtime"
)

func getConfigDir() string {
	dir := filepath.Join(os.Getenv("HOME"), ".config")
	switch runtime.GOOS {
	case "linux":
		if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
			dir = d
		}
	case "darwin":
		dir = filepath.Join(os.Getenv("HOME"), "Library", "Preferences")
	case "windows":
		if d := os.Getenv("APPDATA"); d != "" {
			dir = d
		}
	}
	return filepath.Join(dir, "reco")
}

func getCurrentDir() string {
	dir, _ := os.Getwd()
	return dir
}
