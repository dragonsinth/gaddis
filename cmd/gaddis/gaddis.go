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
	fDebug = flag.Bool("d", false, "debug mode")
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
	if *fDebug {
		for _, stmt := range block.Statements {
			log.Println(stmt.String())
		}
	}

	goSrc := gogen.Generate(block)
	if *fDebug {
		log.Println(goSrc)
	}

	return goexec.Run(ctx, goSrc, os.Stdin, os.Stdout, os.Stderr)
}
