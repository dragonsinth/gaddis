package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	fVerbose = flag.Bool("v", false, "verbose logging to stderr")
	fTest    = flag.Bool("t", false, "run in test mode; capture")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			log.Println(err)
			os.Exit(ee.ExitCode())
		} else {
			log.Fatal(err)
		}
	}
}

func run() error {
	if flag.NArg() != 1 {
		return errors.New("gaddis expects exactly one argument -- the program to run")
	}

	filename := flag.Arg(0)
	gadSrc, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("read file %s: %w", filename, err)
	}

	block, comments, errs := gaddis.Compile(gadSrc)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Println(err)
		}
		os.Exit(1)
	}

	if *fVerbose {
		dbgOut := ast.DebugString(block, comments)
		os.Stdout.WriteString(dbgOut)
	}

	goSrc := gogen.GoGenerate(block)
	if *fVerbose {
		os.Stdout.WriteString(goSrc)
	}

	var terminalInput bool
	var input bytes.Buffer
	var stdin io.ReadCloser
	if *fTest {
		if inBytes, err := os.ReadFile(filename + ".in"); err == nil {
			// use input file as input
			terminalInput = false
			input.Write(inBytes)
			stdin = io.NopCloser(&input)
		}
	}
	if stdin == nil {
		// capture input from terminal
		terminalInput = true
		r, w := io.Pipe()
		go func() {
			_, _ = io.Copy(w, os.Stdin)
			fmt.Println("closing")
			_ = w.Close()
		}()

		readCloser := struct {
			io.Reader
			io.Closer
		}{
			Reader: io.TeeReader(r, &input),
			Closer: w,
		}
		stdin = readCloser
	}

	// echo output to terminal if we need terminal input; or if we're not running test mode; or if verbose
	terminalOutput := terminalInput || !*fTest || *fVerbose
	var output bytes.Buffer
	var errput bytes.Buffer

	var stdout io.Writer
	var stderr io.Writer
	if terminalOutput {
		// run on the terminal
		stdout = io.MultiWriter(os.Stdout, &output)
		stderr = io.MultiWriter(os.Stderr, &errput)
	} else {
		// run silent
		stdout = &output
		stderr = &errput
	}

	dir, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err = goexec.Run(ctx, goSrc, dir, stdin, stdout, stderr)
	if err != nil {
		// If we were running silent and anything failed, spit the output to console
		if !terminalOutput {
			if output.Len() > 0 {
				_, _ = os.Stdout.Write(output.Bytes())
			}
			if errput.Len() > 0 {
				_, _ = os.Stderr.Write(errput.Bytes())
			}
		}
		return err
	}

	// if we were running test mode and captured any input, dump it to an input file
	if *fTest && terminalInput && input.Len() > 0 {
		// drop the captured input into a
		_ = os.WriteFile(filename+".in", input.Bytes(), 0644)
	}

	// if we were running test mode, either save or compare output
	if *fTest {
		gotOutput := output.Bytes()
		if expectOut, err := os.ReadFile(filename + ".out"); err != nil {
			// dump the output we got, if any
			if len(gotOutput) > 0 {
				_ = os.WriteFile(filename+".out", gotOutput, 0644)
			}
		} else if !bytes.Equal(output.Bytes(), expectOut) {
			// compare the output
			return fmt.Errorf("wrong output:\n%s", output.String())
		}
	}

	return nil
}
