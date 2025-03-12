package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
)

type SyncWriter interface {
	io.Writer
	Sync() error
}

type IoContext struct {
	Stdin   *bufio.Scanner
	Stdout  SyncWriter
	WorkDir string
}

func (ctx IoContext) Display(args ...any) {
	var sb bytes.Buffer
	tabCount := 0
	for _, arg := range args {
		switch typedArg := arg.(type) {
		case bool:
			if typedArg {
				sb.WriteString("True")
			} else {
				sb.WriteString("False")
			}
		case tabDisplay:
			tabCount++
			for sb.Len() < 8*tabCount {
				sb.WriteByte(' ')
			}
		case string:
			panic(typedArg) // should be impossible
		case []byte:
			sb.Write(typedArg)
		case byte:
			sb.WriteByte(typedArg)
		default:
			// TODO: special formatting for floats maybe?
			_, _ = fmt.Fprint(&sb, arg)
		}
	}
	sb.WriteByte('\n')
	_, _ = ctx.Stdout.Write(sb.Bytes())
	_ = ctx.Stdout.Sync()
}

func (ctx IoContext) InputInteger() int64 {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "integer> ")
		input := ctx.readLine()
		v, err := strconv.ParseInt(string(input), 10, 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, invalid integer, try again")
	}
}

func (ctx IoContext) InputReal() float64 {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "real> ")
		input := ctx.readLine()
		v, err := strconv.ParseFloat(string(input), 64)
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, invalid real, try again")
	}
}

func (ctx IoContext) InputString() []byte {
	_, _ = fmt.Fprint(ctx.Stdout, "string> ")
	input := ctx.readLine()
	return input
}

func (ctx IoContext) InputCharacter() byte {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "character> ")
		input := ctx.readLine()
		if len(input) == 1 {
			return input[0]
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, input exactly 1 character, try again")
	}
}

func (ctx IoContext) InputBoolean() bool {
	for {
		_, _ = fmt.Fprint(ctx.Stdout, "boolean> ")
		input := ctx.readLine()
		v, err := strconv.ParseBool(string(input))
		if err == nil {
			return v
		}
		_, _ = fmt.Fprintln(ctx.Stdout, "error, invalid boolean, try again")
	}
}

func (ctx IoContext) readLine() []byte {
	_ = ctx.Stdout.Sync() // ensure any prompts are flushed
	if !ctx.Stdin.Scan() {
		panic(io.EOF)
	}
	input, err := ctx.Stdin.Bytes(), ctx.Stdin.Err()
	if err != nil {
		panic(err)
	}
	_ = ctx.Stdout.Sync() // ensure user's newline is flushed to the terminal
	return input
}

type tabDisplay struct{}

// TabDisplay is "Magic" when passed directly to [Builtins.Display].
var TabDisplay = tabDisplay{}

// BELOW: Used only by the gogen runtime.

var (
	ioCtx = IoContext{
		Stdin:   bufio.NewScanner(os.Stdin),
		Stdout:  os.Stdout,
		WorkDir: ".",
	}

	Display        = ioCtx.Display
	InputInteger   = ioCtx.InputInteger
	InputReal      = ioCtx.InputReal
	InputString    = ioCtx.InputString
	InputCharacter = ioCtx.InputCharacter
	InputBoolean   = ioCtx.InputBoolean
)
