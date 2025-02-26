package gogen

import (
	_ "embed"
	"strings"
)

var (
	//go:embed builtins/builtins.go
	builtins string
)

func parseGoCode(source string, imports *[]string, code *[]string) {
	for _, line := range strings.Split(source, "\n") {
		if strings.HasPrefix(line, "package ") {
			continue //skip the package statement
		} else if strings.HasPrefix(line, "import (") {
			panic("don't use block import! use individual imports only")
		} else if strings.HasPrefix(line, "import ") {
			*imports = append(*imports, line)
		} else {
			*code = append(*code, line)
		}
	}
}
