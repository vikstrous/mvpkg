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
func setup(t testing.TB) {
	cleanup()
	// lazy test code... using a binary dependency rather than a library one
	err := exec.Command("cp", "-r", templateDir, testDir).Run()
	if err != nil {
		t.Fatalf("failed to create test dir: %s", err)
	}
}

func TestBasic(t *testing.T) {
	setup(t)
	defer cleanup()

	// execute the package move
	err := mvpkg.MvPkg(t.Logf, testDir, "source/testpkg", "destination/testpkg", []string{"-tags=special"}, false, false)
	if err != nil {
		t.Fatalf("failed to run mvpkg: %s", err)
	}

	// validate the results
	_, err = os.Stat(filepath.Join(testDir, "source/testpkg/testpkg.go"))
	if err == nil {
		t.Fatalf("src file still exists")
	}
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg.go"), filepath.Join(testDir, "destination/testpkg/testpkg.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg_tag.go"), filepath.Join(testDir, "destination/testpkg/testpkg_tag.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg_test.go"), filepath.Join(testDir, "destination/testpkg/testpkg_test.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg_ext_test.go"), filepath.Join(testDir, "destination/testpkg/testpkg_ext_test.go"), true)
	diffFiles(t, filepath.Join(templateDir, "destination/destination.go.expected.non-recursive"), filepath.Join(testDir, "destination/destination.go"), true)
	diffFiles(t, filepath.Join(templateDir, "destination/destination_test.go.expected.non-recursive"), filepath.Join(testDir, "destination/destination_test.go"), true)
}

func TestRecursive(t *testing.T) {
	setup(t)
	defer cleanup()

	// execute the package move
	err := mvpkg.MvPkg(t.Logf, testDir, "source/testpkg", "destination/testpkg", []string{"-tags=special"}, false, true)
	if err != nil {
		t.Fatalf("failed to run mvpkg: %s", err)
	}

	// validate the results
	_, err = os.Stat(filepath.Join(testDir, "source/testpkg/testpkg.go"))
	if err == nil {
		t.Fatalf("src file still exists")
	}
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg.go"), filepath.Join(testDir, "destination/testpkg", "testpkg.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg_tag.go"), filepath.Join(testDir, "destination/testpkg/testpkg_tag.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg_test.go"), filepath.Join(testDir, "destination/testpkg", "testpkg_test.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/testpkg_ext_test.go"), filepath.Join(testDir, "destination/testpkg", "testpkg_ext_test.go"), true)
	diffFiles(t, filepath.Join(templateDir, "source/testpkg/nested/nested.go.expected.recursive"), filepath.Join(testDir, "destination/testpkg/nested/nested.go"), true)
	diffFiles(t, filepath.Join(templateDir, "destination/destination.go.expected.recursive"), filepath.Join(testDir, "destination/destination.go"), true)
	diffFiles(t, filepath.Join(templateDir, "destination/destination_test.go.expected.recursive"), filepath.Join(testDir, "destination/destination_test.go"), true)
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
		t.Fatalf("source and destination files %s:\n%s\n----\nexpected vs actual\n----\n%s", comparison, expectedBytes, actualBytes)
	}
}
