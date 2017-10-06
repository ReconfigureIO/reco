package cmd

import (
	"testing"
)

func TestMakeVirualGoPathWorks(t *testing.T) {
	err := recocheckDep{}.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}
}
