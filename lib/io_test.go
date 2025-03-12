package lib_test

import (
	"bufio"
	"bytes"
	"github.com/dragonsinth/gaddis"
	"github.com/dragonsinth/gaddis/lib"
	"strings"
	"testing"
)

func TestDisplay(t *testing.T) {
	inbuf := strings.NewReader("")
	var outbuf bytes.Buffer
	ctx := lib.IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.NoopSyncWriter(&outbuf),
	}
	ctx.Display([]byte("the score is "), 17, []byte(" to "), 21.5, []byte(" is "), false)
	assertEqual(t, "the score is 17 to 21.5 is False\n", outbuf.String())
}

func TestInputInteger(t *testing.T) {
	inbuf := strings.NewReader("not a number\n123\n")
	var outbuf bytes.Buffer
	ctx := lib.IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.NoopSyncWriter(&outbuf),
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
	var outbuf bytes.Buffer
	ctx := lib.IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.NoopSyncWriter(&outbuf),
	}

	got := ctx.InputReal()
	assertEqual(t, float64(123.456), got)
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "real> error, invalid real, try again\nreal> ", outbuf.String())
}

func TestInputString(t *testing.T) {
	inbuf := strings.NewReader("David\n")
	var outbuf bytes.Buffer
	ctx := lib.IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.NoopSyncWriter(&outbuf),
	}

	got := ctx.InputString()
	assertEqual(t, "David", string(got))
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "string> ", outbuf.String())
}

func TestInputCharacter(t *testing.T) {
	inbuf := strings.NewReader("\ntrue\nc")
	var outbuf bytes.Buffer
	ctx := lib.IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.NoopSyncWriter(&outbuf),
	}

	got := ctx.InputCharacter()
	assertEqual(t, 'c', got)
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "character> error, input exactly 1 character, try again\ncharacter> error, input exactly 1 character, try again\ncharacter> ", outbuf.String())
}

func TestInputBoolean(t *testing.T) {
	inbuf := strings.NewReader("not a boolean\ntrue\n")
	var outbuf bytes.Buffer
	ctx := lib.IoContext{
		Stdin:  bufio.NewScanner(inbuf),
		Stdout: gaddis.NoopSyncWriter(&outbuf),
	}

	got := ctx.InputBoolean()
	assertEqual(t, true, got)
	if ctx.Stdin.Scan() {
		t.Error("extra input:", ctx.Stdin.Text())
	}
	assertEqual(t, "boolean> error, invalid boolean, try again\nboolean> ", outbuf.String())
}

func assertEqual[T comparable](t *testing.T, want T, got T) {
	if want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
