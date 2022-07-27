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
	"regexp"
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

var errNoGoMod = fmt.Errorf("couldn't find go.mod file")

func (p *pkgMover) init(pwd string, flags []string) error {
	mod, modDir, usingModules := goModuleNameAndPath(pwd)
	if !usingModules {
		return errNoGoMod
	}

	p.modulePkgPath = mod
	p.moduleDir = modDir
	loadPath := mod + "/..."
	p.log("Loading %s\n", loadPath)

	pkgs, err := packages.Load(&packages.Config{Tests: true, BuildFlags: flags, Dir: modDir, Mode: packages.NeedName | packages.NeedFiles | packages.NeedImports}, loadPath)
	if err != nil {
		return fmt.Errorf("error loading packages %s: %w", loadPath, err)
	}

	p.log("Loaded %d packages\n", len(pkgs))
	p.pkgs = pkgs

	return nil
}

func makeRenamer(src, dst string) func(filename string) error {
	renameFrom := path.Base(src)
	renameTo := path.Base(dst)

	if renameFrom == renameTo {
		return func(filename string) error {
			return nil
		}
	}

	return func(filename string) error {
		packageRenamer := regexp.MustCompile(fmt.Sprintf("(?m)^package %s$", regexp.QuoteMeta(renameFrom)))
		testPackageRenamer := regexp.MustCompile(fmt.Sprintf("(?m)^package %s_test$", regexp.QuoteMeta(renameFrom)))

		fileBytes, err := ioutil.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read after move of %s: %w", filename, err)
		}

		fileBytes = packageRenamer.ReplaceAll(fileBytes, []byte(fmt.Sprintf("package %s", renameTo)))
		fileBytes = testPackageRenamer.ReplaceAll(fileBytes, []byte(fmt.Sprintf("package %s_test", renameTo)))

		err = ioutil.WriteFile(filename, fileBytes, 0o600)
		if err != nil {
			return fmt.Errorf("failed to write after package rename of %s: %w", filename, err)
		}

		return nil
	}
}

