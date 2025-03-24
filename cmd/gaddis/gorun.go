package main

import (
	"context"
	"fmt"
	"github.com/dragonsinth/gaddis/ast"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"os"
	"os/signal"
	"syscall"
)

func runGo(src *source, opts runOpts, isTest bool, streams *procStreams, prog *ast.Program) error {
	goSrc := gogen.GoGenerate(prog, isTest)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	br, err := goexec.Build(ctx, goSrc, src.desc())
	defer func() {
		if !opts.leaveBuildOutputs {
			if br.GoFile != "" {
				_ = os.Remove(br.GoFile)
			}
			if br.ExeFile != "" {
				_ = os.Remove(br.ExeFile)
			}
		}
	}()
	if err != nil {
		return fmt.Errorf("building %s: %w", src.desc(), err)
	}

	if opts.stopAfterBuild {
		return nil
	}

	if err := goexec.Run(ctx, ".", br.ExeFile, streams.Stdin, streams.Stdout, os.Stderr); err != nil {
		if streams.Silent {
			_, _ = os.Stdout.Write(streams.Output.Bytes())
		}
		return err
	}
	return nil
}
