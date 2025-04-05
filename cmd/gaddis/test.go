package main

import (
	"bytes"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"os"
)

func test(args []string, opts runOpts) error {
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

	outFile := src.desc() + ".out"
	isCaptureMode := false
	wantOutput, err := os.ReadFile(outFile)
	if os.IsNotExist(err) {
		isCaptureMode = true
	} else if err != nil {
		return fmt.Errorf("reading %s: %w", outFile, err)
	}

	if isCaptureMode {
		fmt.Println("Capturing...")
	}

	var streams *procStreams
	if isCaptureMode {
		streams = captureStreams(src)
	} else {
		streams = testStreams(src)
	}

	if !opts.goGen {
		err = runInterp(src, opts, true, streams, prog)
	} else {
		err = runGo(src, opts, true, streams, prog)
	}
	if err != nil {
		return err
	}

	gotOutput := streams.Output.Bytes()

	// if we were running capture mode and captured any input, dump it to an input file
	if isCaptureMode {
		// drop the captured input into an input file
		if streams.Input.Len() > 0 {
			inFile := src.desc() + ".in"
			if err := os.WriteFile(inFile, streams.Input.Bytes(), 0644); err != nil {
				return fmt.Errorf("writing to %s: %w", inFile, err)
			}
		}
		// create an output file
		if err := os.WriteFile(outFile, gotOutput, 0644); err != nil {
			return fmt.Errorf("writing to %s: %w", outFile, err)
		}
		fmt.Println("SAVED new test output")
	} else {
		// compare the output instead
		if !bytes.Equal(gotOutput, wantOutput) {
			// compare the output
			_, _ = fmt.Fprintf(os.Stderr, "wrong output: got=\n%s\nwant=\n%s\n", gotOutput, wantOutput)
			os.Exit(1)
		}
		fmt.Println("PASSED")
	}

	return nil
}
