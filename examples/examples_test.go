package examples

import (
	"bytes"
	"context"
	"github.com/dragonsinth/gaddis/goexec"
	"github.com/dragonsinth/gaddis/gogen"
	"github.com/dragonsinth/gaddis/parser"
	"os"
	"strings"
	"testing"
)

func TestExamples(t *testing.T) {
	files, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".gad") {
			t.Run(file.Name(), func(t *testing.T) {
				testExample(t, file.Name())
			})
		}
	}
}

func testExample(t *testing.T, name string) {
	src, err := os.ReadFile(name)
	if err != nil {
		t.Fatal("cannot read", err)
	}

	block, err := parser.Parse(src)
	if err != nil {
		t.Fatal("cannot parse", err)
	}
	goSrc := gogen.Generate(block)

	input, err := os.ReadFile(name + ".in")
	if err != nil {
		t.Skip("no input file")
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	err = goexec.Run(context.Background(), goSrc, bytes.NewReader(input), &stdout, &stderr)
	if err != nil {
		t.Error(err)
	}
	if stderr.Len() > 0 {
		t.Error("stderr:", stderr.String())
	}
	output := stdout.Bytes()

	expectOut, err := os.ReadFile(name + ".out")
	if err != nil {
		os.WriteFile(name+".out", output, 0644)
	} else {
		if !bytes.Equal(output, expectOut) {
			t.Error("wrong output:\n", string(output))
		}
	}
}
