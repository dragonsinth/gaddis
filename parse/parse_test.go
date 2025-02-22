package parse

import (
	_ "embed"
	"github.com/dragonsinth/gaddis/astprint"
	"os"
	"testing"
)

var (
	//go:embed parse_test.gad
	program string
)

func TestParse(t *testing.T) {
	block, comments, errs := Parse([]byte(program))
	if len(errs) > 0 {
		for _, err := range errs {
			t.Error(err)
		}
		t.Fatal("parse errors")
	}

	out := astprint.Print(block, comments)
	os.Stdout.WriteString(out)
}
