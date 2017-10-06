package cmd

import (
	"testing"
)

func TestMakeVirualGoPathWorks(t *testing.T) {
	srcDir = getCurrentDir()

	err := recocheckDep{}.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}
}
