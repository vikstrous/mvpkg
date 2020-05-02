package mvpkg

import (
	"io/ioutil"
	"path/filepath"
	"regexp"
)

// adapted from https://github.com/99designs/gqlgen/blob/8ed2ec599b8faed3751177fd4335b1b3c3a79922/internal/code/imports.go
// goModuleNameAndPath returns the name of the current go module if there is a go.mod file in the directory tree
// If not, it returns false.
func goModuleNameAndPath(dir string) (string, string, bool) {
	modregex := regexp.MustCompile("module (.*)\n")

	dir, err := filepath.Abs(dir)
	if err != nil {
		panic(err)
	}

	dir = filepath.ToSlash(dir)
	modDir := dir

	for {
		f, err := ioutil.ReadFile(filepath.Join(modDir, "go.mod"))
		if err == nil {
			// found it, stop searching
			return string(modregex.FindSubmatch(f)[1]), modDir, true
		}

		parentDir, err := filepath.Abs(filepath.Join(modDir, ".."))
		if err != nil {
			panic(err)
		}

		if parentDir == modDir {
			// Walked all the way to the root and didnt find anything :'(
			break
		}

		modDir = parentDir
	}

	return "", "", false
}
