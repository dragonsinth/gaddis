package main

import (
	"fmt"
	"github.com/dragonsinth/gaddis"
	"os"
)

type runOpts struct {
	stopAfterBuild    bool
	leaveBuildOutputs bool
	goGen             bool
}

func runCmd(args []string, opts runOpts) error {
	src, err := readSourceFromArgs(args)
	if err != nil {
		return err
	}

	prog, outSrc, errs := gaddis.Compile(src.src)
	reportErrors(errs, src.desc(), *fJson, os.Stdout)
	if len(errs) > 0 {
		os.Exit(1)
	}

	// auto format on success only
	if !src.isStdin && src.src != outSrc {
		if err = os.WriteFile(src.filename, []byte(outSrc), 0666); err != nil {
			return fmt.Errorf("writing to %s: %w", src.filename, err)
		}
	}

	streams := runStreams(src)

	if !opts.goGen {
		return runInterp(src, opts, false, streams, prog)
	} else {
		return runGo(src, opts, false, streams, prog)
	}
}
