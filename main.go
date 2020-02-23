package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vikstrous/mvpkg/internal/mvpkg"
)

var dryRunFlag bool
var recursiveFlag bool
var verboseFlag bool
var buildFlagsFlag arrayFlags

func init() {
	flag.BoolVar(&verboseFlag, "v", false, "verbose, print status while running")
	flag.BoolVar(&dryRunFlag, "dry-run", false, "print planned actions without executing them")
	flag.BoolVar(&recursiveFlag, "recursive", false, "recursively move all packages nested under the source package")
	flag.Var(&buildFlagsFlag, "build-flags", "build tags to use while parsing source packages, can be specified morethan once; ex: -build-flags='-tags=foo bar'")
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <src> <dst>\n", os.Args[0], os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  mvpkg takes two positional arguments: a source and destination path\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  It works only withing a single go module and only with go module support enabled.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  The source and destination paths must be relative to the root of the go module\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	printf := func(s string, args ...interface{}) {}
	if verboseFlag || dryRunFlag {
		printf = func(s string, args ...interface{}) {
			fmt.Printf(s, args...)
		}
	}
	err = mvpkg.MvPkg(printf, pwd, flag.Arg(0), flag.Arg(1), []string(buildFlagsFlag), dryRunFlag, recursiveFlag)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}
