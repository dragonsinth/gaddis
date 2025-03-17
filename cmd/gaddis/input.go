package main

import (
	"bytes"
	"io"
	"os"
)

type procStreams struct {
	Stdin  io.Reader
	Stdout io.Writer

	Input  bytes.Buffer // either prefilled with data, or captures all input
	Output bytes.Buffer // captured output

	Silent bool
}

func runStreams(src *source) *procStreams {
	ret := procStreams{
		Stdin:  os.Stdin,
		Stdout: stdoutSyncWriter{},
	}

	if src.isStdin {
		// there can be no input, we already consumed it all
		ret.Stdin = &ret.Input
	}

	return &ret
}

func testStreams(src *source) *procStreams {
	var ret procStreams
	ret.Stdin = &ret.Input
	ret.Stdout = &ret.Output
	ret.Silent = true
	if inBytes, err := os.ReadFile(src.filename + ".in"); err == nil {
		// use input file as input
		ret.Input.Write(inBytes)
	}
	return &ret
}

func captureStreams(src *source) *procStreams {
	var ret procStreams
	if src.isStdin {
		// there can be no input, we already consumed it all
		ret.Stdin = &ret.Input
	} else {
		// capture input from terminal
		ret.Stdin = io.TeeReader(os.Stdin, &ret.Input)
	}
	ret.Stdout = io.MultiWriter(stdoutSyncWriter{}, &ret.Output)
	return &ret
}

type stdoutSyncWriter struct{}

func (stdoutSyncWriter) Write(p []byte) (n int, err error) {
	n, err = os.Stdout.Write(p)
	if err != nil {
		err = os.Stdout.Sync()
	}
	return
}
