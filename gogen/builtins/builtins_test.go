package main_template

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

type noopSyncWriter struct {
	io.Writer
}

func (noopSyncWriter) Sync() error {
	return nil
}

func TestDisplay(t *testing.T) {
	var outbuf bytes.Buffer
	oldStdout := stdout
	defer func() { stdout = oldStdout }()
	stdout = noopSyncWriter{&outbuf}

	Builtin.Display("the score is ", 17, " to ", 21.5, " is ", false)
	assertEqual(t, "the score is 17 to 21.5 is False\n", outbuf.String())
}

func TestInputInteger(t *testing.T) {
	var outbuf bytes.Buffer
	oldStdout := stdout
	defer func() { stdout = oldStdout }()
	stdout = noopSyncWriter{&outbuf}

	oldStdin := stdin
	defer func() { stdin = oldStdin }()
	stdin = bufio.NewScanner(strings.NewReader("not a number\n123\n"))

	got := Builtin.InputInteger()
	assertEqual(t, int64(123), got)
	if stdin.Scan() {
		t.Error("extra input:", stdin.Text())
	}
	assertEqual(t, "integer> error, invalid integer, try again\ninteger> ", outbuf.String())
}

func TestInputReal(t *testing.T) {
	var outbuf bytes.Buffer
	oldStdout := stdout
	defer func() { stdout = oldStdout }()
	stdout = noopSyncWriter{&outbuf}

	oldStdin := stdin
	defer func() { stdin = oldStdin }()
	stdin = bufio.NewScanner(strings.NewReader("not a number\n123.456\n"))

	got := Builtin.InputReal()
	assertEqual(t, float64(123.456), got)
	if stdin.Scan() {
		t.Error("extra input:", stdin.Text())
	}
	assertEqual(t, "real> error, invalid real, try again\nreal> ", outbuf.String())
}

func TestInputBoolean(t *testing.T) {
	var outbuf bytes.Buffer
	oldStdout := stdout
	defer func() { stdout = oldStdout }()
	stdout = noopSyncWriter{&outbuf}

	oldStdin := stdin
	defer func() { stdin = oldStdin }()
	stdin = bufio.NewScanner(strings.NewReader("not a boolean\ntrue\n"))

	got := Builtin.InputBoolean()
	assertEqual(t, true, got)
	if stdin.Scan() {
		t.Error("extra input:", stdin.Text())
	}
	assertEqual(t, "boolean> error, invalid boolean, try again\nboolean> ", outbuf.String())
}

func TestInputString(t *testing.T) {
	var outbuf bytes.Buffer
	oldStdout := stdout
	defer func() { stdout = oldStdout }()
	stdout = noopSyncWriter{&outbuf}

	oldStdin := stdin
	defer func() { stdin = oldStdin }()
	stdin = bufio.NewScanner(strings.NewReader("David\n"))

	got := Builtin.InputString()
	assertEqual(t, "David", got)
	if stdin.Scan() {
		t.Error("extra input:", stdin.Text())
	}
	assertEqual(t, "string> ", outbuf.String())
}

func TestModInteger(t *testing.T) {
	assertEqual(t, 0, Builtin.ModInteger(5, 1))
	assertEqual(t, 1, Builtin.ModInteger(5, 2))
	assertEqual(t, 2, Builtin.ModInteger(5, 3))
	// TODO: zero, negative behavior spec?
}

func TestModReal(t *testing.T) {
	assertEqual(t, 0.5, Builtin.ModReal(5.5, 1))
	assertEqual(t, 1.5, Builtin.ModReal(5.5, 2))
	assertEqual(t, 2.5, Builtin.ModReal(5.5, 3))
	// TODO: zero, negative behavior spec?
}

func TestExpInteger(t *testing.T) {
	assertEqual(t, 5, Builtin.ExpInteger(5, 1))
	assertEqual(t, 25, Builtin.ExpInteger(5, 2))
	assertEqual(t, 125, Builtin.ExpInteger(5, 3))
	// TODO: zero, negative behavior spec?
}

func TestExpReal(t *testing.T) {
	assertEqual(t, 0.5, Builtin.ExpReal(0.5, 1))
	assertEqual(t, 0.25, Builtin.ExpReal(0.5, 2))
	assertEqual(t, 0.125, Builtin.ExpReal(0.5, 3))
	// TODO: zero, negative behavior spec?
}

func TestStepInteger(t *testing.T) {
	assertEqual(t, true, Builtin.StepInteger(0, 1, 1))
	assertEqual(t, true, Builtin.StepInteger(1, 1, 1))
	assertEqual(t, false, Builtin.StepInteger(2, 1, 1))

	assertEqual(t, true, Builtin.StepInteger(2, 1, -1))
	assertEqual(t, true, Builtin.StepInteger(1, 1, -1))
	assertEqual(t, false, Builtin.StepInteger(0, 1, -1))
}

func TestStepReal(t *testing.T) {
	assertEqual(t, true, Builtin.StepReal(0, 1, 1))
	assertEqual(t, true, Builtin.StepReal(1, 1, 1))
	assertEqual(t, false, Builtin.StepReal(2, 1, 1))

	assertEqual(t, true, Builtin.StepReal(2, 1, -1))
	assertEqual(t, true, Builtin.StepReal(1, 1, -1))
	assertEqual(t, false, Builtin.StepReal(0, 1, -1))
}

func assertEqual[T comparable](t *testing.T, want T, got T) {
	if want != got {
		t.Errorf("want=%v, got=%v", want, got)
	}
}
