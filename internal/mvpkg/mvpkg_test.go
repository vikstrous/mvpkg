package mvpkg_test

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/vikstrous/mvpkg/internal/mvpkg"
)

const templateDir = "testtemplate"
const testDir = "testdir"

func cleanup() {
	// ignore errors
	os.RemoveAll(testDir)
}

func TestBasic(t *testing.T) {
	cleanup()
	defer cleanup()

	// Set up the test package structure
	// lazy test code... using a binary dependency rather than a library one
	err := exec.Command("cp", "-r", templateDir, testDir).Run()
	if err != nil {
		t.Fatalf("failed to create test dir: %s", err)
	}

	// execute the package move
	pwd := testDir
	src := "source/testpkg"
	dst := "destination/testpkg"
	err = mvpkg.MvPkg(pwd, src, dst, false, false, true)
	if err != nil {
		t.Fatalf("failed to run mvpkg: %s", err)
	}

	// validate the results
	_, err = os.Stat(filepath.Join(pwd, src, "testpkg.go"))
	if err == nil {
		t.Fatalf("src file still exists")
	}
	diffFiles(t, filepath.Join(templateDir, src, "testpkg.go"), filepath.Join(pwd, dst, "testpkg.go"), true)
	diffFiles(t, filepath.Join(templateDir, "destination", "destination.go"), filepath.Join(pwd, "destination", "destination.go"), false)
	diffFiles(t, filepath.Join(templateDir, "destination", "destination.go.expected"), filepath.Join(pwd, "destination", "destination.go"), true)
}

func diffFiles(t testing.TB, expected, actual string, shouldEqual bool) {
	expectedBytes, err := ioutil.ReadFile(expected)
	if err != nil {
		t.Fatalf("failed to read destination file: %s", err)
	}
	actualBytes, err := ioutil.ReadFile(actual)
	if err != nil {
		t.Fatalf("failed to read destination file: %s", err)
	}
	if !bytes.Equal(expectedBytes, actualBytes) == shouldEqual {
		comparison := "equal"
		if shouldEqual {
			comparison = "differ"
		}
		t.Fatalf("source and destination files %s: %s\n----\nexpected vs actual\n----\n%s", comparison, expectedBytes, actualBytes)
	}
}
