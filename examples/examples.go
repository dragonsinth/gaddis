package examples

import (
	"bytes"
	"context"
	"fmt"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"github.com/dragonsinth/gaddis/parser"
	"io"
	"os"
	"path/filepath"
)

func RunTest(ctx context.Context, filename string) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	src, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	block, err := parser.Parse(src)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", filename, err)
	}
	goSrc := gogen.Generate(block)

	runTerminal := false
	inBytes, err := os.ReadFile(filename + ".in")
	if err != nil {
		runTerminal = true
	}

	var input bytes.Buffer
	var output bytes.Buffer
	var errput bytes.Buffer

	var stdin io.Reader
	var stdout io.Writer
	var stderr io.Writer
	if runTerminal {
		// run on the terminal
		stdin = io.TeeReader(os.Stdin, &input)
		stdout = io.MultiWriter(os.Stdout, &output)
		stderr = io.MultiWriter(os.Stderr, &errput)
	} else {
		// run automated
		input.Write(inBytes)
		stdin = &input
		stdout = &output
		stderr = &errput
	}

	err = goexec.Run(ctx, goSrc, filepath.Dir(filename), stdin, stdout, stderr)
	if err != nil {
		return fmt.Errorf("failed to exec %s: %w", filename, err)
	}
	if errput.Len() > 0 {
		return fmt.Errorf("stderr:\n%s", errput.String())
	}

	gotOutput := output.Bytes()
	if runTerminal {
		// dump the input and output into files
		// for some reason in this mode we need a double blank line
		inBytes = input.Bytes()
		if bytes.HasSuffix(inBytes, []byte("\n\n")) {
			inBytes = bytes.TrimSuffix(inBytes, []byte("\n"))
		}
		_ = os.WriteFile(filename+".in", inBytes, 0644)
		_ = os.WriteFile(filename+".out", gotOutput, 0644)
	} else if expectOut, err := os.ReadFile(filename + ".out"); err != nil {
		// dump just the output
		_ = os.WriteFile(filename+".out", gotOutput, 0644)
	} else if !bytes.Equal(gotOutput, expectOut) {
		// compare the output
		return fmt.Errorf("wrong output:\n%s", string(gotOutput))
	}
	return nil
}
