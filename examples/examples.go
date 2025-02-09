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

	var input bytes.Buffer
	var output bytes.Buffer
	var errput bytes.Buffer

	if inBytes, err := os.ReadFile(filename + ".in"); err == nil {
		input.Write(inBytes)
	}
	expectOut, err := os.ReadFile(filename + ".out")
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", filename+".out", err)
	}

	err = goexec.Run(ctx, goSrc, filepath.Dir(filename), io.NopCloser(&input), &output, &errput)
	if err != nil {
		return fmt.Errorf("failed to exec %s: %w", filename, err)
	}
	if errput.Len() > 0 {
		return fmt.Errorf("stderr:\n%s", errput.String())
	}

	if !bytes.Equal(output.Bytes(), expectOut) {
		// compare the output
		return fmt.Errorf("wrong output:\n%s", output.String())
	}
	return nil
}
