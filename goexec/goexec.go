package goexec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func Run(ctx context.Context, goSrc string, dir string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	goBytes := []byte(goSrc)
	hash := sha256.New()
	hash.Write(goBytes)
	sum := hash.Sum(nil)
	sha := hex.EncodeToString(sum)

	var goFile = filepath.Join(dir, fmt.Sprintf("main-%s.tmp.go", sha))
	if err := os.WriteFile(goFile, goBytes, 0644); err != nil {
		return fmt.Errorf("could not write %s: %w", goFile, err)
	}
	defer func() {
		_ = os.Remove(goFile)
	}()

	// Compile the Go program
	var execFile = filepath.Join(dir, fmt.Sprintf("main-%s.tmp.exec", sha))
	compileCmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, goFile)
	compileCmd.Dir = dir
	if compileOut, err := compileCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile failed: %w\n%s", err, string(compileOut))
	}
	defer func() {
		_ = os.Remove(execFile)
	}()

	// Run the compiled binary
	runCmd := exec.CommandContext(ctx, execFile)
	runCmd.Dir = dir
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

	var wgIn sync.WaitGroup
	defer wgIn.Wait()
	wgIn.Add(1)
	go func() {
		defer wgIn.Done()
		wgOut.Wait()
		_ = stdinPipe.Close()
	}()
	wgIn.Add(1)
	go func() {
		defer wgIn.Done()
		_, _ = io.Copy(stdinPipe, stdin)
	}()

	if err := runCmd.Start(); err != nil {
		return fmt.Errorf("could not start %s: %w", execFile, err)
	}
	return runCmd.Wait()
}
