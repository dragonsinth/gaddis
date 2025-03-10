package main

import (
	"github.com/dragonsinth/gaddis"
	"os"
)

func checkCmd(args []string) error {
	src, err := readSourceFromArgs(args)
	if err != nil {
		return err
	}

	// Only check errors; output to stdout
	_, _, errs := gaddis.Compile(src.src)
	reportErrors(errs, src.desc(), *fJson, os.Stdout)
	return nil
}
