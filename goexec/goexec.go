package goexec

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
)

func Run(ctx context.Context, goSrc string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {
	goBytes := []byte(goSrc)
	hash := sha256.New()
	hash.Write(goBytes)
	sum := hash.Sum(nil)
	sha := hex.EncodeToString(sum)

	var goFile = fmt.Sprintf("main-%s.tmp.go", sha)
	if err := os.WriteFile(goFile, goBytes, 0644); err != nil {
		return fmt.Errorf("could not write %s: %w", goFile, err)
	}
	defer func() {
		_ = os.Remove(goFile)
	}()

	// Compile the Go program
	var execFile = fmt.Sprintf("main-%s.tmp.exec", sha)
	compileCmd := exec.CommandContext(ctx, "go", "build", "-o", execFile, "./"+goFile)
	if compileOut, err := compileCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("compile failed: %w\n%s", err, string(compileOut))
	}
	defer func() {
		_ = os.Remove(execFile)
	}()

	// Run the compiled binary
	runCmd := exec.Command("./" + execFile)
	runCmd.Stdin = stdin
	runCmd.Stdout = stdout
	runCmd.Stderr = stderr
	if err := runCmd.Start(); err != nil {
		return fmt.Errorf("could not start %s: %w", execFile, err)
	}

	return runCmd.Wait()
}
