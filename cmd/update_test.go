// +build integration

package cmd

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
)

func TestLatestRelease(t *testing.T) {
	latest, err := latestRelease(github.NewClient(nil))
	if err != nil {
		t.Error(err)
	}
	if latest == "" {
		t.Error("Returned string is empty")
	}
	fmt.Println(latest)
}
