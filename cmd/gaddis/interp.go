package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/gogen/builtins"
	"github.com/dragonsinth/gaddis/interp"
	"math/rand"
	"os"
	"strings"
	"time"
)

func runInterp(src *source, opts runOpts, isTest bool, streams *procStreams, prog *ast.Program) error {
	cp := interp.Compile(prog)
	if opts.leaveBuildOutputs {
		asmFile := src.desc() + ".asm"
		var sb bytes.Buffer
		for i, inst := range cp.Code {
			si := inst.GetSourceInfo()
			line := si.Start.Line + 1
			text := strings.TrimSpace(src.src[si.Start.Pos:si.End.Pos])
			lhs := fmt.Sprintf("%3d: %s", i, inst)
			_, _ = fmt.Fprintf(&sb, "%-40s; %3d: %s\n", lhs, line, strings.SplitN(text, "\n", 2)[0])
		}
		if err := os.WriteFile(asmFile, sb.Bytes(), 0644); err != nil {
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

	ec := &interp.ExecutionContext{
		Rng: rand.New(rand.NewSource(seed)),
		IoContext: builtins.IoContext{
			Stdin:  bufio.NewScanner(streams.Stdin),
			Stdout: streams.Stdout,
			Stderr: os.Stderr,
		},
	}

	p := cp.NewProgram(ec)
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
