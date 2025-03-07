package main

import (
	"fmt"
	"github.com/dragonsinth/gaddis/astprint"
	"github.com/dragonsinth/gaddis/parse"
	"os"
)

func format(args []string) error {
	src, err := readSourceFromArgs(args)
	if err != nil {
		return err
	}

	// parse and report lex/parse errors only
	prog, comments, errs := parse.Parse(src.src)
	reportErrors(errs, src.desc(), *fJson, os.Stderr)
	if len(errs) > 0 {
		os.Exit(1)
	}

	// dump formatted source
	outSrc := astprint.Print(prog, comments)
	if src.isStdin {
		_, _ = os.Stdout.WriteString(outSrc)
	} else if src.src != outSrc {
		if err = os.WriteFile(src.filename, []byte(outSrc), 0666); err != nil {
			return fmt.Errorf("writing to %s: %w", src.filename, err)
		}
	}

	return nil
}
