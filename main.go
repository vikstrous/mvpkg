package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/vikstrous/mvpkg/internal/mvpkg"
)

type flagsStruct struct {
	dryRun     bool
	recursive  bool
	verbose    bool
	buildFlags arrayFlags
}

func parseFlags() flagsStruct {
	flags := flagsStruct{}
	flag.BoolVar(&flags.verbose, "v", false, "verbose, print status while running")
	flag.BoolVar(&flags.dryRun, "dry-run", false, "print planned actions without executing them")
	flag.BoolVar(&flags.recursive, "recursive", false, "recursively move all packages nested under the source package")
	flag.Var(&flags.buildFlags, "build-flags", "build tags to use while parsing source packages, can be specified morethan once\n"+
		"ex: -build-flags='-tags=foo bar'")

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s <src> <dst>\n", os.Args[0])
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  mvpkg takes two positional arguments: a source and destination path\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  It works only within a single go module and only with go module support enabled.\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  The source and destination paths must be relative to the root of the go module\n")
		fmt.Fprintf(flag.CommandLine.Output(), "\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	return flags
}

func main() {
	flags := parseFlags()

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
	if flags.verbose || flags.dryRun {
		printf = func(s string, args ...interface{}) {
			fmt.Printf(s, args...)
		}
	}

	err = mvpkg.MvPkg(printf, pwd, flag.Arg(0), flag.Arg(1), []string(flags.buildFlags), flags.dryRun, flags.recursive)
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
