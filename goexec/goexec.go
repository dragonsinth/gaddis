package goexec

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

type BuildResult struct {
	GoFile  string
	ExeFile string
}

func Build(ctx context.Context, goSrc string, basename string) (BuildResult, error) {
	var ret BuildResult
	var goFile = basename + ".go"
	if err := os.WriteFile(goFile, []byte(goSrc), 0644); err != nil {
		return ret, fmt.Errorf("could not write %s: %w", goFile, err)
	}
	ret.GoFile = goFile

	execFile, err := filepath.Abs(basename + ".exe")
	if err != nil {
		return ret, fmt.Errorf("could not resolve path to %s: %w", basename, err)
	}
	compileCmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, goFile)
	if compileOut, err := compileCmd.CombinedOutput(); err != nil {
		return ret, fmt.Errorf("compile failed: %w\n%s", err, string(compileOut))
	}
	ret.ExeFile = execFile
	return ret, nil
}

func Run(ctx context.Context, execFile string, stdin io.ReadCloser, stdout io.Writer, stderr io.Writer) error {
	// Run the compiled binary
	runCmd := exec.CommandContext(ctx, execFile)
	stdinPipe, _ := runCmd.StdinPipe()
	stdoutPipe, _ := runCmd.StdoutPipe()
	stderrPipe, _ := runCmd.StderrPipe()

	var wgOut sync.WaitGroup
	defer wgOut.Wait()
	wgOut.Add(1)
	go func() {
		defer wgOut.Done()
		_, _ = io.Copy(stdout, stdoutPipe)
	}()
	wgOut.Add(1)
	go func() {
		defer wgOut.Done()
		_, _ = io.Copy(stderr, stderrPipe)
	}()

	go func() {
		wgOut.Wait()
		_ = stdin.Close()
	}()

	var wgIn sync.WaitGroup
	defer wgIn.Wait()
	wgIn.Add(1)
	go func() {
		defer wgIn.Done()
		_, _ = io.Copy(stdinPipe, stdin)
		_ = stdinPipe.Close()
	}()

	if err := runCmd.Start(); err != nil {
		return fmt.Errorf("could not start %s: %w", execFile, err)
	}
	return runCmd.Wait()
}
