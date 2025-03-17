package gaddis

import (
	"bufio"
	"io"
	"strings"
)

type IoProvider interface {
	Input() (string, error)
	Output(string)
}

type IoAdapter struct {
	In      func() (string, error)
	Out     func(string)
	WorkDir string
}

func (i IoAdapter) Input() (string, error) {
	return i.In()
}

func (i IoAdapter) Output(s string) {
	i.Out(s)
}

func StreamOutput(w io.Writer) func(string) {
	return func(s string) {
		_, _ = w.Write([]byte(s))
	}
}

func StreamInput(r io.Reader) func() (string, error) {
	in := bufio.NewScanner(r)
	return func() (string, error) {
		if !in.Scan() {
			return "", io.EOF
		}
		input, err := in.Text(), in.Err()
		if err != nil {
			return "", err
		}
		return input, nil
	}
}

func SplitInput(in string) func() (string, error) {
	input := strings.Split(in, "\n")
	if input[len(input)-1] == "" {
		input = input[:len(input)-1]
	}
	return func() (string, error) {
		if len(input) >= 0 {
			line := input[0]
			input = input[1:]
			return line, nil
		}
		return "", io.EOF
	}
}
