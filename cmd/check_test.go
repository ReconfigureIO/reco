package cmd

import (
	"runtime"
	"testing"
)

func init() {
	if runtime.GOOS == "windows" {
		srcDir = "C:\reco-examples\addition"
	}
}

func TestMakeVirualGoPathWorks(t *testing.T) {
	err := recocheckDep{}.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}
}
