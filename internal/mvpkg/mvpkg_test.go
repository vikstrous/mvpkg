package mvpkg_test

import (
	"bytes"
	"os"
	"os/exec"
	"testing"

	"github.com/vikstrous/mvpkg/internal/mvpkg"
)

const (
	templateDir = "testtemplate"
	testDir     = "testdir"
)

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

func compare(t testing.TB, expected, actual string) {
	// lazy test code... using a binary dependency rather than a library one
	cmd := exec.Command("diff", "-r", expected, actual)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("test comparison failed: %s\n%s\n%s\n", err, stdout.String(), stderr.String())
	}
}

func TestBasic(t *testing.T) {
	setup(t)

	defer cleanup()

	// execute the package move
	err := mvpkg.MvPkg(t.Logf, testDir+"/destination", "source/testpkg", "destination/testpkg2", []string{"-tags=special"}, false, false)
	if err != nil {
		t.Fatalf("failed to run mvpkg: %s", err)
	}

	// validate the results
	compare(t, "expected", testDir)
}

func TestRecursive(t *testing.T) {
	setup(t)

	defer cleanup()

	// execute the package move
	err := mvpkg.MvPkg(t.Logf, testDir+"/destination", "source/testpkg", "destination/testpkg2", []string{"-tags=special"}, false, true)
	if err != nil {
		t.Fatalf("failed to run mvpkg: %s", err)
	}

	// validate the results
	compare(t, "expected-recursive", testDir)
}
