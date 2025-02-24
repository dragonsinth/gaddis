package examples

import (
	"bytes"
	"context"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/astprint"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"testing"
)

func RunTest(t *testing.T, filename string) error {
	src, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", filename, err)
	}

	prog, comments, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		for _, err := range ast.ErrorSort(errs) {
			t.Error(err)
		}
		t.Fatalf("%s: failed to compile", filename)
	}

	goSrc := gogen.GoGenerate(prog)

	var input bytes.Buffer
	var output bytes.Buffer
	var errput bytes.Buffer

	if inBytes, err := os.ReadFile(filename + ".in"); err == nil {
		input.Write(inBytes)
	}
	expectOut, err := os.ReadFile(filename + ".out")
	if err != nil {
		t.Fatalf("failed to read file %s: %v", filename+".out", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	err = goexec.Run(ctx, goSrc, filepath.Dir(filename), io.NopCloser(&input), &output, &errput)
	if err != nil {
		t.Fatalf("failed to exec %s: %v", filename, err)
	}
	if errput.Len() > 0 {
		t.Fatalf("stderr:\n%s", errput.String())
	}

	if !bytes.Equal(output.Bytes(), expectOut) {
		// compare the output
		t.Fatalf("wrong output, got=\n%s\nwant=%s", output.String(), string(expectOut))
	}

	// Also check/update format.
	inSrc := string(src)
	outSrc := astprint.Print(prog, comments)
	if inSrc != outSrc {
		t.Error("format changed! updating source file...")
		_ = os.WriteFile(filename, []byte(outSrc), 0666)
	}

	return nil
}
