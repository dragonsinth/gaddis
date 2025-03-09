package debug

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"math/rand"
	"os"
	"time"
)

type DebugHost interface {
}

type DebugOpts struct {
	IsTest bool
	Stdout gaddis.SyncWriter
	Stderr gaddis.SyncWriter
}

type DebugSession struct {
	Host   DebugHost
	Prog   *ast.Program
	Exec   *asm.Execution
	File   string
	Source string
	Asm    string
}

func New(
	filename string,
	compileErr func(err ast.Error),
	host DebugHost,
	opts DebugOpts,
) (*DebugSession, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %w", filename, err)
	}
	src := string(buf)

	prog, _, errs := gaddis.Compile(src)
	if len(errs) > 0 {
		for _, err := range errs {
			compileErr(err)
		}
		return nil, fmt.Errorf("compile errors in %s: %w", filename, err)
	}

	assembled := asm.Assemble(prog)
	asmSrc := assembled.AsmDump(src)
	_ = asmSrc

	var seed int64
	if !opts.IsTest {
		seed = time.Now().UnixNano()
	}

	ec := &asm.ExecutionContext{
		Rng: rand.New(rand.NewSource(seed)),
		IoContext: builtins.IoContext{
			Stdin:  bufio.NewScanner(bytes.NewReader(nil)),
			Stdout: opts.Stdout,
			Stderr: opts.Stderr,
		},
	}

	exec := assembled.NewExecution(ec)

	// TODO: source/asm line mapping.

	return &DebugSession{
		Host:   host,
		Prog:   prog,
		Exec:   exec,
		File:   filename,
		Source: src,
		Asm:    asmSrc,
	}, nil
}
