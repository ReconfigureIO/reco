package reco

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/mholt/archiver"
)

// M is a convenience wrapper for map[string]interface{}.
type M map[string]interface{}

// String returns value of key as string.
func (m M) String(key string) string {
	return String(m[key])
}

// Int returns value of key as int.
func (m M) Int(key string) int {
	return Int(m[key])
}

// Bool returns value of key as bool.
func (m M) Bool(key string) bool {
	return Bool(m[key])
}

// HasKey checks if key exists in the map.
func (m M) HasKey(key string) bool {
	_, ok := m[key]
	return ok
}

// Int returns int value of v if v is an int,
// or 0 otherwise.
func Int(v interface{}) int {
	if v1, ok := v.(int); ok {
		return v1
	}
	if v1, ok := v.(string); ok {
		n, _ := strconv.Atoi(v1)
		return n
	}
	return 0
}

// String returns string value of v if v is a
// string, or "" otherwise.
func String(v interface{}) string {
	if v1, ok := v.(string); ok {
		return v1
	}
	return ""
}

// Bool returns bool value of v if v is a bool,
// or false otherwise.
func Bool(v interface{}) bool {
	if v1, ok := v.(bool); ok {
		return v1
	}
	if v1, ok := v.(int); ok {
		return v1 > 0
	}
	if v1, ok := v.(string); ok {
		switch strings.ToLower(v1) {
		case "1", "true", "yes":
			return true
		}
	}
	return false
}

// StringSlice returns strings slice of v
// if v is a string slice, otherwise returns
// nil.
func StringSlice(v interface{}) []string {
	if s, ok := v.([]string); ok {
		return s
	}
	return nil
}

// Args is convenience wrapper for []interface
// to fetch element at without wrong index errors.
type Args []interface{}

// At returns element at i.
func (a Args) At(i int) interface{} {
	if len(a) <= i {
		return nil
	}
	return a[i]
}

// Last returns last element.
func (a Args) Last() interface{} {
	if len(a) > 0 {
		return a[len(a)-1]
	}
	return nil
}

// First returns first element.
func (a Args) First() interface{} {
	if len(a) > 0 {
		return a[0]
	}
	return nil
}

func archiveDir(dir string) (string, error) {
	tmp, err := tmpDir()
	if err != nil {
		return "", err
	}
	tmpArchive := path.Join(tmp, "source.tar.gz")

	ignoredFiles := []string{
		".reco-work",
		".reco",
	}
	files := ignoreFiles(dir, ignoredFiles)
	if len(files) == 0 {
		return "", fmt.Errorf("'%s' is empty", dir)
	}

	return tmpArchive, archiver.TarGz.Make(tmpArchive, files)
}

// tmpDir wraps ioutil.TempDir for reco.
func tmpDir() (string, error) {
	tmp := "./.reco-work/.tmp"
	if err := os.MkdirAll(tmp, os.FileMode(0775)); err != nil {
		return "", err
	}
	return ioutil.TempDir(tmp, "reco")
}

// ignoreFiles returns files in src without ignored.
func ignoreFiles(src string, ignored []string) (fs []string) {
	dir, err := os.Open(src)
	if err != nil {
		return
	}
	files, err := dir.Readdirnames(-1)
	if err != nil {
		return
	}
	for _, file := range files {
		found := false
		for _, i := range ignored {
			if _, f := path.Split(file); f == i {
				found = true
				break
			}
		}
		if !found {
			fs = append(fs, file)
		}
	}
	return
}

type timeRounder time.Duration

// Nearest rounds to the nearest duration and returns the string.
func (t timeRounder) Nearest(d time.Duration) string {
	duration := time.Duration(t)
	str := "-"
	if t > 0 {
		// round to duration e.g. sec, min
		// easy trick with integer division
		duration /= d
		duration *= d

		str = duration.String()
	}
	return str
}
