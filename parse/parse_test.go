package parse

import (
	_ "embed"
	"fmt"
	"github.com/dragonsinth/gaddis/astprint"
	"os"
	"path/filepath"
	"runtime"
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
		t.Error("format changed! updating format file...")
		_, file, _, _ := runtime.Caller(0)
		root := filepath.Dir(file)
		_ = os.WriteFile(filepath.Join(root, "parse_test_fmt.gad"), []byte(out), 0666)
	}
}
