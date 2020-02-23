package mvpkg

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"time"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

var errNoModules = fmt.Errorf("Not using go modules! Couldn't find go.mod file")

// MvPkg moves a package from a source to a destination path within the same go module
func MvPkg(pwd, src, dst string, dryRun bool, recursive bool, verbose bool) error {
	start := time.Now()
	defer func() {
		fmt.Printf("done in %s\n", time.Now().Sub(start))
	}()
	printf := func(s string, args ...interface{}) {}
	if verbose || dryRun {
		printf = func(s string, args ...interface{}) {
			fmt.Printf(s, args...)
		}
	}
	if path.Base(src) != path.Base(dst) {
		return fmt.Errorf("Soruce and destination package names are not the same. Renaming not supported yet.")
	}
	mod, modDir, usingModules := goModuleNameAndPath(pwd)
	if !usingModules {
		return errNoModules
	}
	loadPath := mod + "/..."
	printf("Loading %s\n", loadPath)
	pkgs, err := packages.Load(&packages.Config{Dir: pwd, Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports}, loadPath)
	if err != nil {
		return err
	}
	printf("Loaded %d packages\n", len(pkgs))
	srcPath := path.Clean(path.Join(mod, src))
	dstPath := path.Clean(path.Join(mod, dst))
	var srcPkg *packages.Package

	packagesToFix := []*packages.Package{}
	for _, pkg := range pkgs {
		if pkg.PkgPath == srcPath {
			srcPkg = pkg
		}
		for imp := range pkg.Imports {
			if imp == srcPath {
				packagesToFix = append(packagesToFix, pkg)
				break
			}
		}
	}
	if srcPkg == nil {
		return fmt.Errorf("Couldn't find source package %s", srcPath)
	}
	printf("Updating packages: %d\n", len(packagesToFix))

	fset := token.NewFileSet()

	printConfig := &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	for _, pkg := range packagesToFix {
		for _, filename := range pkg.GoFiles {
			srcBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				return err
			}

			file, err := parser.ParseFile(fset, filename, srcBytes, parser.ParseComments)
			if err != nil {
				return err
			}

			if astutil.RewriteImport(fset, file, srcPath, dstPath) {
				ast.SortImports(fset, file)
				var buf bytes.Buffer
				err := printConfig.Fprint(&buf, fset, file)
				if err != nil {
					return err
				}
				if dryRun {
					printf("would rewrite %s\n", filename)
				} else {
					printf("rewriting %s\n", filename)
					err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	dstDir := path.Join(modDir, dst)
	if dryRun {
		printf("would create directory %s\n", dstDir)
	} else {
		printf("creating directory %s\n", dstDir)
		err = os.MkdirAll(dstDir, 0755)
		if err != nil {
			return err
		}
	}
	for _, filename := range srcPkg.GoFiles {
		newPath := path.Join(dstDir, path.Base(filename))
		if dryRun {
			printf("would move %s to %s\n", filename, newPath)
		} else {
			printf("moving %s to %s\n", filename, newPath)
			err = os.Rename(filename, newPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