func (p *pkgMover) move(src, dst string) error {
	srcPkgPath := path.Clean(path.Join(p.modulePkgPath, src))
	dstPkgPath := path.Clean(path.Join(p.modulePkgPath, dst))
	// avoid duplication in case the package files show up more than once
	srcFiles := map[string]struct{}{}

	for _, pkg := range p.pkgs {
		if p.getPkgPath(pkg.PkgPath) == srcPkgPath || p.getPkgPath(pkg.PkgPath) == srcPkgPath+"_test" {
			for _, file := range pkg.GoFiles {
				srcFiles[file] = struct{}{}
			}
		}
	}

	if len(srcFiles) == 0 {
		// nothing to move
		return nil
	}

	dstDir := path.Join(p.moduleDir, dst)

	if p.dryRun {
		p.log("would create directory %s\n", dstDir)
	} else {
		p.log("creating directory %s\n", dstDir)
		err := os.MkdirAll(dstDir, 0o755)
		if err != nil {
			return fmt.Errorf("error creating directory %s: %w", dstDir, err)
		}
	}

	renamer := makeRenamer(src, dst)

	for filename := range srcFiles {
		newPath := path.Join(dstDir, path.Base(filename))

		if p.dryRun {
			p.log("would move %s to %s\n", filename, newPath)
		} else {
			p.log("moving %s to %s\n", filename, newPath)
			err := os.Rename(filename, newPath)
			if err != nil {
				return fmt.Errorf("error moving %s to %s: %w", filename, newPath, err)
			}
			err = renamer(newPath)
			if err != nil {
				return fmt.Errorf("renamer failed: %w", err)
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
			filename = p.getFilePath(filename)

			err := p.fixImportsInFile(fset, src, dst, filename)
			if err != nil {
				return fmt.Errorf("failed to fix imports in %s: %w", filename, err)
			}
		}
	}

	return nil
}

func (p *pkgMover) fixImportsInFile(fset *token.FileSet, src, dst, filename string) error {
	srcPkgPath := p.getPkgPath(path.Clean(path.Join(p.modulePkgPath, src)))
	dstPkgPath := p.getPkgPath(path.Clean(path.Join(p.modulePkgPath, dst)))
	renameFrom := path.Base(src)
	renameTo := path.Base(dst)

	srcBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading file %s: %w", filename, err)
	}

	astFile, err := parser.ParseFile(fset, filename, srcBytes, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing file %s: %w", filename, err)
	}

	if !astutil.RewriteImport(fset, astFile, srcPkgPath, dstPkgPath) {
		return nil
	}

	ast.SortImports(fset, astFile)

	newFile := astutil.Apply(astFile, func(c *astutil.Cursor) bool {
		selExpr, ok := c.Node().(*ast.SelectorExpr)
		if !ok {
			return true
		}

		ident, ok := selExpr.X.(*ast.Ident)
		if !ok {
			return true
		}

		if ident.Name == renameFrom && ident.Obj == nil {
			c.Replace(
				&ast.SelectorExpr{
					Sel: selExpr.Sel,
					X: &ast.Ident{
						NamePos: selExpr.Sel.NamePos,
						Name:    renameTo,
					},
				})
		}

		return true
	}, nil)

	var buf bytes.Buffer

	err = p.printConfig.Fprint(&buf, fset, newFile)
	if err != nil {
		return fmt.Errorf("error formatting file %s: %w", astFile.Name.Name, err)
	}

	if p.dryRun {
		p.log("would rewrite %s\n", filename)
	} else {
		p.log("rewriting %s\n", filename)
		err = ioutil.WriteFile(filename, buf.Bytes(), 0o600)
		if err != nil {
			return fmt.Errorf("error writing file %s: %w", filename, err)
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

// getFilePath is meant to be used on files (not directories) with file paths relvative to the root of the module.
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

// MvPkg moves a package from a source to a destination path within the same go module.
func MvPkg(printf func(s string, args ...interface{}), pwd, rootSrc, rootDst string, flags []string, dryRun bool, recursive bool) error {
	start := time.Now()

	defer func() {
		printf("done in %s\n", time.Since(start))
	}()

	rootSrc = filepath.Clean(rootSrc)
	rootDst = filepath.Clean(rootDst)
	mover := &pkgMover{log: printf, dryRun: dryRun, alreadyMovedPkgs: map[string]string{}, alreadyMovedFiles: map[string]string{}, printConfig: &printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}}

	err := mover.init(pwd, flags)
	if err != nil {
		return fmt.Errorf("failed to initialize mover: %w", err)
	}

	mPairs, err := findMovePairs(rootSrc, rootDst, mover.moduleDir, recursive)
	if err != nil {
		return fmt.Errorf("failed to find move pairs: %w", err)
	}

	for _, mPair := range mPairs {
		printf("Move plan: %s -> %s\n", mPair.src, mPair.dst)
	}

	for _, mPair := range mPairs {
		printf("Processing %s -> %s\n", mPair.src, mPair.dst)

		err = mover.fixImports(mPair.src, mPair.dst)
		if err != nil {
			return fmt.Errorf("failed to fix imports for %s -> %s: %w", mPair.src, mPair.dst, err)
		}

		err = mover.move(mPair.src, mPair.dst)
		if err != nil {
			return fmt.Errorf("failed to move %s to %s: %w", mPair.src, mPair.dst, err)
		}
	}

	return nil
}

func findMovePairs(rootSrc, rootDst, moduleDir string, recursive bool) ([]movePair, error) {
	mPairs := []movePair{{src: rootSrc, dst: rootDst}}

	// add additional packages to the list of packages to move if we are using recursive mode
	if !recursive {
		return mPairs, nil
	}

	err := filepath.Walk(filepath.Join(moduleDir, rootSrc), func(filePath string, info os.FileInfo, iterationErr error) error {
		if info == nil || iterationErr != nil {
			return fmt.Errorf("iteration error: %w", iterationErr)
		}
		if !info.IsDir() {
			return nil
		}
		srcFilePath, err := filepath.Rel(moduleDir, filePath)
		if err != nil {
			return fmt.Errorf("failed to parse walk path %s as relative to module root %s: %w", filePath, moduleDir, err)
		}
		// XXX: we are comparing a filepath with what's supposed to be a relvative module path. That's not great.
		if srcFilePath == rootSrc {
			return nil
		}
		// modify dst to include the suffix from src being a subdirectory
		// XXX: we are still treating filepaths and module paths interchanably here
		srcSuffix, err := filepath.Rel(rootSrc, srcFilePath)
		if err != nil {
			return fmt.Errorf("failed to parse srcFilePath %s as relative to rootSrc %s: %w", srcFilePath, rootSrc, err)
		}
		dst := path.Join(rootDst, srcSuffix)
		mPairs = append(mPairs, movePair{src: srcFilePath, dst: dst})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk src tree: %w", err)
	}

	return mPairs, nil
}
