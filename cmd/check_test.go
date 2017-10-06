package cmd

import (
	"testing"
)

func testMakeVirualGoPathWorks(t *testing.T) {
	err := recocheckDep{}.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}
}
