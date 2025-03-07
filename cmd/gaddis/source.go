package main

import (
	"errors"
	"fmt"
	"io"
	"os"
)

type source struct {
	src      string
	filename string
	isStdin  bool
}

func (s *source) desc() string {
	if s.isStdin {
		return "stdin"
	} else {
		return s.filename
	}
}

func readSourceFromArgs(args []string) (*source, error) {
	switch len(args) {
	case 0:
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("reading stdin: %w", err)
		}
		return &source{
			src:      string(buf),
			filename: "",
			isStdin:  true,
		}, nil
	case 1:
		filename := args[0]
		buf, err := os.ReadFile(filename)
		if err != nil {
			return nil, fmt.Errorf("read file %s: %w", filename, err)
		}
		return &source{
			src:      string(buf),
			filename: filename,
			isStdin:  false,
		}, nil
	default:
		return nil, errors.New("expects 0 or 1 arguments: the source to parse")
	}
}
