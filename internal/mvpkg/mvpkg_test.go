package mvpkg_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/vikstrous/mvpkg/internal/mvpkg"
)

const testDir = "testdir"

func cleanup() {
	// ignore errors
	os.RemoveAll(testDir)
}

func TestBasic(t *testing.T) {
	cleanup()
	defer cleanup()
	pwd := testDir
	src := "source/testpkg"
	dst := "destination/testpkg"
	for _, dir := range []string{pwd, filepath.Join(pwd, src), filepath.Join(pwd, dst)} {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			t.Fatalf("failed to create test dir %s: %s", dir, err)
		}
	}
	err := ioutil.WriteFile(filepath.Join(pwd, "go.mod"), []byte(`module example.com

go 1.13
`), 0644)
	if err != nil {
		t.Fatalf("failed to create test dir: %s", err)
	}
	testpkgFileBytes := []byte(`package testpkg

import "fmt"

func exampleFunc() {
	fmt.Println("doing nothing")
}
`)

	srcFilename := filepath.Join(pwd, src, "testpkg.go")
	err = ioutil.WriteFile(srcFilename, testpkgFileBytes, 0644)
	if err != nil {
		t.Fatalf("failed to create test dir: %s", err)
	}
	err = mvpkg.MvPkg(pwd, src, dst, false, false, true)
	if err != nil {
		t.Fatalf("failed to run mvpkg: %s", err)
	}
	_, err = os.Stat(srcFilename)
	if err == nil {
		t.Fatalf("src file still exists")
	}
	dstFilename := filepath.Join(pwd, dst, "testpkg.go")
	actualBytes, err := ioutil.ReadFile(dstFilename)
	if err != nil {
		t.Fatalf("failed to read destination file: %s", err)
	}

	if !bytes.Equal(testpkgFileBytes, actualBytes) {
		t.Fatalf("source and destination files differ: %s\n----\nexpected vs actual\n----\n%s", testpkgFileBytes, actualBytes)
	}
}
