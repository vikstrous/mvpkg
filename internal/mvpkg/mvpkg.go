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
	"time"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type pkgMover struct {
	log               func(s string, args ...interface{})
	dryRun            bool
	modulePkgPath     string
	moduleDir         string
	pkgs              []*packages.Package
	alreadyMovedPkgs  map[string]string
	alreadyMovedFiles map[string]string
	printConfig       *printer.Config
}

func (p *pkgMover) init(pwd string, flags []string) error {
	mod, modDir, usingModules := goModuleNameAndPath(pwd)
	if !usingModules {
		return fmt.Errorf("Not using go modules! Couldn't find go.mod file")
	}
	p.modulePkgPath = mod
	p.moduleDir = modDir
	loadPath := mod + "/..."
	p.log("Loading %s\n", loadPath)
	pkgs, err := packages.Load(&packages.Config{Tests: true, BuildFlags: flags, Dir: modDir, Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports}, loadPath)
	if err != nil {
		return fmt.Errorf("error loading packages %s: %s", loadPath, err)
	}
	p.log("Loaded %d packages\n", len(pkgs))
	p.pkgs = pkgs
	return nil
}

func (p *pkgMover) move(src, dst string) error {
	srcPkgPath := path.Clean(path.Join(p.modulePkgPath, src))
	dstPkgPath := path.Clean(path.Join(p.modulePkgPath, dst))
	var srcPkg *packages.Package
	var srcExtTestPkg *packages.Package
	for _, pkg := range p.pkgs {
		if p.getPkgPath(pkg.PkgPath) == srcPkgPath {
			srcPkg = pkg
		}
		if p.getPkgPath(pkg.PkgPath) == srcPkgPath+"_test" {
			srcExtTestPkg = pkg
		}
	}
	if srcPkg == nil {
		return fmt.Errorf("Couldn't find source package %s", srcPkgPath)
	}
	dstDir := path.Join(p.moduleDir, dst)
	if p.dryRun {
		p.log("would create directory %s\n", dstDir)
	} else {
		p.log("creating directory %s\n", dstDir)
		err := os.MkdirAll(dstDir, 0755)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %s", dstDir, err)
		}
	}
	srcFiles := srcPkg.GoFiles
	if srcExtTestPkg != nil {
		srcFiles = append(srcFiles, srcExtTestPkg.GoFiles...)
	}
	for _, filename := range srcFiles {
		newPath := path.Join(dstDir, path.Base(filename))
		if p.dryRun {
			p.log("would move %s to %s\n", filename, newPath)
		} else {
			p.log("moving %s to %s\n", filename, newPath)
			err := os.Rename(filename, newPath)
			if err != nil {
				return fmt.Errorf("error moving %s to %s: %s", filename, newPath, err)
			}
		}
		p.alreadyMovedFiles[filename] = newPath
	}
	p.alreadyMovedPkgs[srcPkgPath] = dstPkgPath
	p.alreadyMovedPkgs[srcPkgPath+"_test"] = dstPkgPath + "_test"
	return nil
}

func (p *pkgMover) fixImports(src, dst string) error {
	srcPkgPath := p.getPkgPath(path.Clean(path.Join(p.modulePkgPath, src)))
	dstPkgPath := p.getPkgPath(path.Clean(path.Join(p.modulePkgPath, dst)))

	packagesToFix := []*packages.Package{}
	for _, pkg := range p.pkgs {
		for imp := range pkg.Imports {
			if imp == srcPkgPath {
				packagesToFix = append(packagesToFix, pkg)
				break
			}
		}
	}
	p.log("Updating packages: %d\n", len(packagesToFix))

	fset := token.NewFileSet()

	for _, pkg := range packagesToFix {
		for _, filename := range pkg.GoFiles {
			filename := p.getFilePath(filename)
			srcBytes, err := ioutil.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("error reading file %s: %s", filename, err)
			}

			file, err := parser.ParseFile(fset, filename, srcBytes, parser.ParseComments)
			if err != nil {
				return fmt.Errorf("error parsing file %s: %s", filename, err)
			}

			if astutil.RewriteImport(fset, file, srcPkgPath, dstPkgPath) {
				ast.SortImports(fset, file)
				var buf bytes.Buffer
				err := p.printConfig.Fprint(&buf, fset, file)
				if err != nil {
					return fmt.Errorf("error formatting file %s: %s", file.Name.Name, err)
				}
				if p.dryRun {
					p.log("would rewrite %s\n", filename)
				} else {
					p.log("rewriting %s\n", filename)
					err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
					if err != nil {
						return fmt.Errorf("error writing file %s: %s", filename, err)
					}
				}
			}
		}
	}
	return nil
}

