package examples

import (
	"bufio"
	"bytes"
	"context"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/astprint"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"github.com/dragonsinth/gaddis/interp"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"testing"
)

func RunTestGo(t *testing.T, filename string) error {
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", filename, err)
	}
	src := string(srcBytes)

	prog, comments, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		for _, err := range ast.ErrorSort(errs) {
			t.Error(err)
		}
		t.Fatalf("%s: failed to compile", filename)
	}

	goSrc := gogen.GoGenerate(prog, true)

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

	br, err := goexec.Build(ctx, goSrc, filename)
	defer func() {
		if br.GoFile != "" {
			_ = os.Remove(br.GoFile)
		}
		if br.ExeFile != "" {
			_ = os.Remove(br.ExeFile)
		}
	}()
	if err != nil {
		t.Fatalf("failed to build %s: %v", filename, err)
	}

	err = goexec.Run(ctx, br.ExeFile, io.NopCloser(&input), &output, &errput)
	if err != nil {
		t.Fatalf("failed to exec %s: %v", br.ExeFile, err)
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

func RunTestInterp(t *testing.T, filename string) error {
	srcBytes, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("failed to read file %s: %v", filename, err)
	}
	src := string(srcBytes)

	prog, comments, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		for _, err := range ast.ErrorSort(errs) {
			t.Error(err)
		}
		t.Fatalf("%s: failed to compile", filename)
	}

	cp := interp.Compile(prog)

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

	p := cp.NewProgram(&interp.ExecutionContext{
		Rng: rand.New(rand.NewSource(0)),
		IoContext: builtins.IoContext{
			Stdin:  bufio.NewScanner(&input),
			Stdout: gaddis.Synced(&output),
			Stderr: gaddis.Synced(&errput),
		},
	})

	err = p.Run()
	if err != nil {
		t.Errorf("failed to exec %s: %v", filename, err)
		t.Log(p.GetStackTrace(filename))
	}
	if errput.Len() > 0 {
		t.Errorf("stderr:\n%s", errput.String())
	}
	if t.Failed() {
		t.FailNow()
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
