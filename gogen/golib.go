package gogen

import (
	_ "embed"
	"strconv"
	"strings"
)

var (
	//go:embed builtins/builtins.go
	builtins string
)

func parseGoCode(source string, imports *[]string, code *[]string) {
	lines := strings.Split(source, "\n")
	idx := 0
	for idx < len(lines) {
		line := lines[idx]
		if strings.HasPrefix(line, "package ") {
			// skip the package statement
		} else if strings.HasPrefix(line, "import (") {
			// dive into the import block
			idx++
			parseImportBlock(&idx, lines, imports)
		} else if strings.HasPrefix(line, `import "`) {
			line := strings.TrimPrefix(line, "import ")
			imp, err := strconv.Unquote(line)
			if err != nil {
				panic(err)
			}
			*imports = append(*imports, imp)
		} else {
			*code = append(*code, line)
		}
		idx++
	}
}

func parseImportBlock(idx *int, lines []string, imports *[]string) {
	for {
		line := strings.TrimSpace(lines[*idx])
		if line == ")" {
			return
		}
		if line != "" {
			imp, err := strconv.Unquote(line)
			if err != nil {
				panic(err)
			}
			*imports = append(*imports, imp)
		}
		*idx++
	}
}
