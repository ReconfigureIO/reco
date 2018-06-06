package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ReconfigureIO/cobra"
	"github.com/ReconfigureIO/reco/downloader"
	"github.com/ReconfigureIO/reco/logger"
	"github.com/abiosoft/goutils/env"
	"github.com/reconfigureio/archiver"
)

var dependencies = map[string]dep{
	"check": recocheckDep{},
}

// checkCmd represents the version command
var checkCmd = &cobra.Command{
	Use:    "check",
	Short:  "Verify that the Reconfigure.io compiler can build your Go source code.",
	Run:    check,
	PreRun: initializeCmd,
	Annotations: map[string]string{
		"type": "dev",
	},
}

var checkCmdVars struct {
	update bool
}

func init() {
	checkCmd.PersistentFlags().BoolVar(&checkCmdVars.update, "update", false, "check for and install dependency updates")
	RootCmd.AddCommand(checkCmd)
}

func check(_ *cobra.Command, args []string) {
	dep := dependencies["check"]
	update := func() {
		if err := dep.Fetch(); err != nil {
			exitWithError(err)
		}
	}
	switch dep.Installed() {
	case statusUpdateAvailable:
		fmt.Println("check update available, updating...")
		update()
	case statusNotInstalled:
		fmt.Println("check not found, installing...")
		update()
	}
	srcFile := filepath.Join(srcDir, "main.go")
	if len(args) > 0 {
		// ignore other args
		srcFile = args[0]
	}
	if _, err := os.Stat(srcFile); err != nil {
		exitWithError(fmt.Errorf("%s not found", srcFile))
	}

	//reco-check expects absolute paths so convert relative to absolute here
	absSrcFile, err := filepath.Abs(srcFile)
	if err != nil {
		exitWithError(err)
	}
	var out bytes.Buffer
	cmd := exec.Command(filepath.Join(dep.Dir(), "bin", "reco-check"), absSrcFile)
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
		exitWithError(fmt.Errorf("error(s) found while checking %s\n\n%s", absSrcFile, out.String()))
	} else {
		fmt.Println(srcFile, "checked successfully")
	}
}

type depStatus int8

const (
	statusNotInstalled depStatus = iota
	statusUpdateAvailable
	statusInstalled
)

type dep interface {
	Name() string
	Fetch() error
	Installed() depStatus
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

const (
	// TODO make this version configurable
	recocheckURL = "https://s3.amazonaws.com/reconfigure.io/reco-check/bundles/reco-check-bundle-latest-x86_64-%s.zip"
	eTag         = "ETag"
)

type recocheckDep struct {
	once sync.Once
}

var _ dep = recocheckDep{}

func (r recocheckDep) Name() string {
	return "reco-check"
}

func (r recocheckDep) downloadURL() (string, error) {
	var label string
	switch runtime.GOOS {
	case "linux":
		label = "unknown-linux"
	case "darwin":
		label = "apple-darwin"
	case "windows":
		label = "pc-windows-msvc"
	default:
		return "", fmt.Errorf("unsupported platform %v", runtime.GOOS)
	}
	return fmt.Sprintf(recocheckURL, label), nil
}

func (r recocheckDep) Fetch() error {
	downloadURL, err := r.downloadURL()
	if err != nil {
		return err
	}
	resp, err := http.Get(downloadURL)
	if err != nil {
		return err
	}
	dlFile, err := downloader.FromReader(resp.Body, resp.ContentLength)
	if err != nil {
		return err
	}
	defer os.Remove(dlFile)

	err = ioutil.WriteFile(r.eTagFile(), []byte(resp.Header.Get(eTag)), 0644)
	if err != nil {
		logger.Info.Println("could not persist dependency version ", err)
	}
	return archiver.Zip.Open(dlFile, r.Dir())
}

func (r recocheckDep) latestETag() (string, error) {
	downloadURL, err := r.downloadURL()
	if err != nil {
		return "", err
	}
	resp, err := http.Get(downloadURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Header.Get(eTag), nil
}

func (r recocheckDep) eTagFile() string {
	return filepath.Join(r.Dir(), eTag)
}

func (r recocheckDep) Installed() depStatus {
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
		return statusNotInstalled
	}
	stats, err := dir.Readdir(-1)
	if err != nil {
		return statusNotInstalled
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
			return statusNotInstalled
		}
	}

	// update
	if checkCmdVars.update {
		// validate tag
		b, err := ioutil.ReadFile(r.eTagFile())
		if err != nil {
			return statusUpdateAvailable
		}
		latestETag, err := r.latestETag()
		if err != nil {
			logger.Info.Println("could not check for updates")
		}

		if string(b) != latestETag {
			return statusUpdateAvailable
		}
	}

	return statusInstalled
}

func (r recocheckDep) Dir() string {
	return filepath.Join(getConfigDir(), "plugins", r.Name())
}

func (r recocheckDep) VendorDir() string {
	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		exitWithError(err)
	}
	return filepath.Join(srcDir, ".reco-work", "gopath")
}

func (r recocheckDep) makeVirtualGoPath() error {
	os.MkdirAll(r.VendorDir(), 0755)
	srcDir, err := filepath.Abs(srcDir)
	if err != nil {
		exitWithError(err)
	}
	vendorDir := filepath.Join(srcDir, "vendor")
	stat, err := os.Stat(vendorDir)

	if err != nil {
		if pErr, ok := err.(*os.PathError); ok {
			switch pErr.Err.Error() {
			case os.ErrNotExist.Error():
				return nil
			case "no such file or directory":
				return nil
			case "The system cannot find the file specified.":
				return nil
			}
		}

		return err
	}

	if stat.IsDir() {
		virtualVendorDir := filepath.Join(r.VendorDir(), "src")
		return symlink(vendorDir, virtualVendorDir)
	}
	return nil
}

func symlink(src, dest string) error {
	if runtime.GOOS != "windows" {
		return os.Symlink(src, dest)
	}

	// windows
	// regular symlink doesn't give desired result,
	// file copying seems to be the best workaround so far.
	// But there are concerns for when the vendor directory
	// grows really large.
	os.RemoveAll(dest)
	cmd := exec.Command("cmd.exe", "/C", fmt.Sprintf("xcopy /E /Y /I %s %s", src, dest))
	return cmd.Run()
}