func (p *pkgMover) getPkgPath(path string) string {
	newPath, ok := p.alreadyMovedPkgs[path]
	if ok {
		return newPath
	}
	return path
}

// getFilePath is meant to be used on files (not directories) with file paths relvative to the root of the module
func (p *pkgMover) getFilePath(filePath string) string {
	newPath, ok := p.alreadyMovedFiles[filePath]
	if ok {
		return newPath
	}
	return filePath
}

type movePair struct {
	src string
	dst string
}

// MvPkg moves a package from a source to a destination path within the same go module
func MvPkg(printf func(s string, args ...interface{}), pwd, rootSrc, rootDst string, flags []string, dryRun bool, recursive bool) error {
	start := time.Now()
	defer func() {
		printf("done in %s\n", time.Since(start))
	}()
	rootSrc = filepath.Clean(rootSrc)
	rootDst = filepath.Clean(rootDst)
	if path.Base(rootSrc) != path.Base(rootDst) {
		return fmt.Errorf("Soruce and destination package names are not the same. Renaming not supported yet.")
	}
	mover := &pkgMover{log: printf, dryRun: dryRun, alreadyMovedPkgs: map[string]string{}, alreadyMovedFiles: map[string]string{}, printConfig: &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}}
	err := mover.init(pwd, flags)
	if err != nil {
		return fmt.Errorf("failed to initialize mover: %w", err)
	}
	mPairs := []movePair{{src: rootSrc, dst: rootDst}}
	// add additional packages to the list of packages to move if we are using recursive mode
	// TODO: consider using the ./... syntax instead in the future instead of a flag
	// TODO: consider allowing multiple sources and destinations to be specified with , separators
	if recursive {
		err := filepath.Walk(filepath.Join(mover.moduleDir, rootSrc), func(filePath string, info os.FileInfo, err error) error {
			if info == nil || err != nil {
				return err
			}
			if !info.IsDir() {
				return nil
			}
			// TODO: we should turn the relative filePath in to a relative module path before processing the rest here
			srcFilePath, err := filepath.Rel(mover.moduleDir, filePath)
			if err != nil {
				return fmt.Errorf("failed to parse walk path as relative to module root: %s : %s", filePath, mover.moduleDir)
			}
			// XXX: we are comparing a filepath with what's supposed to be a relvative module path. That's not great.
			if srcFilePath == rootSrc {
				return nil
			}
			// modify dst to include the suffix from src being a subdirectory
			// XXX: we are still treating filepaths and module paths interchanably here
			srcSuffix, err := filepath.Rel(rootSrc, srcFilePath)
			dst := path.Join(rootDst, srcSuffix)
			if err != nil {
				return fmt.Errorf("failed to srcFilePath as relative to rootSrc: %s : %s", srcFilePath, rootSrc)
			}
			mPairs = append(mPairs, movePair{src: srcFilePath, dst: dst})
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to walk src tree: %s", err)
		}
	}
	for _, mPair := range mPairs {
		printf("Move plan: %s -> %s\n", mPair.src, mPair.dst)
	}

	for _, mPair := range mPairs {
		printf("Processing %s -> %s\n", mPair.src, mPair.dst)
		err = mover.fixImports(mPair.src, mPair.dst)
		if err != nil {
			return fmt.Errorf("failed to fix imports for %s -> %s: %s", mPair.src, mPair.dst, err)
		}
		err = mover.move(mPair.src, mPair.dst)
		if err != nil {
			return fmt.Errorf("failed to move %s to %s: %s", mPair.src, mPair.dst, err)
		}
	}

	return nil
}
