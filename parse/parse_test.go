package parse

import (
	_ "embed"
	"github.com/dragonsinth/gaddis/ast"
	"os"
	"testing"
)

var (
	//go:embed parse_test.gad
	program string
)

func TestParse(t *testing.T) {
	block, comments, err := Parse([]byte(program))
	if err != nil {
		t.Fatal(err)
	}

	out := ast.DebugString(block, comments)
	os.Stdout.WriteString(out)
}
