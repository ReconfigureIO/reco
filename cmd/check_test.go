package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMakeVirualGoPathWorksIfNoVendor(t *testing.T) {
	srcDir = getCurrentDir()
	vendorDir := filepath.Join(srcDir, "vendor")
	os.RemoveAll(vendorDir)

	err := recocheckDep{}.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}

	os.RemoveAll(vendorDir)
}

func TestMakeVirualGoPathWorksIfVendor(t *testing.T) {
	srcDir = getCurrentDir()
	vendorDir := filepath.Join(srcDir, "vendor")

	os.RemoveAll(vendorDir)
	os.RemoveAll(recocheckDep{}.VendorDir())

	os.MkdirAll(vendorDir, 0755)

	err := recocheckDep{}.makeVirtualGoPath()
	if err != nil {
		t.Error(err)
	}

	os.RemoveAll(vendorDir)

}
