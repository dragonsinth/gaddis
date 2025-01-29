package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"github.com/dragonsinth/gaddis/parser"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	fVerbose = flag.Bool("v", false, "verbose logging to stderr")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			log.Println(ee)
			os.Exit(ee.ExitCode())
		} else {
			log.Fatal(err)
		}
	}
}

/*
Test mode:
- output only: run and verify, assume no input
- output and input: run and verify, assume input
- input only: run and produce output
- neither: run and prodcue output, reading from stdin
*/

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var input []byte
	for _, fname := range flag.Args() {
		buf, err := os.ReadFile(fname)
		if err != nil {
			return fmt.Errorf("read file %s: %w", fname, err)
		}
		input = append(input, buf...)
	}

	block, err := parser.Parse(input)
	if err != nil {
		log.Fatal(err)
	}
	if *fVerbose {
		for _, stmt := range block.Statements {
			log.Println(stmt.String())
		}
	}

	goSrc := gogen.Generate(block)
	if *fVerbose {
		log.Println(goSrc)
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	return goexec.Run(ctx, goSrc, dir, os.Stdin, os.Stdout, os.Stderr)
}
