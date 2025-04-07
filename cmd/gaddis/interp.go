package main

import (
	"fmt"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/asm"
	"github.com/dragonsinth/gaddis/asmgen"
	"github.com/dragonsinth/gaddis/ast"
	"math/rand"
	"os"
	"time"
)

func runInterp(src *source, opts runOpts, isTest bool, streams *procStreams, prog *ast.Program) error {
	assembled := asmgen.Assemble(prog)
	if opts.leaveBuildOutputs {
		asmFile := src.desc() + ".asm"
		asmDump := assembled.Dump(src.src)
		if err := os.WriteFile(asmFile, []byte(asmDump), 0644); err != nil {
			return fmt.Errorf("writing to %s: %w", asmFile, err)
		}
	}

	if opts.stopAfterBuild {
		return nil
	}

	var seed int64
	if !isTest {
		seed = time.Now().UnixNano()
	}

	ec := &asm.ExecutionContext{
		Rng: rand.New(rand.NewSource(seed)),
		IoProvider: gaddis.IoAdapter{
			In:      gaddis.StreamInput(streams.Stdin),
			Out:     gaddis.StreamOutput(streams.Stdout),
			WorkDir: ".",
		},
	}

	p := assembled.NewExecution(ec)
	if err := p.Run(); err != nil {
		if streams.Silent {
			_, _ = os.Stdout.Write(streams.Output.Bytes())
		}
		_, _ = fmt.Fprintln(os.Stderr, err)
		_, _ = fmt.Fprintln(os.Stderr, p.GetStackTrace(src.desc()))
		os.Exit(1)
	}
	return nil
}
