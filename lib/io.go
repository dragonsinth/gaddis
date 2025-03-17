package lib

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
)

type IoProvider interface {
	Input() (string, error)
	Output(string)
}

type ioContext struct {
	provider IoProvider
}

func (ctx ioContext) Display(args ...any) {
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
	sb.WriteRune('\n')
	ctx.provider.Output(sb.String())
}

func (ctx ioContext) InputInteger() int64 {
	for {
		ctx.provider.Output("integer> ")
		input := ctx.readLine()
		v, err := strconv.ParseInt(string(input), 10, 64)
		if err == nil {
			return v
		}
		ctx.provider.Output("error, invalid integer, try again\n")
	}
}

func (ctx ioContext) InputReal() float64 {
	for {
		ctx.provider.Output("real> ")
		input := ctx.readLine()
		v, err := strconv.ParseFloat(string(input), 64)
		if err == nil {
			return v
		}
		ctx.provider.Output("error, invalid real, try again\n")
	}
}

func (ctx ioContext) InputString() []byte {
	ctx.provider.Output("string> ")
	input := ctx.readLine()
	return input
}

func (ctx ioContext) InputCharacter() byte {
	for {
		ctx.provider.Output("character> ")
		input := ctx.readLine()
		if len(input) == 1 {
			return input[0]
		}
		ctx.provider.Output("error, input exactly 1 character, try again\n")
	}
}

func (ctx ioContext) InputBoolean() bool {
	for {
		ctx.provider.Output("boolean> ")
		input := ctx.readLine()
		v, err := strconv.ParseBool(string(input))
		if err == nil {
			return v
		}
		ctx.provider.Output("error, invalid boolean, try again\n")
	}
}

func (ctx ioContext) readLine() []byte {
	in, err := ctx.provider.Input()
	if err != nil {
		panic(err)
	}
	return []byte(in)
}

type tabDisplay struct{}

// TabDisplay is "Magic" when passed directly to [Builtins.Display].
var TabDisplay = tabDisplay{}

// BELOW: Used only by the gogen runtime.

type defaultIo struct {
	in *bufio.Scanner
}

func (dio defaultIo) Input() (string, error) {
	if !dio.in.Scan() {
		return "", io.EOF
	}
	input, err := dio.in.Text(), dio.in.Err()
	if err != nil {
		return "", err
	}
	return input, nil
}

func (dio defaultIo) Output(text string) {
	_, _ = os.Stdout.Write([]byte(text))
	_ = os.Stdout.Sync()
}

var (
	ioCtx = ioContext{provider: defaultIo{in: bufio.NewScanner(os.Stdin)}}

	Display        = ioCtx.Display
	InputInteger   = ioCtx.InputInteger
	InputReal      = ioCtx.InputReal
	InputString    = ioCtx.InputString
	InputCharacter = ioCtx.InputCharacter
	InputBoolean   = ioCtx.InputBoolean
)
