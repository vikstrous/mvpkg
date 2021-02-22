package mvpkg_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

func setup(tb testing.TB) {
	cleanup()
	// lazy test code... using a binary dependency rather than a library one
	err := exec.Command("cp", "-r", templateDir, testDir).Run()
	if err != nil {
		tb.Fatalf("failed to create test dir: %s", err)
	}
}

func compare(tb testing.TB, expected, actual string) {
	// lazy test code... using a binary dependency rather than a library one
	cmd := exec.Command("diff", "-r", expected, actual)
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	if err != nil {
		tb.Fatalf("test comparison failed: %s\n%s\n%s\n", err, stdout.String(), stderr.String())
	}
}

func TestGeneric(t *testing.T) {
	testsDir := "tests"

	testFiles, err := ioutil.ReadDir(testsDir)
	if err != nil {
		t.Fatalf("failed to read list of tests: %s", err)
	}

	for _, testFile := range testFiles {
		if !testFile.IsDir() {
			continue
		}
		// Drop the number from the name
		testSrcDir := filepath.Join(testsDir, testFile.Name())
		testName := strings.Join(strings.Split(testFile.Name(), "_")[1:], "_")

		// clean up before starting and between tests
		cleanup()
		t.Run(testName, func(t *testing.T) {
			originalPath := filepath.Join(testSrcDir, "original")
			// using a binary dependency rather than a library one
			err := exec.Command("cp", "-r", originalPath, testDir).Run()
			if err != nil {
				t.Fatalf("failed to create test dir running: %s", err)
			}
			defer func() {
				// leave the resulting failed output directory in place if the tests failed so we can inspect it
				if !t.Failed() {
					cleanup()
				}
			}()

			expectedPath := filepath.Join(testSrcDir, "expected")
			testInfoFilename := filepath.Join(testSrcDir, "testInfo.json")
			testInfoStr, err := ioutil.ReadFile(testInfoFilename)
			if err != nil {
				t.Fatalf("failed to read command file from %s: %s", testInfoFilename, err)
			}
			var testInfo struct {
				PWD         string   `json:"pwd"`
				Source      string   `json:"source"`
				Destination string   `json:"destination"`
				BuildFlags  []string `json:"build_flags"`
			}
			err = json.Unmarshal(testInfoStr, &testInfo)
			if err != nil {
				t.Fatalf("failed to JSON unmarshal file from %s: %s", testInfoFilename, err)
			}

			// run the tool
			err = mvpkg.MvPkg(t.Logf, filepath.Join(testDir, testInfo.PWD), testInfo.Source, testInfo.Destination, testInfo.BuildFlags, false, false)
			if err != nil {
				t.Fatalf("MvPkg fialed: %s", err)
			}

			// validate the results
			compare(t, expectedPath, testDir)
		})
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
