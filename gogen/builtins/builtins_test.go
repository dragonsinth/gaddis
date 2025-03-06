package builtins

import (
	"bufio"
	"bytes"
	"github.com/dragonsinth/gaddis"
	"strings"
	"testing"
)

func TestDisplay(t *testing.T) {
	inbuf := strings.NewReader("")
	var outbuf, errbuf bytes.Buffer
	ctx := IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.Synced(&outbuf),
		Stderr: gaddis.Synced(&errbuf),
	}
	ctx.Display([]byte("the score is "), 17, []byte(" to "), 21.5, []byte(" is "), false)
	assertEqual(t, "the score is 17 to 21.5 is False\n", outbuf.String())
}

func TestInputInteger(t *testing.T) {
	inbuf := strings.NewReader("not a number\n123\n")
	var outbuf, errbuf bytes.Buffer
	ctx := IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.Synced(&outbuf),
		Stderr: gaddis.Synced(&errbuf),
	}

	got := ctx.InputInteger()
	assertEqual(t, int64(123), got)
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "integer> error, invalid integer, try again\ninteger> ", outbuf.String())
}

func TestInputReal(t *testing.T) {
	inbuf := strings.NewReader("not a number\n123.456\n")
	var outbuf, errbuf bytes.Buffer
	ctx := IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.Synced(&outbuf),
		Stderr: gaddis.Synced(&errbuf),
	}

	got := ctx.InputReal()
	assertEqual(t, float64(123.456), got)
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "real> error, invalid real, try again\nreal> ", outbuf.String())
}

func TestInputBoolean(t *testing.T) {
	inbuf := strings.NewReader("not a boolean\ntrue\n")
	var outbuf, errbuf bytes.Buffer
	ctx := IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.Synced(&outbuf),
		Stderr: gaddis.Synced(&errbuf),
	}

	got := ctx.InputBoolean()
	assertEqual(t, true, got)
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "boolean> error, invalid boolean, try again\nboolean> ", outbuf.String())
}

func TestInputString(t *testing.T) {
	inbuf := strings.NewReader("David\n")
	var outbuf, errbuf bytes.Buffer
	ctx := IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.Synced(&outbuf),
		Stderr: gaddis.Synced(&errbuf),
	}

	got := ctx.InputString()
	assertEqual(t, "David", string(got))
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "string> ", outbuf.String())
}

func TestModInteger(t *testing.T) {
	assertEqual(t, 0, ModInteger(5, 1))
	assertEqual(t, 1, ModInteger(5, 2))
	assertEqual(t, 2, ModInteger(5, 3))
	// TODO: zero, negative behavior spec?
}

func TestModReal(t *testing.T) {
	assertEqual(t, 0.5, ModReal(5.5, 1))
	assertEqual(t, 1.5, ModReal(5.5, 2))
	assertEqual(t, 2.5, ModReal(5.5, 3))
	// TODO: zero, negative behavior spec?
}

func TestExpInteger(t *testing.T) {
	assertEqual(t, 5, ExpInteger(5, 1))
	assertEqual(t, 25, ExpInteger(5, 2))
	assertEqual(t, 125, ExpInteger(5, 3))
	// TODO: zero, negative behavior spec?
}

func TestExpReal(t *testing.T) {
	assertEqual(t, 0.5, ExpReal(0.5, 1))
	assertEqual(t, 0.25, ExpReal(0.5, 2))
	assertEqual(t, 0.125, ExpReal(0.5, 3))
	// TODO: zero, negative behavior spec?
}

//func TestStepInteger(t *testing.T) {
//	assertEqual(t, true, StepInteger(0, 1, 1))
//	assertEqual(t, true, StepInteger(1, 1, 1))
//	assertEqual(t, false, StepInteger(2, 1, 1))
//
//	assertEqual(t, true, StepInteger(2, 1, -1))
//	assertEqual(t, true, StepInteger(1, 1, -1))
//	assertEqual(t, false, StepInteger(0, 1, -1))
//}
//
//func TestStepReal(t *testing.T) {
//	assertEqual(t, true, StepReal(0, 1, 1))
//	assertEqual(t, true, StepReal(1, 1, 1))
//	assertEqual(t, false, StepReal(2, 1, 1))
//
//	assertEqual(t, true, StepReal(2, 1, -1))
//	assertEqual(t, true, StepReal(1, 1, -1))
//	assertEqual(t, false, StepReal(0, 1, -1))
//}

func assertEqual[T comparable](t *testing.T, want T, got T) {
	if want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
