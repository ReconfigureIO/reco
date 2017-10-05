package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ReconfigureIO/reco/downloader"
	"github.com/abiosoft/goutils/env"
	"github.com/mholt/archiver"
	"github.com/spf13/cobra"
)

var dependencies = map[string]dep{
	"check": recocheckDep{},
}

// checkCmd represents the version command
var checkCmd = &cobra.Command{
	Use:    "check",
	Short:  "Verify that the compiler can build your Go source",
	Run:    check,
	PreRun: initializeCmd,
	Annotations: map[string]string{
		"type": "dev",
	},
}

func init() {
	RootCmd.AddCommand(checkCmd)
}

func check(_ *cobra.Command, args []string) {
	dep := dependencies["check"]
	if !dep.Installed() {
		fmt.Println("check dependency not found, installing...")
		if err := dep.Fetch(); err != nil {
			exitWithError(err)
		}
	}
	srcFile := filepath.Join(srcDir, "main.go")
	if len(args) > 0 {
		// ignore other args
		srcFile = args[0]
	}
	if _, err := os.Stat(srcFile); err != nil {
		exitWithError(fmt.Errorf("%s not found", srcFile))
	}

	var out bytes.Buffer
	cmd := exec.Command(filepath.Join(dep.Dir(), "bin", "reco-check"), srcFile)
	cmd.Stdout = &out
	cmd.Stderr = &out

	envs := env.EnvVar(os.Environ())
	gopath := []string{
		filepath.Join(dep.Dir(), "gopath"),
		dep.VendorDir(),
	}
	envs.Set("GOPATH", strings.Join(gopath, string([]rune{filepath.ListSeparator})))
	envs.Set("PATH", filepath.Join(dep.Dir(), "bin")+
		string([]byte{filepath.ListSeparator})+
		envs.Get("PATH"))
	cmd.Env = envs

	if cmd.Run() != nil {
		exitWithError(fmt.Errorf("error(s) found while checking %s\n\n%s", srcFile, out.String()))
	} else {
		fmt.Println(srcFile, "checked successfully")
	}
}

type dep interface {
	Name() string
	Fetch() error
	Installed() bool
	Dir() string
	VendorDir() string
}

// File is a convenience wrapper to get platform
// independent filename. e.g. strip .exe from filename
// for windows.
type File string

func (f File) Name() string {
	if runtime.GOOS == "windows" {
		if strings.HasSuffix(string(f), ".exe") {
			return strings.TrimSuffix(string(f), ".exe")
		}
	}
	return string(f)
}

// TODO make this version configurable
const recocheckURL = "https://s3.amazonaws.com/reconfigure.io/reco-check/bundles/reco-check-bundle-latest-x86_64-%s.zip"

type recocheckDep struct {
	once sync.Once
}

var _ dep = recocheckDep{}

func (r recocheckDep) Name() string {
	return "reco-check"
}

func (r recocheckDep) Fetch() error {
	var label string
	switch runtime.GOOS {
	case "linux":
		label = "unknown-linux"
	case "darwin":
		label = "apple-darwin"
	case "windows":
		label = "pc-windows-msvc"
	default:
		return fmt.Errorf("unsupported platform %v", runtime.GOOS)
	}
	downloadURL := fmt.Sprintf(recocheckURL, label)
	dlFile, err := downloader.FromURL(downloadURL)
	if err != nil {
		return err
	}
	defer os.Remove(dlFile)
	return archiver.Zip.Open(dlFile, r.Dir())
}

func (r recocheckDep) Installed() bool {
	r.once.Do(func() {
		r.makeVirtualGoPath()
	})

	type reqFile struct {
		name  string
		dir   bool
		found bool
	}
	files := []reqFile{
		{name: "bin", dir: true},
		{name: "gopath", dir: true},
	}
	dir, err := os.Open(r.Dir())
	if err != nil {
		return false
	}
	stats, err := dir.Readdir(-1)
	if err != nil {
		return false
	}
outer:
	for _, stat := range stats {
		for i := range files {
			if File(stat.Name()).Name() == files[i].name {
				fileName := filepath.Join(r.Dir(), files[i].name)
				stat1, err := os.Stat(fileName)
				if err == nil && files[i].dir == stat1.IsDir() {
					files[i].found = true
				}
				continue outer
			}
		}
	}
	for i := range files {
		if !files[i].found {
			return false
		}
	}

	return true
}

func (r recocheckDep) Dir() string {
	return filepath.Join(getConfigDir(), "plugins", r.Name())
}

func (r recocheckDep) VendorDir() string {
	return filepath.Join(srcDir, ".reco-work", "gopath")
}

func (r recocheckDep) makeVirtualGoPath() {
	os.MkdirAll(r.VendorDir(), 0755)
	vendorDir := filepath.Join(srcDir, "vendor")
	stat, err := os.Stat(vendorDir)
	if err == nil && stat.IsDir() {
		os.Symlink(vendorDir, filepath.Join(r.VendorDir(), "src"))
	}
}
