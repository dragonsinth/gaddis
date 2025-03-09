package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	fVerbose = flag.Bool("v", false, "verbose logging")
	fDebug   = flag.Bool("d", false, "don't delete any generated files, leave for inspection")
	fTest    = flag.Bool("t", false, "legacy: run in test mode; capture")
	fNoRun   = flag.Bool("no-run", false, "compile only, do not run")
	fJson    = flag.Bool("json", false, "emit errors as json")
	fGogen   = flag.Bool("gogen", false, "run using go compile")
	fPort    = flag.Int("port", -1, "port to listen on; only valid with debug")
)

const help = `Usage: gaddis <command> [options] [arguments]

Available commands:

format: parse and format the input file
check:  parse and error check the input file
build:  parse, check, and build the input file
run:    everything, including format
test:   run in test mode
help:   print this help message
*.gad:  legacy: run the given file
`

func main() {
	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: gaddis <command> [options] [arguments]")
		os.Exit(1)
	}

	opts := runOpts{
		stopAfterBuild:    *fNoRun,
		leaveBuildOutputs: *fDebug,
		goGen:             *fGogen,
	}

	var err error
	switch args[0] {
	case "help":
		fmt.Print(help)
		// TODO: details subcommand help
		os.Exit(0)
	case "format":
		err = format(args[1:])
	case "check":
		err = check(args[1:])
	case "build":
		// always leave build outputs on build command
		opts.stopAfterBuild = true
		opts.leaveBuildOutputs = true
		err = run(args[1:], opts)
	case "test":
		err = test(args[1:], opts)
	case "debug":
		err = debug(*fPort, *fVerbose)
	case "run":
		err = run(args[1:], opts)
	default:
		if *fTest {
			err = test(args, opts)
		} else {
			err = run(args, opts)
		}
	}

	type hasExitCode interface {
		ExitCode() int
	}

	var he hasExitCode
	if errors.As(err, &he) {
		log.Println(err)
		os.Exit(he.ExitCode())
	} else if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
	}
}
