package parser

import (
	_ "embed"
	"github.com/dragonsinth/gaddis/ast"
	"os"
	"testing"
)

var (
	//go:embed parser_test.gad
	program string
)

func TestParse(t *testing.T) {
	block, err := Parse([]byte(program))
	if err != nil {
		t.Fatal(err)
	}

	out := ast.DebugString(block)
	os.Stdout.WriteString(out)
}
