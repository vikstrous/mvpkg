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
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// MvPkg moves a package from a source to a destination path within the same go module
func MvPkg(printf func(s string, args ...interface{}), pwd, rootSrc, rootDst string, dryRun bool, recursive bool) error {
	rootSrc = filepath.Clean(rootSrc)
	rootDst = filepath.Clean(rootDst)
	start := time.Now()
	defer func() {
		printf("done in %s\n", time.Since(start))
	}()
	if path.Base(rootSrc) != path.Base(rootDst) {
		return fmt.Errorf("Soruce and destination package names are not the same. Renaming not supported yet.")
	}
	sources := []string{rootSrc}
	if recursive {
		filepath.Walk(filepath.Join(pwd, rootSrc), func(path string, info os.FileInfo, err error) error {
			if info == nil {
				return nil
			}
			newPath := strings.TrimPrefix(path, pwd+"/")
			if newPath == rootSrc {
				return nil
			}
			if info.IsDir() {
				sources = append(sources, newPath)
			}
			return nil
		})
	}
	printf("Moving sources: %s\n", sources)

	for _, src := range sources {
		// XXX: We reload packages that moved after every time we move a package, but we could be smarter about it and remember what we moved!
		mod, modDir, usingModules := goModuleNameAndPath(pwd)
		if !usingModules {
			return fmt.Errorf("Not using go modules! Couldn't find go.mod file")
		}
		loadPath := mod + "/..."
		printf("Loading %s\n", loadPath)
		pkgs, err := packages.Load(&packages.Config{Dir: pwd, Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports}, loadPath)
		if err != nil {
			return fmt.Errorf("error loading packages %s: %s", loadPath, err)
		}
		printf("Loaded %d packages\n", len(pkgs))
		// modify dst to include the suffix from src being a subdirectory
		dst := rootDst
		srcSuffix := strings.TrimPrefix(src, rootSrc)
		dst += srcSuffix
		printf("Processing %s -> %s\n", src, dst)
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
					return fmt.Errorf("error reading file %s: %s", filename, err)
				}

				file, err := parser.ParseFile(fset, filename, srcBytes, parser.ParseComments)
				if err != nil {
					return fmt.Errorf("error parsing file %s: %s", filename, err)
				}

				if astutil.RewriteImport(fset, file, srcPath, dstPath) {
					ast.SortImports(fset, file)
					var buf bytes.Buffer
					err := printConfig.Fprint(&buf, fset, file)
					if err != nil {
						return fmt.Errorf("error formatting file %s: %s", file.Name.Name, err)
					}
					if dryRun {
						printf("would rewrite %s\n", filename)
					} else {
						printf("rewriting %s\n", filename)
						err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
						if err != nil {
							return fmt.Errorf("error writing file %s: %s", filename, err)
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
				return fmt.Errorf("error creating directory %s: %s", dstDir, err)
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
					return fmt.Errorf("error moving %s to %s: %s", filename, newPath, err)
				}
			}
		}
	}

	return nil
}
