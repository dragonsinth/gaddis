package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/dragonsinth/gaddis/gogen"
	"github.com/dragonsinth/gaddis/parser"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	fDebug = flag.Bool("d", false, "debug mode")
)

func main() {
	flag.Parse()

	if err := run(); err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			log.Println(ee)
			os.Exit(ee.ExitCode())
		} else {
			log.Fatal(err)
		}
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	var input []byte
	for _, fname := range flag.Args() {
		buf, err := os.ReadFile(fname)
		if err != nil {
			return fmt.Errorf("read file %s: %w", fname, err)
		}
		input = append(input, buf...)
	}

	block, err := parser.Parse(input)
	if err != nil {
		log.Fatal(err)
	}
	if *fDebug {
		for _, stmt := range block.Statements {
			log.Println(stmt.String())
		}
	}

	src := gogen.Generate(block)
	if *fDebug {
		log.Println(src)
	}

	const goFile = "main.tmp.go"
	if err := os.WriteFile(goFile, []byte(src), 0644); err != nil {
		return fmt.Errorf("could not write %s: %w", goFile, err)
	}
	defer func() {
		_ = os.Remove(goFile)
	}()

	// Compile the Go program
	const execFile = "__a.out"
	compileCmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, "./"+goFile)
	if compileOut, err := compileCmd.CombinedOutput(); err != nil {
		log.Print(string(compileOut))
		return fmt.Errorf("compile failed: %w", err)
	}
	defer func() {
		_ = os.Remove(execFile)
	}()

	// Run the compiled binary
	runCmd := exec.Command("./" + execFile)
	runCmd.Stdin = os.Stdin
	runCmd.Stdout = os.Stdout
	runCmd.Stderr = os.Stderr
	if err := runCmd.Start(); err != nil {
		return fmt.Errorf("could not start %s: %w", execFile, err)
	}

	return runCmd.Wait()
}
