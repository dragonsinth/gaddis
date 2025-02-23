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
		t.Error("format changed! updating output file...")
		_, thisFile, _, _ := runtime.Caller(0)
		dirPath := filepath.Dir(thisFile)
		t.Log(thisFile)
		update := filepath.Join(dirPath, "parse_test_fmt.gad")
		_ = os.WriteFile(update, []byte(out), 0666)
	}
}
