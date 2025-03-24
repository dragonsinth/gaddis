package lib

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

func makeIoContext(in string) (ioContext, *testProvider) {
	split := strings.Split(in, "\n")
	if split[len(split)-1] == "" {
		split = split[:len(split)-1]
	}
	tp := testProvider{input: split}
	return ioContext{provider: &tp}, &tp
}

type testProvider struct {
	outbuf bytes.Buffer
	input  []string
}

func (tp *testProvider) Output(s string) {
	tp.outbuf.WriteString(s)
}

func (tp *testProvider) Input() (string, error) {
	if len(tp.input) >= 0 {
		in := tp.input[0]
		tp.input = tp.input[1:]
		return in, nil
	}
	return "", io.EOF
}

func (tp *testProvider) Dir() string {
	return ","
}

func (tp *testProvider) String() string {
	return tp.outbuf.String()
}

func (tp *testProvider) assertEmpty(t *testing.T) {
	t.Helper()
	if len(tp.input) > 0 {
		t.Errorf("reamining input: %v", tp.input)
	}
}

func TestDisplay(t *testing.T) {
	ctx, tp := makeIoContext("")
	ctx.Display([]byte("the score is "), 17, []byte(" to "), 21.5, []byte(" is "), false)
	assertEqual(t, "the score is 17 to 21.5 is False\n", tp.String())
}

func TestInputInteger(t *testing.T) {
	ctx, tp := makeIoContext("not a number\n123\n")
	got := ctx.InputInteger()
	assertEqual(t, int64(123), got)
	tp.assertEmpty(t)
	assertEqual(t, "integer> error, invalid integer, try again\ninteger> ", tp.String())
}

func TestInputReal(t *testing.T) {
	ctx, tp := makeIoContext("not a number\n123.456\n")
	got := ctx.InputReal()
	assertEqual(t, float64(123.456), got)
	tp.assertEmpty(t)
	assertEqual(t, "real> error, invalid real, try again\nreal> ", tp.String())
}

func TestInputString(t *testing.T) {
	ctx, tp := makeIoContext("David\n")
	got := ctx.InputString()
	assertEqual(t, "David", string(got))
	tp.assertEmpty(t)
	assertEqual(t, "string> ", tp.String())
}

func TestInputCharacter(t *testing.T) {
	ctx, tp := makeIoContext("\ntrue\nc")
	got := ctx.InputCharacter()
	assertEqual(t, 'c', got)
	tp.assertEmpty(t)
	assertEqual(t, "character> error, input exactly 1 character, try again\ncharacter> error, input exactly 1 character, try again\ncharacter> ", tp.String())
}

func TestInputBoolean(t *testing.T) {
	ctx, tp := makeIoContext("not a boolean\ntrue\n")
	got := ctx.InputBoolean()
	assertEqual(t, true, got)
	tp.assertEmpty(t)
	assertEqual(t, "boolean> error, invalid boolean, try again\nboolean> ", tp.String())
}

func assertEqual[T comparable](t *testing.T, want T, got T) {
	if want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
