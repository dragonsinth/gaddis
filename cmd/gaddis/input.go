package main

import (
	"bytes"
	"github.com/dragonsinth/gaddis"
	"io"
	"os"
)

type procStreams struct {
	Stdin  io.Reader
	Stdout gaddis.SyncWriter

	Input  bytes.Buffer // either prefilled with data, or captures all input
	Output bytes.Buffer // captured output

	Silent bool
}

func runStreams(src *source) *procStreams {
	ret := procStreams{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
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
	ret.Stdout = gaddis.NoopSyncWriter(&ret.Output)
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
	ret.Stdout = gaddis.WriteSync(io.MultiWriter(os.Stdout, &ret.Output), os.Stdout)
	return &ret
}
