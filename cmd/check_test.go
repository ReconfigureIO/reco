package cmd

import (
	"testing"
)

func testMakeVirualGoPathWorks(t *testing.T) {
	err := dep.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}
}
