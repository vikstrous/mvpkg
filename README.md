mvpkg
-----

mvpkg is a refactoring tool for Go codebases that allows you to move a package
or a set of packages from one path to another within the same go module. It's
written for go modules and with performance in mind, so let me know if it's not
fast enough for you.


Installation:

```
go get github.com/vikstrous/mvpkg
```

Usage:

```
Usage of mvpkg: mvpkg <src> <dst>

  mvpkg takes two positional arguments: a source and destination path
  It works only withing a single go module and only with go module support enabled.
  The source and destination paths must be relative to the root of the go module

  -dry-run
        print planned actions without executing them
  -recursive
        recursively move all packages nested under the source package
  -v    verbose, print status while running
```
