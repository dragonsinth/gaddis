package parse

import (
	_ "embed"
	"fmt"
	"github.com/dragonsinth/gaddis/astprint"
	"os"
	"testing"
)

var (
	//go:embed parse_test.gad
	program string
	//go:embed parse_test_fmt.gad
	expectOut string
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
	if out != expectOut {
		fmt.Println(out)
		t.Error("format changed")
		os.WriteFile("parse_test_fmt.gad", []byte(out), 0666)
	}
}
